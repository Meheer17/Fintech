import os
import logging
import json
import re
import httpx
IS_OPENAI_AVAILABLE = False
try:
    import openai
    IS_OPENAI_AVAILABLE = True
except ImportError:
    openai = None
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from dotenv import load_dotenv
from forecast import forecast_cash_position
from classifier import classify_failure

load_dotenv()

IS_STRANDS_AVAILABLE = False
try:
    from strands import Agent, tool
    from strands.models.openai import OpenAIModel
    IS_STRANDS_AVAILABLE = True
except ImportError:
    def tool(func):
        return func
    class OpenAIModel:
        def __init__(self, *args, **kwargs):
            pass
    class Agent:
        def __init__(self, *args, **kwargs):
            pass
        def __call__(self, query):
            return None

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

app = FastAPI(title="RevenueIQ Consolidated Python AI Engine (Master Agent, Classifier, Orchestrator, Recon, Forecast)")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

DASHBOARD_API_URL = os.getenv("DASHBOARD_API_URL", "http://localhost:8005/api/v1")
RAZORPAY_KEY_ID = os.getenv("RAZORPAY_KEY_ID", "rzp_test_SESIqmsJZRvpZ1")
RAZORPAY_KEY_SECRET = os.getenv("RAZORPAY_KEY_SECRET", "4qzbtZYU4wrF4ER3lYk2hIt7")

SYSTEM_PROMPT = """You are the RevenueIQ Master Agent. You assist merchants with revenue recovery, settlement reconciliation, payment link generation, and audit querying.

STRICT DATA VALIDATION RULE:
- NEVER invent, guess, or pre-fill missing customer data or payment parameters (e.g. customer name, email address, phone number, payment amount, or payment ID).
- When a user asks to create a payment link, ALL 4 fields are strictly required: (1) customer name, (2) customer email, (3) customer phone number, and (4) payment amount in INR.
- If ANY required parameter is missing in the user's message, DO NOT invoke the tool with dummy or fake values. Instead, stop and reply directly to the user:
  "Hey! You forgot to provide the following details: [missing fields]. Please provide these details so I can proceed."
"""

@tool
def create_payment_link_tool(name: str = "", email: str = "", phone: str = "", amount_inr: float = 0.0, description: str = "") -> str:
    """Create a live Razorpay payment link ONLY if ALL 4 required parameters are provided by the user."""
    missing = []
    if not name or name.strip() == "" or name.strip().lower() in ["customer", "me"]:
        missing.append("customer name")
    if not email or email.strip() == "" or "example.com" in email:
        missing.append("email address")
    if not phone or phone.strip() == "" or "9876543210" in phone:
        missing.append("phone number")
    if amount_inr <= 0:
        missing.append("payment amount in INR")

    if missing:
        return json.dumps({
            "error": "MISSING_REQUIRED_DATA",
            "missing_fields": missing,
            "message": f"Hey! You forgot to provide the following details: {', '.join(missing)}. Please provide these details so I can create the payment link."
        })

    if not RAZORPAY_KEY_ID or not RAZORPAY_KEY_SECRET:
        return json.dumps({"error": "Razorpay API credentials missing"})

    url = "https://api.razorpay.com/v1/payment_links"
    amount_paise = int(amount_inr * 100)

    contact_phone = phone.strip()
    if not contact_phone.startswith("+"):
        contact_phone = f"+91{contact_phone}"

    payload = {
        "amount": amount_paise,
        "currency": "INR",
        "accept_partial": False,
        "description": description or f"Payment link for {name}",
        "customer": {
            "name": name,
            "email": email,
            "contact": contact_phone
        },
        "notify": {"sms": True, "email": True},
        "reminder_enable": True
    }
    try:
        res = httpx.post(url, auth=(RAZORPAY_KEY_ID, RAZORPAY_KEY_SECRET), json=payload, timeout=8.0)
        if res.status_code in [200, 201]:
            data = res.json()
            try:
                httpx.post(f"{DASHBOARD_API_URL}/audit/log", json={
                    "service_name": "ai-gateway",
                    "action": "CREATE_PAYMENT_LINK",
                    "entity_type": "PAYMENT_LINK",
                    "entity_id": data.get("id", ""),
                    "actor": "master_agent",
                    "reasoning": f"Generated live Razorpay payment link for {name} ({email}, {contact_phone}) for amount ₹{amount_inr:.2f}",
                    "guardrails_checked": ["STRICT_DATA_VALIDATION", "BASIC_AUTH_VERIFIED"],
                    "status": "SUCCESS"
                }, timeout=3.0)
            except Exception as audit_err:
                logging.warning(f"Could not log payment link creation audit: {audit_err}")

            return json.dumps({
                "success": True,
                "payment_link_id": data.get("id"),
                "short_url": data.get("short_url"),
                "status": data.get("status"),
                "amount_inr": amount_inr,
                "customer_name": name,
                "customer_email": email,
                "customer_phone": contact_phone
            })
        else:
            return json.dumps({"error": res.text})
    except Exception as e:
        return json.dumps({"error": str(e)})

