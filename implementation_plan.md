# RecoverIQ — Final Master Build Document

> **One file. Everything you need to build the complete system.**

---

## WHAT WE'RE BUILDING

**RecoverIQ** — an AI-powered revenue recovery and finance control platform for Razorpay merchants. It detects failed payments, diagnoses root causes, executes bounded recovery workflows (including voice calls, promise tracking, and checkout recovery), reconciles settlements, and reports everything through an auditable dashboard.

**Primary Track**: Track 03 — AI Revenue Recovery
**Secondary Track**: Track 04 — AI Finance Controller

---

## CORE PRINCIPLES

1. **Every money action is explainable** — Full audit trail with reasoning
2. **Every action is bounded** — Hard-coded guardrails (not AI-controlled)
3. **Every action is gated** — Human approval for high-value/uncertain actions
4. **Failure is a feature** — Circuit breakers, DLQs, idempotency, graceful degradation
5. **AI where it matters** — LLM for diagnosis and reasoning; rules for everything else

---

## FEATURE COVERAGE — TRACK 03 & TRACK 04

### Track 03: AI Revenue Recovery (7/7 Directions Covered)

| # | Direction | Status | Implementation |
|---|---|---|---|
| 1 | Payment degradation → root cause → recovery | ✅ Core | `failure-detector` diagnoses → `recovery-orchestrator` executes bounded workflow |
| 2 | Checkout drop-off recovery | ✅ Added | `checkout-tracker` emits synthetic `checkout.abandoned` events → recovery via email + payment link |
| 3 | Failed-subscription recovery | ✅ Added | `subscription.charged.failed` webhook → `SUBSCRIPTION_FAILED` category → retry charge or card-update link |
| 4 | Mandate retry sequencer | ✅ Added | Mandate error codes (`mandate_expired`, `debit_rejected`) mapped → scheduled retry sequence |
| 5 | Hinglish voice recovery | ✅ Added | `voice-recovery-worker` (Go) → TTS generation (AWS Polly / Google TTS) → Exotel IVR mock → Hinglish templates |
| 6 | Promise-to-pay tracker | ✅ Added | Customer commits to pay by date → stored in `promises` collection → follow-up scheduler → compliance tracking |
| 7 | B2B receivables chaser | ⚠️ Partial | Handled via promise-to-pay + escalation ladder. Full invoice management out of scope. |

### Track 03 Evaluation Bar ✅

| Criterion | How We Meet It |
|---|---|
| Measured money recovered across a batch | `RecoveryStats`: `total_at_risk`, `total_recovered`, `recovery_rate` across 60+ synthetic failures |
| Compliant escalation | Guardrails → escalation_agent → human review tickets with severity |
| Stopping rules | Max retries, contact limits, cost caps, circuit breaker, workflow age limit |
| Audit trail | Every action logged to audit-service with reasoning, guardrails checked, actor, duration |

### Track 04: AI Finance Controller (3/4 Directions Covered)

| # | Direction | Status | Implementation |
|---|---|---|---|
| 1 | Multi-source reconciliation | ✅ Core | Three-way match: Orders ↔ Payments ↔ Settlements. Tiered: exact → fuzzy → AI → exception |
| 2 | Settlement Q&A agent | ✅ Core | `master_agent` answers: "Why was settlement short?", "Which orders aren't settled?" |
| 3 | Forward cash forecaster | ✅ Added | `forecast_cash_position` tool — extrapolate from historical settlements + pending recoveries |
| 4 | Tax-line matcher | ❌ Skip | Requires GST filing data. Out of scope. |

### Track 04 Evaluation Bar ✅

| Criterion | How We Meet It |
|---|---|
| Throughput | 50+ records in single reconciliation run |
| Measured accuracy | `match_rate` with breakdown: exact/fuzzy/AI/unmatched |
| Honest exception list | Typed exceptions with AI suggestions, amounts, IDs |

---

## COMPLETE FEATURE LIST (29 Features)

```
F1.  Webhook Ingestion & Signature Validation        [webhook-receiver]      Go     ❌ AI
F2.  Payment Failure Classification (Rule-Based)      [failure-detector]      Python ❌ AI
F3.  AI Root Cause Diagnosis                          [failure-detector]      Python ✅ AI
F4.  Systemic Pattern Detection                       [failure-detector]      Python ✅ AI
F5.  Recovery Strategy Selection                      [recovery-orchestrator] Python ✅ AI
F6.  Payment Retry Execution                          [recovery-orchestrator] Python ❌ AI
F7.  Payment Link Generation                          [recovery-orchestrator] Python ❌ AI
F8.  Recovery Email Notifications                     [notification-worker]   Go     ❌ AI
F9.  Escalation Management                            [recovery-orchestrator] Python ✅ AI
F10. Recovery Workflow Orchestration (GraphBuilder)    [recovery-orchestrator] Python ✅ AI
F11. Guardrails & Compliance (Hard-coded)             [recovery-orchestrator] Python ❌ AI
F12. Subscription Failure Recovery                    [failure-detector]      Python ❌ AI
F13. Mandate/AutoPay Retry Handling                   [failure-detector]      Python ❌ AI
F14. Checkout Drop-off Recovery                       [checkout-tracker]      Go     ❌ AI
F15. Promise-to-Pay Tracker                           [recovery-orchestrator] Python ✅ AI
F16. Hinglish Voice Recovery                          [voice-recovery-worker] Go     ❌ AI
F17. Multi-Source Data Ingestion (Recon)              [reconciliation-engine] Python ❌ AI
F18. Exact Matching                                   [reconciliation-engine] Python ❌ AI
F19. Fuzzy Matching                                   [reconciliation-engine] Python ❌ AI
F20. AI Exception Resolution                         [reconciliation-engine] Python ✅ AI
F21. Reconciliation Reporting                         [reconciliation-engine] Python ❌ AI
F22. Exception Management                            [reconciliation-engine] Python ❌ AI
F23. Settlement Q&A                                   [ai-gateway]           Python ✅ AI
F24. Cash Position Forecast                           [ai-gateway]           Python ✅ AI
F25. Full Audit Trail                                 [audit-service]        Go     ❌ AI
F26. AI Chat Interface                                [ai-gateway]           Python ✅ AI
F27. Real-Time Dashboard                              [dashboard-api]        Go     ❌ AI
F28. Synthetic Data Engine                            [data/]                Python ❌ AI
F29. Circuit Breaker + Idempotency                    [redis-service]        Go     ❌ AI
```

**AI count**: 11 features use AI, 18 do not — showing strong "AI judgment" (right tool, right place).

---

## ARCHITECTURE

```
                    ┌──────────────────────────────────────────┐
                    │           EXTERNAL BOUNDARY              │
                    │                                          │
                    │  Razorpay Webhooks ──HTTP──▶ webhook-receiver
                    │  Browser/Frontend  ──HTTP──▶ dashboard-api
                    │  Browser/Frontend  ──SSE───▶ ai-gateway
                    │  Frontend Session  ──HTTP──▶ checkout-tracker
                    └──────────────┬───────────────────────────┘
                                  │
       ┌──────────────────────────▼──────────────────────────────────────┐
       │                 gRPC SERVICE MESH (backend network)             │
       │                                                                 │
       │  ┌─── GO SERVICES (7) ──────────────────────────────────────┐  │
       │  │                                                          │  │
       │  │  mongo-service          :50010    MongoDB access layer   │  │
       │  │  redis-service          :50011    Redis access layer     │  │
       │  │  webhook-receiver       :50001/:8001  Webhook ingestion  │  │
       │  │  audit-service          :50007    Audit trail            │  │
       │  │  notification-worker    :8083     Email dispatch         │  │
       │  │  checkout-tracker       :50008/:8008  Session tracking   │  │
       │  │  voice-recovery-worker  :8009     Hinglish IVR calls    │  │
       │  │  dashboard-api          :50005/:8005  REST API gateway   │  │
       │  │                                                          │  │
       │  └──────────────────────────────────────────────────────────┘  │
       │                                                                 │
       │  ┌─── PYTHON SERVICES (4 — AI) ─────────────────────────────┐  │
       │  │                                                          │  │
       │  │  failure-detector         :50002   Classification + AI   │  │
       │  │  recovery-orchestrator    :50003   Strategy + execution  │  │
       │  │  reconciliation-engine    :50004   Settlement matching   │  │
       │  │  ai-gateway               :50006/:8006  Master agent    │  │
       │  │                                                          │  │
       │  └──────────────────────────────────────────────────────────┘  │
       └─────────────────────────────────────────────────────────────────┘

       ┌─────────────────────────────────────────────────────────────────┐
       │                      INFRASTRUCTURE                             │
       │  MongoDB :27017  │  Redis :6379  │  Redpanda :9092              │
       └─────────────────────────────────────────────────────────────────┘
```

### Service Registry

| # | Service | gRPC | HTTP | Lang | AI | Role |
|---|---|---|---|---|---|---|
| 1 | `mongo-service` | 50010 | — | Go | ❌ | MongoDB CRUD + aggregation |
| 2 | `redis-service` | 50011 | — | Go | ❌ | Cache, idempotency, rate limiting, circuit breaker |
| 3 | `webhook-receiver` | 50001 | 8001 | Go | ❌ | Razorpay webhook ingestion + signature validation |
| 4 | `audit-service` | 50007 | — | Go | ❌ | Audit trail logging + querying (all comm via gRPC) |
| 5 | `notification-worker` | — | 8083 | Go | ❌ | Email dispatch (Redpanda consumer → SES) |
| 6 | `checkout-tracker` | 50008 | 8008 | Go | ❌ | Checkout session tracking + abandonment detection |
| 7 | `voice-recovery-worker` | — | 8009 | Go | ❌ | Hinglish TTS generation + IVR dispatch |
| 8 | `dashboard-api` | 50005 | 8005 | Go | ❌ | REST API gateway for frontend |
| 9 | `failure-detector` | 50002 | — | Python | ✅ | Failure classification + pattern detection |
| 10 | `recovery-orchestrator` | 50003 | — | Python | ✅ | Recovery strategy, execution, promises, guardrails |
| 11 | `reconciliation-engine` | 50004 | — | Python | ✅ | Settlement matching (exact → fuzzy → AI) |
| 12 | `ai-gateway` | 50006 | 8006 | Python | ✅ | Master agent + chat + forecast |
| 13 | `frontend` | — | 3000 | TS | ❌ | Next.js minimalist dashboard |

---

## TECHNOLOGY STACK

| Layer | Go Services | Python Services |
|---|---|---|
| **gRPC** | google.golang.org/grpc | grpcio, grpcio-tools |
| **HTTP** | github.com/gin-gonic/gin | FastAPI + uvicorn |
| **MongoDB** | go.mongodb.org/mongo-driver/v2 | *(via mongo-service gRPC)* |
| **Redis** | github.com/redis/go-redis/v9 | *(via redis-service gRPC)* |
| **Kafka** | github.com/twmb/franz-go | confluent-kafka / aiokafka |
| **Razorpay** | *(HTTP client)* | razorpay Python SDK |
| **AI Agents** | — | strands-agents, strands-agents-tools |
| **LLM** | — | OpenAIModel → Bedrock Mantle (mistral.ministral-3-8b-instruct) |
| **TTS** | AWS Polly SDK / Google TTS | — |
| **Telephony** | Exotel Go SDK (or mock) | — |

