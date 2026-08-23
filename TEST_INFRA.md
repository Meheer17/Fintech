# RevenueIQ End-to-End (E2E) Test Infrastructure & Methodology

## Overview
RevenueIQ is an AI Revenue Recovery & Finance Control microservices platform for Razorpay. This document describes the opaque-box, requirement-driven 4-tier E2E testing infrastructure created under Milestone M3.

The E2E test suite validates real application logic across all microservices (AI Gateway, Failure Detector, Recovery Orchestrator, Reconciliation Engine, Dashboard API, Webhook Receiver, and MongoDB Service) without simulated test stubs or hardcoded result shortcuts.

---

## 4-Tier Testing Methodology

```
+-------------------------------------------------------------------------+
|                  TIER 4: REAL-WORLD APPLICATION SCENARIOS              |
|  (End-to-End Integration Workflows: Dunning Lifecycle, e-Mandate, etc.) |
+-------------------------------------------------------------------------+
                                    ^
                                    |
+-------------------------------------------------------------------------+
|                TIER 3: CROSS-FEATURE PAIRWISE COMBINATIONS              |
|  (Inter-service interactions: Failure -> Recovery -> Audit, Cash -> API)|
+-------------------------------------------------------------------------+
                                    ^
                                    |
+-------------------------------------------------------------------------+
|                TIER 2: BOUNDARY & CORNER CASE VALIDATION                |
|  (Empty inputs, negative amounts, extreme dates, malformed webhooks)    |
+-------------------------------------------------------------------------+
                                    ^
                                    |
+-------------------------------------------------------------------------+
|                 TIER 1: FEATURE COVERAGE (HAPPY PATHS)                  |
|  (>=5 test cases per feature across all 7 platform capability areas)    |
+-------------------------------------------------------------------------+
```

### Tier 1: Feature Coverage (Opaque-Box Requirement Testing)
Validates core happy-path functionality across all 7 primary features (>=5 test cases per feature, 35 tests total):
1. **Razorpay Payment Link Creation**: Name/email/phone/amount validation, paise conversion, +91 phone formatting, direct tool execution, and AI chat trigger.
2. **AI Failure Diagnosis**: Rule engine & LLM classification of Razorpay failure codes (`SUBSCRIPTION_CHARGED_FAILED`, `MANDATE_EXPIRED`, `CARD_EXPIRED`, `BAD_REQUEST_PAYMENT_FAILED`, `GATEWAY_ERROR`).
3. **Recovery Workflow Orchestration**: Max retry guardrails, cost ratio caps (<=20%), automated payment link generation, and human escalation.
4. **Settlement Reconciliation**: 3-way matcher algorithm across Orders, Payments, and Settlements (`EXACT_MATCH`, `FUZZY_MATCH` within 2% fee tolerance / ±₹5.00, `UNMATCHED`).
5. **Cash Position Forecasting**: N-day forward cash position forecasts (7, 14, 30 days) with 15% pending recovery boost calculations.
6. **Audit Trail Logging**: Structured audit entries, entity filtering, limit parameter capping, and recovery guardrail inspection.
7. **B2B Receivables & Promise-to-Pay**: Commitment tracking, schema validation, status transitions (`PENDING`, `FULFILLED`, `BROKEN`), and active receivable aggregations.

### Tier 2: Boundary & Corner Cases
Validates system resilience against invalid inputs and boundary conditions (>=5 test cases per feature, 35 tests total):
- Empty strings, zero/negative amounts, missing email/phone, and placeholder value rejection.
- Empty error codes, unmapped failure codes, whitespace stripping, XSS/SQL injection text in error descriptions.
- Retries at boundary (2 vs 3), cost ratio at boundary (20.0% vs 20.1%), zero payment amount cost cap check.
- Settlement fee tolerance boundary (500 vs 501 paise difference), empty batch lists, missing dictionary keys.
- Forecast with zero days, negative days, extreme 365-day horizons, and zero daily settlement averages.
- Audit trail zero limit, negative limit, nonexistent entity ID filtering, and large limit boundaries.
- Webhook security HMAC SHA256 signature verification, tampered signature rejection, empty secret test mode, unsupported event rejection, and 10 supported events inventory verification.

