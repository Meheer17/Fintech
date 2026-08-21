# RecoverIQ — Master Build Prompt & Task List

> **What**: AI-powered Revenue Recovery & Finance Control Agent for Razorpay Merchants
> **Architecture**: Python microservices + gRPC + Strands Agents + Next.js + MongoDB + Redis + Redpanda
> **Tracks**: Razorpay Hackathon Track 03 (Revenue Recovery) + Track 04 (Finance Controller)

---

## THE PROMPT

You are building **RecoverIQ**, a production-grade microservice platform that detects failed payments, diagnoses root causes using AI agents, executes bounded recovery workflows, reconciles settlements, and reports everything through an auditable dashboard.

### Core Principles

1. **Every money action is explainable** — Full audit trail with reasoning for every decision
2. **Every action is bounded** — Hard-coded guardrails: retry limits, contact windows, cost caps
3. **Every action is gated** — Human approval for high-value or uncertain actions
4. **Failure is a feature** — Circuit breakers, dead-letter queues, idempotency, graceful degradation
5. **AI where it matters** — LLM for diagnosis and complex reasoning; rules for everything else

### Technology Stack

| Layer | Technology | Version |
|---|---|---|
| **Language** | Python | 3.12+ |
| **API Framework** | FastAPI | latest |
| **Agent Framework** | strands-agents, strands-agents-tools | latest |
| **LLM Provider** | OpenAIModel (Bedrock Mantle endpoint) | mistral.ministral-3-8b-instruct |
| **Inter-service Comm** | gRPC (grpcio, grpcio-tools) | latest |
| **Database** | MongoDB | 7.0 |
| **Cache** | Redis | 8.0-alpine |
| **Event Broker** | Redpanda (Kafka-compatible) | v24.1.1 |
| **Payment API** | Razorpay Python SDK (test mode) | latest |
| **Frontend** | Next.js 15 + React 19 | latest |
| **Charts** | Recharts | latest |
| **Containerization** | Docker + Docker Compose | latest |
| **Package Manager** | uv (Python), npm (Node) | latest |

---

## ARCHITECTURE

```
                    ┌──────────────────────────────────────┐
                    │         EXTERNAL BOUNDARY            │
                    │                                      │
                    │  Razorpay ──HTTP──▶ webhook-receiver │
                    │  Browser  ──HTTP──▶ dashboard-api    │
                    │  Browser  ──SSE───▶ ai-gateway       │
                    └──────────────┬───────────────────────┘
                                  │
                    ┌─────────────▼───────────────────────────────────────┐
                    │              gRPC SERVICE MESH (backend network)    │
                    │                                                     │
                    │  webhook-receiver :50001                            │
                    │       │ publishes ──▶ Redpanda [PAYMENT_EVENTS]    │
                    │       │ stores ────▶ mongo-service (gRPC :50010)   │
                    │       │ logs ──────▶ audit-service (gRPC :50007)   │
                    │                                                     │
                    │  failure-detector :50002                            │
                    │       │ consumes ◀── Redpanda [PAYMENT_EVENTS]     │
                    │       │ reads ────▶ mongo-service (gRPC :50010)    │
                    │       │ caches ───▶ redis-service (gRPC :50011)    │
                    │       │ publishes ▶ Redpanda [FAILURE_DIAGNOSED]   │
                    │       │ logs ─────▶ audit-service (gRPC :50007)    │
                    │       │ AI ───────▶ Strands diagnosis_agent        │
                    │                                                     │
                    │  recovery-orchestrator :50003                       │
                    │       │ consumes ◀── Redpanda [FAILURE_DIAGNOSED]  │
                    │       │ reads/writes ▶ mongo-service (gRPC)        │
                    │       │ idempotency ─▶ redis-service (gRPC)        │
                    │       │ payments ────▶ Razorpay API (HTTPS)        │
                    │       │ publishes ───▶ Redpanda [NOTIFICATIONS]    │
                    │       │ publishes ───▶ Redpanda [RECOVERY_EVENTS]  │
                    │       │ logs ────────▶ audit-service (gRPC)        │
                    │       │ AI ──────────▶ Strands GraphBuilder flow   │
                    │                                                     │
                    │  reconciliation-engine :50004                       │
                    │       │ reads ──▶ mongo-service (gRPC)             │
                    │       │ logs ───▶ audit-service (gRPC)             │
                    │       │ AI ────▶ Strands reconciliation_agent      │
                    │                                                     │
                    │  dashboard-api :50005 / :8005                       │
                    │       │ aggregates ◀── all services via gRPC       │
                    │       │ serves ────▶ REST JSON to frontend         │
                    │                                                     │
                    │  ai-gateway :50006 / :8006                          │
                    │       │ delegates ◀──▶ all services via gRPC       │
                    │       │ serves ────▶ SSE streaming to frontend     │
                    │       │ AI ────────▶ Strands master_agent          │
                    │                                                     │
                    │  audit-service :50007                               │
                    │       │ stores ──▶ mongo-service (gRPC)            │
                    │       │ caches ──▶ redis-service (gRPC)            │
                    │                                                     │
                    │  notification-worker (no port, consumer only)       │
                    │       │ consumes ◀── Redpanda [NOTIFICATIONS]      │
                    │       │ sends ────▶ AWS SES (email)                │
                    │       │ logs ─────▶ audit-service (gRPC)           │
                    │                                                     │
                    │  mongo-service :50010                               │
                    │       │ connects ──▶ MongoDB :27017                │
                    │                                                     │
                    │  redis-service :50011                               │
                    │       │ connects ──▶ Redis :6379                   │
                    │                                                     │
                    └─────────────────────────────────────────────────────┘
                    
                    ┌─────────────────────────────────────────────────────┐
                    │              INFRASTRUCTURE                         │
                    │  MongoDB :27017  │  Redis :6379  │  Redpanda :9092 │
                    └─────────────────────────────────────────────────────┘
```

### Service Registry

| # | Service | gRPC Port | HTTP Port | Language | Has AI? | Role |
|---|---|---|---|---|---|---|
| 1 | `mongo-service` | 50010 | — | Python | ❌ | MongoDB access layer (CRUD, queries) |
| 2 | `redis-service` | 50011 | — | Python | ❌ | Redis access layer (cache, idempotency, rate limit) |
| 3 | `webhook-receiver` | 50001 | 8001 | Python | ❌ | Razorpay webhook ingestion |
| 4 | `failure-detector` | 50002 | — | Python | ✅ | Payment failure classification |
| 5 | `recovery-orchestrator` | 50003 | — | Python | ✅ | Recovery strategy + execution |
| 6 | `reconciliation-engine` | 50004 | — | Python | ✅ | Settlement matching |
| 7 | `dashboard-api` | 50005 | 8005 | Python | ❌ | REST API for frontend |
| 8 | `ai-gateway` | 50006 | 8006 | Python | ✅ | Master agent + chat |
| 9 | `audit-service` | 50007 | — | Python | ❌ | Audit trail logging |
| 10 | `notification-worker` | — | — | Python | ❌ | Email dispatch (Redpanda consumer) |
| 11 | `frontend` | — | 3000 | TypeScript | ❌ | Next.js dashboard |

---

## PROTOBUF CONTRACTS

### proto/common/common.proto

