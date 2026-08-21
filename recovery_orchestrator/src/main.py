import os
import logging
from guardrails import RecoveryGuardrails

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

def process_workflow(payment_id: str, current_retries: int, amount_paise: int) -> dict:
    can_retry = RecoveryGuardrails.check_retry_eligibility(current_retries)
    if not can_retry:
        return {
            "status": "WF_FAILED",
            "action": "ACTION_ESCALATE_TO_HUMAN",
            "reason": f"Max retries ({RecoveryGuardrails.MAX_RETRIES}) reached for payment {payment_id}"
        }
    
    return {
        "status": "WF_IN_PROGRESS",
        "action": "ACTION_RETRY_PAYMENT",
        "reason": f"Guardrails passed. Triggering attempt #{current_retries + 1}"
    }

from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn

app = FastAPI(title="RevenueIQ Recovery Orchestrator Service")

class WorkflowRequest(BaseModel):
    payment_id: str
    current_retries: int = 0
    amount_paise: int = 0

@app.get("/healthz")
def healthz():
    return {"status": "ok", "service": "recovery-orchestrator"}

@app.post("/orchestrate")
def orchestrate_endpoint(req: WorkflowRequest):
    return process_workflow(req.payment_id, req.current_retries, req.amount_paise)

if __name__ == "__main__":
    port = int(os.getenv("PORT", "50003"))
    logging.info(f"Starting RevenueIQ Recovery Orchestrator on port {port}")
    uvicorn.run(app, host="0.0.0.0", port=port)