| Frontend | Technology |
|---|---|
| Framework | Next.js 15 + React 19 |
| Charts | Recharts |
| Icons | Lucide React |
| Font | Inter (Google Fonts) |

---

## PROJECT STRUCTURE

```
recoveriq/
├── docker-compose.yml
├── .env.example
├── .env
├── README.md
├── AUDIT_TRAIL.md
├── go.work                                 # Links all Go modules
│
├── proto/                                  # Shared protobuf contracts (9 files)
│   ├── common/common.proto
│   ├── mongo/mongo_service.proto
│   ├── redis/redis_service.proto
│   ├── webhook/webhook.proto
│   ├── failure/failure.proto
│   ├── recovery/recovery.proto
│   ├── reconciliation/reconciliation.proto
│   ├── audit/audit.proto
│   ├── notification/notification.proto
│   ├── checkout/checkout.proto             # NEW
│   ├── voice/voice.proto                   # NEW
│   ├── compile_go.sh
│   └── compile_python.sh
│
├── pkg/                                    # Shared Go package
│   ├── go.mod
│   ├── constants/
│   │   ├── topics.go
│   │   ├── namespaces.go
│   │   └── error_codes.go                 # 40+ Razorpay error codes + mandate + subscription codes
│   ├── grpcutil/
│   │   ├── interceptors.go
│   │   ├── health.go
│   │   └── errors.go
│   └── config/
│       └── env.go
│
├── shared/                                 # Shared Python package
│   ├── __init__.py
│   ├── config.py
│   ├── constants.py
│   ├── guardrails.py
│   ├── grpc_utils.py
│   ├── razorpay_client.py
│   ├── logging_config.py
│   └── generated/                          # Proto Python stubs
│
├── services/
│   │
│   │  ═══ GO SERVICES (8) ═══
│   │
│   ├── mongo_service/                      # Standard Go service structure
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   ├── cmd/server/main.go
│   │   └── internal/
│   │       ├── grpc_server/server.go
│   │       ├── handler/{insert,find,update,delete,aggregate}.go
│   │       ├── db/{connection,serialization}.go
│   │       └── config/config.go
│   │
│   ├── redis_service/
│   │   ├── go.mod, Dockerfile, cmd/server/main.go
│   │   └── internal/
│   │       ├── grpc_server/server.go
│   │       ├── handler/{basic,counter,rate_limit,idempotency}.go
│   │       ├── client/redis.go
│   │       ├── namespace/validator.go
│   │       └── config/config.go
│   │
│   ├── webhook_receiver/
│   │   ├── go.mod, Dockerfile, cmd/server/main.go
│   │   └── internal/
│   │       ├── grpc_server/server.go
│   │       ├── handler/webhook.go          # POST /webhook/razorpay
│   │       ├── validator/signature.go      # HMAC SHA256
│   │       ├── publisher/redpanda.go       # → PAYMENT_EVENTS topic
│   │       ├── grpc_clients/{mongo,audit}.go
│   │       └── config/config.go
│   │
│   ├── audit_service/
│   │   ├── go.mod, Dockerfile, cmd/server/main.go
│   │   └── internal/
│   │       ├── grpc_server/server.go       # LogAction, GetAuditTrail, GetActionsByEntity
│   │       ├── handler/{log,query}.go
│   │       ├── grpc_clients/{mongo,redis}.go
│   │       ├── consumer/audit_events.go    # Redpanda batch consumer
│   │       └── config/config.go
│   │
│   ├── notification_worker/
│   │   ├── go.mod, Dockerfile, cmd/server/main.go
│   │   └── internal/
│   │       ├── consumer/notifications.go   # Redpanda → NOTIFICATIONS
│   │       ├── email/{ses,mock,templates}.go
│   │       ├── grpc_clients/audit.go
│   │       ├── health/health.go            # HTTP /healthz :8083
│   │       └── config/config.go
│   │
│   ├── checkout_tracker/                   # NEW — Checkout abandonment
│   │   ├── go.mod, Dockerfile, cmd/server/main.go
│   │   └── internal/
│   │       ├── grpc_server/server.go       # StartSession, Heartbeat, Complete, GetAbandoned
│   │       ├── handler/
│   │       │   ├── session.go              # POST /checkout/start, POST /checkout/heartbeat
│   │       │   └── complete.go             # POST /checkout/complete
│   │       ├── detector/
│   │       │   └── abandonment.go          # Background goroutine: check for stale sessions
│   │       ├── publisher/redpanda.go       # → CHECKOUT_EVENTS topic
│   │       ├── grpc_clients/{mongo,redis,audit}.go
│   │       └── config/config.go
│   │
│   ├── voice_recovery_worker/              # NEW — Hinglish voice calls
│   │   ├── go.mod, Dockerfile, cmd/server/main.go
│   │   └── internal/
│   │       ├── consumer/voice_tasks.go     # Redpanda → VOICE_RECOVERY
│   │       ├── tts/
│   │       │   ├── polly.go                # AWS Polly TTS client
│   │       │   └── mock.go                 # Mock TTS (logs text)
│   │       ├── ivr/
│   │       │   ├── exotel.go               # Exotel IVR client
│   │       │   └── mock.go                 # Mock IVR (logs call)
│   │       ├── templates/
│   │       │   └── hinglish.go             # Hinglish message templates
│   │       ├── grpc_clients/audit.go
│   │       ├── health/health.go            # HTTP /healthz :8009
│   │       └── config/config.go
│   │
│   ├── dashboard_api/
│   │   ├── go.mod, Dockerfile, cmd/server/main.go
│   │   └── internal/
│   │       ├── grpc_server/server.go
│   │       ├── handler/{overview,failures,recoveries,reconciliation,audit,analytics,promises}.go
│   │       ├── middleware/{cors,request_id,error_handler}.go
│   │       ├── grpc_clients/{failure_detector,recovery,reconciliation,audit,mongo,checkout}.go
│   │       └── config/config.go
│   │
│   │  ═══ PYTHON SERVICES (4 — AI only) ═══
│   │
│   ├── failure_detector/
│   │   ├── Dockerfile, pyproject.toml
│   │   └── src/
│   │       ├── main.py, config.py
│   │       ├── grpc_server.py              # FailureDetectorService RPCs
│   │       ├── grpc_clients.py
│   │       ├── consumer.py                 # PAYMENT_EVENTS + CHECKOUT_EVENTS
│   │       ├── classifier.py               # 40+ error codes + subscription + mandate
│   │       ├── pattern_detector.py
│   │       └── agents/
│   │           ├── tools.py
│   │           └── diagnosis_agent.py
│   │
│   ├── recovery_orchestrator/
│   │   ├── Dockerfile, pyproject.toml
│   │   └── src/
│   │       ├── main.py, config.py
│   │       ├── grpc_server.py              # RecoveryService RPCs
│   │       ├── grpc_clients.py
│   │       ├── consumer.py                 # FAILURE_DIAGNOSED
│   │       ├── guardrails.py
│   │       ├── circuit_breaker.py
│   │       ├── state_machine.py
│   │       ├── promise_tracker.py          # NEW — Promise-to-pay logic
│   │       └── agents/
│   │           ├── tools.py                # All @tool definitions
│   │           ├── strategy_agent.py
│   │           ├── executor_agents.py
│   │           ├── promise_agent.py        # NEW — Promise follow-up
│   │           └── recovery_graph.py       # GraphBuilder workflow
│   │
│   ├── reconciliation_engine/
│   │   ├── Dockerfile, pyproject.toml
│   │   └── src/
│   │       ├── main.py, config.py
│   │       ├── grpc_server.py
│   │       ├── grpc_clients.py
│   │       ├── exact_matcher.py
│   │       ├── fuzzy_matcher.py
│   │       ├── reporter.py
│   │       └── agents/
│   │           ├── tools.py
│   │           └── reconciliation_agent.py
│   │
│   └── ai_gateway/
│       ├── Dockerfile, pyproject.toml
│       └── src/
│           ├── main.py, config.py
│           ├── grpc_server.py
│           ├── grpc_clients.py
│           ├── session_manager.py
│           └── agents/
│               ├── tools.py                # All gRPC wrapper tools
│               └── master_agent.py
│
├── frontend/                               # Next.js 15 — Minimalist Light Theme
│   ├── package.json, next.config.ts, tsconfig.json
│   ├── Dockerfile
│   ├── app/
│   │   ├── globals.css                     # Light theme design system
│   │   ├── layout.tsx                      # Root: Inter font, sidebar
│   │   ├── page.tsx                        # Overview dashboard
│   │   ├── failures/page.tsx
│   │   ├── failures/[id]/page.tsx          # Detail + timeline
│   │   ├── recoveries/page.tsx
│   │   ├── promises/page.tsx               # NEW — Promise tracker view
│   │   ├── reconciliation/page.tsx
│   │   ├── audit/page.tsx
│   │   ├── chat/page.tsx
│   │   └── settings/page.tsx
│   ├── components/
│   │   ├── Sidebar.tsx
│   │   ├── MetricsCard.tsx
│   │   ├── DataTable.tsx
│   │   ├── StatusBadge.tsx
│   │   ├── Timeline.tsx
│   │   ├── ChatInterface.tsx
│   │   ├── PromiseCard.tsx                 # NEW
│   │   └── Charts/{TrendChart,DonutChart}.tsx
│   └── lib/{api,types}.ts
│
├── data/
│   ├── synthetic_generator.py
│   ├── init_db.py
│   ├── razorpay_simulator.py
│   └── seed_data/
│
└── tests/
    ├── integration/
    └── e2e/
```

---

## PROTOBUF CONTRACTS

### proto/common/common.proto

```protobuf
syntax = "proto3";
package recoveriq.common;

import "google/protobuf/timestamp.proto";

message StatusResponse {
  bool success = 1;
  string error_code = 2;
  string error_message = 3;
  string request_id = 4;
}

message Money {
  int64 amount_paise = 1;
  string currency = 2;    // "INR"
}

message PaginationRequest {
  int32 page = 1;
  int32 page_size = 2;
}

message PaginationResponse {
  int32 total = 1;
  int32 page = 2;
  int32 page_size = 3;
  bool has_more = 4;
}

message TimeRange {
  google.protobuf.Timestamp start = 1;
  google.protobuf.Timestamp end = 2;
}
```

### proto/mongo/mongo_service.proto