```protobuf
syntax = "proto3";
package recoveriq.common;

import "google/protobuf/timestamp.proto";
import "google/protobuf/struct.proto";

message StatusResponse {
  bool success = 1;
  string error_code = 2;
  string error_message = 3;
  string request_id = 4;
}

message Money {
  int64 amount_paise = 1;    // Always in paise (₹100 = 10000 paise)
  string currency = 2;       // "INR"
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
  // Generic CRUD
  rpc InsertOne(InsertRequest) returns (InsertResponse);
  rpc FindOne(FindRequest) returns (FindResponse);
  rpc FindMany(FindManyRequest) returns (FindManyResponse);
  rpc UpdateOne(UpdateRequest) returns (UpdateResponse);
  rpc DeleteOne(DeleteRequest) returns (common.StatusResponse);
  
  // Aggregation
  rpc Aggregate(AggregateRequest) returns (AggregateResponse);
  
  // Bulk operations
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
  
  // Sliding window rate limiting
  rpc CheckRateLimit(RateLimitRequest) returns (RateLimitResponse);
  
  // Idempotency
  rpc CheckAndSetIdempotency(IdempotencyRequest) returns (IdempotencyResponse);
}

message SetRequest {
  string request_id = 1;
  string namespace = 2;
  string key = 3;
  string value = 4;
}

message GetRequest {
  string request_id = 1;
  string namespace = 2;
  string key = 3;
}

message GetResponse {
  common.StatusResponse status = 1;
  string value = 2;
  bool found = 3;
}

message DeleteRequest {
  string request_id = 1;
  string namespace = 2;
  string key = 3;
}

message ExistsRequest {
  string request_id = 1;
  string namespace = 2;
  string key = 3;
}

message ExistsResponse {
  common.StatusResponse status = 1;
  bool exists = 2;
}

message SetWithTTLRequest {
  string request_id = 1;
  string namespace = 2;
  string key = 3;
  string value = 4;
  int32 ttl_seconds = 5;
}

message IncrByRequest {
  string request_id = 1;
  string namespace = 2;
  string key = 3;
  int64 increment = 4;
}

message IncrByResponse {
  common.StatusResponse status = 1;
  int64 new_value = 2;
}

message ExpireRequest {
  string request_id = 1;
  string namespace = 2;
  string key = 3;
  int32 ttl_seconds = 4;
}

message RateLimitRequest {
  string request_id = 1;
  string key = 2;               // e.g., "recovery:merchant_123"
  int32 max_requests = 3;       // e.g., 5
  int32 window_seconds = 4;     // e.g., 60
}

message RateLimitResponse {
  common.StatusResponse status = 1;
  bool allowed = 2;
  int32 remaining = 3;
  int32 retry_after_seconds = 4;
}

message IdempotencyRequest {
  string request_id = 1;
  string idempotency_key = 2;
  string value = 3;
  int32 ttl_seconds = 4;       // Default: 86400 (24h)
}

message IdempotencyResponse {
  common.StatusResponse status = 1;
  bool is_new = 2;              // true = first time, false = duplicate
  string existing_value = 3;    // If duplicate, return the stored result
}
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
  string event_type = 3;          // "payment.failed", "payment.captured", etc.
  google.protobuf.Struct payload = 4;
  google.protobuf.Timestamp received_at = 5;
  string signature_valid = 6;     // "VALID" | "INVALID" | "SKIPPED"
  string processing_status = 7;   // "RECEIVED" | "PROCESSING" | "PROCESSED" | "FAILED"
}

message GetEventRequest {
  string request_id = 1;
  string event_id = 2;
}

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
}

message DiagnoseRequest {
  string request_id = 1;
  string payment_id = 2;
  string razorpay_error_code = 3;
  string razorpay_error_description = 4;
  string payment_method = 5;    // "upi", "card", "netbanking", "wallet"
  common.Money amount = 6;
  string customer_id = 7;
  string merchant_id = 8;
}

message DiagnosisResult {
  common.StatusResponse status = 1;
  string diagnosis_id = 2;
  FailureCategory category = 3;
  RecoverySuggestion suggestion = 4;
  string root_cause = 5;           // Human-readable explanation
  string ai_reasoning = 6;         // LLM's reasoning (if AI was used)
  float confidence = 7;            // 0.0 - 1.0
  bool used_ai = 8;                // true if LLM was invoked, false if rule-based
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
  string recovery_status = 11;    // "PENDING" | "IN_PROGRESS" | "RECOVERED" | "FAILED" | "ABANDONED"
  google.protobuf.Timestamp failed_at = 12;
  google.protobuf.Timestamp diagnosed_at = 13;
}

message GetFailureRequest {
  string request_id = 1;
  string failure_id = 2;
}

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

message StatsRequest {
  string request_id = 1;
  string merchant_id = 2;
  common.TimeRange time_range = 3;
}

message FailureStats {
  common.StatusResponse status = 1;
  int32 total_failures = 2;
  common.Money total_amount_at_risk = 3;
  map<string, int32> by_category = 4;
  map<string, int32> by_payment_method = 5;
  map<string, int32> by_recovery_status = 6;
}

message PatternRequest {
  string request_id = 1;
  string merchant_id = 2;
  common.TimeRange time_range = 3;
}

message PatternReport {
  common.StatusResponse status = 1;
  repeated Pattern patterns = 2;
}

message Pattern {
  string pattern_type = 1;        // "BANK_OUTAGE", "TIME_CLUSTER", "METHOD_SPIKE"
  string description = 2;
  float significance = 3;         // Statistical significance 0-1
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
}

enum RecoveryAction {
  ACTION_UNKNOWN = 0;
  ACTION_RETRY_PAYMENT = 1;
  ACTION_CREATE_PAYMENT_LINK = 2;
  ACTION_SEND_REMINDER_EMAIL = 3;
  ACTION_WAIT_AND_RETRY = 4;
  ACTION_ESCALATE_TO_HUMAN = 5;
  ACTION_ABANDON = 6;
}

enum WorkflowStatus {
  WF_CREATED = 0;
  WF_IN_PROGRESS = 1;
  WF_RECOVERED = 2;
  WF_FAILED = 3;
  WF_ABANDONED = 4;
  WF_ESCALATED = 5;
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
  string recovered_payment_id = 13;   // If recovery succeeded
}

message RecoveryStep {
  string step_id = 1;
  RecoveryAction action = 2;
  string status = 3;                  // "PENDING" | "EXECUTING" | "SUCCESS" | "FAILED" | "SKIPPED"
  string result_detail = 4;
  string ai_reasoning = 5;
  map<string, string> metadata = 6;   // e.g., payment_link_url, new_payment_id
  google.protobuf.Timestamp executed_at = 7;
  int32 duration_ms = 8;
  repeated string guardrails_checked = 9;
}

message ExecuteStepRequest {
  string request_id = 1;
  string workflow_id = 2;
}

message StepResult {
  common.StatusResponse status = 1;
  RecoveryStep step = 2;
  WorkflowStatus new_workflow_status = 3;
  string next_action = 4;
}

message GetWorkflowRequest {
  string request_id = 1;
  string workflow_id = 2;
}

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

message RecoveryStatsRequest {
  string request_id = 1;
  string merchant_id = 2;
  common.TimeRange time_range = 3;
}

message RecoveryStats {
  common.StatusResponse status = 1;
  common.Money total_at_risk = 2;
  common.Money total_recovered = 3;
  common.Money total_cost = 4;
  float recovery_rate = 5;            // 0.0 - 1.0
  int32 total_workflows = 6;
  map<string, int32> by_status = 7;
  map<string, int32> by_action = 8;
  map<string, float> recovery_rate_by_method = 9;
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

enum MatchType {
  MATCH_UNKNOWN = 0;
  EXACT_MATCH = 1;
  FUZZY_MATCH = 2;
  AI_MATCH = 3;
  UNMATCHED = 4;
}

enum ExceptionType {
  EXC_UNKNOWN = 0;
  AMOUNT_MISMATCH = 1;
  MISSING_SETTLEMENT = 2;
  MISSING_ORDER = 3;
  DUPLICATE_SETTLEMENT = 4;
  FEE_DISCREPANCY = 5;
  TIMING_MISMATCH = 6;
}

message ReconciliationRequest {
  string request_id = 1;
  string merchant_id = 2;
  common.TimeRange time_range = 3;
}

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
  string id = 1;
  string order_id = 2;
  string payment_id = 3;
  string settlement_id = 4;
  common.Money order_amount = 5;
  common.Money payment_amount = 6;
  common.Money settled_amount = 7;
  MatchType match_type = 8;
  float confidence = 9;
  string notes = 10;
}

message GetReportRequest {
  string request_id = 1;
  string report_id = 2;
}

message ListExceptionsRequest {
  string request_id = 1;
  string report_id = 2;
  ExceptionType exception_type = 3;
  common.PaginationRequest pagination = 4;
}

message ExceptionList {
  common.StatusResponse status = 1;
  repeated ExceptionRecord exceptions = 2;
  common.PaginationResponse pagination = 3;
}

message ExceptionRecord {
  string id = 1;
  ExceptionType type = 2;
  string description = 3;
  string order_id = 4;
  string payment_id = 5;
  string settlement_id = 6;
  common.Money expected_amount = 7;
  common.Money actual_amount = 8;
  string ai_suggestion = 9;
  string resolution_status = 10;   // "OPEN" | "RESOLVED" | "IGNORED"
  string resolved_by = 11;
  google.protobuf.Timestamp resolved_at = 12;
}

message ResolveExceptionRequest {
  string request_id = 1;
  string exception_id = 2;
  string resolution = 3;           // "ACCEPT_AI_SUGGESTION" | "MANUAL_MATCH" | "IGNORE"
  string notes = 4;
}
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
  string request_id = 1;
  string service = 2;             // "recovery-orchestrator", "failure-detector", etc.
  string action = 3;              // "RETRY_PAYMENT", "DIAGNOSE_FAILURE", etc.
  string entity_type = 4;         // "payment", "workflow", "reconciliation"
  string entity_id = 5;           // payment_id, workflow_id, etc.
  string actor = 6;               // "strategy_agent", "system", "user"
  google.protobuf.Struct input = 7;
  google.protobuf.Struct output = 8;
  string reasoning = 9;           // Why this action was taken
  repeated string guardrails_checked = 10;
  int32 duration_ms = 11;
  string parent_request_id = 12;
  string status = 13;             // "SUCCESS" | "FAILED" | "BLOCKED"
  string error_detail = 14;
}

message LogResponse {
  common.StatusResponse status = 1;
  string audit_id = 2;
  google.protobuf.Timestamp logged_at = 3;
}

message GetTrailRequest {
  string request_id = 1;
  string service = 2;
  string action = 3;
  common.TimeRange time_range = 4;
  common.PaginationRequest pagination = 5;
}

message EntityTrailRequest {
  string request_id = 1;
  string entity_type = 2;
  string entity_id = 3;
}

message AuditTrail {
  common.StatusResponse status = 1;
  repeated AuditRecord entries = 2;
  common.PaginationResponse pagination = 3;
}

message AuditRecord {
  string audit_id = 1;
  string service = 2;
  string action = 3;
  string entity_type = 4;
  string entity_id = 5;
  string actor = 6;
  google.protobuf.Struct input = 7;
  google.protobuf.Struct output = 8;
  string reasoning = 9;
  repeated string guardrails_checked = 10;
  int32 duration_ms = 11;
  string status = 12;
  string error_detail = 13;
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
  string request_id = 1;
  string to_email = 2;
  string subject = 3;
  string body_html = 4;
  string body_text = 5;
  string template_id = 6;
  map<string, string> template_vars = 7;
}

message DeliveryStatusRequest {
  string request_id = 1;
  string notification_id = 2;
}

message DeliveryStatus {
  common.StatusResponse status = 1;
  string delivery_state = 2;      // "QUEUED" | "SENT" | "DELIVERED" | "BOUNCED" | "FAILED"
  string error_detail = 3;
}
```

