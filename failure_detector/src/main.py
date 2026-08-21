import os
import time
import logging
from concurrent import futures
from classifier import classify_failure

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

def diagnose(payment_id: str, error_code: str, description: str = "") -> dict:
    logging.info(f"Diagnosing payment failure: {payment_id} | Code: {error_code}")
    res = classify_failure(error_code, description)
    res["payment_id"] = payment_id
    res["diagnosed_at"] = time.strftime("%Y-%m-%dT%H:%M:%SZ")
    return res

from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn

app = FastAPI(title="RevenueIQ Failure Detector Service")

class DiagnoseRequest(BaseModel):
    payment_id: str
    error_code: str
    description: str = ""

@app.get("/healthz")
def healthz():
    return {"status": "ok", "service": "failure-detector"}

@app.post("/diagnose")
def diagnose_endpoint(req: DiagnoseRequest):
    return diagnose(req.payment_id, req.error_code, req.description)

if __name__ == "__main__":
    port = int(os.getenv("PORT", "50002"))
    logging.info(f"Starting RevenueIQ Failure Detector Service on port {port}")
    uvicorn.run(app, host="0.0.0.0", port=port)