```protobuf
syntax = "proto3";
package recoveriq.mongo;

import "google/protobuf/struct.proto";
import "common/common.proto";

service MongoService {
  rpc InsertOne(InsertRequest) returns (InsertResponse);
  rpc FindOne(FindRequest) returns (FindResponse);
  rpc FindMany(FindManyRequest) returns (FindManyResponse);
  rpc UpdateOne(UpdateRequest) returns (UpdateResponse);
  rpc DeleteOne(DeleteRequest) returns (common.StatusResponse);
  rpc Aggregate(AggregateRequest) returns (AggregateResponse);
  rpc BulkInsert(BulkInsertRequest) returns (BulkInsertResponse);
}

message InsertRequest {
  string request_id = 1;
  string collection = 2;
  google.protobuf.Struct document = 3;
}
message InsertResponse {
  common.StatusResponse status = 1;
  string inserted_id = 2;
}
message FindRequest {
  string request_id = 1;
  string collection = 2;
  google.protobuf.Struct filter = 3;
  google.protobuf.Struct projection = 4;
}
message FindResponse {
  common.StatusResponse status = 1;
  google.protobuf.Struct document = 2;
  bool found = 3;
}
message FindManyRequest {
  string request_id = 1;
  string collection = 2;
  google.protobuf.Struct filter = 3;
  google.protobuf.Struct projection = 4;
  google.protobuf.Struct sort = 5;
  int32 limit = 6;
  int32 skip = 7;
}
message FindManyResponse {
  common.StatusResponse status = 1;
  repeated google.protobuf.Struct documents = 2;
  int32 total_count = 3;
}
message UpdateRequest {
  string request_id = 1;
  string collection = 2;
  google.protobuf.Struct filter = 3;
  google.protobuf.Struct update = 4;
  bool upsert = 5;
}
message UpdateResponse {
  common.StatusResponse status = 1;
  int32 matched_count = 2;
  int32 modified_count = 3;
}
message AggregateRequest {
  string request_id = 1;
  string collection = 2;
  repeated google.protobuf.Struct pipeline = 3;
}
message AggregateResponse {
  common.StatusResponse status = 1;
  repeated google.protobuf.Struct results = 2;
}
message BulkInsertRequest {
  string request_id = 1;
  string collection = 2;
  repeated google.protobuf.Struct documents = 3;
}
message BulkInsertResponse {
  common.StatusResponse status = 1;
  repeated string inserted_ids = 2;
}
```

### proto/redis/redis_service.proto

```protobuf
syntax = "proto3";
package recoveriq.redis;

import "common/common.proto";

service RedisService {
  rpc Set(SetRequest) returns (common.StatusResponse);
  rpc Get(GetRequest) returns (GetResponse);
  rpc Delete(DeleteRequest) returns (common.StatusResponse);
  rpc Exists(ExistsRequest) returns (ExistsResponse);
  rpc SetWithTTL(SetWithTTLRequest) returns (common.StatusResponse);
  rpc IncrBy(IncrByRequest) returns (IncrByResponse);
  rpc Expire(ExpireRequest) returns (common.StatusResponse);
  rpc CheckRateLimit(RateLimitRequest) returns (RateLimitResponse);
  rpc CheckAndSetIdempotency(IdempotencyRequest) returns (IdempotencyResponse);
}

message SetRequest { string request_id = 1; string namespace = 2; string key = 3; string value = 4; }
message GetRequest { string request_id = 1; string namespace = 2; string key = 3; }
message GetResponse { common.StatusResponse status = 1; string value = 2; bool found = 3; }
message DeleteRequest { string request_id = 1; string namespace = 2; string key = 3; }
message ExistsRequest { string request_id = 1; string namespace = 2; string key = 3; }
message ExistsResponse { common.StatusResponse status = 1; bool exists = 2; }
message SetWithTTLRequest { string request_id = 1; string namespace = 2; string key = 3; string value = 4; int32 ttl_seconds = 5; }
message IncrByRequest { string request_id = 1; string namespace = 2; string key = 3; int64 increment = 4; }
message IncrByResponse { common.StatusResponse status = 1; int64 new_value = 2; }
message ExpireRequest { string request_id = 1; string namespace = 2; string key = 3; int32 ttl_seconds = 4; }
message RateLimitRequest { string request_id = 1; string key = 2; int32 max_requests = 3; int32 window_seconds = 4; }
message RateLimitResponse { common.StatusResponse status = 1; bool allowed = 2; int32 remaining = 3; int32 retry_after_seconds = 4; }
message IdempotencyRequest { string request_id = 1; string idempotency_key = 2; string value = 3; int32 ttl_seconds = 4; }
message IdempotencyResponse { common.StatusResponse status = 1; bool is_new = 2; string existing_value = 3; }
```

### proto/webhook/webhook.proto

```protobuf
syntax = "proto3";
package recoveriq.webhook;

import "google/protobuf/timestamp.proto";
import "google/protobuf/struct.proto";
import "common/common.proto";

service WebhookService {
  rpc GetWebhookEvent(GetEventRequest) returns (WebhookEvent);
  rpc ListWebhookEvents(ListEventsRequest) returns (ListEventsResponse);
}

message WebhookEvent {
  string id = 1;
  string razorpay_event_id = 2;
  string event_type = 3;          // payment.failed, payment.captured, subscription.charged.failed, etc.
  google.protobuf.Struct payload = 4;
  google.protobuf.Timestamp received_at = 5;
  string signature_valid = 6;
  string processing_status = 7;
}
message GetEventRequest { string request_id = 1; string event_id = 2; }
message ListEventsRequest {
  string request_id = 1;
  string event_type = 2;
  common.TimeRange time_range = 3;
  common.PaginationRequest pagination = 4;
}
message ListEventsResponse {
  common.StatusResponse status = 1;
  repeated WebhookEvent events = 2;
  common.PaginationResponse pagination = 3;
}
```

### proto/failure/failure.proto

```protobuf
syntax = "proto3";
package recoveriq.failure;

import "google/protobuf/timestamp.proto";
import "common/common.proto";

service FailureDetectorService {
  rpc DiagnoseFailure(DiagnoseRequest) returns (DiagnosisResult);
  rpc GetFailureById(GetFailureRequest) returns (FailureRecord);
  rpc ListFailures(ListFailuresRequest) returns (ListFailuresResponse);
  rpc GetFailureStats(StatsRequest) returns (FailureStats);
  rpc DetectPatterns(PatternRequest) returns (PatternReport);
}

enum FailureCategory {
  UNKNOWN = 0;
  INSUFFICIENT_FUNDS = 1;
  CARD_EXPIRED = 2;
  BANK_DECLINE = 3;
  NETWORK_ERROR = 4;
  AUTHENTICATION_FAILED = 5;
  FRAUD_SUSPECTED = 6;
  INVALID_CARD = 7;
  INTERNATIONAL_BLOCKED = 8;
  LIMIT_EXCEEDED = 9;
  SUBSCRIPTION_FAILED = 10;        // NEW
  MANDATE_FAILED = 11;             // NEW
  CHECKOUT_ABANDONED = 12;         // NEW
}

enum RecoverySuggestion {
  NO_SUGGESTION = 0;
  RETRY_SAME_METHOD = 1;
  RETRY_DIFFERENT_METHOD = 2;
  SEND_PAYMENT_LINK = 3;
  WAIT_AND_RETRY = 4;
  CONTACT_CUSTOMER = 5;
  ESCALATE_TO_HUMAN = 6;
  DO_NOT_RETRY = 7;
  RETRY_SUBSCRIPTION = 8;          // NEW
  UPDATE_CARD_LINK = 9;            // NEW
  RENEW_MANDATE = 10;              // NEW
  VOICE_RECOVERY = 11;             // NEW
  CHECKOUT_NUDGE = 12;             // NEW
}

message DiagnoseRequest {
  string request_id = 1;
  string payment_id = 2;
  string razorpay_error_code = 3;
  string razorpay_error_description = 4;
  string payment_method = 5;
  common.Money amount = 6;
  string customer_id = 7;
  string merchant_id = 8;
  string subscription_id = 9;      // NEW — optional, for subscription failures
  string mandate_id = 10;          // NEW — optional, for mandate failures
  string checkout_session_id = 11; // NEW — optional, for checkout abandonment
}

message DiagnosisResult {
  common.StatusResponse status = 1;
  string diagnosis_id = 2;
  FailureCategory category = 3;
  RecoverySuggestion suggestion = 4;
  string root_cause = 5;
  string ai_reasoning = 6;
  float confidence = 7;
  bool used_ai = 8;
  google.protobuf.Timestamp diagnosed_at = 9;
}

message FailureRecord {
  string id = 1;
  string payment_id = 2;
  string razorpay_error_code = 3;
  FailureCategory category = 4;
  RecoverySuggestion suggestion = 5;
  common.Money amount = 6;
  string payment_method = 7;
  string customer_id = 8;
  string merchant_id = 9;
  string root_cause = 10;
  string recovery_status = 11;
  google.protobuf.Timestamp failed_at = 12;
  google.protobuf.Timestamp diagnosed_at = 13;
  string subscription_id = 14;     // NEW
  string mandate_id = 15;          // NEW
  string checkout_session_id = 16; // NEW
}

message GetFailureRequest { string request_id = 1; string failure_id = 2; }
message ListFailuresRequest {
  string request_id = 1;
  string merchant_id = 2;
  FailureCategory category = 3;
  string recovery_status = 4;
  common.TimeRange time_range = 5;
  common.PaginationRequest pagination = 6;
}
message ListFailuresResponse {
  common.StatusResponse status = 1;
  repeated FailureRecord failures = 2;
  common.PaginationResponse pagination = 3;
}
message StatsRequest { string request_id = 1; string merchant_id = 2; common.TimeRange time_range = 3; }
message FailureStats {
  common.StatusResponse status = 1;
  int32 total_failures = 2;
  common.Money total_amount_at_risk = 3;
  map<string, int32> by_category = 4;
  map<string, int32> by_payment_method = 5;
  map<string, int32> by_recovery_status = 6;
}
message PatternRequest { string request_id = 1; string merchant_id = 2; common.TimeRange time_range = 3; }
message PatternReport {
  common.StatusResponse status = 1;
  repeated Pattern patterns = 2;
}
message Pattern {
  string pattern_type = 1;
  string description = 2;
  float significance = 3;
  map<string, string> metadata = 4;
}
```

### proto/recovery/recovery.proto

