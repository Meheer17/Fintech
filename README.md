# RevenueIQ — AI Revenue Recovery & Finance Control Platform

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Python](https://img.shields.io/badge/Python-3.11+-3776AB?style=flat&logo=python)](https://python.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![gRPC](https://img.shields.io/badge/gRPC-Mesh-244c5a?style=flat&logo=grpc)](https://grpc.io/)
[![Redpanda](https://img.shields.io/badge/Redpanda-Kafka-FF3E00?style=flat)](https://redpanda.com/)
[![Razorpay](https://img.shields.io/badge/Razorpay-Track_03_%26_04-0C2340?style=flat)](https://razorpay.com/)

An autonomous AI-powered **Revenue Recovery** and **Finance Control Platform** built for Razorpay merchants. Built on a polyglot microservice mesh (Go + Python AI) connected via gRPC and Redpanda/Kafka event streaming.

GitHub Repository: [github.com/Meheer17/fintech](https://github.com/Meheer17/fintech)

---

## About The Application

**RevenueIQ** is an enterprise financial agent operating 24/7 to safeguard merchant revenue. It actively monitors payment gateways, automatically recovers failed or abandoned transactions via AI-driven multi-channel strategies, reconciles bank settlement payouts down to the penny, and predicts forward cash liquidity.

---

## Why This Project? (The Problem)

1. **Revenue Leakage from Payment Failures**: Up to 15-20% of online checkout attempts fail due to bank outages, network timeouts, or insufficient funds. Merchants lose customers permanently without intelligent, timely recovery.
2. **Tedious Manual Reconciliation**: Finance teams spend dozens of hours every month manually matching bank deposits with merchant order books, missing hidden Merchant Discount Rate (MDR) and GST fee discrepancies.
3. **Regulatory Non-Compliance**: Blind automated customer spamming during non-business hours violates TRAI DND regulations in India, damaging brand reputation.

---

## Who Is The Customer?

* **D2C & E-Commerce Merchants**: High-volume checkout businesses looking to automatically recover abandoned checkouts and failed UPI/card transactions.
* **SaaS & Subscription Businesses**: Companies with recurring billing needing smart retries on failed subscription renewals to reduce involuntary churn.
* **Finance Leads & CFOs**: Accounting teams seeking automated 3-way settlement reconciliation, dispute monitoring, and 7-day predictive cash flow forecasting.

---

## Architecture & System Design

### Polyglot gRPC Microservice Mesh
RevenueIQ combines high-concurrency compiled microservices in **Go** with specialized AI/ML workflows in **Python**:
- **Go Services**: `webhook-receiver`, `checkout-tracker`, `notification-service`, `mongo-service`, `redis-service`, `audit-service`, `voice-recovery-worker`, `dashboard-api`.
- **Python Services**: `failure-detector`, `recovery-orchestrator`, `reconciliation-engine`, `ai-gateway`.
- **Inter-Service Communication**: Low-latency binary gRPC calls over shared Protocol Buffer (`.proto`) schemas defined in `revenueiq_dev_kit/proto`, bypassing HTTP JSON parsing overhead.

### Scalable Event-Driven Ingestion
- **Redpanda (Kafka) Event Bus**: Webhooks are ingested asynchronously into high-throughput Kafka topics (`PAYMENT_EVENTS`). Ingestion is decoupled from AI execution, ensuring zero dropped events during flash sale traffic spikes.
- **Stateless Microservices**: Backend services scale horizontally behind an Envoy gRPC Proxy and Kubernetes ingress.

### MongoDB Data Layer & Centralized Auditing
- **MongoDB Database**: Serves as the primary document store for payment events, recovery execution states, settlement ledgers, and raw webhook payloads with flexible schema evolution.
- **Centralized Audit Service (`audit-service`)**: Dedicated Go gRPC microservice storing immutable, timestamped audit logs for every system action, recovery dispatch, AI prompt evaluation, and settlement discrepancy flag.
- **TRAI & Regulatory Compliance**: Hardcoded guardrails enforce DND contact windows (9 AM - 9 PM IST), max contact frequency limits, and recovery cost ratio caps.

---

## Service Registry & AI Capabilities

### Core Backend & Infrastructure Microservices (Go)

1. **`webhook-receiver`** (HTTP :8001 / gRPC :50001)
   - Ingests live Razorpay webhook events (`payment.failed`, `subscription.charged.failed`, `settlement.processed`, `refund.created`, `order.paid`).
   - Validates HMAC signatures against the webhook secret.
   - Publishes verified events to the Redpanda/Kafka `PAYMENT_EVENTS` topic for asynchronous processing.

2. **`dashboard-api`** (HTTP :8005 / gRPC :50005)
   - Serves as the central REST API gateway for the React dashboard.
   - Provides aggregated endpoints for overview metrics, failure feeds, recovery statuses, 3-way reconciliation results, dispute logs, and refund reports.

3. **`audit-service`** (gRPC :50007)
   - Stores immutable, timestamped audit logs for every system event, recovery attempt, AI decision, and settlement adjustment.
   - Enables compliance auditing and historical trail verification for merchant accounting teams.

4. **`checkout-tracker`** (HTTP :8008 / gRPC :50008)
   - Tracks active user checkout sessions on merchant websites.
   - Detects cart abandonments in real-time and triggers proactive recovery workflows.

5. **`notification-service`** (HTTP :8083)
   - Consumes payment events from Redpanda.
   - Dispatches automated multi-channel customer communications via Email, SMS, and WhatsApp.

6. **`voice-recovery-worker`** (HTTP :8009)
   - Dispatches automated Hinglish IVR voice calls for high-value failed transactions.
   - Handles call scheduling, response logging, and retry state tracking.

7. **`mongo-service`** (gRPC :50010)
   - Acts as the unified MongoDB abstraction layer.
   - Handles CRUD operations and complex aggregations across payment failures, recovery records, settlement ledgers, and audit collections.

8. **`redis-service`** (gRPC :50011)
   - Provides distributed caching, rate-limiting, and idempotency locks (`SETNX`) to prevent duplicate webhook processing.

---

### AI Microservices & Specialized AI Agents (Python)

9. **`failure-detector` — Payment Failure Diagnosis Agent** (gRPC :50002)
   - Analyzes raw Razorpay payment error codes, bank responses, and customer history.
   - Classifies failure root causes (e.g., bank server downtime, insufficient balance, e-Mandate failure, expired card).
   - Recommends the optimal recovery action and retry delay window.

10. **`recovery-orchestrator` — Recovery Workflow DAG Agent** (gRPC :50003)
    - Constructs and executes bounded recovery Directed Acyclic Graphs (DAGs).
    - Enforces hardcoded TRAI DND regulatory guardrails (9 AM–9 PM IST contact hours, maximum 2 contacts/day per customer, and 20% recovery cost ratio cap).
    - Manages multi-step escalation paths (Smart Retry -> Payment Link -> Voice IVR).

11. **`reconciliation-engine` — Settlement Reconciliation Agent** (gRPC :50004)
    - Performs automated 3-way matching between Merchant Internal Ledgers, Razorpay Gateway Records, and Bank Settlement Payouts.
    - Calculates exact Merchant Discount Rate (MDR) fees and GST taxes to identify missing payouts, overcharges, or unmatched transactions.

12. **`ai-gateway` — Master Agent & Cash Flow Forecaster** (HTTP :8006 / gRPC :50006)
    - Serves as the interactive natural language Copilot for merchant teams.
    - Executes tools dynamically (e.g., creating live Razorpay payment links, querying audit logs).
    - Generates 7-day predictive cash flow forecasts based on historical settlement velocity and pending recovery pipelines.

---

## Languages, Frameworks & Tech Stack

| Layer | Languages | Frameworks & Core Tools | Purpose |
| :--- | :--- | :--- | :--- |
| **High-Performance Services** | **Go** (v1.22+) | Gin Web Framework, gRPC Go, Segmentio `kafka-go` | Webhook ingestion, audit storage, checkout tracking, worker services |
| **AI & Finance Engine** | **Python** (v3.11+) | FastAPI, LangGraph / Workflow DAGs, Pydantic, Strands Agent SDK | Failure classification, recovery orchestration, 3-way settlement matcher, AI Copilot |
| **Inter-Service Protocol** | Protobuf | gRPC (Proto3) | Low-latency binary RPC mesh between Go and Python microservices |
| **Data & Persistence** | — | MongoDB, Redis | Document store for audit logs & event ledgers; Redis for rate limiting & idempotency |
| **Messaging & Bus** | — | Redpanda (Kafka Event Bus) | Asynchronous stream processing for live payment events |
| **Frontend UI** | **TypeScript**, HTML5, CSS3 | React 18, Vite, Lucide Icons | Real-time merchant dashboard, analytics control panel, AI chat copilot |
| **Infra & DevOps** | Docker | Docker Compose, Envoy gRPC Proxy, Kubernetes Manifests | Production container orchestration & gRPC service mesh |

---

## Core Features

### Track 03: AI Revenue Recovery
- **Smart Payment Failure Diagnosis**: AI classifies payment failures (bank outages, insufficient funds, expired cards) and determines optimal retry strategies.
- **Automated Recovery DAGs**: Multi-channel recovery via smart retries, instant Razorpay payment links, and Hinglish IVR voice call dispatches.
- **TRAI Regulatory Guardrails**: Hardcoded DND compliance (9 AM–9 PM IST contact hours), max 2 contacts/day per customer, and 20% recovery cost caps.

### Track 04: AI Finance Controller
- **Three-Way Settlement Matcher**: Reconciles Merchant Ledger vs. Razorpay Gateway vs. Bank Settlements (MDR & GST aware).
- **Forward Cash Forecasting**: 7-day predictive cash flow model for working capital management.
- **AI Copilot & Master Agent**: Natural language querying for dispute tracking, settlement queries, and automated payment link creation.

---

## High-Level Architecture

```
                    ┌───────────────────────────────┐
                    │    Razorpay Webhooks / REST   │
                    └───────────────┬───────────────┘
                                    │
                  ┌─────────────────▼─────────────────┐
                  │   Webhook Receiver & Dashboard    │ (Go HTTP/gRPC)
                  └─────────────────┬─────────────────┘
                                    │ (Redpanda Kafka Events)
          ┌─────────────────────────┼─────────────────────────┐
          │                         │                         │
┌─────────▼─────────┐     ┌─────────▼─────────┐     ┌─────────▼─────────┐
│ Failure Detector  │     │ Recovery Engine   │     │ Reconciliation    │
│   (Python AI)     │     │   (Python DAG)    │     │   (Python Match)  │
└───────────────────┘     └───────────────────┘     └───────────────────┘
```

---

## Quick Start

### 1. Launch via Docker Compose
```bash
docker compose -f revenueiq_infrastructure/docker-compose/docker-compose.yml up --build -d
```

### 2. Simulate Live Webhooks
Simulate live Razorpay payment failures, captures, and settlement events:
```bash
python3 data/razorpay_simulator.py
```

### 3. Access Endpoints
- **Dashboard API**: `http://localhost:8005/api/v1/overview`
- **AI Gateway Chat**: `POST http://localhost:8006/chat`
- **Cash Forecast**: `GET http://localhost:8006/forecast?days=7`
- **Redpanda Console**: `http://localhost:8082`

---

## Running Tests

Execute the 5-tier end-to-end test suite:
```bash
python3 run_e2e_tests.py
```