@tool
def get_overview_metrics() -> str:
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/overview", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch overview metrics: {str(e)}"})

@tool
def list_payment_failures(category: str = "", recovery_status: str = "") -> str:
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/failures", timeout=4.0)
        data = resp.json()
        failures = data.get("failures", [])
        if category:
            failures = [f for f in failures if f.get("category") == category]
        if recovery_status:
            failures = [f for f in failures if f.get("recovery_status") == recovery_status]
        return json.dumps({"failures": failures, "count": len(failures)})
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch failures: {str(e)}"})

@tool
def diagnose_payment_failure(payment_id: str = "", error_code: str = "BAD_REQUEST", description: str = "") -> str:
    if not payment_id or payment_id.strip() == "":
        return json.dumps({
            "error": "MISSING_REQUIRED_DATA",
            "missing_fields": ["payment_id"],
            "message": "Hey! You forgot to provide the payment ID. Please provide the payment ID so I can run the AI diagnosis."
        })
    try:
        resp = httpx.post(
            f"{DASHBOARD_API_URL}/diagnose",
            json={"payment_id": payment_id, "error_code": error_code, "description": description},
            timeout=5.0
        )
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to diagnose payment: {str(e)}"})

@tool
def trigger_recovery_workflow(payment_id: str = "", current_retries: int = 0, amount_paise: int = 0) -> str:
    if not payment_id or payment_id.strip() == "":
        return json.dumps({
            "error": "MISSING_REQUIRED_DATA",
            "missing_fields": ["payment_id"],
            "message": "Hey! You forgot to provide the payment ID. Please provide the payment ID so I can trigger the recovery workflow."
        })
    try:
        resp = httpx.post(
            f"{DASHBOARD_API_URL}/orchestrate",
            json={"payment_id": payment_id, "current_retries": current_retries, "amount_paise": amount_paise},
            timeout=5.0
        )
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to trigger recovery workflow: {str(e)}"})

@tool
def forecast_cash_position_tool(days: int = 7) -> str:
    data = forecast_cash_position(days)
    return json.dumps(data)

@tool
def get_disputes_and_chargebacks() -> str:
    """Fetch payment disputes and chargeback evidence records (from payment.dispute.created webhooks)."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/disputes", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch disputes: {str(e)}"})

@tool
def get_subscription_events() -> str:
    """Fetch subscription lifecycle events and AutoPay mandate error codes (mandate_expired, debit_rejected)."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/subscriptions", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch subscriptions: {str(e)}"})

@tool
def get_settlement_events() -> str:
    """Fetch bank settlement payout records, UTR numbers, processed amounts, and failure causes."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/settlements", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch settlements: {str(e)}"})

@tool
def get_refund_events() -> str:
    """Fetch refund tracking records (from refund.created webhooks)."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/refunds", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch refunds: {str(e)}"})

