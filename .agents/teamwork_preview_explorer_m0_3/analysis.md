# RevenueIQ Comprehensive Audit Report: AI Gateway, Cash Forecast gRPC & B2B Promises

**Agent ID**: Explorer 3 (AI Gateway, Cash Forecast gRPC & B2B Promises Audit)  
**Working Directory**: `/home/mahi17/Github/fintech/.agents/teamwork_preview_explorer_m0_3`  
**Date**: 2026-08-23  

---

## Executive Summary

This audit evaluates the current implementation state of four key architectural pillars of RevenueIQ:
1. **AI Gateway Service**: Cash forecasting logic, chat endpoint handling, Strands tool integrations (`create_payment_link_tool`), and missing gRPC client connections.
2. **`mongo-service` gRPC & Settlement Aggregations**: Proto definitions (`mongodb_service.proto`), gRPC server implementations, MongoDB settlement collections, and missing gRPC settlement calculation endpoints needed for live cash forecasting.
3. **B2B Receivables / Promise-to-Pay**: `PromisesTab.tsx` frontend component, `/api/v1/promises` REST endpoint in `dashboard_api`, and the underlying `promises` MongoDB collection schema in `revenueiq_db`.
4. **Razorpay REST API Integration**: `create_payment_link_tool` execution flow, basic auth authentication, payload structures, and native Go Razorpay data sync in `dashboard_api`.

---

## Section 1: `ai_gateway` Service Audit

### 1.1 Key Files & Paths
- **Service Root**: `/home/mahi17/Github/fintech/ai_gateway`
- **Main FastAPI Entrypoint**: `/home/mahi17/Github/fintech/ai_gateway/src/main.py`
- **Forecast Logic Module**: `/home/mahi17/Github/fintech/ai_gateway/src/forecast.py`
- **Proto Stubs Directory**: `/home/mahi17/Github/fintech/ai_gateway/src/proto/mongo_service/`
  - `mongodb_service_pb2.py`
  - `mongodb_service_pb2_grpc.py`

### 1.2 `forecast_cash_position` Analysis
- **Location**: `ai_gateway/src/forecast.py:3-21` and `ai_gateway/src/main.py:267-270, 459-461`
- **Current Signature & Code**:
  ```python
  def forecast_cash_position(days: int = 7, avg_daily_settlement_paise: int = 2500000) -> dict:
      projected_settlements = days * avg_daily_settlement_paise
      pending_recovery_estimate = int(projected_settlements * 0.15) # 15% estimated recovery boost
      net_position = projected_settlements + pending_recovery_estimate
      ...
  ```
- **Audit Findings**:
  - The function relies on hardcoded default parameters (`avg_daily_settlement_paise = 2500000` = ₹25,000/day) and a static 15% recovery multiplier.
  - It does NOT query `mongodb_service` or MongoDB database to compute the actual historical daily settlement average or real pending recovery amounts.
  - **Endpoint**: `@app.get("/forecast")` in `ai_gateway/src/main.py:459-461` exposes this static function to HTTP clients.

### 1.3 Chat Tools & `create_payment_link_tool`
- **Location**: `ai_gateway/src/main.py:61-122`
- **Signature**: `create_payment_link_tool(name: str = "", email: str = "", phone: str = "", amount_inr: float = 0.0, description: str = "") -> str`
- **Validation Logic**:
  - Strictly requires all 4 parameters: customer name, email address, phone number (+91 formatting), and positive amount in INR.
  - If any parameter is missing, returns JSON error payload:
    ```json
    {
      "error": "MISSING_REQUIRED_DATA",
      "missing_fields": ["email address", "phone number"],
      "message": "Hey! You forgot to provide the following details: ..."
    }
    ```
- **Chat Request Interceptor**:
  - `ai_gateway/src/main.py:360-423` in `@app.post("/chat")`.
  - Parses regex matches for emails, 10-digit phone numbers, INR amounts, and customer names.
  - If details are missing, directly returns: `"Hey! You forgot to provide the following details: [missing fields]. Please provide these details so I can create the payment link."`
  - If all 4 details are extracted, calls `create_payment_link_tool` directly and formats markdown response with `short_url` and `payment_link_id`.

### 1.4 gRPC Client Setup in `ai_gateway`
- **Status**: **NOT IMPLEMENTED**
- Generated Python stubs exist in `ai_gateway/src/proto/mongo_service/`, but `src/main.py` and `src/forecast.py` do not import or create a `grpc.insecure_channel`.
- **Requirement**: `ai_gateway` needs a gRPC client wrapper to connect to `mongodb_service` (port `50010` or `50051`) to fetch historical daily settlement totals for N-day forecasting.

