# Project: RevenueIQ Production Integration & Feature Implementation

## Architecture
RevenueIQ is a microservices platform:
- **Frontend**: Next.js / Vite React Dashboard (`frontend/src/components/`):
  - `OverviewTab.tsx`, `ReconciliationTab.tsx`, `RecoveriesTab.tsx`, `FailuresTab.tsx`, `AuditTab.tsx`, `PromisesTab.tsx`, `ChatDrawer.tsx`
  - Newly added views: `DisputesTab.tsx`, `RefundsTab.tsx`
  - API Client: `frontend/src/lib/api.ts`
- **Dashboard API**: FastAPI service (`dashboard_api/`) providing REST endpoints `/api/v1/overview`, `/api/v1/failures`, `/api/v1/recoveries`, `/api/v1/reconciliation`, `/api/v1/disputes`, `/api/v1/refunds`, `/api/v1/promises`, `/api/v1/audit`.
- **Webhook Receiver**: Go server (`webhook_receiver/`) receiving Razorpay webhooks and publishing events (`subscription.charged.failed` fully supported).
- **Failure Detector**: Python service (`failure_detector/src/classifier.py`) mapping failure codes and webhooks to `failure.proto` categories (`SUBSCRIPTION_FAILED`, `MANDATE_FAILED`, etc.) and suggested actions.
- **Recovery Orchestrator**: Go/Python service (`recovery_orchestrator/src/main.py`) orchestrating workflow actions (`ACTION_CREATE_PAYMENT_LINK`, `ACTION_RETRY_SUBSCRIPTION`, `ACTION_SEND_CARD_UPDATE_LINK`, `ACTION_RENEW_MANDATE`).
- **AI Gateway**: Python FastAPI service (`ai_gateway/`) running chat tools (`create_payment_link_tool`) and dynamic cash position forecasting (`forecast_cash_position` over gRPC).
- **MongoDB Service**: Go service (`mongodb_service/`) with gRPC interface (`mongodb_service.proto`, `GetSettlementAggregations`) querying MongoDB `revenueiq_db` (`failures`, `recoveries`, `settlements`, `promises`, `audit_logs`).

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| 0 | Phase 0: Codebase Audit & Discovery | Deep audit of frontend, backend classifier/webhooks, AI gateway & gRPC | none | DONE |
| 1 | M1: Dynamic Frontend & Mock Elimination | Remove mockData.ts, replace synthetic fallbacks in Overview/Recon/Recoveries/Failures/Audit, build Disputes & Refunds tabs | M0 | DONE |
| 2 | M2: Feature Spec Alignment (F12, F13, F21, Promises) | Map subscription.charged.failed webhooks, e-Mandate error codes, gRPC cash forecast queries, Promises verification | M0 | DONE |
| 3 | M3: Dual Track E2E Test Suite Creation | Opaque-box E2E test suite covering Razorpay payment link, AI diagnosis, recovery workflow, settlement reconciliation, cash forecasting, audit trail | M0 | DONE |
| 4 | M4: Final Integration & E2E Pass + Hardening | Pass 100% E2E test suite (Tiers 1-4) and adversarial coverage hardening (Tier 5) with Forensic Auditor verification | M1, M2, M3 | DONE |

## Verification & Audit Summary
- **Reviewer Verdict**: PASS (APPROVE) — all services compile cleanly (`go build ./...`, `npm run build`, Python checks).
- **Challenger Verdict**: 106/106 Test Cases Passed (100% Pass Rate across Tiers 1-5).
- **Forensic Auditor Verdict**: CLEAN (BINARY VETO PASSED) — 0 hardcoded test values, 0 facades.