---

## STRANDS AGENTS SPECIFICATION

### Agent 1: Failure Diagnosis Agent (failure-detector service)

```python
from strands import Agent, tool
from strands.models.openai import OpenAIModel

model = OpenAIModel(
    model_id="mistral.ministral-3-8b-instruct",
    client_args={"base_url": "https://bedrock-mantle.us-east-1.api.aws/v1"}
)

@tool
def lookup_error_code(error_code: str) -> str:
    """Look up a Razorpay error code and return its meaning and typical cause.
    Args:
        error_code: The Razorpay error code (e.g., 'BAD_REQUEST_ERROR')
    """
    # Lookup from hardcoded map of known Razorpay error codes
    pass

@tool
def get_customer_payment_history(customer_id: str, limit: int = 10) -> str:
    """Get a customer's recent payment history to understand patterns.
    Args:
        customer_id: The customer ID
        limit: Number of recent payments to fetch
    """
    # gRPC call to mongo-service
    pass

@tool
def check_bank_health(bank_name: str) -> str:
    """Check if a bank is experiencing known issues or outages.
    Args:
        bank_name: Name of the bank (e.g., 'HDFC', 'SBI', 'ICICI')
    """
    # Check recent failure rates from mongo, cached in redis
    pass

@tool
def get_failure_rate_by_method(payment_method: str, hours: int = 1) -> str:
    """Get the failure rate for a specific payment method in the last N hours.
    Args:
        payment_method: Payment method (e.g., 'upi', 'card', 'netbanking')
        hours: Time window in hours
    """
    # Aggregate query via mongo-service
    pass

diagnosis_agent = Agent(
    name="failure_diagnoser",
    model=model,
    system_prompt="""You are a Razorpay payment failure diagnosis specialist.
Given a payment failure event with error code and context, determine:
1. The failure category (one of: INSUFFICIENT_FUNDS, CARD_EXPIRED, BANK_DECLINE,
   NETWORK_ERROR, AUTHENTICATION_FAILED, FRAUD_SUSPECTED, INVALID_CARD, UNKNOWN)
2. The root cause explanation
3. The recommended recovery action

Rules:
- Use lookup_error_code FIRST to understand the error
- Check customer history for patterns
- Check bank health for systemic issues  
- Be precise. If you're less than 70% confident, say category is UNKNOWN
- Never fabricate data. Only report what the tools return.""",
    tools=[lookup_error_code, get_customer_payment_history,
           check_bank_health, get_failure_rate_by_method]
)
```

### Agent 2: Recovery Strategy Agent (recovery-orchestrator service)

```python
@tool
def check_retry_eligibility(payment_id: str) -> str:
    """Check if a payment is eligible for retry (hasn't exceeded limits).
    Args:
        payment_id: The failed payment ID
    """
    # Check redis for retry count, check guardrails
    pass

@tool
def check_contact_window() -> str:
    """Check if the current time is within allowed contact hours (9 AM - 9 PM IST)."""
    pass

@tool
def calculate_recovery_cost(amount_paise: int, action: str) -> str:
    """Calculate the cost of a recovery action and check if it's within the cost cap.
    Args:
        amount_paise: Original payment amount in paise
        action: Recovery action type ('retry', 'payment_link', 'email', 'escalate')
    """
    pass

@tool
def check_customer_dnd(customer_id: str) -> str:
    """Check if a customer has opted out of recovery communications.
    Args:
        customer_id: The customer ID
    """
    pass

strategy_agent = Agent(
    name="recovery_strategist",
    model=model,
    system_prompt="""You are a payment recovery strategist.
Given a diagnosed payment failure, decide the best recovery action.

Available actions:
- RETRY_PAYMENT: Retry the same payment method (only if eligible)
- CREATE_PAYMENT_LINK: Generate a new payment link for the customer
- SEND_REMINDER_EMAIL: Send an email with payment instructions
- WAIT_AND_RETRY: Schedule a retry after a delay
- ESCALATE_TO_HUMAN: Flag for human review
- ABANDON: Stop recovery (cost exceeds benefit)

Rules:
- ALWAYS check retry eligibility before suggesting retry
- ALWAYS check contact window before suggesting email
- ALWAYS check cost cap before any action
- ALWAYS check DND before customer contact
- If amount < ₹50, prefer ABANDON over expensive recovery
- If fraud suspected, ALWAYS escalate
- Explain your reasoning clearly""",
    tools=[check_retry_eligibility, check_contact_window,
           calculate_recovery_cost, check_customer_dnd]
)
```

### Agent 3: Recovery Executor Agents (recovery-orchestrator service)

```python
@tool
def retry_razorpay_payment(payment_id: str, idempotency_key: str) -> str:
    """Retry a failed Razorpay payment using the Razorpay API.
    Args:
        payment_id: The original failed payment ID
        idempotency_key: Unique key to prevent double-charging
    """
    # Call Razorpay API with idempotency, log to audit
    pass

@tool  
def create_razorpay_payment_link(amount_paise: int, customer_email: str,
                                  description: str, order_id: str) -> str:
    """Create a Razorpay payment link for customer to pay via alternate method.
    Args:
        amount_paise: Amount in paise
        customer_email: Customer's email address
        description: Payment description
        order_id: Associated order ID
    """
    # Call Razorpay Payment Links API
    pass

@tool
def send_recovery_email(customer_email: str, template: str,
                        payment_link_url: str, amount_display: str) -> str:
    """Send a recovery email to the customer via notification service.
    Args:
        customer_email: Recipient email
        template: Template name ('payment_failed', 'retry_reminder', 'payment_link')
        payment_link_url: URL of the payment link (if applicable)
        amount_display: Human-readable amount (e.g., '₹1,500.00')
    """
    # Publish to Redpanda NOTIFICATIONS topic
    pass

@tool
def create_escalation_ticket(workflow_id: str, reason: str,
                              severity: str) -> str:
    """Create an escalation ticket for human review.
    Args:
        workflow_id: The recovery workflow ID
        reason: Why this is being escalated
        severity: 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL'
    """
    # Store in MongoDB, log to audit
    pass
```

