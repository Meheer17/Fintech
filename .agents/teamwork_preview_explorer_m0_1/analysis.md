# Comprehensive Frontend & Dynamic Endpoints Audit Report

**Target Workspace**: `/home/mahi17/Github/fintech`  
**Agent**: Explorer 1 (Frontend & Dynamic Endpoints Auditor)  
**Date**: 2026-08-23  

---

## 1. Executive Summary

This audit examined the entire frontend codebase located in `frontend/src/` alongside its backend API integrations in `dashboard_api/` (port 8005) and `ai_gateway/` (port 8006).

### Key Findings:
1. **Unused Legacy Mock File (`mockData.ts`)**: Contains `MOCK_FAILURES`, `MOCK_WORKFLOWS`, and `MOCK_AUDIT`. It is currently unused by UI components (components fetch live data from `lib/api.ts`), but exists as orphaned mock code.
2. **Static Fallback Arrays & Synthetic Fallback Defaults**:
   - `OverviewTab.tsx`: Hardcoded `trendData` weekly performance array (lines 15–23), static string `"60 failed transactions detected"` (line 48), and static engine health/guardrail indicators (lines 156–169).
   - `ReconciliationTab.tsx`: Hardcoded numeric fallbacks (`total_records || 50`, `exact_matches || 42`, `fuzzy_matches || 5`, `ai_matches || 2`, `unmatched || 1`, `match_rate || 0.942`) (lines 38–43). This causes the UI to render synthetic 42/5/2 match splits whenever real data is 0 or empty.
   - `RecoveriesTab.tsx`: Fallback `guardrails` (`['MAX_RETRIES', 'CONTACT_WINDOW', 'COST_CAP']`) and fallback `steps` array (`st_1`, `st_2` with hardcoded 2026 timestamps) (lines 60–69).
   - `FailuresTab.tsx`: Fallback strings for missing `root_cause`, `order_id` (`'ord_001'`), and `suggestion` (`'RETRY_PAYMENT'`).
   - `AuditTab.tsx`: Fallback `guardrails` array (`['HMAC_VERIFICATION']`) (line 79).
3. **Missing Frontend View Components (Disputes & Refunds)**:
   - Backend `dashboard_api` implements `GET /api/v1/disputes` and `GET /api/v1/refunds`.
   - `lib/api.ts` implements `fetchDisputes()` and `fetchRefunds()`.
   - However, **no frontend components (`DisputesTab.tsx` or `RefundsTab.tsx`) exist** in `frontend/src/components/`, nor are they integrated into `Sidebar.tsx` navigation or `App.tsx`.

---

## 2. Frontend Architecture & Data Flow Overview

The frontend is a Vite + React + TypeScript single-page application:
- `App.tsx`: Controls main tab state, polling timer (10s interval for `refreshAllData`), and top-level data fetching via `lib/api.ts` (`fetchOverviewMetrics`, `fetchFailures`, `fetchWorkflows`, `fetchAuditLogs`).
- `lib/api.ts`: API wrapper connecting to `http://localhost:8005/api/v1` (`dashboard_api`) and `http://localhost:8006` (`ai_gateway`).
- Tab Components:
  - Data supplied via Props from `App.tsx`: `OverviewTab`, `FailuresTab`, `RecoveriesTab`, `AuditTab`.
  - Self-fetching Data in `useEffect`: `SettlementsTab`, `SubscriptionsTab`, `ReconciliationTab`, `PromisesTab`, `ChatTab`.

---

## 3. View-by-View Detailed Audit

### 3.1 `mockData.ts` (Standalone Mock Module)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/mockData.ts`
- **Lines**: 1 to 135
- **Content**:
  - Lines 3–52: `MOCK_FAILURES: FailureRecord[]` (4 synthetic objects).
  - Lines 54–83: `MOCK_WORKFLOWS: WorkflowRecord[]` (2 synthetic workflow objects).
  - Lines 85–134: `MOCK_AUDIT: AuditEntry[]` (4 synthetic audit log objects).
- **Status**: Orphaned mock file. Not imported by any view component.

---