```protobuf
syntax = "proto3";
package recoveriq.recovery;

import "google/protobuf/timestamp.proto";
import "common/common.proto";

service RecoveryService {
  rpc CreateWorkflow(CreateWorkflowRequest) returns (RecoveryWorkflow);
  rpc ExecuteNextStep(ExecuteStepRequest) returns (StepResult);
  rpc GetWorkflow(GetWorkflowRequest) returns (RecoveryWorkflow);
  rpc ListWorkflows(ListWorkflowsRequest) returns (ListWorkflowsResponse);
  rpc GetRecoveryStats(RecoveryStatsRequest) returns (RecoveryStats);
  // Promise-to-pay
  rpc CreatePromise(CreatePromiseRequest) returns (Promise);
  rpc GetPromise(GetPromiseRequest) returns (Promise);
  rpc ListPromises(ListPromisesRequest) returns (ListPromisesResponse);
  rpc CheckDuePromises(CheckDueRequest) returns (DuePromisesResponse);
}

enum RecoveryAction {
  ACTION_UNKNOWN = 0;
  ACTION_RETRY_PAYMENT = 1;
  ACTION_CREATE_PAYMENT_LINK = 2;
  ACTION_SEND_REMINDER_EMAIL = 3;
  ACTION_WAIT_AND_RETRY = 4;
  ACTION_ESCALATE_TO_HUMAN = 5;
  ACTION_ABANDON = 6;
  ACTION_RETRY_SUBSCRIPTION = 7;   // NEW
  ACTION_SEND_CARD_UPDATE_LINK = 8; // NEW
  ACTION_RENEW_MANDATE = 9;        // NEW
  ACTION_VOICE_CALL = 10;          // NEW
  ACTION_CHECKOUT_NUDGE = 11;      // NEW
  ACTION_PROMISE_FOLLOW_UP = 12;   // NEW
}

enum WorkflowStatus {
  WF_CREATED = 0;
  WF_IN_PROGRESS = 1;
  WF_RECOVERED = 2;
  WF_FAILED = 3;
  WF_ABANDONED = 4;
  WF_ESCALATED = 5;
  WF_PROMISED = 6;                 // NEW — customer promised to pay
}

message CreateWorkflowRequest {
  string request_id = 1;
  string failure_id = 2;
  string payment_id = 3;
  string diagnosis_id = 4;
  string customer_id = 5;
  string merchant_id = 6;
  common.Money amount = 7;
  string payment_method = 8;
  string suggested_action = 9;
  string subscription_id = 10;
  string mandate_id = 11;
  string checkout_session_id = 12;
}

message RecoveryWorkflow {
  common.StatusResponse status = 1;
  string workflow_id = 2;
  string failure_id = 3;
  string payment_id = 4;
  common.Money amount = 5;
  WorkflowStatus workflow_status = 6;
  repeated RecoveryStep steps = 7;
  int32 retry_count = 8;
  int32 contact_count = 9;
  common.Money recovery_cost = 10;
  google.protobuf.Timestamp created_at = 11;
  google.protobuf.Timestamp updated_at = 12;
  string recovered_payment_id = 13;
  string promise_id = 14;          // NEW — linked promise if customer committed
}

message RecoveryStep {
  string step_id = 1;
  RecoveryAction action = 2;
  string status = 3;              // PENDING, EXECUTING, SUCCESS, FAILED, SKIPPED
  string result_detail = 4;
  string ai_reasoning = 5;
  map<string, string> metadata = 6;
  google.protobuf.Timestamp executed_at = 7;
  int32 duration_ms = 8;
  repeated string guardrails_checked = 9;
}

message ExecuteStepRequest { string request_id = 1; string workflow_id = 2; }
message StepResult {
  common.StatusResponse status = 1;
  RecoveryStep step = 2;
  WorkflowStatus new_workflow_status = 3;
  string next_action = 4;
}
message GetWorkflowRequest { string request_id = 1; string workflow_id = 2; }
message ListWorkflowsRequest {
  string request_id = 1;
  string merchant_id = 2;
  WorkflowStatus status = 3;
  common.TimeRange time_range = 4;
  common.PaginationRequest pagination = 5;
}
message ListWorkflowsResponse {
  common.StatusResponse status = 1;
  repeated RecoveryWorkflow workflows = 2;
  common.PaginationResponse pagination = 3;
}
message RecoveryStatsRequest { string request_id = 1; string merchant_id = 2; common.TimeRange time_range = 3; }
message RecoveryStats {
  common.StatusResponse status = 1;
  common.Money total_at_risk = 2;
  common.Money total_recovered = 3;
  common.Money total_cost = 4;
  float recovery_rate = 5;
  int32 total_workflows = 6;
  map<string, int32> by_status = 7;
  map<string, int32> by_action = 8;
  map<string, float> recovery_rate_by_method = 9;
  int32 voice_calls_made = 10;     // NEW
  int32 promises_created = 11;     // NEW
  int32 promises_kept = 12;        // NEW
}

// ── Promise-to-Pay Messages ──

enum PromiseStatus {
  PROMISE_ACTIVE = 0;
  PROMISE_KEPT = 1;                // Customer paid by promised date
  PROMISE_BROKEN = 2;             // Date passed, no payment
  PROMISE_EXTENDED = 3;           // Customer requested extension
  PROMISE_CANCELLED = 4;
}

message CreatePromiseRequest {
  string request_id = 1;
  string workflow_id = 2;
  string customer_id = 3;
  string payment_id = 4;
  common.Money amount = 5;
  google.protobuf.Timestamp promised_date = 6;
  string channel = 7;             // "email", "voice", "chat"
  string notes = 8;
}

message Promise {
  common.StatusResponse status = 1;
  string promise_id = 2;
  string workflow_id = 3;
  string customer_id = 4;
  string payment_id = 5;
  common.Money amount = 6;
  google.protobuf.Timestamp promised_date = 7;
  PromiseStatus promise_status = 8;
  string channel = 9;
  int32 follow_up_count = 10;
  google.protobuf.Timestamp created_at = 11;
  google.protobuf.Timestamp resolved_at = 12;
  string resolution_payment_id = 13;   // Payment that fulfilled the promise
}

message GetPromiseRequest { string request_id = 1; string promise_id = 2; }
message ListPromisesRequest {
  string request_id = 1;
  string merchant_id = 2;
  PromiseStatus status = 3;
  common.PaginationRequest pagination = 4;
}
message ListPromisesResponse {
  common.StatusResponse status = 1;
  repeated Promise promises = 2;
  common.PaginationResponse pagination = 3;
}
message CheckDueRequest { string request_id = 1; }
message DuePromisesResponse {
  common.StatusResponse status = 1;
  repeated Promise due_today = 2;
  repeated Promise overdue = 3;
}
```

### proto/reconciliation/reconciliation.proto

```protobuf
syntax = "proto3";
package recoveriq.reconciliation;

import "google/protobuf/timestamp.proto";
import "common/common.proto";

service ReconciliationService {
  rpc RunReconciliation(ReconciliationRequest) returns (ReconciliationReport);
  rpc GetReport(GetReportRequest) returns (ReconciliationReport);
  rpc ListExceptions(ListExceptionsRequest) returns (ExceptionList);
  rpc ResolveException(ResolveExceptionRequest) returns (common.StatusResponse);
}

enum MatchType { MATCH_UNKNOWN = 0; EXACT_MATCH = 1; FUZZY_MATCH = 2; AI_MATCH = 3; UNMATCHED = 4; }
enum ExceptionType { EXC_UNKNOWN = 0; AMOUNT_MISMATCH = 1; MISSING_SETTLEMENT = 2; MISSING_ORDER = 3; DUPLICATE_SETTLEMENT = 4; FEE_DISCREPANCY = 5; TIMING_MISMATCH = 6; }

message ReconciliationRequest { string request_id = 1; string merchant_id = 2; common.TimeRange time_range = 3; }
message ReconciliationReport {
  common.StatusResponse status = 1;
  string report_id = 2;
  int32 total_records = 3;
  int32 exact_matches = 4;
  int32 fuzzy_matches = 5;
  int32 ai_matches = 6;
  int32 unmatched = 7;
  float match_rate = 8;
  common.Money total_settled = 9;
  common.Money total_expected = 10;
  common.Money discrepancy = 11;
  repeated MatchRecord matches = 12;
  google.protobuf.Timestamp run_at = 13;
}
message MatchRecord {
  string id = 1; string order_id = 2; string payment_id = 3; string settlement_id = 4;
  common.Money order_amount = 5; common.Money payment_amount = 6; common.Money settled_amount = 7;
  MatchType match_type = 8; float confidence = 9; string notes = 10;
}
message GetReportRequest { string request_id = 1; string report_id = 2; }
message ListExceptionsRequest { string request_id = 1; string report_id = 2; ExceptionType exception_type = 3; common.PaginationRequest pagination = 4; }
message ExceptionList { common.StatusResponse status = 1; repeated ExceptionRecord exceptions = 2; common.PaginationResponse pagination = 3; }
message ExceptionRecord {
  string id = 1; ExceptionType type = 2; string description = 3;
  string order_id = 4; string payment_id = 5; string settlement_id = 6;
  common.Money expected_amount = 7; common.Money actual_amount = 8;
  string ai_suggestion = 9; string resolution_status = 10; string resolved_by = 11;
  google.protobuf.Timestamp resolved_at = 12;
}
message ResolveExceptionRequest { string request_id = 1; string exception_id = 2; string resolution = 3; string notes = 4; }
```

### proto/audit/audit.proto

```protobuf
syntax = "proto3";
package recoveriq.audit;

import "google/protobuf/timestamp.proto";
import "google/protobuf/struct.proto";
import "common/common.proto";

service AuditService {
  rpc LogAction(AuditEntry) returns (LogResponse);
  rpc GetAuditTrail(GetTrailRequest) returns (AuditTrail);
  rpc GetActionsByEntity(EntityTrailRequest) returns (AuditTrail);
}

message AuditEntry {
  string request_id = 1; string service = 2; string action = 3;
  string entity_type = 4; string entity_id = 5; string actor = 6;
  google.protobuf.Struct input = 7; google.protobuf.Struct output = 8;
  string reasoning = 9; repeated string guardrails_checked = 10;
  int32 duration_ms = 11; string parent_request_id = 12;
  string status = 13; string error_detail = 14;
}
message LogResponse { common.StatusResponse status = 1; string audit_id = 2; google.protobuf.Timestamp logged_at = 3; }
message GetTrailRequest { string request_id = 1; string service = 2; string action = 3; common.TimeRange time_range = 4; common.PaginationRequest pagination = 5; }
message EntityTrailRequest { string request_id = 1; string entity_type = 2; string entity_id = 3; }
message AuditTrail { common.StatusResponse status = 1; repeated AuditRecord entries = 2; common.PaginationResponse pagination = 3; }
message AuditRecord {
  string audit_id = 1; string service = 2; string action = 3;
  string entity_type = 4; string entity_id = 5; string actor = 6;
  google.protobuf.Struct input = 7; google.protobuf.Struct output = 8;
  string reasoning = 9; repeated string guardrails_checked = 10;
  int32 duration_ms = 11; string status = 12; string error_detail = 13;
  google.protobuf.Timestamp created_at = 14;
}
```

### proto/notification/notification.proto

```protobuf
syntax = "proto3";
package recoveriq.notification;
import "common/common.proto";

service NotificationService {
  rpc SendEmail(EmailRequest) returns (common.StatusResponse);
  rpc GetDeliveryStatus(DeliveryStatusRequest) returns (DeliveryStatus);
}

message EmailRequest {
  string request_id = 1; string to_email = 2; string subject = 3;
  string body_html = 4; string body_text = 5;
  string template_id = 6; map<string, string> template_vars = 7;
}
message DeliveryStatusRequest { string request_id = 1; string notification_id = 2; }
message DeliveryStatus { common.StatusResponse status = 1; string delivery_state = 2; string error_detail = 3; }
```

### proto/checkout/checkout.proto — NEW