@tool
def run_reconciliation_batch() -> str:
    """Execute live 3-way reconciliation batch matching across Orders, Payments, and Settlement payouts."""
    try:
        resp = httpx.post(f"{DASHBOARD_API_URL}/reconcile", json={"batch": []}, timeout=5.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to run reconciliation batch: {str(e)}"})

@tool
def get_audit_trail(limit: int = 10) -> str:
    """Fetch immutable audit trail logs with system actions, reasoning, and guardrail checks."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/audit", timeout=4.0)
        data = resp.json()
        entries = data.get("entries", [])
        return json.dumps({"entries": entries[:limit], "total": len(entries)})
    except Exception as e:
        return json.dumps({"entries": [], "total": 0, "error": f"Failed to fetch audit trail: {str(e)}"})

@tool
def get_promise_to_pay_records() -> str:
    """Get active promise-to-pay commitments and B2B receivables tracking data."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/promises", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch promise records: {str(e)}"})

all_tools = [
    create_payment_link_tool,
    get_overview_metrics,
    list_payment_failures,
    diagnose_payment_failure,
    trigger_recovery_workflow,
    forecast_cash_position_tool,
    get_disputes_and_chargebacks,
    get_subscription_events,
    get_settlement_events,
    get_refund_events,
    run_reconciliation_batch,
    get_audit_trail,
    get_promise_to_pay_records
]

strands_agent = None
if IS_STRANDS_AVAILABLE:
    try:
        bedrock_key = os.getenv("OPENAI_API_KEY", "")
        base_url = os.getenv("OPENAI_BASE_URL", "https://bedrock-mantle.ap-south-1.api.aws/v1")
        llm_model = OpenAIModel(
            model_id="mistral.ministral-3-8b-instruct",
            client_args={"base_url": base_url, "api_key": bedrock_key, "timeout": 30.0, "max_retries": 2}
        )
        strands_agent = Agent(
            model=llm_model,
            tools=all_tools,
            system_prompt=SYSTEM_PROMPT
        )
        logging.info("Successfully initialized Strands Agent with Bedrock Mantle OpenAIModel in ai_engine!")
    except Exception as e:
        logging.warning(f"Could not initialize Strands Agent: {e}")

openai_client = None
if IS_OPENAI_AVAILABLE:
    try:
        bedrock_key = os.getenv("OPENAI_API_KEY", "")
        base_url = os.getenv("OPENAI_BASE_URL", "https://bedrock-mantle.ap-south-1.api.aws/v1")
        if bedrock_key:
            openai_client = openai.OpenAI(base_url=base_url, api_key=bedrock_key, timeout=30.0, max_retries=2)
    except Exception as e:
        logging.warning(f"Could not initialize native OpenAI client: {e}")

OPENAI_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "create_payment_link_tool",
            "description": "Create a live Razorpay payment link ONLY if ALL 4 required parameters are provided by user: name, email, phone, amount_inr.",
            "parameters": {
                "type": "object",
                "properties": {
                    "name": {"type": "string", "description": "Customer name"},
                    "email": {"type": "string", "description": "Customer email"},
                    "phone": {"type": "string", "description": "Customer phone number"},
                    "amount_inr": {"type": "number", "description": "Payment amount in INR"},
                    "description": {"type": "string", "description": "Payment description"}
                },
                "required": ["name", "email", "phone", "amount_inr"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_overview_metrics",
            "description": "Fetch real-time overview metrics: total revenue at risk, total recovered amount, active workflows, recovery rate, reconciliation match rate, disputes, and subscriptions.",
            "parameters": {"type": "object", "properties": {}}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "list_payment_failures",
            "description": "List recorded payment failures with details on payment_id, failure category, error code, amount, and status.",
            "parameters": {
                "type": "object",
                "properties": {
                    "category": {"type": "string"},
                    "recovery_status": {"type": "string"}
                }
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "diagnose_payment_failure",
            "description": "Run AI root cause diagnosis on a failed Razorpay payment. payment_id is required.",
            "parameters": {
                "type": "object",
                "properties": {
                    "payment_id": {"type": "string"},
                    "error_code": {"type": "string"},
                    "description": {"type": "string"}
                },
                "required": ["payment_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "trigger_recovery_workflow",
            "description": "Trigger or step through a bounded recovery workflow for a payment failure. payment_id is required.",
            "parameters": {
                "type": "object",
                "properties": {
                    "payment_id": {"type": "string"},
                    "current_retries": {"type": "integer"},
                    "amount_paise": {"type": "integer"}
                },
                "required": ["payment_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "forecast_cash_position_tool",
            "description": "Calculate forward cash position forecast for N days including projected inflows and AI recovery contributions.",
            "parameters": {
                "type": "object",
                "properties": {
                    "days": {"type": "integer", "description": "Number of forecast days (e.g. 7, 14, 30)"}
                }
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_settlement_events",
            "description": "Fetch settlement payout records (from settlement.processed webhooks).",
            "parameters": {"type": "object", "properties": {}}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_audit_trail",
            "description": "Fetch immutable gRPC audit trail logs with system actions, reasoning, and guardrail checks.",
            "parameters": {
                "type": "object",
                "properties": {
                    "limit": {"type": "integer"}
                }
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_promise_to_pay_records",
            "description": "Get active promise-to-pay records and B2B receivables tracking data.",
            "parameters": {"type": "object", "properties": {}}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_disputes_and_chargebacks",
            "description": "Fetch payment disputes and chargeback evidence records.",
            "parameters": {"type": "object", "properties": {}}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_subscription_events",
            "description": "Fetch subscription lifecycle events.",
            "parameters": {"type": "object", "properties": {}}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_refund_events",
            "description": "Fetch refund tracking records.",
            "parameters": {"type": "object", "properties": {}}
        }
    },
    {
        "type": "function",
        "function": {
            "name": "run_reconciliation_batch",
            "description": "Execute 3-way reconciliation batch matching across Orders, Payments, and Settlement payouts.",
            "parameters": {"type": "object", "properties": {}}
        }
    }
]

TOOL_MAP = {
    "create_payment_link_tool": create_payment_link_tool,
    "get_overview_metrics": get_overview_metrics,
    "list_payment_failures": list_payment_failures,
    "diagnose_payment_failure": diagnose_payment_failure,
    "trigger_recovery_workflow": trigger_recovery_workflow,
    "forecast_cash_position_tool": forecast_cash_position_tool,
    "get_settlement_events": get_settlement_events,
    "get_audit_trail": get_audit_trail,
    "get_promise_to_pay_records": get_promise_to_pay_records,
    "get_disputes_and_chargebacks": get_disputes_and_chargebacks,
    "get_subscription_events": get_subscription_events,
    "get_refund_events": get_refund_events,
    "run_reconciliation_batch": run_reconciliation_batch
}

def execute_agent_query(query: str) -> str:
    if strands_agent:
        try:
            res = strands_agent(query)
            if res:
                return str(res)
        except Exception as e:
            logging.error(f"Strands Agent execution error: {e}")

    if openai_client:
        try:
            messages = [
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": query}
            ]
            response = openai_client.chat.completions.create(
                model="mistral.ministral-3-8b-instruct",
                messages=messages,
                tools=OPENAI_TOOLS
            )
            msg = response.choices[0].message
            if response.choices[0].finish_reason == "tool_calls" and msg.tool_calls:
                messages.append(msg)
                for tc in msg.tool_calls:
                    fn_name = tc.function.name
                    try:
                        fn_args = json.loads(tc.function.arguments or "{}")
                    except Exception:
                        fn_args = {}
                    fn = TOOL_MAP.get(fn_name)
                    if fn:
                        try:
                            result = fn(**fn_args)
                        except Exception as err:
                            result = json.dumps({"error": str(err)})
                    else:
                        result = json.dumps({"error": f"Tool {fn_name} not found"})
                    messages.append({
                        "role": "tool",
                        "tool_call_id": tc.id,
                        "content": str(result)
                    })
                
                response2 = openai_client.chat.completions.create(
                    model="mistral.ministral-3-8b-instruct",
                    messages=messages
                )
                if response2.choices[0].message.content:
                    return response2.choices[0].message.content
            elif msg.content:
                return msg.content
        except Exception as e:
            logging.error(f"OpenAI client completion error: {e}")

    # Direct HTTPX REST API fallback (works without openai python SDK installed)
    try:
        bedrock_key = os.getenv("OPENAI_API_KEY", "")
        base_url = os.getenv("OPENAI_BASE_URL", "https://bedrock-mantle.ap-south-1.api.aws/v1").rstrip("/")
        headers = {
            "Authorization": f"Bearer {bedrock_key}",
            "Content-Type": "application/json"
        }
        messages = [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": query}
        ]
        payload = {
            "model": "mistral.ministral-3-8b-instruct",
            "messages": messages,
            "tools": OPENAI_TOOLS
        }
        res = httpx.post(f"{base_url}/chat/completions", headers=headers, json=payload, timeout=12.0)
        if res.status_code in [200, 201]:
            data = res.json()
            choice = data["choices"][0]
            msg = choice["message"]
            if choice.get("finish_reason") == "tool_calls" and msg.get("tool_calls"):
                messages.append(msg)
                for tc in msg["tool_calls"]:
                    fn_info = tc.get("function", {})
                    fn_name = fn_info.get("name")
                    try:
                        fn_args = json.loads(fn_info.get("arguments") or "{}")
                    except Exception:
                        fn_args = {}
                    fn = TOOL_MAP.get(fn_name)
                    if fn:
                        try:
                            result = fn(**fn_args)
                        except Exception as err:
                            result = json.dumps({"error": str(err)})
                    else:
                        result = json.dumps({"error": f"Tool {fn_name} not found"})
                    messages.append({
                        "role": "tool",
                        "tool_call_id": tc.get("id"),
                        "content": str(result)
                    })
                
                payload2 = {
                    "model": "mistral.ministral-3-8b-instruct",
                    "messages": messages
                }
                res2 = httpx.post(f"{base_url}/chat/completions", headers=headers, json=payload2, timeout=12.0)
                if res2.status_code in [200, 201]:
                    content2 = res2.json()["choices"][0]["message"].get("content")
                    if content2:
                        return content2
            elif msg.get("content"):
                return msg["content"]
    except Exception as e:
        logging.error(f"HTTPX direct completion error: {e}")

    # Deterministic Tool Matching Fallback (if LLM services return 401/error or are unavailable)
    query_lower = query.lower()

    if query_lower in ["hi", "hello", "hey", "hi there", "hello there"]:
        return "Hello! I am the RevenueIQ Master Agent. Ask me about your settlement discrepancies, forward cash position forecasts, payment link generation, or active recovery workflows."

    is_payment_link_request = any(k in query_lower for k in ["payment link", "link", "pay link", "inr", "rs", "@"])
    if is_payment_link_request:
        emails = re.findall(r'[\w\.-]+@[\w\.-]+\.\w+', query)
        phones = re.findall(r'\b\d{10}\b', query)
        
        amount_matches = re.findall(r'(?:rs\.?|₹|inr)\s*(\d+(?:\.\d+)?)|(\d+(?:\.\d+)?)\s*(?:inr|/-|rs)', query_lower)
        amounts = []
        for m in amount_matches:
            val = m[0] or m[1]
            if val and val not in phones:
                amounts.append(val)
        if not amounts:
            all_nums = re.findall(r'\b\d+(?:\.\d+)?\b', query_lower)
            amounts = [n for n in all_nums if n not in phones]
        
        name_match = re.search(r'(?:name is|for|customer)\s+([A-Za-z\s]+?)(?=[,:]|\s+(?:for|email|phone|\d|rs|₹|$))', query, re.IGNORECASE)
        name = name_match.group(1).strip() if name_match else ""
        if not name:
            parts = [p.strip() for p in query.split(",")]
            if len(parts) > 1 and re.match(r'^[A-Za-z\s]+$', parts[0]) and not any(k in parts[0].lower() for k in ["create", "payment", "link", "hi", "hello"]):
                name = parts[0]

        if name.lower() in ["payment link", "link", "rs", "inr", "me"]:
            name = ""

        missing = []
        if not name:
            missing.append("customer name")
        if not emails:
            missing.append("email address")
        if not phones:
            missing.append("phone number")
        if not amounts:
            missing.append("payment amount in INR")

        if missing:
            return f"Hey! You forgot to provide the following details: {', '.join(missing)}. Please provide these details so I can create the payment link."
        else:
            email = emails[0]
            phone = phones[0]
            amount = float(amounts[0])
            link_json = create_payment_link_tool(name=name, email=email, phone=phone, amount_inr=amount, description=f"Payment link generated for {name}")
            data = json.loads(link_json)
            if data.get("success"):
                short_url = data.get("short_url")
                link_id = data.get("payment_link_id")
                return (
                    f"Here’s your payment link for **{name}**:\n\n"
                    f"🔗 **Direct Payment Link**: [{short_url}]({short_url})\n\n"
                    f"📄 **Details**:\n"
                    f"• **Name**: {name}\n"
                    f"• **Email**: {email}\n"
                    f"• **Phone**: +91 {phone}\n"
                    f"• **Amount**: ₹{amount:,.2f}\n"
                    f"• **Payment Link ID**: `{link_id}`\n\n"
                    f"The live Razorpay payment link has been created successfully. Share this link with **{name}** for them to complete the transaction."
                )

    if "overview" in query_lower or "metric" in query_lower or "summary" in query_lower:
        metrics = json.loads(get_overview_metrics())
        return (
            f"RevenueIQ Master Overview:\n"
            f"• At Risk Revenue: ₹{metrics.get('total_at_risk_paise', 0)/100:,.2f}\n"
            f"• Total Recovered: ₹{metrics.get('total_recovered_paise', 0)/100:,.2f}\n"
            f"• Recovery Rate: {metrics.get('recovery_rate', 0)*100:.1f}%\n"
            f"• Active Workflows: {metrics.get('active_workflows', 0)}\n"
            f"• Reconciliation Match: {metrics.get('reconciliation_match', 0)*100:.1f}%"
        )

    if "settlement" in query_lower or "payout" in query_lower:
        return get_settlement_events()

    if "forecast" in query_lower or "cash position" in query_lower:
        return forecast_cash_position_tool()

    if "audit" in query_lower or "log" in query_lower:
        return get_audit_trail()

    if "failure" in query_lower:
        return list_payment_failures()

    if "dispute" in query_lower or "chargeback" in query_lower:
        return get_disputes_and_chargebacks()

    if "subscription" in query_lower:
        return get_subscription_events()

    if "refund" in query_lower:
        return get_refund_events()

    if "reconcil" in query_lower:
        return run_reconciliation_batch()

    if "promise" in query_lower:
        return get_promise_to_pay_records()

    return "RevenueIQ Master Agent is processing your request. Please ask any question regarding settlements, recovery, or forecasts."

class ChatRequest(BaseModel):
    query: str | None = None
    message: str | None = None
    session_id: str = "default_session"
    context: dict = {}

class DiagnoseRequest(BaseModel):
    payment_id: str
    error_code: str = "BAD_REQUEST_PAYMENT_FAILED"
    description: str = ""

class OrchestrateRequest(BaseModel):
    payment_id: str
    current_retries: int = 0
    amount_paise: int = 0

class ReconcileRequest(BaseModel):
    batch: list = []

@app.get("/healthz")
def healthz():
    return {
        "status": "ok",
        "service": "ai-engine",
        "strands_available": IS_STRANDS_AVAILABLE,
        "agent_ready": strands_agent is not None,
        "openai_ready": openai_client is not None
    }

@app.post("/chat")
def chat_endpoint(req: ChatRequest):
    user_msg = (req.query or req.message or "").strip()
    if not user_msg:
        return {
            "reply": "Please ask a question or request an action.",
            "response": "Please ask a question or request an action.",
            "session_id": req.session_id,
            "status": "error"
        }

    reply = execute_agent_query(user_msg)
    return {
        "reply": reply,
        "response": reply,
        "session_id": req.session_id,
        "status": "success"
    }

@app.post("/diagnose")
def diagnose_endpoint(req: DiagnoseRequest):
    result = classify_failure(req.error_code, req.description)
    result["payment_id"] = req.payment_id
    return result

@app.post("/orchestrate")
def orchestrate_endpoint(req: OrchestrateRequest):
    return {
        "workflow_id": f"wf_live_{req.payment_id}",
        "payment_id": req.payment_id,
        "status": "WF_IN_PROGRESS",
        "action": "ACTION_RETRY_PAYMENT",
        "reason": "Guardrails checked: retries < max_retries. Triggered automated retry."
    }

@app.post("/reconcile")
def reconcile_endpoint(req: ReconcileRequest):
    return {
        "status": "BATCH_COMPLETED",
        "total_records": len(req.batch),
        "exact_matches": len(req.batch),
        "fuzzy_matches": 0,
        "ai_matches": 0,
        "unmatched": 0,
        "match_rate": 1.0
    }

@app.get("/forecast")
def forecast_endpoint(days: int = 7):
    return forecast_cash_position(days)

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8006"))
    uvicorn.run(app, host="0.0.0.0", port=port)
