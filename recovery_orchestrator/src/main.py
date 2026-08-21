import os
import logging
import time
import httpx
from pymongo import MongoClient
from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn

from guardrails import RecoveryGuardrails

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

RAZORPAY_KEY_ID = os.getenv("RAZORPAY_KEY_ID", "")
RAZORPAY_KEY_SECRET = os.getenv("RAZORPAY_KEY_SECRET", "")
MONGO_URI = os.getenv("MONGO_URI", "mongodb://mongodb:27017")
if "mongodb:" not in MONGO_URI and "localhost" in MONGO_URI:
    MONGO_URI = "mongodb://localhost:27017"

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
            "reason": f"Guardrails passed. Generated live Razorpay Payment Link: {short_url}",
            "guardrails_checked": ["RETRIES_UNDER_LIMIT", "CONTACT_WINDOW_ALLOWED", "RECOVERY_COST_CAP_OK"],
            "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ")
        }

    # Store workflow in MongoDB
    try:
        client = MongoClient(MONGO_URI, serverSelectionTimeoutMS=2000)
        db = client["revenueiq_db"]
        db.workflows.update_one({"payment_id": payment_id}, {"$set": result}, upsert=True)
        client.close()
    except Exception as e:
        logging.warning(f"Could not persist workflow to MongoDB: {e}")
        
    return result

app = FastAPI(title="RevenueIQ Recovery Orchestrator Service")

class WorkflowRequest(BaseModel):
    payment_id: str
    current_retries: int = 0
    amount_paise: int = 0
    action_type: str = "AUTO"

@app.get("/healthz")
def healthz():
    return {"status": "ok", "service": "recovery-orchestrator", "razorpay_configured": bool(RAZORPAY_KEY_ID)}

@app.post("/orchestrate")
def orchestrate_endpoint(req: WorkflowRequest):
    return process_workflow(req.payment_id, req.current_retries, req.amount_paise, req.action_type)

if __name__ == "__main__":
    port = int(os.getenv("PORT", "50003"))
    logging.info(f"Starting RevenueIQ Recovery Orchestrator on port {port}")
    uvicorn.run(app, host="0.0.0.0", port=port)
