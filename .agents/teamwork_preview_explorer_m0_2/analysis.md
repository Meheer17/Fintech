# RevenueIQ Explorer 2: Backend Failure Classifier & Webhook Recovery Audit Report

## 1. Executive Summary
This report provides a comprehensive, line-by-line audit of backend microservices (`failure_detector`, `recovery_orchestrator`, `webhook_receiver`, `dashboard_api`) and protobuf definitions (`failure.proto`, `recovery.proto`) in the RevenueIQ codebase.

Key findings:
1. **Webhook Event Omission**: `webhook_receiver/cmd/server/main.go` is missing `subscription.charged.failed` in its `SupportedWebhookEvents` map (line 28-39).
2. **Failure Detector Rule Matcher Mismatch**: `failure_detector/src/classifier.py` defines `"SUBSCRIPTION_CHARGED_FAILED"` in `ERROR_CODE_MAP`, but Razorpay webhooks pass dot-separated strings (`subscription.charged.failed` / `SUBSCRIPTION.CHARGED.FAILED`), causing substring matching (`if key in code_upper:`) to fail and fall through to default/LLM handling.
3. **Missing e-Mandate Error Code Mappings**: `classifier.py` maps only `"MANDATE_EXPIRED"`. The e-Mandate / AutoPay failure error codes `debit_rejected`, `mandate_not_active`, and `insufficient_balance_mandate` are missing from `ERROR_CODE_MAP` and fail to map to `MANDATE_FAILED`.
4. **Recovery Orchestrator Single-Action Hardcoding**: `recovery_orchestrator/src/main.py` ignores `action_type` or failure category and hardcodes `ACTION_CREATE_PAYMENT_LINK` for all workflows. It lacks API/branching logic for `ACTION_RETRY_SUBSCRIPTION` and `ACTION_SEND_CARD_UPDATE_LINK`.

---

## 2. Webhook Ingestion Audit (`webhook_receiver`)

### File Location
- File: `/home/mahi17/Github/fintech/webhook_receiver/cmd/server/main.go`
- Lines: 28-39

### Code Observation
```go
// Supported 10 Active Razorpay Webhook Events
var SupportedWebhookEvents = map[string]bool{
	"payment.authorized":     true,
	"payment.failed":         true,
	"payment.captured":       true,
	"payment.dispute.created": true,
	"order.paid":              true,
	"subscription.pending":    true,
	"subscription.charged":    true,
	"subscription.cancelled":  true,
	"settlement.processed":    true,
	"refund.created":         true,
}
```

### Analysis & Impact
- `subscription.charged.failed` is **missing** from `SupportedWebhookEvents`.
- When Razorpay POSTs `subscription.charged.failed` to `/webhook/razorpay` or `/webhook`:
  - `isKnown := SupportedWebhookEvents[eventType]` evaluates to `false` (line 75).
  - The response returned to caller is `{"supported": false, "status": "RECEIVED"}`.
  - The Redpanda/Kafka message payload contains `"is_supported": false`.

### Proposed Fix
Add `"subscription.charged.failed": true` to `SupportedWebhookEvents` in `webhook_receiver/cmd/server/main.go`.

---

## 3. Failure Proto & Failure Detector Classification Audit (`failure.proto`, `classifier.py`, `main.py`)

### Proto Definitions Audit (`revenueiq_dev_kit/proto/failure/failure.proto`)
- **Status**: Complete in Proto, but check fields.
- **Enum `FailureCategory` (lines 17-31)**:
  - `SUBSCRIPTION_FAILED = 10;` is present.
  - `MANDATE_FAILED = 11;` is present.
  - `CHECKOUT_ABANDONED = 12;` is present.
- **Enum `RecoverySuggestion` (lines 33-47)**:
  - `RETRY_SUBSCRIPTION = 8;` is present.
  - `UPDATE_CARD_LINK = 9;` is present.
  - `RENEW_MANDATE = 10;` is present.
- **Messages `DiagnoseRequest` & `FailureRecord` (lines 49-92)**:
  - `subscription_id` is present (`field 9` in request, `field 14` in record).
  - `mandate_id` is present (`field 10` in request, `field 15` in record).

### Classifier Rule Engine Audit (`failure_detector/src/classifier.py`)
- File: `/home/mahi17/Github/fintech/failure_detector/src/classifier.py`
- Lines: 23-34

