# Original User Request

## Initial Request — 2026-08-23T12:53:07Z

Complete the production integration and feature implementation of RevenueIQ (AI Revenue Recovery & Finance Control Agent for Razorpay), eliminating mock data, static charts, and simulated fallbacks while integrating real Razorpay APIs, live Mongo aggregations, and full tool suites across microservices. Note: Voice recovery telephony and AWS SES live email can be ignored.

Working directory: /home/mahi17/Github/fintech
Integrity mode: development

## Requirements

### R1. Live API & Real Data Integration (Mock Data Elimination)
- Replace static chart fallbacks (e.g. trendData in OverviewTab.tsx) with dynamic API endpoints querying real daily aggregations from MongoDB.
- Ensure all dashboard views (Failures, Recoveries, Settlements, Subscriptions, Disputes, Refunds, Audit, Chat) render live data from MongoDB and Razorpay REST APIs.
- Remove obsolete mockData.ts references.

### R2. Feature Coverage & Specification Alignment
- **F12 Subscription Failure Recovery**: Map subscription.charged.failed webhooks to SUBSCRIPTION_FAILED category in failure_detector and failure.proto, and support subscription retry / update-card links in recovery_orchestrator.
- **F13 Mandate/AutoPay Error Codes**: Map e-Mandate/AutoPay error codes (mandate_expired, debit_rejected, mandate_not_active, insufficient_balance_mandate) in classifier.py.
- **F21 Live Cash Position Forecast**: Upgrade forecast_cash_position in ai_gateway to query historical settlement data from mongo-service via gRPC instead of static simulated values.
- **B2B Receivables & Promise-to-Pay**: Connect PromisesTab.tsx and /api/v1/promises to backend MongoDB collection promises.

### R3. Comprehensive Verification & End-to-End Testing
- Build and compile all services (Go modules, Python FastAPI services, Next.js/Vite frontend).
- Run end-to-end testing across Razorpay payment link creation, AI diagnosis, recovery workflow orchestration, settlement reconciliation, cash forecasting, and audit trail logging.

## Acceptance Criteria

### Real Data & Features
- [ ] OverviewTab.tsx renders dynamic weekly revenue performance based on real MongoDB payment failure and recovery dates/amounts.
- [ ] subscription.charged.failed and mandate error codes correctly map to SUBSCRIPTION_FAILED and MANDATE_FAILED with appropriate AI suggestions.
- [ ] forecast_cash_position queries historical settlements from MongoDB via gRPC and computes dynamic N-day forward forecast.
- [ ] PromisesTab renders live payment commitment records from MongoDB database revenueiq_db.
- [ ] create_payment_link_tool successfully creates live Razorpay links when invoked via AI Gateway chat.

### Verification
- [ ] All Python and Go microservices pass health checks (/healthz).
- [ ] Automated or script-driven verification confirms end-to-end functionality for diagnosis, orchestration, reconciliation, forecasting, and chat.
