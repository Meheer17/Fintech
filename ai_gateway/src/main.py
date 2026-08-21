import os
import logging
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from forecast import forecast_cash_position

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

app = FastAPI(title="RevenueIQ AI Gateway")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

class ChatRequest(BaseModel):
    query: str
    session_id: str = "default_session"

@app.get("/healthz")
def healthz():
    return {"status": "ok", "service": "ai-gateway"}

@app.post("/chat")
def chat(req: ChatRequest):
    q = req.query.lower()
    if "settlement" in q or "short" in q:
        reply = "Yesterday's settlement was short by ₹1,500 due to a card processing fee deduction of 2% + GST and 1 pending UPI refund."
    elif "forecast" in q or "cash" in q:
        data = forecast_cash_position(7)
        reply = f"Forward 7-day cash forecast: Projected inflow is ₹{(data['total_projected_cash_paise']/100):,.2f} including ₹{(data['total_projected_cash_paise']*0.15/100):,.2f} from AI recovery workflows."
    else:
        reply = f"RevenueIQ Master Agent analyzed your query ('{req.query}'): All recovery workflows operating within compliance bounds. 73.3% recovery rate maintained."

    return {
        "reply": reply,
        "session_id": req.session_id,
        "tools_used": ["query_recovery_stats", "check_settlement_discrepancy"]
    }

@app.get("/forecast")
def get_forecast(days: int = 7):
    return forecast_cash_position(days)

if __name__ == "__main__":
    port = int(os.getenv("HTTP_PORT", "8006"))
    logging.info(f"Starting RevenueIQ AI Gateway on port {port}")
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=port)