#### Existing `ERROR_CODE_MAP`:
```python
ERROR_CODE_MAP = {
    "BAD_REQUEST_PAYMENT_FAILED": ("INSUFFICIENT_FUNDS", "RETRY_SAME_METHOD"),
    "GATEWAY_ERROR": ("NETWORK_ERROR", "WAIT_AND_RETRY"),
    "CARD_EXPIRED": ("CARD_EXPIRED", "SEND_PAYMENT_LINK"),
    "AUTHENTICATION_FAILED": ("AUTHENTICATION_FAILED", "RETRY_DIFFERENT_METHOD"),
    "BANK_TECHNICAL_GLITCH": ("BANK_DECLINE", "WAIT_AND_RETRY"),
    "FRAUD_SUSPECTED": ("FRAUD_SUSPECTED", "ESCALATE_TO_HUMAN"),
    "LIMIT_EXCEEDED": ("LIMIT_EXCEEDED", "CONTACT_CUSTOMER"),
    "SUBSCRIPTION_CHARGED_FAILED": ("SUBSCRIPTION_FAILED", "RETRY_SUBSCRIPTION"),
    "MANDATE_EXPIRED": ("MANDATE_FAILED", "RENEW_MANDATE"),
    "CHECKOUT_ABANDONED": ("CHECKOUT_ABANDONED", "CHECKOUT_NUDGE"),
}
```

#### Defect 1: Dot vs. Underscore Mismatch in Subscription Webhook Error Codes
- Webhooks send event name `subscription.charged.failed` or error code `subscription.charged.failed`.
- Line 37 converts input: `code_upper = error_code.upper() if error_code else "UNKNOWN"` -> `"SUBSCRIPTION.CHARGED.FAILED"`.
- Match logic: `if key in code_upper:` where `key` is `"SUBSCRIPTION_CHARGED_FAILED"`.
- `"SUBSCRIPTION_CHARGED_FAILED" in "SUBSCRIPTION.CHARGED.FAILED"` returns `False` due to underscore vs dot differences.
- **Result**: `subscription.charged.failed` fails to match the rule dictionary.

#### Defect 2: Missing e-Mandate / AutoPay Error Codes
- Required error codes:
  - `mandate_expired`: Mapped (`"MANDATE_EXPIRED": ("MANDATE_FAILED", "RENEW_MANDATE")`).
  - `debit_rejected`: **Missing**.
  - `mandate_not_active`: **Missing**.
  - `insufficient_balance_mandate`: **Missing**.
- **Result**: e-Mandate debit failures with codes like `debit_rejected` fall through to LLM/UNKNOWN.

#### Defect 3: Incomplete LLM Prompt Enum Definitions
- In `classifier.py` (lines 51-56):
  ```python
  prompt = (
      f"Diagnose payment failure code '{error_code}' with description '{description}'. "
      f"Classify into category (INSUFFICIENT_FUNDS, BANK_DECLINE, CARD_EXPIRED, NETWORK_ERROR, AUTHENTICATION_FAILED, FRAUD_SUSPECTED) "
      f"and recovery suggestion (RETRY_SAME_METHOD, RETRY_DIFFERENT_METHOD, SEND_PAYMENT_LINK, WAIT_AND_RETRY, ESCALATE_TO_HUMAN). "
      f"Return short diagnosis."
  )
  ```
- Missing categories in prompt: `SUBSCRIPTION_FAILED`, `MANDATE_FAILED`, `CHECKOUT_ABANDONED`.
- Missing suggestions in prompt: `RETRY_SUBSCRIPTION`, `UPDATE_CARD_LINK`, `RENEW_MANDATE`, `CHECKOUT_NUDGE`.

### Proposed Fixes for `classifier.py`
Update `ERROR_CODE_MAP` and normalization logic:
```python
ERROR_CODE_MAP = {
    "BAD_REQUEST_PAYMENT_FAILED": ("INSUFFICIENT_FUNDS", "RETRY_SAME_METHOD"),
    "GATEWAY_ERROR": ("NETWORK_ERROR", "WAIT_AND_RETRY"),
    "CARD_EXPIRED": ("CARD_EXPIRED", "SEND_PAYMENT_LINK"),
    "AUTHENTICATION_FAILED": ("AUTHENTICATION_FAILED", "RETRY_DIFFERENT_METHOD"),
    "BANK_TECHNICAL_GLITCH": ("BANK_DECLINE", "WAIT_AND_RETRY"),
    "FRAUD_SUSPECTED": ("FRAUD_SUSPECTED", "ESCALATE_TO_HUMAN"),
    "LIMIT_EXCEEDED": ("LIMIT_EXCEEDED", "CONTACT_CUSTOMER"),
    "SUBSCRIPTION_CHARGED_FAILED": ("SUBSCRIPTION_FAILED", "RETRY_SUBSCRIPTION"),
    "SUBSCRIPTION.CHARGED.FAILED": ("SUBSCRIPTION_FAILED", "RETRY_SUBSCRIPTION"),
    "SUBSCRIPTION_CARD_INVALID": ("SUBSCRIPTION_FAILED", "UPDATE_CARD_LINK"),
    "MANDATE_EXPIRED": ("MANDATE_FAILED", "RENEW_MANDATE"),
    "DEBIT_REJECTED": ("MANDATE_FAILED", "RENEW_MANDATE"),
    "MANDATE_NOT_ACTIVE": ("MANDATE_FAILED", "RENEW_MANDATE"),
    "INSUFFICIENT_BALANCE_MANDATE": ("MANDATE_FAILED", "WAIT_AND_RETRY"),
    "CHECKOUT_ABANDONED": ("CHECKOUT_ABANDONED", "CHECKOUT_NUDGE"),
}
```