### Agent 4: Reconciliation Agent (reconciliation-engine service)

```python
@tool
def search_orders_by_amount(amount_paise: int, tolerance_paise: int = 0,
                             date_range_days: int = 7) -> str:
    """Search orders by amount with optional tolerance for fuzzy matching.
    Args:
        amount_paise: Target amount in paise
        tolerance_paise: Acceptable deviation in paise (e.g., 500 for ±₹5)
        date_range_days: How many days back to search
    """
    pass

@tool
def search_payments_by_settlement(settlement_id: str) -> str:
    """Find all payments associated with a settlement.
    Args:
        settlement_id: The Razorpay settlement ID
    """
    pass

@tool
def check_fee_schedule(payment_method: str, amount_paise: int) -> str:
    """Calculate expected Razorpay fees for a payment to explain amount differences.
    Args:
        payment_method: Payment method ('upi', 'card', 'netbanking')
        amount_paise: Payment amount in paise
    """
    pass

@tool
def flag_as_exception(record_id: str, exception_type: str,
                       description: str) -> str:
    """Flag a record as an unresolvable exception for human review.
    Args:
        record_id: The match record ID
        exception_type: Type of exception
        description: Explanation of why it can't be resolved
    """
    pass

reconciliation_agent = Agent(
    name="reconciler",
    model=model,
    system_prompt="""You are a financial reconciliation specialist.
You receive unmatched transactions that automated exact and fuzzy matching 
could not resolve. Your job is to find the correct match or explain why 
one doesn't exist.

Common reasons for mismatches:
- Razorpay processing fees (2% for cards, 0% for UPI)
- GST on fees (18%)
- Partial refunds reducing settled amounts
- Split settlements across multiple days
- Currency rounding

Rules:
- NEVER fabricate a match. If uncertain, flag as exception.
- Always explain your reasoning
- Check fee schedules when amounts are close but not exact
- Consider that settlements may be net of fees""",
    tools=[search_orders_by_amount, search_payments_by_settlement,
           check_fee_schedule, flag_as_exception]
)
```

### Agent 5: Master Orchestrator Agent (ai-gateway service)

```python
# All sub-agents exposed as tools via gRPC wrappers
@tool
def analyze_payment_failure(payment_id: str) -> str:
    """Analyze why a specific payment failed and get diagnosis.
    Args:
        payment_id: The Razorpay payment ID
    """
    # gRPC call to failure-detector
    pass

@tool
def start_recovery(failure_id: str) -> str:
    """Start a recovery workflow for a diagnosed failure.
    Args:
        failure_id: The failure record ID
    """
    # gRPC call to recovery-orchestrator
    pass

@tool
def get_recovery_dashboard() -> str:
    """Get current recovery dashboard metrics (at risk, recovered, rate)."""
    # gRPC call to dashboard-api
    pass

@tool
def run_reconciliation(merchant_id: str, days: int = 7) -> str:
    """Run settlement reconciliation for a merchant.
    Args:
        merchant_id: The merchant ID
        days: Number of days to reconcile
    """
    # gRPC call to reconciliation-engine
    pass

@tool
def get_audit_trail(entity_id: str) -> str:
    """Get the full audit trail for any entity (payment, workflow, etc.).
    Args:
        entity_id: The entity ID to trace
    """
    # gRPC call to audit-service
    pass

master_agent = Agent(
    name="recoveriq_master",
    model=model,
    system_prompt="""You are RecoverIQ, an AI-powered revenue recovery assistant.
You help merchants understand, recover, and reconcile their payment revenue.

Capabilities:
- Analyze why payments failed
- Start and monitor recovery workflows
- Run settlement reconciliation
- Query audit trails
- Provide dashboard analytics

Rules:
- Always be precise with monetary amounts (show in ₹)
- Cite the data source for every claim
- If you don't have data, say so — never guess
- Explain complex financial concepts simply""",
    tools=[analyze_payment_failure, start_recovery,
           get_recovery_dashboard, run_reconciliation, get_audit_trail]
)
```

---

## GUARDRAILS (Hard-coded, NOT AI-controlled)

```python
# shared/guardrails.py — These limits are NEVER overridable by AI

class RecoveryGuardrails:
    # Retry limits
    MAX_RETRIES_PER_PAYMENT: int = 3
    RETRY_BACKOFF_SECONDS: list = [60, 3600, 86400]  # 1min, 1hr, 24hr
    
    # Contact limits
    MAX_CONTACTS_PER_CUSTOMER_PER_DAY: int = 2
    NO_CONTACT_BEFORE_HOUR: int = 9     # 9 AM IST
    NO_CONTACT_AFTER_HOUR: int = 21     # 9 PM IST
    
    # Cost limits
    MAX_RECOVERY_COST_PERCENT: float = 0.20  # Stop if cost > 20% of payment
    MIN_RECOVERABLE_AMOUNT_PAISE: int = 5000  # Don't recover < ₹50
    
    # Circuit breaker
    CIRCUIT_BREAKER_FAILURE_THRESHOLD: int = 5  # failures in window
    CIRCUIT_BREAKER_WINDOW_SECONDS: int = 60
    CIRCUIT_BREAKER_COOLDOWN_SECONDS: int = 300
    
    # Idempotency
    IDEMPOTENCY_TTL_SECONDS: int = 86400  # 24 hours
    
    # Workflow limits
    MAX_WORKFLOW_AGE_DAYS: int = 7  # Abandon after 7 days
    MAX_ACTIVE_WORKFLOWS_PER_MERCHANT: int = 100
```

---

## MONGODB COLLECTIONS

| Collection | Purpose | Key Fields |
|---|---|---|
| `webhook_events` | Raw Razorpay webhook events | `razorpay_event_id`, `event_type`, `payload`, `received_at` |
| `failures` | Diagnosed payment failures | `payment_id`, `category`, `suggestion`, `root_cause`, `recovery_status` |
| `recovery_workflows` | Recovery workflow state | `workflow_id`, `failure_id`, `status`, `steps[]`, `retry_count` |
| `reconciliation_reports` | Recon run results | `report_id`, `match_rate`, `matches[]`, `exceptions[]` |
| `orders` | Synthetic order data | `order_id`, `amount`, `customer_id`, `status`, `created_at` |
| `payments` | Synthetic payment data | `payment_id`, `order_id`, `amount`, `status`, `method`, `error_code` |
| `settlements` | Synthetic settlement data | `settlement_id`, `amount`, `utr`, `payment_ids[]`, `settled_at` |
| `customers` | Customer profiles | `customer_id`, `email`, `dnd`, `contact_count_today` |
| `audit_log` | All system actions | `service`, `action`, `entity_id`, `reasoning`, `created_at` |
| `escalations` | Human review tickets | `workflow_id`, `reason`, `severity`, `status` |

### Indexes

```javascript
// webhook_events
db.webhook_events.createIndex({ "razorpay_event_id": 1 }, { unique: true })
db.webhook_events.createIndex({ "event_type": 1, "received_at": -1 })

// failures
db.failures.createIndex({ "payment_id": 1 }, { unique: true })
db.failures.createIndex({ "category": 1, "recovery_status": 1 })
db.failures.createIndex({ "merchant_id": 1, "failed_at": -1 })

// recovery_workflows
db.recovery_workflows.createIndex({ "workflow_id": 1 }, { unique: true })
db.recovery_workflows.createIndex({ "failure_id": 1 })
db.recovery_workflows.createIndex({ "status": 1, "created_at": -1 })

// audit_log
db.audit_log.createIndex({ "entity_type": 1, "entity_id": 1 })
db.audit_log.createIndex({ "service": 1, "created_at": -1 })
db.audit_log.createIndex({ "created_at": -1 }, { expireAfterSeconds: 7776000 }) // 90 days TTL
```

