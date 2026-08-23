# Handoff Report: Explorer 3 (AI Gateway, Cash Forecast gRPC & B2B Promises Audit)

## 1. Observation

- **AI Gateway & Forecast Logic**:
  - `ai_gateway/src/forecast.py:3-21`: `forecast_cash_position` takes a default static `avg_daily_settlement_paise: int = 2500000` (₹25,000) and computes projected settlements as `days * avg_daily_settlement_paise` with a static 15% recovery boost.
  - `ai_gateway/src/main.py:61-122`: `@tool def create_payment_link_tool(...)` validates all 4 mandatory input parameters (`name`, `email`, `phone`, `amount_inr` > 0). When missing parameters, returns JSON with `"error": "MISSING_REQUIRED_DATA"`. When valid, sends HTTP POST to `https://api.razorpay.com/v1/payment_links` via `httpx` with basic auth using `RAZORPAY_KEY_ID` (`rzp_test_SESIqmsJZRvpZ1`) and `RAZORPAY_KEY_SECRET` (`4qzbtZYU4wrF4ER3lYk2hIt7`).
  - `ai_gateway/src/main.py:360-423`: `@app.post("/chat")` extracts payment link details using regex, returns warning message if parameters are missing, or invokes `create_payment_link_tool` when all 4 parameters are present.
  - `ai_gateway/src/proto/mongo_service/mongodb_service_pb2.py` and `mongodb_service_pb2_grpc.py` exist in Python, but no gRPC channel is created in `ai_gateway/src/main.py` or `forecast.py`.

- **Mongo Service & Settlement Aggregations**:
  - `revenueiq_dev_kit/proto/mongo_service/mongodb_service.proto`: Defines 10 gRPC services (`UserService`, `OrderMongoService`, `DroneMongoService`, `FlightMongoService`, `OperationsMongoService`, `DeliveryMongoService`, `PaymentMongoService`, `WarehouseMongoService`, `AnalyticsMongoService`, `WorkflowMongoService`). It currently lacks any RPC method for querying historical settlement aggregations.
  - `mongodb_service/internal/grpc_servers/mongo_service.go:26-35`: Registers all 10 gRPC services on TCP listener (default port 50051 / 50010).
  - MongoDB database `revenueiq_db` / `mongodb_service_db` contains collection `settlements` (seeded from `data/synthetic_dataset.json` with 100 settlement documents).

- **B2B Receivables / Promise-to-Pay**:
  - `frontend/src/components/PromisesTab.tsx:1-82`: Component mounts, calls `fetchPromises()`, and displays promise cards (customer name, status tag, amount in INR, promised date, workflow ID).
  - `frontend/src/lib/api.ts:39-44`: `fetchPromises()` issues HTTP GET request to `http://localhost:8005/api/v1/promises`.
  - `dashboard_api/cmd/server/main.go:485-502`: `api.GET("/promises", ...)` queries MongoDB collection `promises` in `revenueiq_db` and returns `{ "promises": list, "total": len(list) }`.
  - `data/seed_mongodb.py:135-164`: Seeds `promises` collection with 3 initial records (`"prm_301"`, `"prm_302"`, `"prm_303"`).

---

## 2. Logic Chain

1. **AI Gateway Forecasting Incoherence**:
   - *Observation*: `forecast_cash_position` in `ai_gateway/src/forecast.py` uses hardcoded default `2500000` paise (₹25,000) and static 15% multiplier without reading database records.
   - *Logic*: To satisfy F21 Live Cash Forecast specifications, `forecast_cash_position` must query actual historical settlement aggregations from MongoDB via `mongodb_service` over gRPC.

2. **Proto Schema Deficit for Settlements**:
   - *Observation*: `mongodb_service.proto` defines services for payments, orders, analytics, workflows, etc., but lacks any `GetSettlementAggregations` or `GetHistoricalSettlements` RPC.
   - *Logic*: Before `ai_gateway` can query historical settlements over gRPC, `mongodb_service.proto` must be extended with a settlement aggregation RPC, and the Go handler implemented in `mongodb_service/internal/server/` and `mongodb_service/internal/db/`.

3. **Payment Link Creation Readiness**:
   - *Observation*: `create_payment_link_tool` in `ai_gateway/src/main.py` is fully implemented with strict 4-field validation and live Razorpay REST API invocation (`https://api.razorpay.com/v1/payment_links`).
   - *Logic*: The tool is ready for live use; any chat query containing all 4 required fields (name, email, phone, amount) will immediately trigger live Razorpay link generation.

4. **B2B Promises Protocol Integrity**:
   - *Observation*: `PromisesTab.tsx` calls `fetchPromises()`, which hits `/api/v1/promises` in `dashboard_api`, which reads collection `promises` from `revenueiq_db`.
   - *Logic*: The end-to-end data pipeline from React UI through Go REST API down to MongoDB collection `promises` is already connected and operational.

---

## 3. Caveats

- **Network Isolation**: Operating in CODE_ONLY mode, so actual live HTTP requests to `https://api.razorpay.com` were not executed during this read-only audit.
- **Service Ports**: `mongodb_service` runs on port 50051 by default, but `dashboard_api` configures `MONGO_SERVICE_ADDR` default to `localhost:50010`. The gRPC connection string in `ai_gateway` must align with the running container port configuration.

---

## 4. Conclusion

1. **B2B Promises & Receivables**: Fully wired end-to-end from `PromisesTab.tsx` -> `fetchPromises()` -> `/api/v1/promises` -> MongoDB `promises` collection.
2. **Razorpay REST API Integration**: `create_payment_link_tool` in `ai_gateway/src/main.py` is fully implemented with strict parameter guardrails and live Razorpay REST API payload formatting.
3. **Cash Position Forecast (F21)**: Requires implementation of a new gRPC endpoint (e.g. `GetSettlementAggregations`) in `mongodb_service.proto` and `mongodb_service` Go server, along with gRPC client initialization in `ai_gateway/src/forecast.py` to replace static simulated values with dynamic historical settlement queries.

---

## 5. Verification Method

To independently verify these observations and conclusions:
1. **Inspect Forecast Logic**:
   - View `ai_gateway/src/forecast.py` (lines 3-21) to confirm static hardcoded parameters.
2. **Inspect Payment Link Tool**:
   - View `ai_gateway/src/main.py` (lines 61-122) to confirm Razorpay REST API endpoint, payload structure, and basic auth credentials.
3. **Inspect B2B Promises Pipeline**:
   - View `frontend/src/components/PromisesTab.tsx` (lines 9-14) for `fetchPromises()` call.
   - View `dashboard_api/cmd/server/main.go` (lines 485-502) for Mongo `promises` query.
   - View `data/seed_mongodb.py` (lines 135-164) for `promises` collection document structure.
4. **Inspect gRPC Proto Definitions**:
   - View `revenueiq_dev_kit/proto/mongo_service/mongodb_service.proto` to confirm absence of settlement RPCs.
