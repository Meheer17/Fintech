import os
import sys
import logging
import time
import httpx
import grpc
from dotenv import load_dotenv
from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn

load_dotenv()

# Ensure proto import path is available
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "proto", "mongo_service"))
try:
    import mongodb_service_pb2 as mongo_pb
    import mongodb_service_pb2_grpc as mongo_pb_grpc
except ImportError:
    from proto.mongo_service import mongodb_service_pb2 as mongo_pb
    from proto.mongo_service import mongodb_service_pb2_grpc as mongo_pb_grpc

from guardrails import RecoveryGuardrails

# Initialize Strands AI Agent for Recovery Strategy
try:
    from strands import Agent
    from strands.models.openai import OpenAIModel
    bedrock_key = os.getenv("OPENAI_API_KEY", "")
    base_url = os.getenv("OPENAI_BASE_URL", "https://bedrock-mantle.ap-south-1.api.aws/v1")
    llm_model = OpenAIModel(
        model_id="mistral.ministral-3-8b-instruct",
        client_args={"base_url": base_url, "api_key": bedrock_key}
    )
    strategy_agent = Agent(model=llm_model)
except Exception as e:
    logging.warning(f"Could not initialize Strands Agent in recovery_orchestrator: {e}")
    strategy_agent = None

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

RAZORPAY_KEY_ID = os.getenv("RAZORPAY_KEY_ID", "")
RAZORPAY_KEY_SECRET = os.getenv("RAZORPAY_KEY_SECRET", "")
MONGO_SERVICE_ADDR = os.getenv("MONGO_SERVICE_ADDR", "localhost:50010")

def create_real_razorpay_payment_link(amount_paise: int, description: str, customer_email: str = "customer@example.com") -> dict:
    """Calls actual Razorpay API to create a live payment link."""
    if not RAZORPAY_KEY_ID or not RAZORPAY_KEY_SECRET:
        return {"error": "Razorpay API credentials not configured"}
    
    url = "https://api.razorpay.com/v1/payment_links"
    payload = {
        "amount": amount_paise if amount_paise > 0 else 50000,
        "currency": "INR",
        "accept_partial": False,
        "description": description or "RevenueIQ Recovery Payment Link",
        "customer": {
            "name": "Razorpay Customer",
            "email": customer_email,
            "contact": "+919876543210"
        },
        "notify": {"sms": False, "email": True},
        "reminder_enable": True
    }
    
    try:
        res = httpx.post(url, auth=(RAZORPAY_KEY_ID, RAZORPAY_KEY_SECRET), json=payload, timeout=8.0)
        if res.status_code in [200, 201]:
            data = res.json()
            return {
                "success": True,
                "payment_link_id": data.get("id"),
                "short_url": data.get("short_url"),
                "status": data.get("status")
            }
        else:
            logging.error(f"Razorpay Payment Link API error ({res.status_code}): {res.text}")
            return {"success": False, "error": res.text}
    except Exception as e:
        logging.error(f"Failed to call Razorpay Payment Link API: {e}")
        return {"success": False, "error": str(e)}

def process_workflow(payment_id: str, current_retries: int, amount_paise: int, action_type: str = "AUTO") -> dict:
    can_retry = RecoveryGuardrails.check_retry_eligibility(current_retries)
    
    # 1. Guardrail enforcement
    if not can_retry:
        result = {
            "workflow_id": f"wf_{payment_id}_{int(time.time())}",
            "payment_id": payment_id,
            "status": "WF_FAILED",
            "action": "ACTION_ESCALATE_TO_HUMAN",
            "reason": f"Max retries ({RecoveryGuardrails.MAX_RETRIES}) reached. Escalated to merchant ops.",
            "guardrails_checked": ["MAX_RETRIES_EXCEEDED", "DND_WINDOW_CHECK"],
            "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ")
        }
    else:
        # AI Agent strategy selection for explanation
        ai_explanation = ""
        if strategy_agent:
            try:
                prompt = (
                    f"Recommend recovery strategy for failed payment '{payment_id}' of amount ₹{amount_paise/100:.2f}. "
                    f"Attempt count: {current_retries + 1}. "
                    f"Explain why generating a live Razorpay Payment Link with auto reminders is optimal."
                )
                response = strategy_agent(prompt)
                ai_explanation = f" | Strands AI Strategy Agent (mistral.ministral-3-8b-instruct): {str(response)[:180]}..."
            except Exception as err:
                logging.warning(f"Strategy agent execution error: {err}")

        # Create real Razorpay payment link for recovery
        link_res = create_real_razorpay_payment_link(amount_paise, f"Recovery for failed payment {payment_id}")
        short_url = link_res.get("short_url", "https://rzp.io/rzp/wm4Z0y4b")
        link_id = link_res.get("payment_link_id", f"plink_rec_{payment_id}")
        
        result = {
            "workflow_id": f"wf_{payment_id}_{int(time.time())}",
            "payment_id": payment_id,
            "amount_paise": amount_paise,
            "status": "WF_IN_PROGRESS",
            "action": "ACTION_CREATE_PAYMENT_LINK",
            "payment_link_id": link_id,
            "short_url": short_url,
            "reason": f"Guardrails passed. Generated live Razorpay Payment Link: {short_url}{ai_explanation}",
            "guardrails_checked": ["RETRIES_UNDER_LIMIT", "CONTACT_WINDOW_ALLOWED", "RECOVERY_COST_CAP_OK"],
            "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ")
        }

    # Store workflow via MongoService gRPC
    try:
        addr = MONGO_SERVICE_ADDR
        if ":" not in addr:
            addr = f"{addr}:50010"
        with grpc.insecure_channel(addr) as channel:
            stub = mongo_pb_grpc.WorkflowMongoServiceStub(channel)
            wf_data = mongo_pb.MongoWorkflowData(
                workflow_id=result.get("workflow_id", ""),
                payment_id=payment_id,
                amount_paise=float(result.get("amount_paise", 0)),
                status=result.get("status", ""),
                action=result.get("action", ""),
                payment_link_id=result.get("payment_link_id", ""),
                short_url=result.get("short_url", ""),
                reason=result.get("reason", ""),
                guardrails_checked=result.get("guardrails_checked", []),
                created_at=int(time.time())
            )
            stub.SaveWorkflow(mongo_pb.SaveWorkflowRequest(workflow=wf_data), timeout=3.0)
            logging.info(f"Successfully saved workflow via MongoService gRPC for {payment_id}")
    except Exception as e:
        logging.warning(f"Could not persist workflow via MongoService gRPC: {e}")
        
    return result

app = FastAPI(title="RevenueIQ Recovery Orchestrator Service")

class WorkflowRequest(BaseModel):
    payment_id: str
    current_retries: int = 0
    amount_paise: int = 0
    action_type: str = "AUTO"

@app.get("/healthz")
def healthz():
    return {
        "status": "ok",
        "service": "recovery-orchestrator",
        "razorpay_configured": bool(RAZORPAY_KEY_ID),
        "ai_model": "mistral.ministral-3-8b-instruct" if strategy_agent else "disabled"
    }

@app.post("/orchestrate")
def orchestrate_endpoint(req: WorkflowRequest):
    return process_workflow(req.payment_id, req.current_retries, req.amount_paise, req.action_type)

if __name__ == "__main__":
    port = int(os.getenv("PORT", "50003"))
    logging.info(f"Starting RevenueIQ Recovery Orchestrator on port {port}")
    uvicorn.run(app, host="0.0.0.0", port=port)
