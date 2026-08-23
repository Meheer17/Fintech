# Handoff Report — Explorer 2 (Backend Failure Classifier & Webhook Recovery Audit)

## 1. Observation

### Webhook Event Handling in `webhook_receiver`
- **File**: `webhook_receiver/cmd/server/main.go` (lines 28-39)
- **Direct Code Quote**:
  ```go
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
- **Finding**: `"subscription.charged.failed"` is completely absent from `SupportedWebhookEvents`.

### Protobuf Definitions in `failure.proto` and `recovery.proto`
- **File**: `revenueiq_dev_kit/proto/failure/failure.proto` (lines 17-31, 33-47, 49-92)
- **Direct Code Quote**:
  ```protobuf
  enum FailureCategory {
    ...
    SUBSCRIPTION_FAILED = 10;
    MANDATE_FAILED = 11;
    CHECKOUT_ABANDONED = 12;
  }
  ```
- **Finding**: `failure.proto` has `SUBSCRIPTION_FAILED = 10;`, `MANDATE_FAILED = 11;`, `RETRY_SUBSCRIPTION = 8;`, `UPDATE_CARD_LINK = 9;`, `RENEW_MANDATE = 10;`, `subscription_id`, and `mandate_id`.
- **File**: `revenueiq_dev_kit/proto/recovery/recovery.proto` (lines 20-34)
- **Direct Code Quote**:
  ```protobuf
  enum RecoveryAction {
    ...
    ACTION_RETRY_SUBSCRIPTION = 7;
    ACTION_SEND_CARD_UPDATE_LINK = 8;
    ACTION_RENEW_MANDATE = 9;
  }
  ```
- **Finding**: `recovery.proto` has action enums for subscription retry and card update links.

### Failure Classification Logic in `classifier.py`
- **File**: `failure_detector/src/classifier.py` (lines 23-34, 37-46, 51-56)
- **Direct Code Quote**:
  ```python
  ERROR_CODE_MAP = {
      ...
      "SUBSCRIPTION_CHARGED_FAILED": ("SUBSCRIPTION_FAILED", "RETRY_SUBSCRIPTION"),
      "MANDATE_EXPIRED": ("MANDATE_FAILED", "RENEW_MANDATE"),
      ...
  }

  def classify_failure(error_code: str, description: str = "") -> dict:
      code_upper = error_code.upper() if error_code else "UNKNOWN"
      for key, (category, suggestion) in ERROR_CODE_MAP.items():
          if key in code_upper:
              return ...
  ```
- **Finding 1 (Dot vs Underscore)**: Key `"SUBSCRIPTION_CHARGED_FAILED"` (with underscores) does not match `"SUBSCRIPTION.CHARGED.FAILED"` (with dots) produced by uppercase normalization of `subscription.charged.failed`.
- **Finding 2 (Missing e-Mandate Codes)**: Error codes `debit_rejected`, `mandate_not_active`, and `insufficient_balance_mandate` are missing from `ERROR_CODE_MAP`.
- **Finding 3 (LLM Prompt Incompleteness)**: Prompt string (lines 53-54) only lists original categories/suggestions, omitting `SUBSCRIPTION_FAILED`, `MANDATE_FAILED`, `RETRY_SUBSCRIPTION`, `UPDATE_CARD_LINK`, `RENEW_MANDATE`.

### Recovery Workflow Logic in `recovery_orchestrator`
- **File**: `recovery_orchestrator/src/main.py` (lines 83-128, 158-162)
- **Direct Code Quote**:
  ```python
  class WorkflowRequest(BaseModel):
      payment_id: str
      current_retries: int = 0
      amount_paise: int = 0
      action_type: str = "AUTO"
  ```
- **Finding**: `process_workflow()` ignores `action_type` or failure category and unconditionally sets `"action": "ACTION_CREATE_PAYMENT_LINK"`. No handling or API calls exist for `ACTION_RETRY_SUBSCRIPTION` or `ACTION_SEND_CARD_UPDATE_LINK`.

---

## 2. Logic Chain

1. **Observation**: Razorpay webhook payloads send `event: "subscription.charged.failed"`.
   **Step 1**: In `webhook_receiver/cmd/server/main.go`, `SupportedWebhookEvents["subscription.charged.failed"]` evaluates to `false` because the key is missing in line 28-39. This causes the webhook receiver to output `is_supported: false`.
   **Step 2**: When `subscription.charged.failed` is forwarded to `failure_detector/src/classifier.py`, `classify_failure("subscription.charged.failed")` converts the code to `"SUBSCRIPTION.CHARGED.FAILED"`.
   **Step 3**: `ERROR_CODE_MAP` contains key `"SUBSCRIPTION_CHARGED_FAILED"`. Checking `"SUBSCRIPTION_CHARGED_FAILED" in "SUBSCRIPTION.CHARGED.FAILED"` evaluates to `False`. The rule engine fails and falls back to LLM/UNKNOWN.

2. **Observation**: e-Mandate AutoPay failures send error codes such as `debit_rejected`, `mandate_not_active`, or `insufficient_balance_mandate`.
   **Step 1**: `classifier.py`'s `ERROR_CODE_MAP` only contains `"MANDATE_EXPIRED"`.
   **Step 2**: When `debit_rejected`, `mandate_not_active`, or `insufficient_balance_mandate` are received, they match no key in `ERROR_CODE_MAP`.
   **Step 3**: The rule engine fails and falls back to LLM or default `UNKNOWN` classification instead of returning `MANDATE_FAILED`.

3. **Observation**: `recovery_orchestrator/src/main.py` handles workflow orchestration requests.
   **Step 1**: `WorkflowRequest` model only accepts `payment_id`, `current_retries`, `amount_paise`, and `action_type`.
   **Step 2**: `process_workflow` creates a payment link via `create_real_razorpay_payment_link()` and returns `ACTION_CREATE_PAYMENT_LINK` for every incoming request.
   **Step 3**: There is no conditional routing or implementation for `ACTION_RETRY_SUBSCRIPTION` or `ACTION_SEND_CARD_UPDATE_LINK` or `ACTION_RENEW_MANDATE`.

---

## 3. Caveats

- **External Razorpay Subscriptions API**: Real execution of Razorpay subscription retries (`POST /v1/subscriptions/{id}/retry`) or card update link generation requires valid Razorpay API keys (`RAZORPAY_KEY_ID` and `RAZORPAY_KEY_SECRET`).
- **No live code modification**: As Explorer 2 (read-only role), no source files outside `.agents/teamwork_preview_explorer_m0_2/` were modified.

---

## 4. Conclusion

1. **`webhook_receiver`**: Must add `"subscription.charged.failed": true` to `SupportedWebhookEvents` in `webhook_receiver/cmd/server/main.go`.
2. **`failure_detector/src/classifier.py`**:
   - Must normalize dots to underscores or add `"SUBSCRIPTION.CHARGED.FAILED"` to `ERROR_CODE_MAP`.
   - Must add `DEBIT_REJECTED`, `MANDATE_NOT_ACTIVE`, `INSUFFICIENT_BALANCE_MANDATE` mapping to `MANDATE_FAILED`.
   - Must update LLM fallback prompt string with new categories and suggestions.
3. **`recovery_orchestrator/src/main.py`**:
   - Must update `WorkflowRequest` Pydantic model to include `subscription_id`, `mandate_id`, `failure_category`, and `suggested_action`.
   - Must add branching in `process_workflow()` to handle `ACTION_RETRY_SUBSCRIPTION` (calling Razorpay subscription charge retry) and `ACTION_SEND_CARD_UPDATE_LINK` (generating hosted card update link).

---

## 5. Verification Method

To verify these findings independently:

1. **Inspect Webhook Events**:
   ```bash
   grep -n "SupportedWebhookEvents" webhook_receiver/cmd/server/main.go -A 15
   ```
2. **Inspect Classifier Error Code Map**:
   ```bash
   grep -n "ERROR_CODE_MAP" failure_detector/src/classifier.py -A 15
   ```
3. **Inspect Recovery Orchestrator Process Workflow**:
   ```bash
   grep -n "def process_workflow" recovery_orchestrator/src/main.py -A 40
   ```
4. **Inspect Failure Proto Enums**:
   ```bash
   grep -n "enum FailureCategory" revenueiq_dev_kit/proto/failure/failure.proto -A 20
   ```