### 3.2 Overview View (`OverviewTab.tsx`)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/components/OverviewTab.tsx`
- **Prop / API Data**: Accepts `metrics` from `App.tsx` (`fetchOverviewMetrics()` -> `GET /api/v1/overview`).
- **Static Mock Data / Fallbacks Identified**:
  - **Lines 15–23**:
    ```typescript
    const trendData = [
      { day: 'Mon', atRisk: 42000, recovered: 31000 },
      { day: 'Tue', atRisk: 38000, recovered: 29000 },
      { day: 'Wed', atRisk: 55000, recovered: 41000 },
      { day: 'Thu', atRisk: 48000, recovered: 36000 },
      { day: 'Fri', atRisk: 62000, recovered: 48000 },
      { day: 'Sat', atRisk: 29000, recovered: 24000 },
      { day: 'Sun', atRisk: 234500, recovered: 172000 },
    ];
    ```
    Used directly in `<AreaChart data={trendData}>` (Line 134).
  - **Line 48**: `"60 failed transactions detected"` (hardcoded text count).
  - **Lines 156–169**: Engine Health & Guardrails panel text:
    - Line 158: `Circuit Breaker Status: CLOSED (Healthy — 0 failures/min)`
    - Line 162: `Contact Window Guardrail: ACTIVE (09:00 - 21:00 IST Enforced)`
    - Line 166: `Recovery Cost Cap: HARD CAP (Max 20% of Payment Value)`
- **Replacement Strategy**:
  1. Add dynamic daily breakdown array (`trend_data`) in `GET /api/v1/overview` backend response by grouping MongoDB failures by weekday/date.
  2. Include `failed_transactions_count` in `metrics` object returned by `GET /api/v1/overview`.
  3. Include `circuit_breaker_status`, `contact_window_active`, and `cost_cap_rule` in `metrics` or fetch from a dedicated `GET /api/v1/health` endpoint.

---

### 3.3 Failures View (`FailuresTab.tsx`)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/components/FailuresTab.tsx`
- **Prop / API Data**: Accepts `failures` array from `App.tsx` (`fetchFailures()` -> `GET /api/v1/failures`). Triggers live AI diagnosis via `triggerDiagnosis()` -> `POST /api/v1/diagnose`.
- **Static Fallbacks Identified**:
  - **Line 149**: `"{diagnosisResult || selectedFailure.root_cause || 'Click below to run real-time AI diagnosis against failure_detector service.'}"`
  - **Line 177**: `{selectedFailure.order_id || 'ord_001'}`
  - **Line 185**: `{selectedFailure.suggestion || 'RETRY_PAYMENT'}`
- **Replacement Strategy**:
  1. Update `dashboard_api` MongoDB `failures` collection schema to persist `order_id`, `suggestion`, and `root_cause`.
  2. In UI, display `selectedFailure.order_id || 'N/A'` and `selectedFailure.suggestion || 'NO_SUGGESTION'` rather than synthetic ID defaults like `'ord_001'`.

---

### 3.4 Recoveries View (`RecoveriesTab.tsx`)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/components/RecoveriesTab.tsx`
- **Prop / API Data**: Accepts `workflows` array from `App.tsx` (`fetchWorkflows()` -> `GET /api/v1/recoveries`). Triggers live recovery orchestration via `triggerWorkflow()` -> `POST /api/v1/orchestrate`.
- **Static Fallbacks Identified**:
  - **Lines 60–62**:
    ```typescript
    const guardrails = Array.isArray(wf.guardrails_checked) && wf.guardrails_checked.length > 0
      ? wf.guardrails_checked
      : ['MAX_RETRIES', 'CONTACT_WINDOW', 'COST_CAP'];
    ```
  - **Lines 64–69**:
    ```typescript
    const steps = Array.isArray(wf.steps) && wf.steps.length > 0
      ? wf.steps
      : [
          { step_id: 'st_1', action: 'DIAGNOSE_FAILURE', status: 'SUCCESS', result: 'Category: BANK_DECLINE', executed_at: '2026-08-21T08:14:05Z' },
          { step_id: 'st_2', action: 'RETRY_PAYMENT', status: 'SUCCESS', result: `Recovered via ${wf.recovered_id || 'retry attempt'}`, executed_at: '2026-08-21T08:15:20Z' }
        ];
    ```