### Tier 3: Cross-Feature Pairwise Combinations
Validates inter-service interactions across microservice boundaries (10 pairwise tests):
1. Failure Classifier -> Recovery Orchestrator -> Audit Trail Logging
2. Settlement Event -> gRPC Cash Position Forecast Calculation
3. Webhook Ingestion (`payment.failed`) -> Failure Detector Diagnosis API
4. Webhook Ingestion (`subscription.charged.failed`) -> Classifier `SUBSCRIPTION_FAILED` -> Recovery Retry Workflow
5. Webhook Ingestion (`MANDATE_EXPIRED`) -> Classifier `MANDATE_FAILED` -> Mandate Renewal Workflow
6. Recovery Payment Link Completion -> Reconciliation 3-Way Matcher
7. Dispute Event Creation -> Dashboard At-Risk Revenue Overview Metrics
8. B2B Promise Recording -> Cash Position Forecast Boost Integration
9. AI Copilot Diagnosis Tool -> Payment Link Tool Chain
10. Reconciliation Batch Execution -> Dashboard Summary Breakdown Report

### Tier 4: Real-World Application Integration Scenarios
Validates complete production workflows (6 end-to-end integration scenarios):
1. **Complete Dunning Lifecycle Workflow**: Webhook `subscription.charged.failed` -> AI Classifier -> Guardrails check -> Razorpay payment link -> Audit log entry.
2. **Mandate Failure & AutoPay Renewal Workflow**: e-Mandate expiry -> Classifier diagnosis -> Contact limit guardrail check -> Renewal workflow dispatch.
3. **Three-Way Settlement Reconciliation Workflow**: Multi-transaction batch processing (Exact, Fuzzy fee deduction, Unmatched discrepancy) -> Summary report & match rate calculation.
4. **B2B Receivables & Cash Forecast Workflow**: Commitment promise ingestion -> Historical settlement query -> 7/30-day forward cash position with recovery boost.
5. **Interactive AI Copilot Payment Link Generation Workflow**: Incomplete prompt validation gate -> Detail collection -> Razorpay payment link creation & markdown response.
6. **Webhook Security & Burst Resilience Workflow**: Ingestion of 10 active Razorpay event types with HMAC SHA256 signature verification.

---

## Feature Inventory & Test Coverage Matrix

| Feature ID | Feature Name | Tier 1 (Happy) | Tier 2 (Boundary) | Tier 3 (Cross) | Tier 4 (E2E) | Total Tests |
|------------|--------------|----------------|-------------------|----------------|--------------|-------------|
| F1 | Razorpay Payment Link Creation | 5 | 5 | 2 | 1 | 13 |
| F2 | AI Failure Diagnosis | 5 | 5 | 2 | 1 | 13 |
| F3 | Recovery Workflow Orchestration | 5 | 5 | 2 | 1 | 13 |
| F4 | Settlement Reconciliation | 5 | 5 | 2 | 1 | 13 |
| F5 | Cash Position Forecasting | 5 | 5 | 2 | 1 | 13 |
| F6 | Audit Trail Logging | 5 | 5 | 1 | 1 | 12 |
| F7 | B2B Promises & Receivables | 5 | 5 | 1 | 1 | 12 |
| **TOTAL** | **All Capabilities** | **35** | **35** | **10** | **6** | **86** |

---

## Running the E2E Test Suite

### Command
```bash
python3 run_e2e_tests.py
```
Or via pytest:
```bash
pytest e2e_tests/ -v
```

### Expected Output & Exit Code
The test runner script outputs per-tier pass counts, total test count, execution time, and returns exit code `0` when all tests pass.
