import os
import logging
import json
import re
import httpx
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from dotenv import load_dotenv
from forecast import forecast_cash_position

load_dotenv()

# Import Strands Agent framework
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

app = FastAPI(title="RevenueIQ AI Gateway with Strands Copilot Tools")

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

# ---------------------------------------------------------
# STRANDS AGENT TOOLS FOR REVENUEIQ PLATFORM
# ---------------------------------------------------------

@tool
def create_payment_link_tool(name: str = "", email: str = "", phone: str = "", amount_inr: float = 0.0, description: str = "") -> str:
    """Create a live Razorpay payment link ONLY if ALL 4 required parameters are provided by the user: name, email, phone, and amount_inr.
    If ANY field is missing or empty or 0, DO NOT create a link. Instead, return an error specifying the missing fields."""
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
    """Fetch real-time overview metrics: total revenue at risk, total recovered amount, active workflows, recovery rate, reconciliation match rate, disputes, and subscriptions."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/overview", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch overview metrics: {str(e)}"})

@tool
def list_payment_failures(category: str = "", recovery_status: str = "") -> str:
    """List recorded payment failures (from payment.failed webhooks) with details on payment_id, failure category, error code, amount, and status."""
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
    """Run AI root cause diagnosis on a failed Razorpay payment. payment_id is required."""
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
    """Trigger or step through a bounded recovery workflow for a payment failure. payment_id is required."""
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
def list_recovery_workflows(status: str = "") -> str:
    """Fetch active recovery workflows and their step execution history."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/recoveries", timeout=4.0)
        data = resp.json()
        workflows = data.get("workflows", [])
        if status:
            workflows = [w for w in workflows if w.get("status") == status]
        return json.dumps({"workflows": workflows, "count": len(workflows)})
    except Exception as e:
        return json.dumps({"error": f"Failed to list recovery workflows: {str(e)}"})

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
    """Fetch subscription lifecycle events (subscription.pending, subscription.charged, subscription.cancelled)."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/subscriptions", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch subscription events: {str(e)}"})

@tool
def get_settlement_events() -> str:
    """Fetch settlement payout records (from settlement.processed webhooks)."""
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
    """Execute 3-way reconciliation batch matching across Orders, Payments, and Settlement payouts."""
    try:
        resp = httpx.post(f"{DASHBOARD_API_URL}/reconcile", json={"batch": []}, timeout=5.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to run reconciliation batch: {str(e)}"})

@tool
def get_reconciliation_report() -> str:
    """Get reconciliation match rate breakdown (exact, fuzzy, AI matches) and unresolved fee discrepancy exceptions."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/reconciliation", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch reconciliation report: {str(e)}"})

