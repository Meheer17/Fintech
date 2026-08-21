import os
import sys
import time
import logging
import json
import grpc
from classifier import classify_failure
from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn

# Ensure proto import path is available
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "proto", "mongo_service"))
try:
    import mongodb_service_pb2 as mongo_pb
    import mongodb_service_pb2_grpc as mongo_pb_grpc
except ImportError:
    from proto.mongo_service import mongodb_service_pb2 as mongo_pb
    from proto.mongo_service import mongodb_service_pb2_grpc as mongo_pb_grpc

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

MONGO_SERVICE_ADDR = os.getenv("MONGO_SERVICE_ADDR", "localhost:50010")

def diagnose(payment_id: str, error_code: str, description: str = "", amount_paise: int = 50000, method: str = "card") -> dict:
    logging.info(f"Diagnosing payment failure: {payment_id} | Code: {error_code}")
    res = classify_failure(error_code, description)
    res["payment_id"] = payment_id
    res["amount_paise"] = amount_paise
    res["payment_method"] = method
    res["recovery_status"] = "PENDING"
    res["diagnosed_at"] = time.strftime("%Y-%m-%dT%H:%M:%SZ")

    # Store via MongoService gRPC
    try:
        addr = MONGO_SERVICE_ADDR
        if ":" not in addr:
            addr = f"{addr}:50010"
        with grpc.insecure_channel(addr) as channel:
            stub = mongo_pb_grpc.AnalyticsMongoServiceStub(channel)
            event_data = mongo_pb.MongoFailureEventData(
                event_id=f"fail_{payment_id}",
                category=res.get("category", "UNKNOWN"),
                service_name="failure-detector",
                description=json.dumps(res),
                timestamp=int(time.time())
            )
            stub.SaveFailureEvent(mongo_pb.SaveFailureEventRequest(event=event_data), timeout=3.0)
            logging.info(f"Successfully saved failure diagnosis via MongoService gRPC for {payment_id}")
    except Exception as e:
        logging.warning(f"Could not persist failure diagnosis via MongoService gRPC: {e}")

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