---

## 4. Recovery Orchestrator Audit (`recovery_orchestrator`, `recovery.proto`)

### Proto Definitions Audit (`revenueiq_dev_kit/proto/recovery/recovery.proto`)
- Lines 20-34: `enum RecoveryAction` contains:
  - `ACTION_RETRY_SUBSCRIPTION = 7;`
  - `ACTION_SEND_CARD_UPDATE_LINK = 8;`
  - `ACTION_RENEW_MANDATE = 9;`
  - `ACTION_CREATE_PAYMENT_LINK = 2;`

### Recovery Orchestrator Audit (`recovery_orchestrator/src/main.py`)
- File: `/home/mahi17/Github/fintech/recovery_orchestrator/src/main.py`
- Lines: 83-154, 158-162

#### Defect 1: Restricted Request Model (`WorkflowRequest`)
```python
class WorkflowRequest(BaseModel):
    payment_id: str
    current_retries: int = 0
    amount_paise: int = 0
    action_type: str = "AUTO"
```
- Missing parameters: `subscription_id`, `mandate_id`, `failure_category`, `suggested_action`.

#### Defect 2: Hardcoded Single Action (`ACTION_CREATE_PAYMENT_LINK`)
- Lines 112-127 in `process_workflow`:
  Regardless of `action_type` or `suggested_action` or failure category, `process_workflow` invokes `create_real_razorpay_payment_link()` and sets `"action": "ACTION_CREATE_PAYMENT_LINK"`.

#### Defect 3: Missing Support for Subscription Retry and Update-Card Links
- No implementation or routing for:
  - `ACTION_RETRY_SUBSCRIPTION`: Retrying a failed subscription charge (calling Razorpay subscription charge retry API `POST /v1/subscriptions/{id}/charge` or creating a retry action).
  - `ACTION_SEND_CARD_UPDATE_LINK`: Creating or returning a hosted card-update / payment-method update link for subscriptions.
  - `ACTION_RENEW_MANDATE`: Initiating mandate re-registration flow.

---

## 5. Summary Table of Audit Findings

| Component | Target File | Current State | Defect / Missing Requirement | Impact |
|---|---|---|---|---|
| Webhook Receiver | `webhook_receiver/cmd/server/main.go` | 10 events supported | `subscription.charged.failed` missing from `SupportedWebhookEvents` | Incoming subscription failure webhooks flagged as unsupported (`is_supported = false`) |
| Failure Classifier | `failure_detector/src/classifier.py` | Key `"SUBSCRIPTION_CHARGED_FAILED"` | String format mismatch (`.` vs `_`) for `subscription.charged.failed` | Webhook error code fails rule match, falls back to UNKNOWN/LLM |
| Failure Classifier | `failure_detector/src/classifier.py` | Only `"MANDATE_EXPIRED"` mapped | Missing `debit_rejected`, `mandate_not_active`, `insufficient_balance_mandate` | e-Mandate error codes fail rule match, fall back to UNKNOWN |
| Failure Classifier | `failure_detector/src/classifier.py` | LLM prompt contains partial enums | Prompt missing `SUBSCRIPTION_FAILED`, `MANDATE_FAILED`, `RETRY_SUBSCRIPTION`, etc. | LLM fallback cannot output new failure categories |
| Recovery Orchestrator | `recovery_orchestrator/src/main.py` | Hardcoded `ACTION_CREATE_PAYMENT_LINK` | No branching for `ACTION_RETRY_SUBSCRIPTION` or `ACTION_SEND_CARD_UPDATE_LINK` | All recoveries attempt payment link creation regardless of subscription failure type |
| Recovery Orchestrator | `recovery_orchestrator/src/main.py` | `WorkflowRequest` model limited | Missing `subscription_id`, `mandate_id`, `suggested_action` | Cannot pass subscription/mandate context into recovery workflow |

---

## 6. Verification Methods & Inspection Steps

1. **Verify Webhook Ingestion**:
   - Inspect `webhook_receiver/cmd/server/main.go` line 28-39.
   - Verify `SupportedWebhookEvents["subscription.charged.failed"]`.

2. **Verify Classifier Rule Engine**:
   - Inspect `failure_detector/src/classifier.py` lines 23-34 and line 39 (`if key in code_upper`).
   - Pass `"subscription.charged.failed"`, `"debit_rejected"`, `"mandate_not_active"`, `"insufficient_balance_mandate"` to `classify_failure()`.
   - Confirm returned `category` is `"SUBSCRIPTION_FAILED"` or `"MANDATE_FAILED"`.

3. **Verify Recovery Orchestrator**:
   - Inspect `recovery_orchestrator/src/main.py` lines 83-128.
   - Confirm `process_workflow` routes `ACTION_RETRY_SUBSCRIPTION` and `ACTION_SEND_CARD_UPDATE_LINK` properly.