```protobuf
syntax = "proto3";
package recoveriq.checkout;

import "google/protobuf/timestamp.proto";
import "common/common.proto";

service CheckoutTrackerService {
  rpc StartSession(StartSessionRequest) returns (CheckoutSession);
  rpc Heartbeat(HeartbeatRequest) returns (common.StatusResponse);
  rpc CompleteSession(CompleteRequest) returns (common.StatusResponse);
  rpc GetAbandonedSessions(AbandonedRequest) returns (AbandonedResponse);
}

enum SessionStatus {
  SESSION_ACTIVE = 0;
  SESSION_COMPLETED = 1;         // Payment successful
  SESSION_ABANDONED = 2;         // No heartbeat for > threshold
  SESSION_RECOVERING = 3;        // Recovery in progress
}

message StartSessionRequest {
  string request_id = 1;
  string customer_id = 2;
  string merchant_id = 3;
  common.Money cart_amount = 4;
  string order_id = 5;
  repeated string cart_items = 6;  // Product names/IDs
}

message CheckoutSession {
  common.StatusResponse status = 1;
  string session_id = 2;
  string customer_id = 3;
  string merchant_id = 4;
  common.Money cart_amount = 5;
  string order_id = 6;
  SessionStatus session_status = 7;
  google.protobuf.Timestamp started_at = 8;
  google.protobuf.Timestamp last_heartbeat = 9;
  int32 heartbeat_count = 10;
  int32 time_on_page_seconds = 11;
}

message HeartbeatRequest {
  string request_id = 1;
  string session_id = 2;
  string current_step = 3;        // "cart", "address", "payment_select", "payment_processing"
}

message CompleteRequest {
  string request_id = 1;
  string session_id = 2;
  string payment_id = 3;
}

message AbandonedRequest {
  string request_id = 1;
  string merchant_id = 2;
  common.TimeRange time_range = 3;
  common.PaginationRequest pagination = 4;
}

message AbandonedResponse {
  common.StatusResponse status = 1;
  repeated CheckoutSession sessions = 2;
  common.PaginationResponse pagination = 3;
  common.Money total_abandoned_value = 4;
}
```

### proto/voice/voice.proto — NEW

```protobuf
syntax = "proto3";
package recoveriq.voice;

import "google/protobuf/timestamp.proto";
import "common/common.proto";

service VoiceRecoveryService {
  rpc GetCallStatus(CallStatusRequest) returns (CallRecord);
  rpc ListCalls(ListCallsRequest) returns (ListCallsResponse);
}

enum CallStatus {
  CALL_QUEUED = 0;
  CALL_IN_PROGRESS = 1;
  CALL_COMPLETED = 2;
  CALL_FAILED = 3;
  CALL_NO_ANSWER = 4;
  CALL_PROMISE_RECEIVED = 5;    // Customer promised to pay
}

message VoiceTask {
  string task_id = 1;
  string workflow_id = 2;
  string customer_id = 3;
  string customer_phone = 4;
  string customer_name = 5;
  common.Money amount = 6;
  string language = 7;            // "hinglish", "hindi", "english"
  string template = 8;            // "payment_failed", "subscription_expired", "promise_reminder"
  map<string, string> template_vars = 9;
}

message CallRecord {
  common.StatusResponse status = 1;
  string call_id = 2;
  string task_id = 3;
  string workflow_id = 4;
  CallStatus call_status = 5;
  int32 duration_seconds = 6;
  string tts_text = 7;            // The Hinglish text spoken
  string tts_audio_url = 8;       // S3/local URL of generated audio
  string customer_response = 9;   // "will_pay", "not_interested", "no_answer"
  google.protobuf.Timestamp called_at = 10;
  google.protobuf.Timestamp promise_date = 11;   // If customer committed
}

message CallStatusRequest { string request_id = 1; string call_id = 2; }
message ListCallsRequest {
  string request_id = 1;
  string merchant_id = 2;
  CallStatus status = 3;
  common.PaginationRequest pagination = 4;
}
message ListCallsResponse {
  common.StatusResponse status = 1;
  repeated CallRecord calls = 2;
  common.PaginationResponse pagination = 3;
}
```

---

## STRANDS AGENTS SPECIFICATION

### Model Configuration (All Agents)

```python
from strands.models.openai import OpenAIModel

model = OpenAIModel(
    model_id="mistral.ministral-3-8b-instruct",
    client_args={
        "base_url": "https://bedrock-mantle.ap-south-1.api.aws/v1",
        "api_key": os.getenv("LLM_API_KEY")
    },
    model_config={"temperature": 0.1, "max_tokens": 4096}
)
```

### Agent 1: Failure Diagnosis Agent — `failure-detector`

```python
@tool
def lookup_error_code(error_code: str) -> str:
    """Look up a Razorpay error code and return its meaning and typical cause.
    Args: error_code: The Razorpay error code (e.g., 'BAD_REQUEST_ERROR')"""

@tool
def get_customer_payment_history(customer_id: str, limit: int = 10) -> str:
    """Get a customer's recent payment history to understand patterns.
    Args: customer_id: The customer ID. limit: Number of recent payments."""

@tool
def check_bank_health(bank_name: str) -> str:
    """Check if a bank is experiencing known issues or outages.
    Args: bank_name: Bank name (e.g., 'HDFC', 'SBI')"""

@tool
def get_failure_rate_by_method(payment_method: str, hours: int = 1) -> str:
    """Get failure rate for a specific payment method in the last N hours.
    Args: payment_method: 'upi', 'card', 'netbanking'. hours: Time window."""

@tool
def check_subscription_status(subscription_id: str) -> str:
    """Check the status of a Razorpay subscription — active, halted, cancelled.
    Args: subscription_id: The Razorpay subscription ID"""

@tool
def check_mandate_status(mandate_id: str) -> str:
    """Check e-Mandate/AutoPay status — confirmed, rejected, expired.
    Args: mandate_id: The mandate token ID"""

diagnosis_agent = Agent(
    name="failure_diagnoser", model=model,
    system_prompt="""You are a Razorpay payment failure diagnosis specialist.
Given a payment failure with error code and context, determine:
1. Category (INSUFFICIENT_FUNDS, CARD_EXPIRED, BANK_DECLINE, NETWORK_ERROR,
   AUTHENTICATION_FAILED, FRAUD_SUSPECTED, SUBSCRIPTION_FAILED, MANDATE_FAILED,
   CHECKOUT_ABANDONED, UNKNOWN)
2. Root cause explanation
3. Recommended recovery action

Rules:
- Use lookup_error_code FIRST
- Check customer history for repeat patterns
- For subscription failures, check subscription status
- For mandate failures, check mandate status
- If confidence < 70%, set category = UNKNOWN
- Never fabricate data""",
    tools=[lookup_error_code, get_customer_payment_history, check_bank_health,
           get_failure_rate_by_method, check_subscription_status, check_mandate_status]
)
```

### Agent 2: Recovery Strategy Agent — `recovery-orchestrator`

```python
@tool
def check_retry_eligibility(payment_id: str) -> str:
    """Check if payment is eligible for retry (count, circuit breaker, age)."""

@tool
def check_contact_window() -> str:
    """Check if current time is within allowed contact hours (9AM-9PM IST)."""

@tool
def calculate_recovery_cost(amount_paise: int, action: str) -> str:
    """Calculate cost of recovery action and check against cost cap.
    Args: amount_paise: Payment amount. action: 'retry','payment_link','email','voice','escalate'"""

@tool
def check_customer_dnd(customer_id: str) -> str:
    """Check if customer opted out of recovery communications."""

@tool
def check_customer_promise(customer_id: str, payment_id: str) -> str:
    """Check if customer already has an active promise-to-pay for this payment."""

@tool
def get_checkout_session_context(session_id: str) -> str:
    """Get checkout session details — cart items, time on page, last step.
    Args: session_id: The checkout session ID"""

strategy_agent = Agent(
    name="recovery_strategist", model=model,
    system_prompt="""You are a payment recovery strategist. Given a diagnosed failure,
decide the best recovery action.

Available actions:
- RETRY_PAYMENT: Retry same method (only if eligible)
- CREATE_PAYMENT_LINK: New payment link for alternate method
- SEND_REMINDER_EMAIL: Email with payment instructions
- WAIT_AND_RETRY: Schedule delayed retry
- VOICE_CALL: Hinglish voice call for high-value payments (> ₹500)
- CHECKOUT_NUDGE: Reminder for abandoned checkout
- RETRY_SUBSCRIPTION: Retry subscription charge
- SEND_CARD_UPDATE_LINK: Link to update expired card on subscription
- RENEW_MANDATE: Send mandate renewal link
- PROMISE_FOLLOW_UP: Follow up on a broken promise
- ESCALATE_TO_HUMAN: Flag for human review
- ABANDON: Stop recovery

Rules:
- ALWAYS check retry eligibility before retry
- ALWAYS check contact window before email/voice
- ALWAYS check cost cap before any action
- ALWAYS check DND before customer contact
- ALWAYS check for existing promise before new contact
- Voice calls ONLY for amounts > ₹500 and during business hours
- Checkout nudges within 30 min of abandonment are most effective
- If fraud suspected → ALWAYS escalate
- If amount < ₹50 → prefer ABANDON""",
    tools=[check_retry_eligibility, check_contact_window, calculate_recovery_cost,
           check_customer_dnd, check_customer_promise, get_checkout_session_context]
)
```

### Agent 3: Recovery Executors — `recovery-orchestrator`

```python
@tool
def retry_razorpay_payment(payment_id: str, idempotency_key: str) -> str:
    """Retry a failed payment with idempotency key."""

@tool
def create_razorpay_payment_link(amount_paise: int, customer_email: str,
                                  description: str, order_id: str) -> str:
    """Create Razorpay payment link for alternate payment."""

@tool
def retry_subscription_charge(subscription_id: str) -> str:
    """Retry a failed subscription charge via Razorpay Subscriptions API."""

@tool
def create_card_update_link(subscription_id: str, customer_email: str) -> str:
    """Create a link for customer to update their card on a subscription."""

@tool
def create_mandate_renewal_link(mandate_id: str, customer_email: str) -> str:
    """Create a link for customer to renew an expired e-Mandate."""

@tool
def send_recovery_email(customer_email: str, template: str,
                        payment_link_url: str, amount_display: str) -> str:
    """Send recovery email. Templates: payment_failed, retry_reminder,
    payment_link, checkout_abandoned, subscription_expired, mandate_expired."""

@tool
def initiate_voice_call(workflow_id: str, customer_id: str, customer_phone: str,
                         customer_name: str, amount_paise: int, template: str) -> str:
    """Initiate a Hinglish voice recovery call for high-value payments.
    Templates: payment_failed, subscription_expired, promise_reminder."""

@tool
def create_promise(workflow_id: str, customer_id: str, payment_id: str,
                   amount_paise: int, promised_date: str, channel: str) -> str:
    """Record a customer's promise to pay by a specific date.
    Args: promised_date: ISO date. channel: 'email','voice','chat'."""

@tool
def create_escalation_ticket(workflow_id: str, reason: str, severity: str) -> str:
    """Create escalation ticket for human review. severity: LOW/MEDIUM/HIGH/CRITICAL."""
```

### Agent 4: Promise Follow-up Agent — `recovery-orchestrator` — NEW

```python
promise_agent = Agent(
    name="promise_tracker", model=model,
    system_prompt="""You are a promise-to-pay follow-up agent.
You check on customer promises and decide next action.

If promise is DUE TODAY: send a gentle reminder email
If promise is OVERDUE (1 day): send a firm reminder email
If promise is OVERDUE (3+ days): escalate to human or try voice call
If promise is KEPT: celebrate and close workflow

Rules:
- Max 2 follow-ups per promise
- Don't contact before 9AM or after 9PM
- Don't contact if DND
- Be respectful — never threaten""",
    tools=[check_customer_dnd, check_contact_window, send_recovery_email,
           initiate_voice_call, create_escalation_ticket]
)
```

