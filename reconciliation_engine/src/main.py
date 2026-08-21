import os
import logging
from matcher import match_record

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

def run_batch_reconciliation(batch: list) -> dict:
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
    match_rate = (exact + fuzzy + ai) / total if total > 0 else 1.0

    return {
        "total_records": total,
        "exact_matches": exact,
        "fuzzy_matches": fuzzy,
        "ai_matches": ai,
        "unmatched": unmatched,
        "match_rate": round(match_rate, 4),
        "results": results
    }

from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn

app = FastAPI(title="RevenueIQ Reconciliation Engine Service")

class BatchReconciliationRequest(BaseModel):
    batch: list

@app.get("/healthz")
def healthz():
    return {"status": "ok", "service": "reconciliation-engine"}

@app.post("/reconcile")
def reconcile_endpoint(req: BatchReconciliationRequest):
    return run_batch_reconciliation(req.batch)

if __name__ == "__main__":
    port = int(os.getenv("PORT", "50004"))
    logging.info(f"Starting RevenueIQ Reconciliation Engine on port {port}")
    uvicorn.run(app, host="0.0.0.0", port=port)