---

## Section 2: `mongo-service` & Cash Forecast gRPC Aggregations Audit

### 2.1 Key Files & Paths
- **Proto Schema**: `/home/mahi17/Github/fintech/revenueiq_dev_kit/proto/mongo_service/mongodb_service.proto`
- **Go gRPC Server Entrypoint**: `/home/mahi17/Github/fintech/mongodb_service/cmd/server/main.go`
- **Go gRPC Server Registration**: `/home/mahi17/Github/fintech/mongodb_service/internal/grpc_servers/mongo_service.go`
- **Database Handlers**: `/home/mahi17/Github/fintech/mongodb_service/internal/db/payment.go`

### 2.2 Existing gRPC Proto Services
In `mongodb_service.proto`:
- `UserService`, `OrderMongoService`, `DroneMongoService`, `FlightMongoService`, `OperationsMongoService`, `DeliveryMongoService`, `PaymentMongoService`, `WarehouseMongoService`, `AnalyticsMongoService`, `WorkflowMongoService`.
- **Gap**: There is currently **no `SettlementMongoService`** or RPC method for historical settlement aggregations (e.g. `GetHistoricalSettlements` or `GetSettlementAggregations`).

### 2.3 MongoDB Settlement Data Schema
- **Database Names**: `revenueiq_db` and `mongodb_service_db`
- **Collection Name**: `settlements`
- **Document Structure**:
  ```json
  {
    "settlement_id": "set_RZP_0001",
    "order_id": "order_RZP_0001",
    "payment_id": "pay_RZP_0001",
    "amount_paise": 245000,
    "utr": "UTR_23995886"
  }
  ```
- Seeded by `data/seed_mongodb.py` from `data/synthetic_dataset.json` (100 settlement records).

### 2.4 gRPC Upgrade Strategy for `forecast_cash_position`
To enable `forecast_cash_position` to calculate dynamic cash position forecasts:
1. Define a `GetSettlementAggregations` RPC in `mongodb_service.proto` (or within `PaymentMongoService` / `AnalyticsMongoService`):
   ```protobuf
   message GetSettlementAggregationsRequest {
     int32 days = 1;
   }
   message GetSettlementAggregationsResponse {
     int64 total_settled_paise = 1;
     int64 avg_daily_settlement_paise = 2;
     int32 count = 3;
   }
   ```
2. Implement the gRPC handler in `mongodb_service/internal/server/` and `mongodb_service/internal/db/` to run a MongoDB aggregation pipeline over `settlements` collection.
3. Recompile proto stubs for Python (`ai_gateway/src/proto/mongo_service/`).
4. Update `ai_gateway/src/forecast.py` to instantiate a gRPC stub to `mongodb_service:50010` and query historical average daily settlement data dynamically.

---

## Section 3: B2B Receivables & Promise-to-Pay Audit

### 3.1 Key Files & Paths
- **Frontend Component**: `/home/mahi17/Github/fintech/frontend/src/components/PromisesTab.tsx`
- **Frontend API Client**: `/home/mahi17/Github/fintech/frontend/src/lib/api.ts` (function `fetchPromises`)
- **Backend API Route**: `/home/mahi17/Github/fintech/dashboard_api/cmd/server/main.go` (endpoint `/api/v1/promises`)
- **Seed Scripts**: `/home/mahi17/Github/fintech/data/seed_mongodb.py` & `/home/mahi17/Github/fintech/data/seed_container_27018.py`

### 3.2 Frontend `PromisesTab.tsx` Implementation
- **Location**: `frontend/src/components/PromisesTab.tsx:1-82`
- **Behavior**:
  - Executes `fetchPromises()` on mount (`useEffect`).
  - Renders a 2-column grid of promise cards.
  - Displays customer name (`p.customer`), status badge (`p.status`: green `#ebfbee` for `KEPT`, yellow `#fff9db` for `PENDING`), amount formatted in INR (`₹(p.amount_paise / 100)`), promised date (`p.promised_date`), and workflow ID (`p.workflow_id`).