### Agent 5: Reconciliation Agent — `reconciliation-engine`

```python
@tool
def search_orders_by_amount(amount_paise: int, tolerance_paise: int = 0, date_range_days: int = 7) -> str:
    """Search orders by amount with tolerance for fuzzy matching."""

@tool
def search_payments_by_settlement(settlement_id: str) -> str:
    """Find all payments associated with a settlement."""

@tool
def check_fee_schedule(payment_method: str, amount_paise: int) -> str:
    """Calculate expected Razorpay fees to explain amount differences.
    UPI: 0%, Card: 2% + 18% GST, Netbanking: 2% + 18% GST."""

@tool
def flag_as_exception(record_id: str, exception_type: str, description: str) -> str:
    """Flag record as unresolvable exception for human review."""

reconciliation_agent = Agent(
    name="reconciler", model=model,
    system_prompt="""You are a financial reconciliation specialist.
Resolve unmatched transactions that automated matching couldn't handle.

Common mismatch reasons:
- Razorpay fees (2% cards, 0% UPI) + 18% GST on fees
- Partial refunds reducing settled amounts
- Split settlements across multiple days

Rules:
- NEVER fabricate a match
- Always explain reasoning
- Check fee schedules for close-but-not-exact amounts
- Flag truly unresolvable as exception""",
    tools=[search_orders_by_amount, search_payments_by_settlement,
           check_fee_schedule, flag_as_exception]
)
```

### Agent 6: Master Orchestrator — `ai-gateway`

```python
@tool
def analyze_payment_failure(payment_id: str) -> str:
    """Analyze why a specific payment failed."""

@tool
def start_recovery(failure_id: str) -> str:
    """Start a recovery workflow for a diagnosed failure."""

@tool
def get_recovery_dashboard() -> str:
    """Get current recovery metrics: at risk, recovered, rate."""

@tool
def run_reconciliation(merchant_id: str, days: int = 7) -> str:
    """Run settlement reconciliation for a merchant."""

@tool
def get_audit_trail(entity_id: str) -> str:
    """Get full audit trail for any entity."""

@tool
def get_abandoned_checkouts(merchant_id: str) -> str:
    """Get list of abandoned checkout sessions with cart values."""

@tool
def get_promise_status(merchant_id: str) -> str:
    """Get all active and overdue promise-to-pay commitments."""

@tool
def forecast_cash_position(merchant_id: str, days: int = 7) -> str:
    """Predict cash position for next N days based on settlement history + pending recoveries."""

master_agent = Agent(
    name="recoveriq_master", model=model,
    system_prompt="""You are RecoverIQ, an AI revenue recovery assistant.
You help merchants understand, recover, and reconcile payment revenue.

Capabilities:
- Analyze payment failures and root causes
- Start and monitor recovery workflows
- Track promise-to-pay commitments
- View abandoned checkout sessions
- Run settlement reconciliation
- Forecast cash position
- Query audit trails

Rules:
- Always be precise with ₹ amounts
- Cite data sources
- Never guess — if no data, say so
- Explain financial concepts simply""",
    tools=[analyze_payment_failure, start_recovery, get_recovery_dashboard,
           run_reconciliation, get_audit_trail, get_abandoned_checkouts,
           get_promise_status, forecast_cash_position]
)
```

---

## GUARDRAILS (Hard-coded, Never AI-overridable)

```python
class RecoveryGuardrails:
    MAX_RETRIES_PER_PAYMENT = 3
    RETRY_BACKOFF_SECONDS = [60, 3600, 86400]       # 1min, 1hr, 24hr
    MAX_CONTACTS_PER_CUSTOMER_PER_DAY = 2
    NO_CONTACT_BEFORE_HOUR = 9                       # 9 AM IST
    NO_CONTACT_AFTER_HOUR = 21                       # 9 PM IST
    MAX_RECOVERY_COST_PERCENT = 0.20                 # Stop if cost > 20% of amount
    MIN_RECOVERABLE_AMOUNT_PAISE = 5000              # Don't recover < ₹50
    MIN_VOICE_CALL_AMOUNT_PAISE = 50000              # Voice only for > ₹500
    CIRCUIT_BREAKER_FAILURE_THRESHOLD = 5
    CIRCUIT_BREAKER_WINDOW_SECONDS = 60
    CIRCUIT_BREAKER_COOLDOWN_SECONDS = 300
    IDEMPOTENCY_TTL_SECONDS = 86400
    MAX_WORKFLOW_AGE_DAYS = 7
    MAX_ACTIVE_WORKFLOWS_PER_MERCHANT = 100
    CHECKOUT_ABANDONMENT_THRESHOLD_SECONDS = 300     # 5 min no heartbeat = abandoned
    CHECKOUT_NUDGE_WINDOW_MINUTES = 30               # Nudge within 30 min
    MAX_PROMISE_FOLLOW_UPS = 2
    PROMISE_OVERDUE_GRACE_DAYS = 1                   # 1 day grace before escalation
```

---

## DATA LAYER

### MongoDB Collections (12)

| Collection | Purpose | New? |
|---|---|---|
| `webhook_events` | Raw Razorpay webhook events | |
| `failures` | Diagnosed payment failures | |
| `recovery_workflows` | Recovery workflow state + steps | |
| `reconciliation_reports` | Recon run results | |
| `orders` | Synthetic order data | |
| `payments` | Synthetic payment data | |
| `settlements` | Synthetic settlement data | |
| `customers` | Customer profiles + DND | |
| `audit_log` | All system actions | |
| `escalations` | Human review tickets | |
| `checkout_sessions` | Checkout tracking sessions | ✅ NEW |
| `promises` | Promise-to-pay records | ✅ NEW |

### Redis Namespaces (9)

| Namespace | Key Pattern | TTL |
|---|---|---|
| `idempotency` | `idempotency:{payment_id}:{action}` | 24h |
| `rate_limit` | `rate_limit:{entity}:{window}` | Dynamic |
| `circuit_breaker` | `cb:{service}:{merchant_id}` | 5 min |
| `retry_count` | `retry:{payment_id}` | 7 days |
| `contact_count` | `contact:{customer_id}:{date}` | 24h |
| `cache` | `cache:{collection}:{hash}` | 5 min |
| `session` | `session:{session_id}` | 1h |
| `checkout` | `checkout:{session_id}:heartbeat` | 10 min | ✅ NEW |
| `promise` | `promise:{customer_id}:{payment_id}` | 30 days | ✅ NEW |

### Redpanda Topics (8)

| Topic | Publisher | Consumer |
|---|---|---|
| `PAYMENT_EVENTS` | webhook-receiver | failure-detector |
| `CHECKOUT_EVENTS` | checkout-tracker | failure-detector | ✅ NEW |
| `FAILURE_DIAGNOSED` | failure-detector | recovery-orchestrator |
| `RECOVERY_EVENTS` | recovery-orchestrator | dashboard-api |
| `NOTIFICATIONS` | recovery-orchestrator | notification-worker |
| `VOICE_RECOVERY` | recovery-orchestrator | voice-recovery-worker | ✅ NEW |
| `AUDIT_EVENTS` | all services | audit-service |
| `DLQ` | any service | monitoring/alerting |

---

## NEW SERVICE DETAILS

### checkout-tracker (Go) — Checkout Drop-off Recovery

**How it works:**
1. Frontend calls `POST /checkout/start` when customer begins checkout → creates session in MongoDB via mongo-service, sets Redis heartbeat key
2. Frontend sends `POST /checkout/heartbeat` every 30 seconds with current step ("cart", "address", "payment_select", "payment_processing")
3. Frontend calls `POST /checkout/complete` when payment succeeds
4. **Background goroutine** runs every 60 seconds: scans Redis for sessions with no heartbeat for > 5 minutes → marks as `ABANDONED` → publishes `checkout.abandoned` event to `CHECKOUT_EVENTS` topic
5. failure-detector consumes the event, classifies as `CHECKOUT_ABANDONED`, recovery-orchestrator sends a nudge email with a payment link

**Failure handling:**
- If Redis heartbeat check fails → query MongoDB for sessions older than threshold with no completion
- If customer returns and completes → `CompleteSession` marks session as completed, cancels any pending recovery

### voice-recovery-worker (Go) — Hinglish Voice Recovery

**How it works:**
1. recovery-orchestrator publishes to `VOICE_RECOVERY` topic when strategy selects `ACTION_VOICE_CALL`
2. Worker consumes task, generates Hinglish TTS audio using AWS Polly (or mock)
3. Initiates outbound IVR call via Exotel API (or mock that logs)
4. If customer responds with "will pay" → creates a promise-to-pay via recovery-orchestrator gRPC
5. Logs call status + recording to audit-service

**Hinglish templates:**

```go
var HinglishTemplates = map[string]string{
    "payment_failed": "Namaste {{.Name}}, aapka {{.Amount}} rupaye ka payment fail ho gaya hai. " +
        "Kya aap abhi dubara try karna chahenge? Hum aapko ek nayi payment link bhej sakte hain. " +
        "Agar aap baad mein pay karna chahte hain, toh please ek date bataiye.",

    "subscription_expired": "Namaste {{.Name}}, aapke subscription ka payment {{.Amount}} rupaye " +
        "charge nahi ho paya. Aapka card expire ho gaya ho sakta hai. " +
        "Hum aapko ek link bhej rahe hain jisse aap apna card update kar sakte hain.",

    "promise_reminder": "Namaste {{.Name}}, aapne {{.PromiseDate}} tak {{.Amount}} rupaye " +
        "pay karne ka promise kiya tha. Kya aap aaj pay kar payenge? " +
        "Agar aapko koi dikkat ho rahi hai toh hum madad kar sakte hain.",
}
```

**Guardrails for voice:**
- Only for amounts > ₹500 (`MIN_VOICE_CALL_AMOUNT_PAISE`)
- Only during business hours (9 AM – 9 PM IST)
- Max 1 voice call per customer per day (included in `MAX_CONTACTS_PER_CUSTOMER_PER_DAY`)
- DND check before calling
- No threatening language — templates are pre-approved, not AI-generated

### Promise-to-Pay Tracker — `recovery-orchestrator`

**How it works:**
1. When a customer responds to recovery (email click "I'll pay later" or voice call "will_pay") → `create_promise` tool records promise with date
2. Background scheduler (in recovery-orchestrator) checks daily for due/overdue promises via `CheckDuePromises` gRPC
3. Due today → `promise_agent` sends gentle reminder email
4. Overdue 1 day → firm reminder
5. Overdue 3+ days → escalate or voice call
6. When payment comes in (webhook `payment.captured` for the order) → automatically matches to promise → marks `PROMISE_KEPT` → closes workflow

**Promise states:** `ACTIVE → KEPT | BROKEN | EXTENDED | CANCELLED`

---

## FAILURE HANDLING MATRIX (14 Scenarios)