---

## REDIS NAMESPACES

| Namespace | Purpose | Key Pattern | TTL |
|---|---|---|---|
| `idempotency` | Prevent double-charges | `idempotency:{payment_id}:{action}` | 24h |
| `rate_limit` | Rate limiting per entity | `rate_limit:{entity}:{window}` | Dynamic |
| `circuit_breaker` | Circuit breaker state | `cb:{service}:{merchant_id}` | 5 min |
| `retry_count` | Track retry attempts | `retry:{payment_id}` | 7 days |
| `contact_count` | Daily contact counter | `contact:{customer_id}:{date}` | 24h |
| `cache` | General query cache | `cache:{collection}:{hash}` | 5 min |
| `session` | AI chat sessions | `session:{session_id}` | 1h |

---

## REDPANDA TOPICS

| Topic | Publisher | Consumer | Payload |
|---|---|---|---|
| `PAYMENT_EVENTS` | webhook-receiver | failure-detector | `{ event_type, payment_id, error_code, amount, ... }` |
| `FAILURE_DIAGNOSED` | failure-detector | recovery-orchestrator | `{ failure_id, category, suggestion, confidence, ... }` |
| `RECOVERY_EVENTS` | recovery-orchestrator | dashboard-api (real-time) | `{ workflow_id, action, status, amount_recovered, ... }` |
| `NOTIFICATIONS` | recovery-orchestrator | notification-worker | `{ type: "email", to, subject, body, template, ... }` |
| `AUDIT_EVENTS` | all services | audit-service | `{ service, action, entity_id, ... }` |
| `DLQ` | any service | monitoring/alerting | Dead-letter events that failed processing |

---

## SYNTHETIC DATA SPECIFICATION

Generate **100 records** with these distributions:

### Payments (60 failed + 40 successful)

```python
FAILURE_DISTRIBUTION = {
    "BAD_REQUEST_ERROR": {           # 15 records
        "sub_codes": ["payment_failed", "card_expired", "invalid_card_number"],
        "methods": ["card"],
        "amounts_range": (10000, 500000),  # ₹100 - ₹5,000
    },
    "GATEWAY_ERROR": {               # 12 records
        "sub_codes": ["gateway_timeout", "gateway_internal_error"],
        "methods": ["card", "netbanking", "upi"],
        "amounts_range": (5000, 200000),
    },
    "SERVER_ERROR": {                # 8 records
        "sub_codes": ["internal_error"],
        "methods": ["upi", "card"],
        "amounts_range": (1000, 100000),
    },
    "BAD_REQUEST_ERROR_FUNDS": {     # 15 records
        "sub_codes": ["insufficient_funds"],
        "methods": ["card", "upi"],
        "amounts_range": (50000, 1000000),   # Higher amounts
    },
    "BAD_REQUEST_ERROR_AUTH": {      # 5 records
        "sub_codes": ["authentication_failed", "3ds_failed"],
        "methods": ["card"],
        "amounts_range": (20000, 300000),
    },
    "BAD_REQUEST_ERROR_FRAUD": {     # 5 records
        "sub_codes": ["suspected_fraud", "risk_check_failed"],
        "methods": ["card"],
        "amounts_range": (100000, 500000),
    },
}
```

### Settlements (30 records)

- 20 matching orders exactly
- 5 with fee deductions (amount differs by 2% + GST)
- 3 with timing mismatches (settled 2-3 days after payment)
- 2 genuinely missing (no matching order)

### Customers (15 unique)

- 10 normal customers (allow contact)
- 3 DND opted-out
- 2 high-value repeat customers (10+ orders each)

---

## NEXT.JS FRONTEND SPECIFICATION

### Design System

- **Theme**: Dark mode with glassmorphism
- **Font**: Inter (Google Fonts)
- **Colors**:
  - Background: `#0a0a0f` (deep dark)
  - Surface: `rgba(255, 255, 255, 0.05)` (glass)
  - Border: `rgba(255, 255, 255, 0.1)`
  - Primary: `#6366f1` (indigo)
  - Success: `#22c55e` (green)
  - Warning: `#f59e0b` (amber)
  - Danger: `#ef4444` (red)
  - Text Primary: `#f1f5f9`
  - Text Secondary: `#94a3b8`
- **Effects**: Backdrop blur, subtle gradients, micro-animations on hover

### Pages

#### 1. Overview Dashboard (`/`)
- **Hero metrics row**: 4 glassmorphism cards
  - ₹ At Risk (red glow pulse)
  - ₹ Recovered (green glow)
  - Recovery Rate % (gauge chart)
  - Recon Match Rate % (gauge chart)
- **Recovery Trend Chart**: Line chart — daily ₹ recovered over last 7 days (Recharts)
- **Failure Distribution**: Donut chart — failures by category
- **Recent Activity Feed**: Live-updating list of latest actions
- **Quick Actions**: "Run Reconciliation", "View Failures", "Start Recovery Batch"

#### 2. Failures Page (`/failures`)
- **Filter bar**: Category dropdown, status dropdown, date range picker, payment method
- **Data table**: Sortable columns — Payment ID, Amount, Method, Error, Category, Status, Time
- **Status badges**: Color-coded (PENDING=amber, IN_PROGRESS=blue, RECOVERED=green, FAILED=red)
- **Click row** → navigate to failure detail

#### 3. Failure Detail Page (`/failures/[id]`)
- **Header**: Payment ID, amount, customer, merchant
- **Diagnosis Card**: Category, root cause, confidence, AI reasoning (if used)
- **Recovery Timeline**: Vertical timeline showing every step:
  - Step 1: Diagnosed → category
  - Step 2: Strategy selected → action
  - Step 3: Action executed → result
  - Step 4: Outcome → recovered/failed
- **Audit Trail Table**: Every audit entry for this payment
- **Actions**: "Retry Recovery", "Escalate", "Abandon"

#### 4. Recoveries Page (`/recoveries`)
- **Tab navigation**: Active | Completed | Escalated | Abandoned
- **Workflow cards**: Each showing workflow status, steps completed, amount, time elapsed
- **Bulk actions**: "Start Recovery for All Pending Failures"

#### 5. Reconciliation Page (`/reconciliation`)
- **Run Reconciliation button** (with loading state)
- **Results summary**: Match rate bar (exact/fuzzy/AI/unmatched segments)
- **Matched records table**: With match type badges
- **Exceptions table**: With "Resolve" action buttons
- **Discrepancy highlight**: Total expected vs. total settled

#### 6. Audit Trail Page (`/audit`)
- **Search bar**: Search by entity ID, service, action
- **Filter chips**: By service, by action type, by status
- **Timeline view**: Chronological list with expandable details
- **Export button**: Download as CSV

#### 7. Chat Page (`/chat`)
- **Chat interface**: Message bubbles, streaming responses
- **Suggested prompts**: "Why did payments fail yesterday?", "Run reconciliation", "Show recovery stats"
- **Tool call visualization**: Show when the agent is calling tools (loading indicators)
- **Context panel**: Side panel showing referenced data

#### 8. Settings Page (`/settings`)
- **API Keys**: Razorpay key display (masked)
- **Guardrails config**: Display current limits (read-only in v1)
- **Notification preferences**: Email templates preview
- **System health**: Service status indicators

---

## FAILURE HANDLING MATRIX