- **Replacement Strategy**:
  1. Ensure `recovery_orchestrator` saves step updates into Mongo `workflows` collection with actual execution step details and guardrail names.
  2. Remove default hardcoded steps array. If `steps` is empty, render `"Workflow pending step execution"`.

---

### 3.5 Settlements View (`SettlementsTab.tsx`)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/components/SettlementsTab.tsx`
- **Prop / API Data**: Fetches live data via `fetchSettlements()` -> `GET /api/v1/settlements`.
- **Static Fallbacks Identified**: None. Clean dynamic rendering of synced Razorpay bank settlements with empty state handling.

---

### 3.6 Subscriptions View (`SubscriptionsTab.tsx`)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/components/SubscriptionsTab.tsx`
- **Prop / API Data**: Fetches live data via `fetchSubscriptions()` -> `GET /api/v1/subscriptions`.
- **Static Fallbacks Identified**: None. Clean dynamic rendering of synced Razorpay subscriptions with empty state handling.

---

### 3.7 Disputes View (MISSING VIEW COMPONENT)
- **Current Status**:
  - Backend `dashboard_api` supports `GET /api/v1/disputes`.
  - `lib/api.ts` defines `export async function fetchDisputes()`.
  - **Missing Component**: No `DisputesTab.tsx` exists.
- **Replacement Strategy**: Create `frontend/src/components/DisputesTab.tsx` to display dispute records (Dispute ID, Payment ID, Amount, Status, Reason, Due Date) synced from Razorpay, and add `disputes` tab to `Sidebar.tsx` and `App.tsx`.

---

### 3.8 Refunds View (MISSING VIEW COMPONENT)
- **Current Status**:
  - Backend `dashboard_api` supports `GET /api/v1/refunds`.
  - `lib/api.ts` defines `export async function fetchRefunds()`.
  - **Missing Component**: No `RefundsTab.tsx` exists.
- **Replacement Strategy**: Create `frontend/src/components/RefundsTab.tsx` to display refund records (Refund ID, Payment ID, Amount, Status, Speed, Created Date) synced from Razorpay, and add `refunds` tab to `Sidebar.tsx` and `App.tsx`.

---

### 3.9 Promises View (`PromisesTab.tsx`)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/components/PromisesTab.tsx`
- **Prop / API Data**: Fetches live data via `fetchPromises()` -> `GET /api/v1/promises`.
- **Static Fallbacks Identified**: None inside component.

---

### 3.10 Reconciliation View (`ReconciliationTab.tsx`)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/components/ReconciliationTab.tsx`
- **Prop / API Data**: Fetches live reconciliation metrics via `fetchReconciliation()` -> `GET /api/v1/reconciliation` and triggers batch reconciliation via `triggerReconciliationBatch()` -> `POST /api/v1/reconcile`.
- **Static Fallbacks Identified**:
  - **Lines 38–43**:
    ```typescript
    const total = data?.total_records || 50;
    const exact = data?.exact_matches || 42;
    const fuzzy = data?.fuzzy_matches || 5;
    const ai = data?.ai_matches || 2;
    const unmatched = data?.unmatched || 1;
    const matchRatePct = ((data?.match_rate || 0.942) * 100).toFixed(1);
    ```
- **Replacement Strategy**: Replace `|| 50`, `|| 42`, `|| 5`, `|| 2`, `|| 1`, `|| 0.942` with nullish coalescing `?? 0`. If `total === 0`, display empty state component or 0% match rate instead of synthetic 42/5/2 match splits.

---

### 3.11 Audit View (`AuditTab.tsx`)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/components/AuditTab.tsx`
- **Prop / API Data**: Accepts `logs` array from `App.tsx` (`fetchAuditLogs()` -> `GET /api/v1/audit`).
- **Static Fallbacks Identified**:
  - **Line 79**: `const guardrails = Array.isArray(log.guardrails_checked) ? log.guardrails_checked : ['HMAC_VERIFICATION'];`