| # | Scenario | Response | Audit |
|---|---|---|---|
| 1 | Razorpay API timeout | Backoff 1s→2s→4s, max 3 | `RAZORPAY_TIMEOUT` |
| 2 | Razorpay rate limit (429) | Respect `Retry-After`, queue | `RATE_LIMITED` |
| 3 | Double-charge risk | Redis idempotency key check | `IDEMPOTENCY_HIT` |
| 4 | LLM hallucination | Validate against enums, fallback rules | `LLM_FALLBACK` |
| 5 | LLM unavailable | Pure rule-based classification | `LLM_UNAVAILABLE` |
| 6 | Retry storm | Circuit breaker: 5/min → 5min pause | `CIRCUIT_BREAKER_OPEN` |
| 7 | Customer harassment | Hard cap: 2 contacts/day | `CONTACT_LIMIT` |
| 8 | Recovery cost > value | 20% cap → abandon | `COST_CAP_HIT` |
| 9 | MongoDB write fail | Retry 2x, then DLQ | `MONGO_WRITE_FAIL` |
| 10 | Redpanda publish fail | Redis fallback, background retry | `REDPANDA_FAIL` |
| 11 | Invalid webhook signature | 401 + log suspicious event | `INVALID_SIGNATURE` |
| 12 | gRPC service unavailable | Retry + graceful degradation | `SERVICE_UNAVAILABLE` |
| 13 | Voice call fails/no answer | Log failure, fall back to email | `VOICE_CALL_FAILED` |
| 14 | Checkout session race condition | Redis atomic ops, idempotent completion | `CHECKOUT_RACE` |

---

## FRONTEND — MINIMALIST LIGHT THEME

### Design System

```css
:root {
  --bg-primary: #ffffff;
  --bg-secondary: #f8f9fa;
  --bg-tertiary: #f1f3f5;
  --border-light: #e9ecef;
  --border-medium: #dee2e6;
  --text-primary: #212529;
  --text-secondary: #495057;
  --text-tertiary: #868e96;
  --accent-primary: #4263eb;
  --accent-success: #2b8a3e;
  --accent-warning: #e67700;
  --accent-danger: #c92a2a;
  --accent-info: #1971c2;
  --card-shadow: 0 1px 3px rgba(0,0,0,0.04), 0 1px 2px rgba(0,0,0,0.06);
  --card-radius: 8px;
  --font-sans: 'Inter', -apple-system, sans-serif;
  --font-mono: 'JetBrains Mono', monospace;
}
```

### Pages (9)

| Page | Route | Key Elements |
|---|---|---|
| Overview | `/` | 4 metric cards, trend line chart, failure donut, activity feed |
| Failures | `/failures` | Filter bar, clean data table, status pills |
| Failure Detail | `/failures/[id]` | Diagnosis card, vertical timeline, audit table |
| Recoveries | `/recoveries` | Tabs (Active/Completed/Escalated), workflow cards |
| Promises | `/promises` | Due today, overdue, kept — promise cards with dates |
| Reconciliation | `/reconciliation` | Run button, match rate bar, exceptions table |
| Audit Trail | `/audit` | Search, filters, expandable entries |
| Chat | `/chat` | Message bubbles, SSE streaming, tool indicators |
| Settings | `/settings` | API keys (masked), guardrails table, service health |

### Visual Rules
- White cards on `#f8f9fa` background
- Thin `1px #e9ecef` borders, subtle shadows
- Muted status badges with soft tinted backgrounds
- Clean alternating-row tables
- Minimal sidebar — text links with optional small Lucide icons
- Only hover transitions (no load animations)
- Inter font, 16px body, clear hierarchy

---

## SYNTHETIC DATA (100+ records)

| Data Type | Count | Notes |
|---|---|---|
| Customers | 15 | 10 normal, 3 DND, 2 high-value repeat |
| Orders | 100 | Various amounts ₹100 – ₹10,000 |
| Payments (failed) | 60 | Distribution: insufficient_funds 25%, bank_decline 20%, card_expired 15%, auth_failed 15%, network_error 10%, fraud 5%, subscription 5%, mandate 3%, unknown 2% |
| Payments (successful) | 40 | Normal completions |
| Settlements | 30 | 20 exact match, 5 fee-adjusted, 3 timing mismatch, 2 missing |
| Checkout sessions | 10 | 7 abandoned, 3 completed |
| Promises | 5 | 2 active, 1 kept, 1 broken, 1 extended |

---

## ENVIRONMENT VARIABLES

```env
# Infrastructure
MONGO_URI=mongodb://mongodb:27017
MONGO_DB_NAME=recoveriq
REDIS_ADDR=redis:6379
REDPANDA_BROKER=redpanda:29092

# Razorpay (Test Mode)
RAZORPAY_KEY_ID=rzp_test_xxxxx
RAZORPAY_KEY_SECRET=xxxxx
RAZORPAY_WEBHOOK_SECRET=xxxxx

# LLM
LLM_BASE_URL=https://bedrock-mantle.ap-south-1.api.aws/v1
LLM_MODEL_ID=mistral.ministral-3-8b-instruct
LLM_API_KEY=your_key

# Notification
AWS_SES_MOCK=true
AWS_REGION=ap-south-1
SENDER_EMAIL=noreply@recoveriq.dev

# Voice (NEW)
EXOTEL_MOCK=true
EXOTEL_SID=
EXOTEL_TOKEN=
EXOTEL_CALLER_ID=
AWS_POLLY_MOCK=true

# App
MODE=dev
LOG_LEVEL=INFO
```

---

# COMPLETE TODO TASK LIST

## Phase 0: Project Scaffolding [9 tasks]
- [ ] Create project root `recoveriq/`
- [ ] Initialize git repository
- [ ] Create `go.work` linking all Go service modules
- [ ] Create `.env.example` with all environment variables
- [ ] Create `.gitignore` (Go, Python, Node, Docker, env)
- [ ] Create `README.md` with project overview
- [ ] Create `docker-compose.yml` with all 16 containers
- [ ] Create Docker networks (frontend, backend) and volumes
- [ ] Create `AUDIT_TRAIL.md` skeleton

## Phase 1: Protobuf Contracts [14 tasks]
- [ ] Create `proto/common/common.proto`
- [ ] Create `proto/mongo/mongo_service.proto`
- [ ] Create `proto/redis/redis_service.proto`
- [ ] Create `proto/webhook/webhook.proto`
- [ ] Create `proto/failure/failure.proto` (with subscription, mandate, checkout categories)
- [ ] Create `proto/recovery/recovery.proto` (with promise-to-pay, voice actions)
- [ ] Create `proto/reconciliation/reconciliation.proto`
- [ ] Create `proto/audit/audit.proto`
- [ ] Create `proto/notification/notification.proto`
- [ ] Create `proto/checkout/checkout.proto` — NEW
- [ ] Create `proto/voice/voice.proto` — NEW
- [ ] Create `proto/compile_go.sh`
- [ ] Create `proto/compile_python.sh`
- [ ] Compile all protos and verify stubs

## Phase 2: Shared Packages [15 tasks]
- [ ] Create `pkg/go.mod`
- [ ] Create `pkg/constants/topics.go` — 8 topics including CHECKOUT_EVENTS, VOICE_RECOVERY
- [ ] Create `pkg/constants/namespaces.go` — 9 namespaces including checkout, promise
- [ ] Create `pkg/constants/error_codes.go` — 40+ codes including subscription + mandate
- [ ] Create `pkg/grpcutil/interceptors.go`
- [ ] Create `pkg/grpcutil/health.go`
- [ ] Create `pkg/grpcutil/errors.go`
- [ ] Create `pkg/config/env.go`
- [ ] Create `shared/__init__.py`
- [ ] Create `shared/config.py`
- [ ] Create `shared/constants.py` (mirrors Go constants)
- [ ] Create `shared/guardrails.py` (full guardrails with voice + promise + checkout limits)
- [ ] Create `shared/grpc_utils.py`
- [ ] Create `shared/razorpay_client.py`
- [ ] Create `shared/logging_config.py`

## Phase 3: Go Infrastructure Services [70 tasks]

### mongo-service [12 tasks]
- [ ] Create go.mod, Dockerfile, cmd/server/main.go, config
- [ ] Create db/connection.go — MongoDB connection pool
- [ ] Create db/serialization.go — protobuf.Struct ↔ bson.M
- [ ] Create grpc_server/server.go
- [ ] Implement InsertOne + BulkInsert
- [ ] Implement FindOne + FindMany (filter, projection, sort, pagination)
- [ ] Implement UpdateOne (with upsert)
- [ ] Implement DeleteOne
- [ ] Implement Aggregate
- [ ] Add health check
- [ ] Write unit tests for serialization
- [ ] Write integration test

### redis-service [11 tasks]
- [ ] Create scaffolding
- [ ] Create client/redis.go — go-redis v9
- [ ] Create namespace/validator.go — validate all 9 namespaces
- [ ] Implement Set / Get / Delete / Exists
- [ ] Implement SetWithTTL
- [ ] Implement IncrBy / Expire
- [ ] Implement CheckRateLimit — sliding window (sorted set)
- [ ] Implement CheckAndSetIdempotency — atomic SETNX + TTL
- [ ] Add health check
- [ ] Write unit tests for rate limiting
- [ ] Write unit tests for idempotency

### audit-service [8 tasks]
- [ ] Create scaffolding
- [ ] Create gRPC clients to mongo-service + redis-service
- [ ] Implement LogAction — store via mongo-service, cache in redis-service
- [ ] Implement GetAuditTrail — query with filters
- [ ] Implement GetActionsByEntity — query by entity
- [ ] Implement background Redpanda consumer for AUDIT_EVENTS (batch insert)
- [ ] Add Redis caching for recent entries
- [ ] Write integration test

### webhook-receiver [9 tasks]
- [ ] Create scaffolding (Gin + gRPC dual server)
- [ ] Implement POST /webhook/razorpay handler
- [ ] Implement signature.go — HMAC SHA256 validation
- [ ] Implement publisher/redpanda.go → PAYMENT_EVENTS
- [ ] Implement gRPC server (GetWebhookEvent, ListWebhookEvents)
- [ ] Create gRPC clients to mongo-service + audit-service
- [ ] Handle: invalid signature → 401 + audit
- [ ] Handle: Redpanda down → Redis fallback
- [ ] Handle: MongoDB down → retry + DLQ

### notification-worker [7 tasks]
- [ ] Create scaffolding
- [ ] Implement Redpanda consumer for NOTIFICATIONS
- [ ] Create email/ses.go (AWS SES v2)
- [ ] Create email/mock.go (logs when AWS_SES_MOCK=true)
- [ ] Create email/templates.go — 6 templates: payment_failed, retry_reminder, payment_link, checkout_abandoned, subscription_expired, promise_reminder
- [ ] Create gRPC client to audit-service
- [ ] Create HTTP /healthz on :8083

### checkout-tracker [10 tasks] — NEW
- [ ] Create scaffolding (Gin + gRPC dual server)
- [ ] Implement POST /checkout/start — create session in MongoDB via mongo-service
- [ ] Implement POST /checkout/heartbeat — update Redis heartbeat key
- [ ] Implement POST /checkout/complete — mark session complete, cancel recovery
- [ ] Implement gRPC server (StartSession, Heartbeat, Complete, GetAbandoned)
- [ ] Implement detector/abandonment.go — background goroutine scanning for stale sessions
- [ ] Implement publisher/redpanda.go → CHECKOUT_EVENTS
- [ ] Create gRPC clients to mongo-service, redis-service, audit-service
- [ ] Handle: race condition (complete arrives after abandon detected) → idempotent
- [ ] Write unit tests for abandonment detection timing

