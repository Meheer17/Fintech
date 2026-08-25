# RevenueIQ Forward Cash Position Forecaster with gRPC MongoService Integration

import os
import logging
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "proto", "mongo_service"))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "proto"))

try:
    import mongodb_service_pb2 as mongo_pb
    import mongodb_service_pb2_grpc as mongo_pb_grpc
except Exception:
    mongo_pb = None
    mongo_pb_grpc = None

MONGO_SERVICE_ADDR = os.getenv("MONGO_SERVICE_ADDR", "localhost:50010")

def forecast_cash_position(days: int = 7, avg_daily_settlement_paise: int = 2500000) -> dict:
    if mongo_pb_grpc and mongo_pb:
        try:
            import grpc
            addr = MONGO_SERVICE_ADDR
            if ":" not in addr:
                addr = f"{addr}:50010"
            with grpc.insecure_channel(addr) as channel:
                stub = mongo_pb_grpc.SettlementMongoServiceStub(channel)
                req = mongo_pb.GetSettlementAggregationsRequest(days=days)
                resp = stub.GetSettlementAggregations(req, timeout=3.0)
                if resp and resp.success and resp.avg_daily_settlement_paise > 0:
                    avg_daily_settlement_paise = int(resp.avg_daily_settlement_paise)
                    logging.info(f"Fetched gRPC settlement aggregation: avg_daily={avg_daily_settlement_paise} paise")
        except Exception as e:
            logging.warning(f"Could not query gRPC MongoService for settlement aggregations: {e}")

    projected_settlements = days * avg_daily_settlement_paise
    pending_recovery_estimate = int(projected_settlements * 0.15)
    net_position = projected_settlements + pending_recovery_estimate

    forecast_items = []
    for day in range(1, days + 1):
        forecast_items.append({
            "day": day,
            "expected_settlement_paise": avg_daily_settlement_paise,
            "pending_recovery_paise": int(avg_daily_settlement_paise * 0.15),
            "projected_cash_paise": int(avg_daily_settlement_paise * 1.15)
        })

    return {
        "horizon_days": days,
        "avg_daily_settlement_paise": avg_daily_settlement_paise,
        "total_projected_cash_paise": net_position,
        "daily_forecast": forecast_items
    }