| # | Failure Scenario | Detection | Response | Audit Entry |
|---|---|---|---|---|
| 1 | Razorpay API timeout | HTTP 504 / timeout exception | Exponential backoff: 1s → 2s → 4s, max 3 retries | `RAZORPAY_TIMEOUT: payment_id=X, attempt=N` |
| 2 | Razorpay API rate limit | HTTP 429 | Respect `Retry-After` header, queue for later | `RATE_LIMITED: retry_after=Ns` |
| 3 | Double-charge risk | Before every retry | Check Redis idempotency key. If exists, return cached result | `IDEMPOTENCY_HIT: key=X, cached_result=Y` |
| 4 | LLM hallucination | Post-LLM validation | Validate output against known categories/enums. Fallback to rule-based | `LLM_FALLBACK: confidence=0.3, used=rule_engine` |
| 5 | LLM timeout/error | API exception | Fallback to rule-based classification entirely | `LLM_UNAVAILABLE: error=X, fallback=rules` |
| 6 | Retry storm | Circuit breaker check | If > 5 failures in 60s for same merchant → open circuit, pause 5min | `CIRCUIT_BREAKER_OPEN: merchant=X, failures=N` |
| 7 | Customer harassment | Contact counter check | Hard cap: 2 contacts/customer/day. Enforce time window. | `CONTACT_LIMIT: customer=X, count=2, action=SKIP` |
| 8 | Recovery cost > value | Cost calculation | If recovery cost > 20% of payment amount → abandon | `COST_CAP_HIT: amount=₹50, cost=₹15, action=ABANDON` |
| 9 | MongoDB write failure | Exception catch | Retry 2x, then publish to DLQ topic for later processing | `MONGO_WRITE_FAIL: collection=X, retries=2, sent_to=DLQ` |
| 10 | Redpanda publish failure | Exception catch | Store event in Redis as fallback, background worker retries | `REDPANDA_FAIL: topic=X, fallback=redis` |
| 11 | Webhook signature invalid | HMAC validation | Reject with 401, log suspicious event | `INVALID_SIGNATURE: event_id=X, ip=Y` |
| 12 | gRPC service unavailable | Connection error | Retry with backoff, degrade gracefully (return partial data) | `SERVICE_UNAVAILABLE: service=X, degraded=true` |

---

## DOCKER COMPOSE SPECIFICATION

```yaml
# Networks
networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge
    internal: true

# Volumes
volumes:
  mongodb-data:
  redis-data:
  redpanda-data:

# Services: 14 containers total
# Infrastructure: mongodb, redis, redpanda, redpanda-console (4)
# Data access: mongo-service, redis-service (2)
# Application: webhook-receiver, failure-detector, recovery-orchestrator,
#              reconciliation-engine, dashboard-api, ai-gateway,
#              audit-service, notification-worker (8)
# Frontend: frontend (1)
```

Each Python service Dockerfile:
```dockerfile
# Multi-stage build
FROM python:3.12-slim AS base
# Install uv
# Copy shared/ and proto/ 
# Compile protos
# Install service dependencies

FROM base AS dev
# Hot-reload with uvicorn --reload

FROM base AS prod
# Optimized production run
```

---

## ENVIRONMENT VARIABLES

```env
# Infrastructure
MONGO_URI=mongodb://mongodb:27017
MONGO_DB_NAME=recoveriq
REDIS_URL=redis://redis:6379
REDPANDA_BROKER=redpanda:29092

# Service addresses (gRPC)
MONGO_SERVICE_ADDR=mongo-service:50010
REDIS_SERVICE_ADDR=redis-service:50011
AUDIT_SERVICE_ADDR=audit-service:50007
FAILURE_DETECTOR_ADDR=failure-detector:50002
RECOVERY_ORCHESTRATOR_ADDR=recovery-orchestrator:50003
RECONCILIATION_ENGINE_ADDR=reconciliation-engine:50004
DASHBOARD_API_ADDR=dashboard-api:50005

# Razorpay (Test Mode)
RAZORPAY_KEY_ID=rzp_test_xxxxx
RAZORPAY_KEY_SECRET=xxxxx
RAZORPAY_WEBHOOK_SECRET=xxxxx

# LLM
LLM_BASE_URL=https://bedrock-mantle.us-east-1.api.aws/v1
LLM_MODEL_ID=mistral.ministral-3-8b-instruct
LLM_API_KEY=your_key

# Notification
AWS_SES_MOCK=true
AWS_REGION=ap-south-1
SENDER_EMAIL=noreply@recoveriq.dev

# App
MODE=dev
LOG_LEVEL=INFO
```

---

# TODO TASK LIST

## Phase 0: Project Scaffolding
- [ ] Create project directory structure
- [ ] Initialize git repository
- [ ] Create root `pyproject.toml` with shared dependencies
- [ ] Create `.env.example` with all environment variables
- [ ] Create `.gitignore` (Python, Node, Docker, env files)
- [ ] Create `README.md` with project overview
- [ ] Create `docker-compose.yml` with all 14 services
- [ ] Create Docker networks (frontend, backend) and volumes

## Phase 1: Protobuf Contracts
- [ ] Create `proto/common/common.proto` — shared types
- [ ] Create `proto/mongo/mongo_service.proto` — MongoDB access layer
- [ ] Create `proto/redis/redis_service.proto` — Redis access layer
- [ ] Create `proto/webhook/webhook.proto` — webhook events
- [ ] Create `proto/failure/failure.proto` — failure detection
- [ ] Create `proto/recovery/recovery.proto` — recovery workflows
- [ ] Create `proto/reconciliation/reconciliation.proto` — settlement matching
- [ ] Create `proto/audit/audit.proto` — audit trail
- [ ] Create `proto/notification/notification.proto` — email notifications
- [ ] Create `proto/compile_protos.sh` — compilation script
- [ ] Run proto compilation and verify generated Python stubs
- [ ] Create `shared/` package with generated proto imports

## Phase 2: Shared Package
- [ ] Create `shared/__init__.py`
- [ ] Create `shared/config.py` — base config with env var loading
- [ ] Create `shared/constants.py` — topics, namespaces, error code maps
- [ ] Create `shared/grpc_utils.py` — interceptors, error handling, health checks
- [ ] Create `shared/guardrails.py` — hard-coded safety limits
- [ ] Create `shared/razorpay_client.py` — Razorpay SDK wrapper with retry + idempotency
- [ ] Create `shared/models.py` — shared Pydantic models
- [ ] Create `shared/logging_config.py` — structured JSON logging
- [ ] Write unit tests for guardrails
- [ ] Write unit tests for razorpay_client (mocked)

## Phase 3: Infrastructure Services

### MongoDB Service (mongo-service)
- [ ] Create `services/mongo_service/Dockerfile`
- [ ] Create `services/mongo_service/pyproject.toml`
- [ ] Create `services/mongo_service/src/main.py` — gRPC server startup
- [ ] Create `services/mongo_service/src/config.py` — MongoDB connection config
- [ ] Create `services/mongo_service/src/connection.py` — Motor async MongoDB client
- [ ] Create `services/mongo_service/src/grpc_server.py` — implement MongoService RPCs
  - [ ] InsertOne — single document insert
  - [ ] FindOne — single document query
  - [ ] FindMany — paginated query with filter/sort
  - [ ] UpdateOne — single document update
  - [ ] DeleteOne — single document delete
  - [ ] Aggregate — aggregation pipeline execution
  - [ ] BulkInsert — batch insert
- [ ] Create `services/mongo_service/src/serialization.py` — Struct ↔ dict conversion
- [ ] Add health check endpoint
- [ ] Write unit tests for serialization
- [ ] Write integration test with real MongoDB
- [ ] Verify service starts and responds to gRPC calls

### Redis Service (redis-service)
- [ ] Create `services/redis_service/Dockerfile`
- [ ] Create `services/redis_service/pyproject.toml`
- [ ] Create `services/redis_service/src/main.py` — gRPC server startup
- [ ] Create `services/redis_service/src/config.py` — Redis connection config
- [ ] Create `services/redis_service/src/connection.py` — aioredis client
- [ ] Create `services/redis_service/src/grpc_server.py` — implement RedisService RPCs
  - [ ] Set / Get / Delete / Exists — basic key-value ops
  - [ ] SetWithTTL — key with expiry
  - [ ] IncrBy — atomic counter
  - [ ] CheckRateLimit — sliding window rate limiter
  - [ ] CheckAndSetIdempotency — atomic check-and-set for dedup
- [ ] Add namespace validation (reject unknown namespaces)
- [ ] Add health check endpoint
- [ ] Write unit tests for rate limiting logic
- [ ] Write unit tests for idempotency logic
- [ ] Verify service starts and responds to gRPC calls

## Phase 4: Core Services