- **Replacement Strategy**: Default to empty array `[]` when `guardrails_checked` is missing/empty, rendering `"None"` instead of synthetic `'HMAC_VERIFICATION'`.

---

### 3.12 Chat View (`ChatTab.tsx`)
- **Path**: `/home/mahi17/Github/fintech/frontend/src/components/ChatTab.tsx`
- **Prop / API Data**: Calls `sendAIChatMessage()` -> `POST http://localhost:8006/chat` (`ai_gateway`).
- **Static Data Identified**:
  - **Lines 24–28**: Hardcoded sample prompt chips (`"Why was yesterday's settlement short?"`, etc.).

---

## 4. Backend Endpoints Summary & Connection Mapping

| View Component | Current API Call | Backend Endpoint | Status / Data Source |
|---|---|---|---|
| Overview | `fetchOverviewMetrics()` | `GET /api/v1/overview` | Dynamic MongoDB calculation |
| Overview Trend Chart | None (static `trendData`) | *Needs `GET /api/v1/overview/trend`* | Missing endpoint |
| Failures | `fetchFailures()`, `triggerDiagnosis()` | `GET /api/v1/failures`, `POST /api/v1/diagnose` | Dynamic MongoDB + Proxy to `failure_detector` |
| Recoveries | `fetchWorkflows()`, `triggerWorkflow()` | `GET /api/v1/recoveries`, `POST /api/v1/orchestrate` | Dynamic MongoDB + Proxy to `recovery_orchestrator` |
| Settlements | `fetchSettlements()` | `GET /api/v1/settlements` | Dynamic MongoDB synced from Razorpay REST API |
| Subscriptions | `fetchSubscriptions()` | `GET /api/v1/subscriptions` | Dynamic MongoDB synced from Razorpay REST API |
| **Disputes** | None in UI (`fetchDisputes()` exists in `lib/api.ts`) | `GET /api/v1/disputes` | Endpoint exists in backend; **UI tab missing** |
| **Refunds** | None in UI (`fetchRefunds()` exists in `lib/api.ts`) | `GET /api/v1/refunds` | Endpoint exists in backend; **UI tab missing** |
| Promises | `fetchPromises()` | `GET /api/v1/promises` | Dynamic MongoDB |
| Reconciliation | `fetchReconciliation()`, `triggerReconciliationBatch()` | `GET /api/v1/reconciliation`, `POST /api/v1/reconcile` | Dynamic MongoDB + Proxy to `reconciliation_engine` |
| Audit | `fetchAuditLogs()` | `GET /api/v1/audit` | Dynamic MongoDB |
| Chat | `sendAIChatMessage()` | `POST http://localhost:8006/chat` | Live AI Gateway service |

---

## 5. Comprehensive Recommendation & Action Plan

1. **Delete / Clean Up `mockData.ts`**: Remove unused static mock file.
2. **Remove Synthetic Fallbacks in Components**:
   - `ReconciliationTab.tsx`: Replace `|| 50`, `|| 42`, etc. with `?? 0`.
   - `OverviewTab.tsx`: Fetch `trend_data` dynamically from backend; replace `"60 failed transactions"` with `metrics.failed_count`.
   - `RecoveriesTab.tsx`: Remove hardcoded fallback steps array (`st_1`, `st_2`).
   - `FailuresTab.tsx`: Replace `'ord_001'` fallback with `f.order_id || 'N/A'`.
   - `AuditTab.tsx`: Replace `['HMAC_VERIFICATION']` fallback with `log.guardrails_checked || []`.
3. **Build Missing View Components**:
   - Implement `DisputesTab.tsx` and `RefundsTab.tsx`.
   - Update `Sidebar.tsx` and `App.tsx` to include Disputes and Refunds navigation items.
4. **Backend Enhancements in `dashboard_api`**:
   - Add `trend_data` array to `GET /api/v1/overview`.
   - Ensure MongoDB seed/sync populates all record fields so no UI default fallbacks are triggered.