### voice-recovery-worker [9 tasks] — NEW
- [ ] Create scaffolding
- [ ] Implement Redpanda consumer for VOICE_RECOVERY
- [ ] Create tts/polly.go — AWS Polly TTS client
- [ ] Create tts/mock.go — mock TTS (logs text)
- [ ] Create ivr/exotel.go — Exotel IVR API client
- [ ] Create ivr/mock.go — mock IVR (logs call details + simulates response)
- [ ] Create templates/hinglish.go — 3 Hinglish templates with Go templating
- [ ] Create gRPC client to audit-service
- [ ] Create HTTP /healthz on :8009

### dashboard-api [12 tasks]
- [ ] Create scaffolding (Gin + gRPC dual server)
- [ ] Create middleware: CORS, request-ID, error handler
- [ ] Create gRPC clients to ALL services (including checkout-tracker)
- [ ] Implement GET /api/dashboard/overview (aggregate from all services)
- [ ] Implement GET /api/dashboard/failures + /failures/:id
- [ ] Implement GET /api/dashboard/recoveries
- [ ] Implement GET /api/dashboard/reconciliation + /exceptions
- [ ] Implement GET /api/dashboard/audit
- [ ] Implement GET /api/dashboard/analytics
- [ ] Implement GET /api/dashboard/promises — NEW
- [ ] Implement GET /api/dashboard/checkouts/abandoned — NEW
- [ ] Write integration tests

## Phase 4: Python AI Services [55 tasks]

### failure-detector [17 tasks]
- [ ] Create scaffolding
- [ ] Create gRPC clients to mongo-service, redis-service, audit-service
- [ ] Implement Redpanda consumer for PAYMENT_EVENTS
- [ ] Implement Redpanda consumer for CHECKOUT_EVENTS — NEW
- [ ] Create classifier.py — 40+ error codes including subscription + mandate + checkout
- [ ] Create @tool lookup_error_code
- [ ] Create @tool get_customer_payment_history
- [ ] Create @tool check_bank_health
- [ ] Create @tool get_failure_rate_by_method
- [ ] Create @tool check_subscription_status — NEW
- [ ] Create @tool check_mandate_status — NEW
- [ ] Create Strands diagnosis_agent with updated system prompt
- [ ] Create pattern_detector.py
- [ ] Implement gRPC server (all 5 RPCs)
- [ ] Publish to FAILURE_DIAGNOSED topic
- [ ] Handle: LLM unavailable → rules; LLM hallucination → validate
- [ ] Write unit tests for classifier (100% code coverage)

### recovery-orchestrator [20 tasks]
- [ ] Create scaffolding
- [ ] Create gRPC clients to all Go services
- [ ] Implement Redpanda consumer for FAILURE_DIAGNOSED
- [ ] Import guardrails from shared
- [ ] Implement circuit_breaker.py (Redis-backed)
- [ ] Create @tool check_retry_eligibility
- [ ] Create @tool check_contact_window
- [ ] Create @tool calculate_recovery_cost (updated with voice cost)
- [ ] Create @tool check_customer_dnd
- [ ] Create @tool check_customer_promise — NEW
- [ ] Create @tool get_checkout_session_context — NEW
- [ ] Create Strands strategy_agent with updated prompt
- [ ] Create @tool retry_razorpay_payment
- [ ] Create @tool create_razorpay_payment_link
- [ ] Create @tool retry_subscription_charge — NEW
- [ ] Create @tool create_card_update_link — NEW
- [ ] Create @tool create_mandate_renewal_link — NEW
- [ ] Create @tool send_recovery_email (6 templates)
- [ ] Create @tool initiate_voice_call — NEW (publish to VOICE_RECOVERY)
- [ ] Create @tool create_promise — NEW (store in MongoDB via mongo-service)
- [ ] Create @tool create_escalation_ticket
- [ ] Create Strands promise_agent — NEW
- [ ] Implement GraphBuilder recovery workflow (updated with voice + promise + checkout nodes)
- [ ] Implement state_machine.py (updated with WF_PROMISED status)
- [ ] Implement promise_tracker.py — daily scheduler for due/overdue promises — NEW
- [ ] Implement gRPC server (all RPCs including promise RPCs)
- [ ] Handle all failure scenarios
- [ ] Write unit tests for guardrails, circuit breaker, state machine, promise tracker

### reconciliation-engine [10 tasks]
- [ ] Create scaffolding
- [ ] Create gRPC clients
- [ ] Implement exact_matcher.py
- [ ] Implement fuzzy_matcher.py
- [ ] Create @tool search_orders_by_amount
- [ ] Create @tool search_payments_by_settlement
- [ ] Create @tool check_fee_schedule
- [ ] Create @tool flag_as_exception
- [ ] Create Strands reconciliation_agent
- [ ] Implement gRPC server (all 4 RPCs)
- [ ] Implement reporter.py

### ai-gateway [8 tasks]
- [ ] Create scaffolding (FastAPI + gRPC)
- [ ] Create gRPC clients to ALL services
- [ ] Create @tool wrappers (8 tools including abandoned checkouts, promises, forecast)
- [ ] Create Strands master_agent with updated prompt
- [ ] Implement POST /api/chat (sync)
- [ ] Implement POST /api/chat/stream (SSE)
- [ ] Implement session management
- [ ] Handle: LLM timeout → graceful error

## Phase 5: Synthetic Data [7 tasks]
- [ ] Create synthetic_generator.py (15 customers, 100 orders, 100 payments, 30 settlements)
- [ ] Add 10 checkout sessions (7 abandoned, 3 completed) — NEW
- [ ] Add 5 promise records — NEW
- [ ] Add subscription + mandate failure records — NEW
- [ ] Create init_db.py — seed via mongo-service gRPC + create indexes
- [ ] Create razorpay_simulator.py — send webhook POSTs
- [ ] Verify seeded data

## Phase 6: Frontend [32 tasks]
- [ ] Initialize Next.js 15
- [ ] Install: recharts, lucide-react
- [ ] Create globals.css — light theme design system
- [ ] Create layout.tsx — Inter font, minimal sidebar
- [ ] Create Sidebar.tsx
- [ ] Create MetricsCard.tsx (white, subtle shadow)
- [ ] Create DataTable.tsx (alternating rows)
- [ ] Create StatusBadge.tsx (soft tinted backgrounds)
- [ ] Create Timeline.tsx (thin vertical line + dots)
- [ ] Create ChatInterface.tsx (clean bubbles + SSE)
- [ ] Create PromiseCard.tsx — NEW
- [ ] Create Charts/TrendChart.tsx
- [ ] Create Charts/DonutChart.tsx
- [ ] Create lib/api.ts
- [ ] Create lib/types.ts
- [ ] Build Overview page — 4 cards + charts + feed
- [ ] Build Failures list page — filters + table
- [ ] Build Failure detail page — diagnosis + timeline + audit
- [ ] Build Recoveries page — tabs + workflow cards
- [ ] Build Promises page — due/overdue/kept sections — NEW
- [ ] Build Reconciliation page — run button + match rate + exceptions
- [ ] Build Audit Trail page — search + expandable entries
- [ ] Build Chat page — messages + streaming + tool indicators
- [ ] Build Settings page — config + health
- [ ] Add responsive design
- [ ] Add loading states (skeleton loaders)
- [ ] Add error states
- [ ] Add subtle hover transitions
- [ ] Create frontend/Dockerfile
- [ ] Verify all pages render

## Phase 7: Integration Testing [16 tasks]
- [ ] docker compose up — verify all 16 containers start
- [ ] Verify gRPC connectivity (Go ↔ Go, Python ↔ Go)
- [ ] Seed database via init_db.py
- [ ] Send 60 payment failure webhooks
- [ ] Send 5 subscription failure webhooks — NEW
- [ ] Send 3 mandate failure webhooks — NEW
- [ ] Verify failure-detector classifies all failures
- [ ] Verify checkout-tracker detects abandoned sessions — NEW
- [ ] Verify recovery-orchestrator creates workflows
- [ ] Verify voice-recovery-worker processes voice tasks — NEW
- [ ] Verify promise-to-pay creates and tracks promises — NEW
- [ ] Verify notification-worker sends emails
- [ ] Verify audit-service has complete trail
- [ ] Run reconciliation — verify match rate > 90%
- [ ] Test chat: "Why did payments fail?"
- [ ] Verify frontend loads with data

## Phase 8: Failure Testing [10 tasks]
- [ ] Kill failure-detector → verify webhook-receiver still works
- [ ] Kill mongo-service → verify graceful degradation
- [ ] Invalid LLM_BASE_URL → verify rule-based fallbacks
- [ ] Invalid webhook signature → verify 401 + audit
- [ ] Trigger contact limit → verify blocked
- [ ] Trigger cost cap → verify ABANDON
- [ ] Trigger circuit breaker → verify pause
- [ ] Send duplicate webhook → verify idempotency
- [ ] Voice call fails → verify email fallback — NEW
- [ ] Checkout race condition (complete after abandon) → verify correct state — NEW

## Phase 9: Polish & Demo [10 tasks]
- [ ] Write comprehensive README.md
- [ ] Complete AUDIT_TRAIL.md with all 14 failure scenarios
- [ ] Create demo script (9 steps — show checkout recovery, voice call, promise tracker too)
- [ ] Code cleanup + formatting
- [ ] Security review (no hardcoded secrets, test mode enforced)
- [ ] Performance review (indexes, TTLs, no N+1)
- [ ] Run all unit tests: `go test ./...` + `pytest`
- [ ] Record demo video (optional)
- [ ] Create .env.example with comments
- [ ] Final full-flow test

---

## TASK SUMMARY

| Phase | Tasks |
|---|---|
| Phase 0: Scaffolding | 9 |
| Phase 1: Protobufs | 14 |
| Phase 2: Shared Packages | 15 |
| Phase 3: Go Services (8) | 78 |
| Phase 4: Python AI Services (4) | 55 |
| Phase 5: Synthetic Data | 7 |
| Phase 6: Frontend | 32 |
| Phase 7: Integration Testing | 16 |
| Phase 8: Failure Testing | 10 |
| Phase 9: Polish | 10 |
| **TOTAL** | **246** |

---

## SERVICE COMMUNICATION SUMMARY

**All internal communication uses gRPC only.** No direct database access from any service except mongo-service (MongoDB) and redis-service (Redis).

```
Every service ──gRPC──▶ audit-service ──gRPC──▶ mongo-service ──▶ MongoDB
Every service ──gRPC──▶ mongo-service ──▶ MongoDB
Every service ──gRPC──▶ redis-service ──▶ Redis
Services publish ──▶ Redpanda ──▶ Consumer services
External ──HTTP──▶ webhook-receiver, dashboard-api, ai-gateway, checkout-tracker
```

**No REST between internal services. No direct DB connections except the two wrapper services. All audit logging via gRPC.**
