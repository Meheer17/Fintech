# RecoverIQ — Feature Coverage & Gap Analysis

## Track 03: AI Revenue Recovery — Feature Mapping

### Problem Statement Requirements

> *"Build an agent that detects revenue at risk, determines the right intervention, and executes a bounded recovery workflow"*

| Requirement | Covered? | Where in RecoverIQ |
|---|---|---|
| Detects revenue at risk | ✅ | `failure-detector` consumes webhooks, classifies failures, calculates ₹ at risk |
| Determines the right intervention | ✅ | `recovery-orchestrator` → `strategy_agent` selects action based on diagnosis |
| Executes a bounded recovery workflow | ✅ | `recovery-orchestrator` → `GraphBuilder` workflow with guardrails |

### Example Directions

| Direction | Covered? | Implementation Details |
|---|---|---|
| **Payment degradation → root cause → recovery action** | ✅ Full | **Core flow.** `webhook-receiver` → `failure-detector` (diagnosis_agent classifies root cause) → `recovery-orchestrator` (strategy_agent picks action → executor agents run it). End-to-end pipeline. |
| **Checkout drop-off recovery** | ⚠️ Partial | We handle `payment.failed` events (payment-level drop-off). **Not covered**: pre-payment cart abandonment (requires session tracking, which Razorpay webhooks don't provide). We could add a `checkout.abandoned` synthetic event type for demo purposes. |
| **Failed-subscription recovery** | ⚠️ Can Add | Razorpay has Subscriptions API. We can treat `subscription.charged.failed` webhook the same as `payment.failed` — the failure-detector can handle it with a new category `SUBSCRIPTION_FAILED` and the recovery-orchestrator can retry the subscription charge or send a payment link to update card. **Add this.** |
| **B2B receivables chaser** | ❌ Not covered | Requires Razorpay Invoices API and a follow-up sequence engine. Different domain — overdue invoices vs failed payments. Could add as a future module but not core to our demo. |
| **Mandate retry sequencer** | ⚠️ Can Add | Razorpay has e-Mandate/AutoPay APIs. Failed mandate debits come as `payment.failed` with specific error codes. Our pipeline already handles this — we just need to map mandate-specific error codes in the classifier. **Partially covered.** |
| **Hinglish voice recovery** | ❌ Not covered | Requires telephony integration (Exotel/Twilio) + multilingual TTS. Out of scope for this build but mentionable as a future direction. |
| **Promise-to-pay tracker** | ❌ Not covered | Requires a scheduling + commitment tracking system. Out of scope. |

### Evaluation Bar

> *"Don't just identify the problem. Show measured money recovered across a batch, with compliant escalation, stopping rules, and an audit trail."*

| Criterion | Covered? | Where |
|---|---|---|
| **Measured money recovered across a batch** | ✅ | `RecoveryStats` gRPC response: `total_at_risk`, `total_recovered`, `recovery_rate`. Dashboard overview shows ₹ figures. Synthetic data has 60 failures = measurable batch. |
| **Compliant escalation** | ✅ | `escalation_agent` creates tickets. Guardrails enforce: max retries, contact limits, cost caps, time windows. DND check before every contact. |
| **Stopping rules** | ✅ | `RecoveryGuardrails` class: `MAX_RETRIES=3`, `MAX_CONTACTS_PER_DAY=2`, `MAX_RECOVERY_COST_PERCENT=20%`, `MAX_WORKFLOW_AGE_DAYS=7`, circuit breaker at 5 failures/min. |
| **Audit trail** | ✅ | `audit-service` logs every action with: service, action, entity, actor, input, output, **reasoning**, guardrails checked, duration, status. Full trace per payment. |

---

## Track 04: AI Finance Controller — Feature Mapping

### Problem Statement Requirements

> *"Build an agent that closes one finance-ops loop across a 50+ record batch of synthetic data, reporting its match rate and the exceptions it could not resolve."*

| Requirement | Covered? | Where in RecoverIQ |
|---|---|---|
| Closes one finance-ops loop | ✅ | `reconciliation-engine`: Orders ↔ Payments ↔ Settlements three-way match |
| 50+ record batch | ✅ | Synthetic data: 100 orders + 60 failed/40 successful payments + 30 settlements |
| Reports match rate | ✅ | `ReconciliationReport.match_rate` + breakdown by match type (exact/fuzzy/AI) |
| Reports exceptions it could not resolve | ✅ | `ExceptionList` with categorized unresolved items + AI suggestion per exception |

### Example Directions

| Direction | Covered? | Implementation Details |
|---|---|---|
| **Multi-source reconciliation** | ✅ Full | **Core feature.** Three sources: Orders (merchant system) ↔ Payments (Razorpay) ↔ Settlements (bank/Razorpay). Tiered pipeline: exact match → fuzzy match (amount tolerance, date tolerance, fee-adjusted) → AI match (Strands `reconciliation_agent`) → exception. |
| **Settlement Q&A agent** | ✅ Full | `ai-gateway` master agent has tools: `run_reconciliation`, `get_audit_trail`, `get_recovery_dashboard`. Users can ask: *"Why was yesterday's settlement ₹15K less?"*, *"Which orders haven't been settled?"*, *"What's the fee breakdown for last week?"* |
| **Forward cash forecaster** | ⚠️ Can Add | We have the data (payment patterns, settlement timing). Could add a `@tool forecast_cash_position` to the `master_agent` that uses simple time-series extrapolation from historical settlement data. **Worth adding as a bonus tool.** |
| **Tax-line matcher** | ❌ Not covered | Requires GST filing data integration. Out of scope. |

### Evaluation Bar

> *"Throughput plus measured accuracy plus an honest exception list. One cherry-picked match proves nothing."*

| Criterion | Covered? | Where |
|---|---|---|
| **Throughput** | ✅ | Process 50+ records in a single reconciliation run. `ReconciliationReport.total_records`. |
| **Measured accuracy** | ✅ | `match_rate` as percentage. Breakdown: `exact_matches`, `fuzzy_matches`, `ai_matches`, `unmatched`. No cherry-picking — full batch results. |
| **Honest exception list** | ✅ | `ExceptionList` with typed exceptions: `AMOUNT_MISMATCH`, `MISSING_SETTLEMENT`, `MISSING_ORDER`, `DUPLICATE_SETTLEMENT`, `FEE_DISCREPANCY`, `TIMING_MISMATCH`. Each with amounts, IDs, and AI suggestion. |

---

## Cross-Track Feature Summary

### What We're Building (Complete Feature List)

```
┌────────────────────────────────────────────────────────────────────────────┐
│                         RecoverIQ FEATURE MAP                              │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  TRACK 03 FEATURES (Revenue Recovery)                                      │
│  ─────────────────────────────────────                                      │
│                                                                            │
│  F1. Webhook Ingestion & Validation                                        │
│      Service:  webhook-receiver (Go)                                       │
│      What:     Receive Razorpay webhooks, validate HMAC signature,         │
│                store raw event, publish to Redpanda                        │
│      Events:   payment.failed, payment.captured, order.paid,               │
│                refund.created, subscription.charged.failed                  │
│                                                                            │
│  F2. Payment Failure Classification (Rule-Based)                           │
│      Service:  failure-detector (Python)                                   │
│      What:     Map Razorpay error codes → FailureCategory enum             │
│      Covers:   30+ error codes across 9 categories                        │
│      AI:       ❌ Pure rules                                               │
│                                                                            │
│  F3. AI Root Cause Diagnosis                                               │
│      Service:  failure-detector (Python)                                   │
│      Agent:    diagnosis_agent (Strands)                                   │
│      Tools:    lookup_error_code, get_customer_payment_history,            │
│                check_bank_health, get_failure_rate_by_method               │
│      What:     Deep analysis for ambiguous failures — considers            │
│                customer history, bank health, payment method trends        │
│      AI:       ✅ LLM with confidence threshold (< 0.7 → fallback)       │
│                                                                            │
│  F4. Systemic Pattern Detection                                            │
│      Service:  failure-detector (Python)                                   │
│      Agent:    pattern_agent (Strands)                                     │
│      What:     Batch analysis to detect: bank outages, time clusters,      │
│                method-specific spikes, geographic patterns                  │
│      AI:       ✅ Statistical + LLM reasoning                             │
│                                                                            │
│  F5. Recovery Strategy Selection                                           │
│      Service:  recovery-orchestrator (Python)                              │
│      Agent:    strategy_agent (Strands)                                    │
│      Tools:    check_retry_eligibility, check_contact_window,              │
│                calculate_recovery_cost, check_customer_dnd                 │
│      What:     Decide optimal action: retry, payment link, email,          │
│                wait, escalate, or abandon                                  │
│      AI:       ✅ LLM decision with guardrail pre-checks                  │
│                                                                            │
│  F6. Payment Retry Execution                                               │
│      Service:  recovery-orchestrator (Python)                              │
│      Tool:     retry_razorpay_payment                                     │
│      What:     Retry failed payment via Razorpay API                       │
│      Safety:   Idempotency key (Redis), retry count check, circuit         │
│                breaker, exponential backoff (1s → 2s → 4s)                │
│      AI:       ❌ Pure execution                                           │
│                                                                            │
│  F7. Payment Link Generation                                               │
│      Service:  recovery-orchestrator (Python)                              │
│      Tool:     create_razorpay_payment_link                               │
│      What:     Create Razorpay Payment Link for alternate payment          │
│                method when original method failed                          │
│      AI:       ❌ Pure execution                                           │
│                                                                            │
│  F8. Recovery Email Notifications                                          │
│      Services: recovery-orchestrator (Python) → notification-worker (Go)   │
│      Flow:     Orchestrator publishes to Redpanda NOTIFICATIONS topic →    │
│                Worker consumes → Sends via AWS SES (or mock)               │
│      Templates: payment_failed, retry_reminder, payment_link,              │
│                 recovery_success                                           │
│      AI:       ❌ Templated emails                                         │
│                                                                            │
│  F9. Escalation Management                                                 │
│      Service:  recovery-orchestrator (Python)                              │
│      Agent:    escalation_agent (Strands)                                  │
│      Tool:     create_escalation_ticket                                   │
│      What:     Flag unrecoverable cases for human review with              │
│                severity and full context                                   │
│      AI:       ✅ Determines severity and explanation                      │
│                                                                            │
│  F10. Recovery Workflow Orchestration (GraphBuilder)                        │
│      Service:  recovery-orchestrator (Python)                              │
│      What:     Strands GraphBuilder DAG: strategy → retry/link/email →     │
│                success/failure → escalate. Full state machine with         │
│                persistent workflow state in MongoDB.                       │
│      AI:       ✅ Multi-agent graph                                       │
│                                                                            │
│  F11. Guardrails & Compliance                                              │
│      Service:  recovery-orchestrator (Python)                              │
│      What:     Hard-coded limits (NOT AI-controlled):                      │
│                • Max 3 retries per payment                                 │
│                • Max 2 contacts per customer per day                       │
│                • No contact 9PM–9AM IST                                    │
│                • Stop if cost > 20% of payment amount                      │
│                • Circuit breaker: 5 failures/min → 5min cooldown          │
│                • Abandon after 7 days                                      │
│      AI:       ❌ Hard rules, never overridable                            │
│                                                                            │
│  F12. Subscription Failure Recovery (NEW)                                   │
│      Service:  failure-detector + recovery-orchestrator                     │
│      What:     Handle subscription.charged.failed webhooks.                │
│                New category SUBSCRIPTION_FAILED. Recovery: retry           │
│                subscription charge or send update-card payment link.       │
│      Razorpay: Subscriptions API                                           │
│                                                                            │
│  F13. Mandate/AutoPay Retry Handling (NEW)                                 │
│      Service:  failure-detector                                            │
│      What:     Map e-Mandate/NACH-specific error codes                     │
│                (mandate_expired, debit_rejected, etc.) to                  │
│                appropriate recovery actions.                               │
│      Razorpay: e-Mandate/AutoPay webhook error codes                       │
│                                                                            │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  TRACK 04 FEATURES (Finance Controller)                                    │
│  ──────────────────────────────────────                                     │
│                                                                            │
│  F14. Multi-Source Data Ingestion                                           │
│      Service:  reconciliation-engine (Python)                              │
│      What:     Ingest from 3 sources via mongo-service gRPC:               │
│                • Orders (merchant's order system)                          │
│                • Payments (Razorpay payments)                              │
│                • Settlements (Razorpay settlements / bank statements)      │
│      AI:       ❌ Pure data fetching                                       │
│                                                                            │
│  F15. Exact Matching                                                       │
│      Service:  reconciliation-engine (Python)                              │
│      What:     Match by order_id + payment_id + exact amount.              │
│                Also match by settlement UTR + payment_id.                  │
│      AI:       ❌ SQL/query-based                                          │
│                                                                            │
│  F16. Fuzzy Matching                                                       │
│      Service:  reconciliation-engine (Python)                              │
│      What:     Tolerance matching:                                         │
│                • Amount: ±₹5 (processing fees, rounding)                  │
│                • Date: ±2 days (settlement delay)                         │
│                • Fee-adjusted: subtract 2% card fee + 18% GST             │
│      AI:       ❌ Algorithmic                                              │
│                                                                            │
│  F17. AI Exception Resolution                                              │
│      Service:  reconciliation-engine (Python)                              │
│      Agent:    reconciliation_agent (Strands)                              │
│      Tools:    search_orders_by_amount, search_payments_by_settlement,     │
│                check_fee_schedule, flag_as_exception                       │
│      What:     Resolve complex mismatches: split payments, partial          │
│                refunds, multi-day settlements. If truly unresolvable,      │
│                flag as exception with explanation.                         │
│      AI:       ✅ LLM reasoning for complex financial cases               │
│                                                                            │
│  F18. Reconciliation Reporting                                             │
│      Service:  reconciliation-engine (Python)                              │
│      What:     Generate report with:                                       │
│                • Total match rate (% and absolute)                         │
│                • Breakdown: exact / fuzzy / AI / unmatched                 │
│                • Total expected vs total settled                           │
│                • Discrepancy amount                                        │
│      AI:       ❌ Aggregation                                              │
│                                                                            │
│  F19. Exception Management                                                 │
│      Service:  reconciliation-engine (Python) + dashboard-api (Go)         │
│      What:     Categorized exception list with types:                      │
│                AMOUNT_MISMATCH, MISSING_SETTLEMENT, MISSING_ORDER,         │
│                DUPLICATE_SETTLEMENT, FEE_DISCREPANCY, TIMING_MISMATCH.     │
│                Each with AI suggestion + manual resolve action.            │
│                                                                            │
│  F20. Settlement Q&A                                                       │
│      Service:  ai-gateway (Python)                                         │
│      Agent:    master_agent (Strands)                                      │
│      What:     Natural language queries:                                   │
│                "Why was yesterday's settlement short?"                      │
│                "Which orders from last week aren't settled?"                │
│                "What's average settlement delay by method?"                │
│      AI:       ✅ LLM + tool calling to reconciliation-engine              │
│                                                                            │
│  F21. Cash Position Forecast (NEW — Bonus)                                 │
│      Service:  ai-gateway (Python)                                         │
│      Tool:     forecast_cash_position (NEW)                               │
│      What:     Predict cash position for next 7/14/30 days based on        │
│                historical settlement patterns and pending payments.        │
│                Simple time-series extrapolation, not deep ML.              │
│      AI:       ✅ LLM interprets forecast data                            │
│                                                                            │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  CROSS-TRACK FEATURES                                                      │
│  ────────────────────                                                       │
│                                                                            │
│  F22. Full Audit Trail                                                     │
│      Service:  audit-service (Go)                                          │
│      Comm:     gRPC ONLY (all services → audit-service via gRPC)           │
│      What:     Every action across every service logged with:              │
│                service, action, entity, actor, input, output,              │
│                reasoning, guardrails_checked, duration, status             │
│      Storage:  MongoDB via mongo-service gRPC                              │
│      Cache:    Redis via redis-service gRPC (recent entries)               │
│                                                                            │
│  F23. AI Chat Interface                                                    │
│      Service:  ai-gateway (Python) + frontend (TypeScript)                 │
│      What:     Conversational interface for merchants to ask about          │
│                failures, recoveries, reconciliation, audit data.           │
│                Streaming responses via SSE.                                │
│      AI:       ✅ master_agent with tool calling                           │
│                                                                            │
│  F24. Real-Time Dashboard                                                  │
│      Service:  dashboard-api (Go) + frontend (TypeScript)                  │
│      What:     Overview metrics, failure lists, recovery workflows,         │
│                reconciliation results, audit trail — all via REST.         │
│      AI:       ❌ Pure data display                                        │
│                                                                            │
│  F25. Synthetic Data Engine                                                │
│      Location: data/ (Python scripts)                                      │
│      What:     Generate 100+ realistic records with deliberate             │
│                failure distributions and settlement mismatches.            │
│                Razorpay webhook simulator for end-to-end testing.          │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
```

### What We're NOT Building (and Why)

| Feature | Track | Why Excluded |
|---|---|---|
| Checkout drop-off (pre-payment) | 03 | Requires session tracking — Razorpay webhooks only fire after payment attempt. We cover payment-level failures which is the closest analog. |
| B2B receivables chaser | 03 | Different domain (invoices, not payments). Would need Razorpay Invoices API + a follow-up sequence engine. Good v2 feature. |
| Hinglish voice recovery | 03 | Requires telephony (Exotel/Twilio) + multilingual TTS. Out of scope, but mentionable as future direction. |
| Promise-to-pay tracker | 03 | Requires commitment scheduling system. Out of scope. |
| Tax-line matcher | 04 | Requires GST filing data integration. Out of scope. |

> [!TIP]
> **We cover 8 of 11 example directions** across both tracks (73%). The 3 excluded ones are either different domains (B2B receivables, tax matching) or require external telephony. This is strong coverage — judges won't expect all directions.

---

## Audit Service Communication — Confirmed gRPC Only

> [!IMPORTANT]
> **All communication with audit-service is via gRPC.** No REST, no direct MongoDB access.

```
                      gRPC :50007
┌──────────────────┐ ──────────────▶ ┌──────────────────┐
│ webhook-receiver │   LogAction()   │                  │     gRPC :50010
│ (Go)             │ ◀────────────── │   audit-service  │ ──────────────▶ ┌──────────────┐
└──────────────────┘   AuditId       │   (Go)           │                 │ mongo-service │
                                     │                  │ ◀────────────── │ (Go)          │
┌──────────────────┐ ──────────────▶ │   gRPC Server:   │                 └──────────────┘
│ failure-detector │   LogAction()   │   • LogAction    │
│ (Python)         │ ◀────────────── │   • GetAuditTrail│     gRPC :50011
└──────────────────┘                 │   • GetActions   │ ──────────────▶ ┌──────────────┐
                                     │     ByEntity     │                 │ redis-service │
┌──────────────────┐ ──────────────▶ │                  │ ◀────────────── │ (Go)          │
│ recovery-orch.   │   LogAction()   │   Also consumes  │                 └──────────────┘
│ (Python)         │ ◀────────────── │   AUDIT_EVENTS   │
└──────────────────┘                 │   from Redpanda  │
                                     │   (batch insert  │
┌──────────────────┐ ──────────────▶ │    optimization) │
│ reconciliation   │   LogAction()   │                  │
│ (Python)         │ ◀────────────── └──────────────────┘
└──────────────────┘                          ▲
                                              │ gRPC
┌──────────────────┐ ──────────────▶          │
│ notification-wkr │   LogAction()   ─────────┘
│ (Go)             │
└──────────────────┘

┌──────────────────┐ ──────────────▶ audit-service
│ dashboard-api    │  GetAuditTrail()
│ (Go)             │  GetActionsByEntity()
└──────────────────┘

┌──────────────────┐ ──────────────▶ audit-service
│ ai-gateway       │  GetAuditTrail()
│ (Python)         │  GetActionsByEntity()
└──────────────────┘
```

**Two ingestion paths (both ultimately gRPC):**

1. **Synchronous gRPC** (primary): Services call `LogAction()` directly for critical actions that need immediate confirmation (payment retries, escalations, guardrail blocks).

2. **Async via Redpanda** (optimization): High-volume, lower-priority events (tool calls, cache hits, health checks) can be published to `AUDIT_EVENTS` topic. The audit-service has a background Redpanda consumer that batch-inserts these via mongo-service gRPC. **This is still gRPC at the storage layer** — the consumer calls mongo-service gRPC `BulkInsert`, not direct MongoDB.

---

## Updated Frontend Spec — Minimalist Light Theme

### Design Direction Change

~~Dark mode, glassmorphism, vibrant glows~~ → **Clean, minimalist, light theme**

### Design System

```css
/* ── Light Minimalist Theme ── */

:root {
  /* Backgrounds */
  --bg-primary: #ffffff;
  --bg-secondary: #f8f9fa;
  --bg-tertiary: #f1f3f5;
  --bg-card: #ffffff;

  /* Borders */
  --border-light: #e9ecef;
  --border-medium: #dee2e6;
  
  /* Text */
  --text-primary: #212529;
  --text-secondary: #495057;
  --text-tertiary: #868e96;
  --text-muted: #adb5bd;

  /* Accent Colors (muted, not vibrant) */
  --accent-primary: #4263eb;      /* Indigo — primary actions */
  --accent-success: #2b8a3e;      /* Green — recovered/matched */
  --accent-warning: #e67700;      /* Amber — in progress/pending */
  --accent-danger: #c92a2a;       /* Red — failed/at risk */
  --accent-info: #1971c2;         /* Blue — informational */

  /* Surfaces */
  --card-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06);
  --card-shadow-hover: 0 4px 6px rgba(0, 0, 0, 0.05), 0 2px 4px rgba(0, 0, 0, 0.06);
  --card-radius: 8px;

  /* Typography */
  --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', monospace;
}
```

### Visual Principles

| Principle | Implementation |
|---|---|
| **Whitespace > decoration** | Generous padding, no gradients or background textures |
| **Subtle borders** | 1px `#e9ecef` borders instead of shadows or glows |
| **Muted color accents** | Colors only for status indicators and CTAs, not backgrounds |
| **Clean typography** | Inter font, clear hierarchy (16px body, 14px secondary, 24px headings) |
| **Flat cards** | White cards on `#f8f9fa` background with thin shadow, no rounded glass |
| **No animations** on load | Only subtle hover transitions (opacity, translateY -1px) |
| **Data-dense tables** | Compact rows, alternating row tint (`#f8f9fa`), clear column headers |
| **Status pills** | Small rounded badges with soft background tints (not bold fills) |

### Component Styles

#### MetricsCard (minimalist)
```
┌──────────────────────────┐
│  Revenue at Risk         │  ← text-tertiary, 12px uppercase
│  ₹2,34,500               │  ← text-primary, 28px semibold
│  ↑ 12% from last week   │  ← accent-danger, 12px
└──────────────────────────┘
   White card, thin border, subtle shadow
```

#### StatusBadge (soft tints)
```
  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐
  │ ● Recovered │  │ ● Pending   │  │ ● In Progress│  │ ● Failed    │
  │   #2b8a3e   │  │   #e67700   │  │   #1971c2    │  │   #c92a2a   │
  │   bg: #ebfbee│  │   bg: #fff9db│  │   bg: #e7f5ff│  │   bg: #fff5f5│
  └─────────────┘  └─────────────┘  └──────────────┘  └─────────────┘
```

#### DataTable (clean)
```
┌───────────────┬──────────┬────────┬───────────────┬───────────┐
│ Payment ID    │ Amount   │ Method │ Category      │ Status    │
├───────────────┼──────────┼────────┼───────────────┼───────────┤
│ pay_xyz123    │ ₹1,500   │ UPI    │ Bank Decline  │ ● Pending │  ← white row
├───────────────┼──────────┼────────┼───────────────┼───────────┤
│ pay_abc456    │ ₹3,200   │ Card   │ Card Expired  │ ● Recover │  ← #f8f9fa row
├───────────────┼──────────┼────────┼───────────────┼───────────┤
│ pay_def789    │ ₹850     │ UPI    │ Insuf. Funds  │ ● Failed  │  ← white row
└───────────────┴──────────┴────────┴───────────────┴───────────┘
   No bold header backgrounds, just semibold text + bottom border
```

#### Sidebar (minimal)
```
┌─────────────────────┐
│  RecoverIQ          │  ← 18px semibold, accent-primary
│                     │
│  ○ Overview         │  ← Selected: accent-primary text + left border
│  ○ Failures         │  ← Unselected: text-secondary
│  ○ Recoveries       │
│  ○ Reconciliation   │
│  ○ Audit Trail      │
│  ○ Chat             │
│                     │
│  ─────────────      │
│  ○ Settings         │
└─────────────────────┘
   White bg, no icons (or minimal Lucide icons), thin right border
```

### Page Updates (Light Theme)

| Page | Key Changes from v1 |
|---|---|
| **Overview** | White metric cards on light grey bg. Recharts with muted line colors. No glows. Clean activity feed with timestamps. |
| **Failures** | White filter bar with dropdowns. Clean table with alternating rows. Subtle status pills. |
| **Failure Detail** | Clean card layout. Timeline uses thin vertical line with dot markers (not fancy). |
| **Recoveries** | Tab bar with underline indicator (not filled tabs). Simple workflow status cards. |
| **Reconciliation** | Segmented bar chart for match rate (muted greens/blues). Clean exception table. |
| **Audit Trail** | Compact log entries. Monospace for IDs/timestamps. Expandable accordion for details. |
| **Chat** | Clean message bubbles (white user, light grey agent). No fancy tool-call animations — just a "🔧 Using tool: analyze_payment_failure" text indicator. |
| **Settings** | Simple form layout. Guardrail values in a clean table. |

---

## New Tasks to Add (from gap analysis)

### F12: Subscription Failure Recovery
- [ ] Add `SUBSCRIPTION_FAILED` to `FailureCategory` enum in `failure.proto`
- [ ] Add subscription error codes to `pkg/constants/error_codes.go` and `shared/constants.py`
- [ ] Add `subscription.charged.failed` to webhook-receiver event type handling
- [ ] Add subscription-specific diagnosis logic in failure-detector classifier
- [ ] Add `@tool retry_subscription_charge` in recovery-orchestrator tools
- [ ] Add `@tool create_card_update_link` for expired card on subscription
- [ ] Add 5 synthetic subscription failure records to data generator

### F13: Mandate/AutoPay Error Codes
- [ ] Add mandate-specific error codes to classifier: `mandate_expired`, `debit_rejected`, `mandate_not_active`, `insufficient_balance_mandate`
- [ ] Map to existing recovery actions (retry, payment link, escalate)

### F21: Cash Position Forecast (Bonus)
- [ ] Add `@tool forecast_cash_position` to ai-gateway master_agent
- [ ] Implement simple forecast: average daily settlement × N days + pending payments
- [ ] Query historical settlement data from mongo-service via gRPC
- [ ] Return structured forecast: { day, expected_inflow, pending_recovery, net_position }

### Frontend Light Theme Updates
- [ ] Replace all dark theme CSS variables with light theme values
- [ ] Remove glassmorphism, backdrop blur, glow effects
- [ ] Use white cards with thin borders and subtle shadows
- [ ] Use muted status badge colors with soft background tints
- [ ] Use clean alternating-row tables
- [ ] Minimal sidebar with text links (optional small Lucide icons)
- [ ] Subtle hover transitions only (no load animations)

---

## Updated Total Task Count

| Phase | Tasks |
|---|---|
| Phase 0: Scaffolding | 8 |
| Phase 1: Protobufs | 12 |
| Phase 2: Shared Packages | 14 |
| Phase 3: Go Services | 57 |
| Phase 4: Python AI Services | 48 |
| Phase 5: Synthetic Data | 6 |
| Phase 6: Frontend | 28 |
| Phase 7: Integration Testing | 14 |
| Phase 8: Failure Testing | 8 |
| Phase 9: Polish | 10 |
| **NEW: Subscription/Mandate** | **8** |
| **NEW: Cash Forecast** | **3** |
| **NEW: Light Theme Updates** | **7** |
| **TOTAL** | **223** |
