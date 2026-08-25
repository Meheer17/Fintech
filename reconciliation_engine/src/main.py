import os
import sys
import logging
import grpc
from dotenv import load_dotenv
from matcher import match_record
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

# Initialize Strands AI Agent for Reconciliation
try:
    from strands import Agent
    from strands.models.openai import OpenAIModel
    bedrock_key = os.getenv("OPENAI_API_KEY", "")
    base_url = os.getenv("OPENAI_BASE_URL", "https://bedrock-mantle.ap-south-1.api.aws/v1")
    llm_model = OpenAIModel(
        model_id="mistral.ministral-3-8b-instruct",
        client_args={"base_url": base_url, "api_key": bedrock_key, "timeout": 30.0, "max_retries": 2}
    )
    recon_agent = Agent(model=llm_model)
except Exception as e:
    logging.warning(f"Could not initialize Strands Agent in reconciliation_engine: {e}")
    recon_agent = None

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

MONGO_SERVICE_ADDR = os.getenv("MONGO_SERVICE_ADDR", "localhost:50010")

def run_batch_reconciliation(batch: list = None) -> dict:
    if not batch:
        batch = []
        try:
            addr = MONGO_SERVICE_ADDR
            if ":" not in addr:
                addr = f"{addr}:50010"
            with grpc.insecure_channel(addr) as channel:
                order_stub = mongo_pb_grpc.OrderMongoServiceStub(channel)
                orders_resp = order_stub.ListOrders(mongo_pb.ListMongoOrdersRequest(limit=100), timeout=3.0)
                orders = [{"order_id": o.order_id, "total_amount": o.total_amount, "status": o.status} for o in orders_resp.orders]

                payment_stub = mongo_pb_grpc.PaymentMongoServiceStub(channel)
                tx_resp = payment_stub.ListUserTransactions(mongo_pb.ListUserTxRequest(limit=100), timeout=3.0)
                payments = [{"transaction_id": t.transaction_id, "amount": t.amount, "status": t.status} for t in tx_resp.transactions]

                for idx, order in enumerate(orders):
                    payment = payments[idx] if idx < len(payments) else {}
                    settlement = {}
                    batch.append({"order": order, "payment": payment, "settlement": settlement})
        except Exception as e:
            logging.warning(f"Could not load records via MongoService gRPC for recon batch: {e}")

    exact, fuzzy, ai, unmatched = 0, 0, 0, 0
    results = []

    for item in batch:
        res = match_record(item.get("order", {}), item.get("payment", {}), item.get("settlement", {}))
        mtype = res["match_type"]
        if mtype == "EXACT_MATCH":
            exact += 1
        elif mtype == "FUZZY_MATCH":
            fuzzy += 1
        elif mtype == "AI_MATCH":
            ai += 1
        else:
            unmatched += 1
        results.append(res)

    total = len(batch)
    if total == 0:
        return {
            "total_records": 0,
            "exact_matches": 0,
            "fuzzy_matches": 0,
            "ai_matches": 0,
            "unmatched": 0,
            "match_rate": 0.0,
            "status": "BATCH_COMPLETED",
            "results": []
        }

    match_rate = (exact + fuzzy + ai) / total if total > 0 else 0.0

    return {
        "total_records": total,
        "exact_matches": exact,
        "fuzzy_matches": fuzzy,
        "ai_matches": ai,
        "unmatched": unmatched,
        "match_rate": round(match_rate, 4),
        "status": "BATCH_COMPLETED",
        "results": results[:10]
    }

app = FastAPI(title="RevenueIQ Reconciliation Engine Service")

class BatchReconciliationRequest(BaseModel):
    batch: list = []

@app.get("/healthz")
def healthz():
    return {
        "status": "ok",
        "service": "reconciliation-engine",
        "ai_model": "mistral.ministral-3-8b-instruct" if recon_agent else "disabled"
    }

@app.post("/reconcile")
def reconcile_endpoint(req: BatchReconciliationRequest):
    return run_batch_reconciliation(req.batch)

if __name__ == "__main__":
    port = int(os.getenv("PORT", "50004"))
    logging.info(f"Starting RevenueIQ Reconciliation Engine on port {port}")
    uvicorn.run(app, host="0.0.0.0", port=port)