### Audit Service (audit-service)
- [ ] Create service scaffolding (Dockerfile, pyproject.toml, src/)
- [ ] Implement gRPC server with RPCs:
  - [ ] LogAction — store audit entry via mongo-service gRPC
  - [ ] GetAuditTrail — query with filters via mongo-service
  - [ ] GetActionsByEntity — query by entity via mongo-service
- [ ] Add Redis caching for recent audit entries
- [ ] Add background Redpanda consumer for AUDIT_EVENTS topic
- [ ] Write unit tests
- [ ] Verify end-to-end: log action → query trail

### Webhook Receiver (webhook-receiver)
- [ ] Create service scaffolding
- [ ] Implement FastAPI HTTP endpoint `POST /webhook/razorpay`
  - [ ] Parse Razorpay webhook body
  - [ ] Validate HMAC SHA256 signature
  - [ ] Extract event type and payment data
- [ ] Implement Redpanda publisher (PAYMENT_EVENTS topic)
- [ ] Implement MongoDB storage via mongo-service gRPC
- [ ] Implement audit logging via audit-service gRPC
- [ ] Implement gRPC server (GetWebhookEvent, ListWebhookEvents)
- [ ] Handle failure: invalid signature → 401 + audit log
- [ ] Handle failure: Redpanda down → fallback to Redis
- [ ] Handle failure: MongoDB down → retry + DLQ
- [ ] Write unit tests for signature validation
- [ ] Write integration test: send fake webhook → verify stored

### Notification Worker (notification-worker)
- [ ] Create service scaffolding
- [ ] Implement Redpanda consumer (NOTIFICATIONS topic)
- [ ] Implement email sender (AWS SES or mock)
- [ ] Create email templates:
  - [ ] `payment_failed` — "Your payment of ₹X failed"
  - [ ] `retry_reminder` — "Please retry your payment"
  - [ ] `payment_link` — "Click here to complete your payment"
  - [ ] `recovery_success` — "Your payment has been processed"
- [ ] Implement audit logging for each send
- [ ] Handle failure: SES error → retry 2x, then log failure
- [ ] Write unit tests with mock SES

## Phase 5: AI-Powered Services

### Failure Detector (failure-detector)
- [ ] Create service scaffolding
- [ ] Implement Redpanda consumer (PAYMENT_EVENTS topic)
- [ ] Implement rule-based error code classifier
  - [ ] Map all Razorpay error codes to FailureCategory enum
  - [ ] Handle ~30 known error codes
  - [ ] Default UNKNOWN for unrecognized codes
- [ ] Implement Strands diagnosis agent
  - [ ] Create `@tool lookup_error_code` — hardcoded error code map
  - [ ] Create `@tool get_customer_payment_history` — gRPC to mongo-service
  - [ ] Create `@tool check_bank_health` — aggregate recent failures by bank
  - [ ] Create `@tool get_failure_rate_by_method` — aggregate by method
  - [ ] Wire agent with OpenAIModel (Bedrock Mantle)
  - [ ] Add confidence threshold: < 0.7 → fallback to rules
- [ ] Implement Strands pattern detection agent
  - [ ] Create `@tool query_failures_by_bank`
  - [ ] Create `@tool query_failures_by_time`
  - [ ] Create `@tool query_failures_by_method`
  - [ ] Create `@tool calculate_baseline_rate`
- [ ] Implement gRPC server:
  - [ ] DiagnoseFailure — classify + store result
  - [ ] GetFailureById — query from MongoDB
  - [ ] ListFailures — paginated query
  - [ ] GetFailureStats — aggregation metrics
  - [ ] DetectPatterns — batch analysis
- [ ] Publish to FAILURE_DIAGNOSED topic after diagnosis
- [ ] Log every diagnosis to audit-service
- [ ] Handle failure: LLM unavailable → pure rule-based
- [ ] Handle failure: LLM hallucination → validate output against enums
- [ ] Write unit tests for rule-based classifier (100% coverage)
- [ ] Write unit tests for agent tools (mocked gRPC)
- [ ] Write integration test: webhook event → diagnosis result

### Recovery Orchestrator (recovery-orchestrator)
- [ ] Create service scaffolding
- [ ] Implement Redpanda consumer (FAILURE_DIAGNOSED topic)
- [ ] Implement guardrails checker (import from shared/guardrails.py)
  - [ ] Retry eligibility check (count from redis-service)
  - [ ] Contact window check (time-based)
  - [ ] Cost cap check (calculate cost vs amount)
  - [ ] DND check (query mongo-service for customer preferences)
  - [ ] Circuit breaker check (query redis-service)
- [ ] Implement Strands strategy agent
  - [ ] Create `@tool check_retry_eligibility`
  - [ ] Create `@tool check_contact_window`
  - [ ] Create `@tool calculate_recovery_cost`
  - [ ] Create `@tool check_customer_dnd`
  - [ ] Wire agent with system prompt
- [ ] Implement recovery executor tools
  - [ ] Create `@tool retry_razorpay_payment` — call Razorpay API with idempotency
  - [ ] Create `@tool create_razorpay_payment_link` — Razorpay Payment Links API
  - [ ] Create `@tool send_recovery_email` — publish to NOTIFICATIONS topic
  - [ ] Create `@tool create_escalation_ticket` — store in MongoDB
- [ ] Implement GraphBuilder recovery workflow
  - [ ] Define strategy node → routes to action nodes
  - [ ] Define retry node → success/failure branches
  - [ ] Define payment_link node → notification node
  - [ ] Define escalation node → ticket creation
- [ ] Implement workflow state management (MongoDB)
  - [ ] Create workflow record on start
  - [ ] Update step status on each action
  - [ ] Update workflow status on completion
- [ ] Implement circuit breaker (Redis-backed)
  - [ ] Track failure count per merchant per window
  - [ ] Open circuit → reject all retries for cooldown period
  - [ ] Half-open → allow one test retry
- [ ] Implement gRPC server:
  - [ ] CreateWorkflow
  - [ ] ExecuteNextStep
  - [ ] GetWorkflow
  - [ ] ListWorkflows
  - [ ] GetRecoveryStats
- [ ] Publish to RECOVERY_EVENTS topic after each step
- [ ] Log every action + reasoning to audit-service
- [ ] Handle failure: Razorpay timeout → exponential backoff
- [ ] Handle failure: Double-charge → idempotency key check
- [ ] Handle failure: Circuit breaker open → skip + log
- [ ] Handle failure: LLM timeout → rule-based strategy fallback
- [ ] Write unit tests for guardrails checks
- [ ] Write unit tests for circuit breaker
- [ ] Write unit tests for workflow state transitions
- [ ] Write integration test: diagnosed failure → recovery workflow → outcome

### Reconciliation Engine (reconciliation-engine)
- [ ] Create service scaffolding
- [ ] Implement data ingestion
  - [ ] Fetch orders from mongo-service
  - [ ] Fetch payments from mongo-service
  - [ ] Fetch settlements from mongo-service
- [ ] Implement exact matcher
  - [ ] Match by order_id + payment_id + exact amount
  - [ ] Match by settlement UTR + payment_id
- [ ] Implement fuzzy matcher
  - [ ] Amount tolerance matching (±₹5 for fees)
  - [ ] Date range matching (±2 days for settlement delay)
  - [ ] Fee-adjusted matching (subtract 2% + 18% GST)
- [ ] Implement Strands reconciliation agent
  - [ ] Create `@tool search_orders_by_amount`
  - [ ] Create `@tool search_payments_by_settlement`
  - [ ] Create `@tool check_fee_schedule`
  - [ ] Create `@tool flag_as_exception`
  - [ ] Wire agent for complex exception resolution
- [ ] Implement tiered pipeline: exact → fuzzy → AI → exception
- [ ] Implement report generation
  - [ ] Match rate calculation
  - [ ] Exception categorization
  - [ ] Total expected vs settled discrepancy
- [ ] Implement gRPC server:
  - [ ] RunReconciliation
  - [ ] GetReport
  - [ ] ListExceptions
  - [ ] ResolveException
- [ ] Log all matching decisions to audit-service
- [ ] Handle failure: LLM unavailable → skip AI tier, mark as exception
- [ ] Write unit tests for exact matcher
- [ ] Write unit tests for fuzzy matcher
- [ ] Write integration test: seed data → run recon → verify match rate