@tool
def get_audit_trail(limit: int = 10, entity_id: str = "") -> str:
    """Fetch immutable gRPC audit trail logs with system actions, reasoning, and guardrail checks."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/audit", timeout=4.0)
        data = resp.json()
        entries = data.get("entries", [])
        if entity_id:
            entries = [e for e in entries if e.get("entity_id") == entity_id]
        return json.dumps({"entries": entries[:limit], "total": len(entries)})
    except Exception as e:
        return json.dumps({"entries": [], "total": 0, "error": f"Failed to fetch audit trail: {str(e)}"})

@tool
def forecast_cash_position_tool(days: int = 7) -> str:
    """Calculate forward cash position forecast for N days including projected inflows and AI recovery contributions."""
    data = forecast_cash_position(days)
    return json.dumps(data)

@tool
def get_promise_to_pay_records() -> str:
    """Get active promise-to-pay records and B2B receivables tracking data."""
    try:
        resp = httpx.get(f"{DASHBOARD_API_URL}/promises", timeout=4.0)
        return json.dumps(resp.json())
    except Exception as e:
        return json.dumps({"error": f"Failed to fetch promise records: {str(e)}"})

@tool
def get_guardrail_config() -> str:
    """Get active recovery guardrail rules and compliance constraints."""
    config = {
        "max_retries_per_payment": 3,
        "max_contacts_per_day": 2,
        "dnd_window": "21:00 to 09:00 IST",
        "max_recovery_cost_percent": 20.0,
        "circuit_breaker_threshold": "5 failures per minute triggers 5min cooldown",
        "max_workflow_age_days": 7
    }
    return json.dumps(config)

# ---------------------------------------------------------
# INITIALIZE STRANDS AGENT WITH AWS BEDROCK MANTLE
# ---------------------------------------------------------

bedrock_key = os.getenv("OPENAI_API_KEY", "")
base_url = os.getenv("OPENAI_BASE_URL", "https://bedrock-mantle.us-east-1.api.aws/v1")

model = OpenAIModel(
    model_id="mistral.ministral-3-8b-instruct",
    client_args={
        "base_url": base_url,
        "api_key": bedrock_key,
        "timeout": 30.0,
        "max_retries": 2
    }
)

tools_list = [
    create_payment_link_tool,
    get_overview_metrics,
    list_payment_failures,
    diagnose_payment_failure,
    trigger_recovery_workflow,
    list_recovery_workflows,
    get_disputes_and_chargebacks,
    get_subscription_events,
    get_settlement_events,
    get_refund_events,
    run_reconciliation_batch,
    get_reconciliation_report,
    get_audit_trail,
    forecast_cash_position_tool,
    get_promise_to_pay_records,
    get_guardrail_config
]

master_agent = Agent(model=model, tools=tools_list, system_prompt=SYSTEM_PROMPT) if IS_STRANDS_AVAILABLE else None

# ---------------------------------------------------------
# FASTAPI CHAT ENDPOINT
# ---------------------------------------------------------

class ChatRequest(BaseModel):
    query: str
    session_id: str = "default_session"

@app.get("/healthz")
def healthz():
    return {"status": "ok", "service": "ai-gateway", "llm_model": "mistral.ministral-3-8b-instruct", "tools_count": len(tools_list)}

@app.post("/chat")
def chat(req: ChatRequest):
    query = req.query.strip()
    if not query:
        return {"reply": "Please ask a question or request an action.", "session_id": req.session_id}

    logging.info(f"Processing chat query: '{query}'")
    query_lower = query.lower()

    # Greetings
    if query_lower in ["hi", "hello", "hey", "hi there", "hello there"]:
        return {
            "reply": "Hello! I am the RevenueIQ Master Agent. Ask me about your settlement discrepancies, forward cash position forecasts, payment link generation, or active recovery workflows.",
            "session_id": req.session_id,
            "status": "success"
        }

    # Extract Payment Link details if present or requested
    is_payment_link_request = any(k in query_lower for k in ["payment link", "link", "pay link", "inr", "rs", "@"])
    if is_payment_link_request:
        emails = re.findall(r'[\w\.-]+@[\w\.-]+\.\w+', query)
        phones = re.findall(r'\b\d{10}\b', query)
        
        # Extract amount excluding 10-digit phone numbers
        amount_matches = re.findall(r'(?:rs\.?|₹|inr)\s*(\d+(?:\.\d+)?)|(\d+(?:\.\d+)?)\s*(?:inr|/-|rs)', query_lower)
        amounts = []
        for m in amount_matches:
            val = m[0] or m[1]
            if val and val not in phones:
                amounts.append(val)
        if not amounts:
            all_nums = re.findall(r'\b\d+(?:\.\d+)?\b', query_lower)
            amounts = [n for n in all_nums if n not in phones]
        
        # Check customer name presence
        name_match = re.search(r'(?:name is|for|customer)\s+([A-Za-z\s]+?)(?=[,:]|\s+(?:for|email|phone|\d|rs|₹|$))', query, re.IGNORECASE)
        name = name_match.group(1).strip() if name_match else ""
        if not name:
            # Check if name is first word in comma separated query e.g. "Meheer J, meherr17.jphotos@gmail.com 8310697451 1290 INR"
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
            reply = f"Hey! You forgot to provide the following details: {', '.join(missing)}. Please provide these details so I can create the payment link."
            return {"reply": reply, "session_id": req.session_id, "status": "missing_data"}
        else:
            # All 4 details are present -> Execute Payment Link Tool directly!
            email = emails[0]
            phone = phones[0]
            amount = float(amounts[0])
            link_json = create_payment_link_tool(name=name, email=email, phone=phone, amount_inr=amount, description=f"Payment link generated for {name}")
            data = json.loads(link_json)
            if data.get("success"):
                short_url = data.get("short_url")
                link_id = data.get("payment_link_id")
                reply = (
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
                return {"reply": reply, "session_id": req.session_id, "status": "success"}

    # Invoke Strands Agent if available
    if master_agent:
        try:
            response = master_agent(query)
            if response:
                reply = str(response)
                if "Processed query:" not in reply:
                    return {"reply": reply, "session_id": req.session_id, "status": "success"}
        except Exception as e:
            logging.error(f"Strands Agent execution error: {e}")

    # Direct Tool Fallback execution for other queries
    if "overview" in query_lower or "metric" in query_lower or "summary" in query_lower:
        metrics = json.loads(get_overview_metrics())
        reply = (
            f"RevenueIQ Master Overview:\n"
            f"• At Risk Revenue: ₹{metrics.get('total_at_risk_paise', 0)/100:,.2f}\n"
            f"• Total Recovered: ₹{metrics.get('total_recovered_paise', 0)/100:,.2f}\n"
            f"• Recovery Rate: {metrics.get('recovery_rate', 0)*100:.1f}%\n"
            f"• Active Workflows: {metrics.get('active_workflows', 0)}\n"
            f"• Reconciliation Match: {metrics.get('reconciliation_match', 0)*100:.1f}%"
        )
        return {"reply": reply, "session_id": req.session_id, "status": "success"}

    if "settlement" in query_lower or "payout" in query_lower:
        settlements = json.loads(get_settlement_events())
        reply = f"Here are your latest bank settlements: {json.dumps(settlements.get('settlements', [])[:5], indent=2)}"
        return {"reply": reply, "session_id": req.session_id, "status": "success"}

    return {
        "reply": f"Hello! I am the RevenueIQ Master Agent. How can I assist you with settlement reconciliation, payment link generation, or recovery workflows today?",
        "session_id": req.session_id,
        "status": "success"
    }

@app.get("/forecast")
def get_forecast(days: int = 7):
    return forecast_cash_position(days)

if __name__ == "__main__":
    port = int(os.getenv("HTTP_PORT", "8006"))
    logging.info(f"Starting RevenueIQ AI Gateway (Strands Agent + Bedrock Mantle) on port {port}")
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=port)
