## 2026-08-23T12:58:21Z

You are Worker M2 (Backend Classifier, Webhooks, gRPC Cash Forecast & Promises Implementer).
Your assigned working directory is: /home/mahi17/Github/fintech/.agents/teamwork_preview_worker_m2_1
Identity: Worker subagent for RevenueIQ.

Scope Document: /home/mahi17/Github/fintech/PROJECT.md
Explorer Reports:
- /home/mahi17/Github/fintech/.agents/teamwork_preview_explorer_m0_2/analysis.md
- /home/mahi17/Github/fintech/.agents/teamwork_preview_explorer_m0_3/analysis.md

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Tasks to implement:
1. F12 Webhook Receiver (`webhook_receiver/cmd/server/main.go`):
   - Add `"subscription.charged.failed"` to `SupportedWebhookEvents` map so incoming subscription failure webhooks set `is_supported = true`.
2. F12 & F13 Failure Detector (`failure_detector/src/classifier.py`):
   - Update `ERROR_CODE_MAP` to map `"SUBSCRIPTION.CHARGED.FAILED"` and `"SUBSCRIPTION_CHARGED_FAILED"` to Category `SUBSCRIPTION_FAILED` (10) and Suggested Actions `RETRY_SUBSCRIPTION` (8) / `UPDATE_CARD_LINK` (9).
   - Map e-Mandate error codes (`mandate_expired`, `debit_rejected`, `mandate_not_active`, `insufficient_balance_mandate`) to Category `MANDATE_FAILED` (11) and Suggested Action `RENEW_MANDATE` (10).
   - Update LLM prompt string in `classifier.py` to include `SUBSCRIPTION_FAILED`, `MANDATE_FAILED`, `CHECKOUT_ABANDONED` and suggestions `RETRY_SUBSCRIPTION`, `UPDATE_CARD_LINK`, `RENEW_MANDATE`.
3. F12 Recovery Orchestrator (`recovery_orchestrator/src/main.py`):
   - Update `WorkflowRequest` model to accept `subscription_id`, `mandate_id`, `failure_category`, `suggested_action`.
   - Add handler logic for `ACTION_RETRY_SUBSCRIPTION` and `ACTION_SEND_CARD_UPDATE_LINK`.
4. F21 Live Cash Position Forecast & gRPC Integration:
   - Update `revenueiq_dev_kit/proto/mongo_service/mongodb_service.proto` to include a gRPC RPC service definition `GetSettlementAggregations` (or `GetSettlementSummary`). Regenerate Python & Go proto bindings if needed.
   - Implement `GetSettlementAggregations` in `mongodb_service` (Go) querying historical settlement records from MongoDB collection `settlements` in `revenueiq_db`.
   - Upgrade `forecast_cash_position` in `ai_gateway/src/forecast.py` to initialize a gRPC client to `mongodb_service`, query historical settlement data, and dynamically compute N-day forward cash position forecasts.
5. B2B Receivables / Promise-to-Pay:
   - Ensure `/api/v1/promises` and `PromisesTab.tsx` are fully connected to MongoDB collection `promises`.
6. Compile and Verify:
   - Run compilation and tests for Go services (`webhook_receiver`, `mongodb_service`, `recovery_orchestrator` if Go).
   - Run compilation/tests for Python microservices (`ai_gateway`, `failure_detector`, `recovery_orchestrator`).
7. Output Requirements:
   - Create `/home/mahi17/Github/fintech/.agents/teamwork_preview_worker_m2_1/progress.md`.
   - Write `/home/mahi17/Github/fintech/.agents/teamwork_preview_worker_m2_1/handoff.md` with modified files list, build/test results, and verification proof.
   - Send completion message to main orchestrator.