## Phase 6: API & Gateway

### Dashboard API (dashboard-api)
- [ ] Create service scaffolding
- [ ] Implement FastAPI REST endpoints:
  - [ ] `GET /api/dashboard/overview` — aggregate from all services
  - [ ] `GET /api/dashboard/failures` — proxy to failure-detector gRPC
  - [ ] `GET /api/dashboard/failures/{id}` — single failure detail
  - [ ] `GET /api/dashboard/recoveries` — proxy to recovery-orchestrator
  - [ ] `GET /api/dashboard/reconciliation` — proxy to reconciliation-engine
  - [ ] `GET /api/dashboard/reconciliation/exceptions`
  - [ ] `GET /api/dashboard/audit` — proxy to audit-service
  - [ ] `GET /api/dashboard/analytics` — aggregated chart data
- [ ] Implement gRPC clients to all internal services
- [ ] Add CORS middleware (allow frontend origin)
- [ ] Add request ID middleware (generate UUID per request)
- [ ] Add error handling middleware (convert gRPC errors to HTTP)
- [ ] Write unit tests for response formatting
- [ ] Write integration test: verify each endpoint returns data

### AI Gateway (ai-gateway)
- [ ] Create service scaffolding
- [ ] Implement master Strands agent
  - [ ] Create `@tool analyze_payment_failure` — gRPC to failure-detector
  - [ ] Create `@tool start_recovery` — gRPC to recovery-orchestrator
  - [ ] Create `@tool get_recovery_dashboard` — gRPC to dashboard-api
  - [ ] Create `@tool run_reconciliation` — gRPC to reconciliation-engine
  - [ ] Create `@tool get_audit_trail` — gRPC to audit-service
- [ ] Implement FastAPI endpoints:
  - [ ] `POST /api/chat` — synchronous chat response
  - [ ] `POST /api/chat/stream` — SSE streaming response using `agent.stream_async()`
- [ ] Implement session management (FileSessionManager or Redis-backed)
- [ ] Add suggested prompts endpoint
- [ ] Handle failure: LLM timeout → return error message gracefully
- [ ] Write unit tests for tool wiring
- [ ] Write integration test: send chat message → get response with tool calls

## Phase 7: Synthetic Data & Seeding
- [ ] Create `data/synthetic_generator.py`
  - [ ] Generate 15 customer profiles
  - [ ] Generate 100 orders with realistic amounts/dates
  - [ ] Generate 60 failed + 40 successful payments
  - [ ] Generate 30 settlements with deliberate mismatches
  - [ ] Apply failure distribution (see spec above)
  - [ ] Ensure referential integrity (orders → payments → settlements)
- [ ] Create `data/init_db.py` — seed MongoDB via mongo-service gRPC
  - [ ] Create all collections
  - [ ] Create all indexes
  - [ ] Insert generated data
- [ ] Create `data/razorpay_simulator.py` — simulate webhook events
  - [ ] Generate realistic webhook payloads matching Razorpay format
  - [ ] Send to webhook-receiver HTTP endpoint
  - [ ] Configurable: send all at once or drip-feed over time
- [ ] Verify seeded data is correct and queryable

## Phase 8: Frontend (Next.js)
- [ ] Initialize Next.js 15 project (`npx -y create-next-app@latest ./`)
- [ ] Install dependencies: recharts, lucide-react
- [ ] Create design system
  - [ ] `app/globals.css` — CSS variables, dark theme, glassmorphism classes
  - [ ] `app/layout.tsx` — root layout with Inter font, sidebar navigation
- [ ] Create shared components
  - [ ] `components/Sidebar.tsx` — navigation sidebar with icons
  - [ ] `components/MetricsCard.tsx` — glassmorphism stat card with glow effects
  - [ ] `components/DataTable.tsx` — sortable, filterable table
  - [ ] `components/StatusBadge.tsx` — color-coded status pills
  - [ ] `components/Timeline.tsx` — vertical timeline for recovery steps
  - [ ] `components/LoadingSpinner.tsx` — animated loader
  - [ ] `components/ErrorBoundary.tsx` — graceful error display
- [ ] Create API client
  - [ ] `lib/api.ts` — fetch wrapper with error handling
  - [ ] Type definitions for all API responses
- [ ] Build pages:
  - [ ] Overview Dashboard (`app/page.tsx`)
    - [ ] Hero metrics cards (4 cards)
    - [ ] Recovery trend line chart (Recharts)
    - [ ] Failure distribution donut chart
    - [ ] Recent activity feed
    - [ ] Quick action buttons
  - [ ] Failures page (`app/failures/page.tsx`)
    - [ ] Filter bar (category, status, method, date)
    - [ ] Paginated data table
    - [ ] Click-through to detail
  - [ ] Failure detail page (`app/failures/[id]/page.tsx`)
    - [ ] Header with payment info
    - [ ] Diagnosis card
    - [ ] Recovery timeline
    - [ ] Audit trail section
    - [ ] Action buttons
  - [ ] Recoveries page (`app/recoveries/page.tsx`)
    - [ ] Tab navigation (Active/Completed/Escalated/Abandoned)
    - [ ] Workflow cards with progress
  - [ ] Reconciliation page (`app/reconciliation/page.tsx`)
    - [ ] Run button with loading state
    - [ ] Match rate segmented bar
    - [ ] Matched records table
    - [ ] Exceptions table with resolve buttons
  - [ ] Audit Trail page (`app/audit/page.tsx`)
    - [ ] Search bar
    - [ ] Filter chips
    - [ ] Expandable timeline entries
  - [ ] Chat page (`app/chat/page.tsx`)
    - [ ] Message list with bubbles
    - [ ] Input with send button
    - [ ] SSE streaming display
    - [ ] Suggested prompt chips
    - [ ] Tool call visualization
  - [ ] Settings page (`app/settings/page.tsx`)
    - [ ] API key display (masked)
    - [ ] Guardrails display (read-only)
    - [ ] Service health indicators
- [ ] Add responsive design (mobile-friendly)
- [ ] Add loading states for all data fetches
- [ ] Add error states for API failures
- [ ] Add micro-animations (hover, transitions, page enters)

## Phase 9: Integration & Testing
- [ ] Docker compose up — verify all 14 containers start
- [ ] Verify gRPC connectivity between all services
- [ ] Run synthetic data seeder
- [ ] Run webhook simulator — send 60 failed payment events
- [ ] Verify failure-detector classifies all 60 failures
- [ ] Verify recovery-orchestrator creates workflows
- [ ] Verify at least 1 recovery succeeds (retry or payment link)
- [ ] Verify notification-worker sends emails
- [ ] Verify audit-service has complete trail
- [ ] Run reconciliation — verify match rate > 90%
- [ ] Test AI Gateway chat — ask "Why did payments fail?"
- [ ] Test frontend — all pages load with data
- [ ] Test failure handling:
  - [ ] Kill failure-detector → verify webhook-receiver still works
  - [ ] Flood retries → verify circuit breaker opens
  - [ ] Send duplicate webhook → verify idempotency
  - [ ] Set LLM_BASE_URL to invalid → verify rule-based fallback
- [ ] Run all unit tests: `pytest --cov`
- [ ] Run integration tests
- [ ] Fix any bugs found

## Phase 10: Polish & Demo
- [ ] Write comprehensive README.md
  - [ ] Project overview and architecture diagram
  - [ ] Quick start guide (docker compose up)
  - [ ] API documentation
  - [ ] Demo walkthrough
- [ ] Create `AUDIT_TRAIL.md` — document all failure handling
- [ ] Create demo script:
  - [ ] Show the problem (₹X in failed payments)
  - [ ] Show detection (dashboard with classified failures)
  - [ ] Show recovery (click into one, watch AI diagnose and recover)
  - [ ] Show failure handled (trigger circuit breaker)
  - [ ] Show results (₹Y recovered, N% rate)
  - [ ] Show reconciliation (match rate)
  - [ ] Show audit trail (full history)
- [ ] Record demo video (optional)
- [ ] Final code cleanup and formatting
- [ ] Security review: no hardcoded secrets, test mode enforced
- [ ] Performance review: no N+1 queries, proper indexing
