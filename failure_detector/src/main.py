import os
import time
import logging
from pymongo import MongoClient
from classifier import classify_failure
from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

MONGO_URI = os.getenv("MONGO_URI", "mongodb://mongodb:27017")
if "mongodb:" not in MONGO_URI and "localhost" in MONGO_URI:
    MONGO_URI = "mongodb://localhost:27017"

def diagnose(payment_id: str, error_code: str, description: str = "", amount_paise: int = 50000, method: str = "card") -> dict:
    logging.info(f"Diagnosing payment failure: {payment_id} | Code: {error_code}")
    res = classify_failure(error_code, description)
    res["payment_id"] = payment_id
    res["amount_paise"] = amount_paise
    res["payment_method"] = method
    res["recovery_status"] = "PENDING"
    res["diagnosed_at"] = time.strftime("%Y-%m-%dT%H:%M:%SZ")

    # Store in MongoDB
    try:
        client = MongoClient(MONGO_URI, serverSelectionTimeoutMS=2000)
        db = client["revenueiq_db"]
        db.failures.update_one({"payment_id": payment_id}, {"$set": res}, upsert=True)
        client.close()
    except Exception as e:
        logging.warning(f"Could not persist failure diagnosis to MongoDB: {e}")

    return res

app = FastAPI(title="RevenueIQ Failure Detector Service")

class DiagnoseRequest(BaseModel):
    payment_id: str
    error_code: str
    description: str = ""
    amount_paise: int = 50000
    method: str = "card"

@app.get("/healthz")
def healthz():
    return {"status": "ok", "service": "failure-detector"}

@app.post("/diagnose")
def diagnose_endpoint(req: DiagnoseRequest):
    return diagnose(req.payment_id, req.error_code, req.description, req.amount_paise, req.method)

if __name__ == "__main__":
    port = int(os.getenv("PORT", "50002"))
    logging.info(f"Starting RevenueIQ Failure Detector Service on port {port}")
    uvicorn.run(app, host="0.0.0.0", port=port)
