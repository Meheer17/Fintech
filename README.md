# RevenueIQ — AI Revenue Recovery & Finance Control Agent

> **What**: Production-grade AI-powered Revenue Recovery and Finance Control Platform for Razorpay Merchants.
> **Architecture**: Polyglot Microservices (Go + Python AI) + gRPC Mesh + Redpanda (Kafka) + MongoDB + Redis.
> **Tracks**: Razorpay Track 03 (AI Revenue Recovery) + Track 04 (AI Finance Controller).

---

## Architecture Overview

```
                      ┌────────────────────────────────────────┐
                      │          EXTERNAL BOUNDARY             │
                      │                                        │
                      │  Razorpay Webhooks ──HTTP──▶ webhook-receiver :8001
                      │  REST Clients      ──HTTP──▶ dashboard-api :8005
                      │  AI Chat / SSE     ──HTTP──▶ ai-gateway :8006
                      │  Checkout Session  ──HTTP──▶ checkout-tracker :8008
                      └──────────────┬─────────────────────────┘
                                     │
         ┌───────────────────────────▼──────────────────────────────────────┐
         │                 gRPC SERVICE MESH (backend network)              │
         │                                                                  │
         │  ┌─── GO SERVICES (8) ──────────────────────────────────────┐   │
         │  │  mongo-service           :50010   MongoDB access layer   │   │
         │  │  redis-service           :50011   Redis access layer     │   │
         │  │  webhook-receiver        :50001   Webhook ingestion      │   │
         │  │  audit-service           :50007   Central audit log store│   │
         │  │  notification-service    :8083    Email/SMS worker       │   │
         │  │  checkout-tracker        :50008   Session tracking       │   │
         │  │  voice-recovery-worker   :8009    Hinglish IVR worker    │   │
         │  │  dashboard-api           :50005   REST API gateway       │   │
         │  └──────────────────────────────────────────────────────────┘   │
         │                                                                  │
         │  ┌─── PYTHON AI SERVICES (4) ──────────────────────────────┐   │
         │  │  failure-detector        :50002   Classifier + Diagnosis │   │
         │  │  recovery-orchestrator   :50003   GraphBuilder + Guard   │   │
         │  │  reconciliation-engine   :50004   Three-Way Matcher      │   │
         │  │  ai-gateway              :50006   Master Agent + Chat    │   │
         │  └──────────────────────────────────────────────────────────┘   │
         └──────────────────────────────────────────────────────────────────┘
```

---

## Service Registry

| # | Service | gRPC Port | HTTP Port | Language | AI Enabled? | Role |
|---|---|---|---|---|---|---|
| 1 | `mongo-service` | 50010 | — | Go | ❌ | MongoDB CRUD & Aggregation |
| 2 | `redis-service` | 50011 | — | Go | ❌ | Cache, Rate Limiting & Idempotency |
| 3 | `webhook-receiver` | 50001 | 8001 | Go | ❌ | Razorpay Webhook Ingestion & HMAC Verification |
| 4 | `audit-service` | 50007 | — | Go | ❌ | Centralized Audit Log Storage |
| 5 | `notification-service` | — | 8083 | Go | ❌ | Redpanda Consumer -> Email/SMS |
| 6 | `checkout-tracker` | 50008 | 8008 | Go | ❌ | Session Tracking & Abandonment Detector |
| 7 | `voice-recovery-worker` | — | 8009 | Go | ❌ | Hinglish Voice IVR Call Dispatch |
| 8 | `dashboard-api` | 50005 | 8005 | Go | ❌ | REST API Gateway for Metrics & Reports |
| 9 | `failure-detector` | 50002 | — | Python | ✅ | Failure Classifier & AI Diagnosis |
| 10 | `recovery-orchestrator` | 50003 | — | Python | ✅ | Bounded Workflow DAG & Guardrails |
| 11 | `reconciliation-engine` | 50004 | — | Python | ✅ | Three-Way Settlement Matching |
| 12 | `ai-gateway` | 50006 | 8006 | Python | ✅ | Master Agent & Cash Forecaster |

---

## How to Run

### 1. Launch Full Stack via Docker Compose

```bash
docker compose -f revenueiq_infrastructure/docker-compose/docker-compose.yml up --build -d
```

### 2. Stream Synthetic Razorpay Webhooks (Simulation)

To simulate live Razorpay payment failures, subscriptions, and captures:

```bash
python3 data/razorpay_simulator.py
```

### 3. Test API & AI Endpoints

- **Dashboard Overview API**: `http://localhost:8005/api/v1/overview`
- **AI Gateway Chat**: `POST http://localhost:8006/chat`
  ```json
  {
    "query": "Why was yesterday's settlement short?"
  }
  ```
- **Forward Cash Forecast**: `GET http://localhost:8006/forecast?days=7`
- **Redpanda Console**: `http://localhost:8082`

---

## Project Structure

```
.
├── go.work                          # Root Go workspace linking all 10 modules
├── .env                             # Environment configuration
├── .env.example                     # Sample environment variables
│
├── revenueiq_dev_kit/               # Shared gRPC protobuf contracts & Go packages
├── revenueiq_infrastructure/        # Docker Compose, Envoy, K8s & Postman configs
│
├── auth_service/                    # Go authentication service
├── mongodb_service/                 # Go MongoDB wrapper
├── redis_service/                   # Go Redis wrapper
├── notification_service/            # Go notification consumer worker
├── webhook_receiver/                # Go Razorpay webhook receiver
├── audit_service/                   # Go gRPC audit logger
├── dashboard_api/                   # Go REST API gateway
├── checkout_tracker/                # Go checkout abandonment tracker
├── voice_recovery_worker/           # Go Hinglish voice worker
│
├── failure_detector/                # Python failure classifier & AI diagnosis
├── recovery_orchestrator/           # Python workflow orchestrator & guardrails
├── reconciliation_engine/           # Python settlement matcher
├── ai_gateway/                      # Python Master Agent & Cash Forecaster
│
└── data/                            # Synthetic dataset generator & Razorpay simulator
```