### 3.3 Backend REST Route `/api/v1/promises`
- **Location**: `dashboard_api/cmd/server/main.go:484-502`
- **Code**:
  ```go
  // DYNAMIC PROMISES
  api.GET("/promises", func(c *gin.Context) {
      if db != nil {
          ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
          defer cancel()

          cursor, err := db.Collection("promises").Find(ctx, bson.M{})
          if err == nil {
              var list []bson.M
              _ = cursor.All(ctx, &list)
              if list == nil {
                  list = []bson.M{}
              }
              c.JSON(http.StatusOK, gin.H{"promises": list, "total": len(list)})
              return
          }
      }
      c.JSON(http.StatusOK, gin.H{"promises": []interface{}{}, "total": 0})
  })
  ```

### 3.4 MongoDB `promises` Collection Schema
- **Database**: `revenueiq_db` (and `mongodb_service_db`)
- **Collection**: `promises`
- **Document Fields**:
  - `id` (string): Unique promise identifier (e.g. `"prm_301"`)
  - `customer` (string): Customer full name (e.g. `"Vikram Mehta"`)
  - `amount_paise` (int64): Promised amount in paise (e.g. `350000` = ₹3,500.00)
  - `promised_date` (string): Promised payment date ISO string (e.g. `"2026-08-25"`)
  - `status` (string): Promise compliance status (`"KEPT"`, `"PENDING"`, `"BROKEN"`)
  - `workflow_id` (string): Linked recovery workflow ID (e.g. `"wf_801"`)

---

## Section 4: Razorpay REST API Integration Audit

### 4.1 Payment Link Tool Integration
- **Location**: `ai_gateway/src/main.py:61-123`
- **API Endpoint**: `https://api.razorpay.com/v1/payment_links`
- **Auth**: Basic HTTP Auth using `RAZORPAY_KEY_ID` and `RAZORPAY_KEY_SECRET`.
  - Defaults: `RAZORPAY_KEY_ID="rzp_test_SESIqmsJZRvpZ1"`, `RAZORPAY_KEY_SECRET="4qzbtZYU4wrF4ER3lYk2hIt7"`.
- **Payload Sent**:
  ```json
  {
    "amount": amount_paise,
    "currency": "INR",
    "accept_partial": false,
    "description": description,
    "customer": {
      "name": name,
      "email": email,
      "contact": contact_phone
    },
    "notify": {"sms": true, "email": true},
    "reminder_enable": true
  }
  ```
- **Response Format**: On 200/201, returns JSON containing `payment_link_id`, `short_url`, `status`, `amount_inr`, `customer_name`, `customer_email`, `customer_phone`.

### 4.2 Native Razorpay Go Data Sync in `dashboard_api`
- **Location**: `dashboard_api/cmd/server/main.go:72-141` (`triggerRazorpaySync`)
- **Endpoints Synced**:
  - `https://api.razorpay.com/v1/settlements` -> MongoDB collection `settlements`
  - `https://api.razorpay.com/v1/subscriptions` -> MongoDB collection `subscriptions`
  - `https://api.razorpay.com/v1/payments` -> MongoDB collection `payments`
  - `https://api.razorpay.com/v1/orders` -> MongoDB collection `orders`
  - `https://api.razorpay.com/v1/disputes` -> MongoDB collection `disputes`
  - `https://api.razorpay.com/v1/refunds` -> MongoDB collection `refunds`
- Performs upserts into `revenueiq_db` database natively using Go MongoDB driver.

---

## Summary Matrix of Component Status

| Component | File Path | Existing Status | Missing / Required Actions |
|---|---|---|---|
| Cash Forecast Function | `ai_gateway/src/forecast.py` | Static default `avg_daily_settlement_paise=2500000` | Query historical settlements from `mongodb_service` via gRPC |
| AI Gateway gRPC Client | `ai_gateway/src/main.py` | Proto stubs present, channel uninitialized | Initialize gRPC channel to `mongodb_service:50010` |
| Settlement gRPC RPC | `revenueiq_dev_kit/proto/mongo_service/mongodb_service.proto` | No settlement service/RPC defined | Add `GetSettlementAggregations` RPC & Go implementation |
| Promise-to-Pay Component | `frontend/src/components/PromisesTab.tsx` | Renders cards via `fetchPromises()` | Complete & functional |
| Promises REST Endpoint | `dashboard_api/cmd/server/main.go:485` | `/api/v1/promises` queries Mongo `promises` col | Complete & functional |
| Razorpay Link Creation Tool | `ai_gateway/src/main.py:61` | Live HTTP POST to Razorpay REST API | Complete & functional with validation |
