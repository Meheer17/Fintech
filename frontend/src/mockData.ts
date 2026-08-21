import { FailureRecord, WorkflowRecord, AuditEntry } from './types';

export const MOCK_FAILURES: FailureRecord[] = [
  {
    id: 'fail_101',
    payment_id: 'pay_RZP_0012',
    order_id: 'order_RZP_0012',
    amount_paise: 250000, // ₹2,500.00
    payment_method: 'upi',
    category: 'BANK_DECLINE',
    root_cause: 'HDFC Bank UPI timeout during peak hours',
    recovery_status: 'RECOVERED',
    failed_at: '2026-08-21T08:14:00Z',
    suggestion: 'RETRY_SAME_METHOD'
  },
  {
    id: 'fail_102',
    payment_id: 'pay_RZP_0015',
    order_id: 'order_RZP_0015',
    amount_paise: 499000, // ₹4,990.00
    payment_method: 'card',
    category: 'CARD_EXPIRED',
    root_cause: 'HDFC Visa Card expired in 07/2026',
    recovery_status: 'IN_PROGRESS',
    failed_at: '2026-08-21T08:22:10Z',
    suggestion: 'SEND_PAYMENT_LINK'
  },
  {
    id: 'fail_103',
    payment_id: 'pay_RZP_0019',
    order_id: 'order_RZP_0019',
    amount_paise: 1000000, // ₹10,000.00
    payment_method: 'netbanking',
    category: 'INSUFFICIENT_FUNDS',
    root_cause: 'ICICI Netbanking account balance exceeded limit',
    recovery_status: 'PENDING',
    failed_at: '2026-08-21T08:35:45Z',
    suggestion: 'WAIT_AND_RETRY'
  },
  {
    id: 'fail_104',
    payment_id: 'pay_RZP_0024',
    order_id: 'order_RZP_0024',
    amount_paise: 120000, // ₹1,200.00
    payment_method: 'card',
    category: 'FRAUD_SUSPECTED',
    root_cause: 'Geographic impossibility (multiple IPs within 2 mins)',
    recovery_status: 'ABANDONED',
    failed_at: '2026-08-21T08:40:00Z',
    suggestion: 'ESCALATE_TO_HUMAN'
  }
];

export const MOCK_WORKFLOWS: WorkflowRecord[] = [
  {
    workflow_id: 'wf_801',
    failure_id: 'fail_101',
    payment_id: 'pay_RZP_0012',
    amount_paise: 250000,
    workflow_status: 'WF_RECOVERED',
    retry_count: 1,
    contact_count: 1,
    guardrails_checked: ['MAX_RETRIES', 'CONTACT_WINDOW', 'COST_CAP'],
    steps: [
      { step_id: 'st_1', action: 'DIAGNOSE_FAILURE', status: 'SUCCESS', result: 'Category: BANK_DECLINE', executed_at: '2026-08-21T08:14:05Z' },
      { step_id: 'st_2', action: 'RETRY_PAYMENT', status: 'SUCCESS', result: 'Payment recovered via UPI retry pay_RZP_0099', executed_at: '2026-08-21T08:15:20Z' }
    ]
  },
  {
    workflow_id: 'wf_802',
    failure_id: 'fail_102',
    payment_id: 'pay_RZP_0015',
    amount_paise: 499000,
    workflow_status: 'WF_IN_PROGRESS',
    retry_count: 0,
    contact_count: 1,
    guardrails_checked: ['MAX_RETRIES', 'CONTACT_WINDOW'],
    steps: [
      { step_id: 'st_1', action: 'DIAGNOSE_FAILURE', status: 'SUCCESS', result: 'Category: CARD_EXPIRED', executed_at: '2026-08-21T08:22:15Z' },
      { step_id: 'st_2', action: 'CREATE_PAYMENT_LINK', status: 'SUCCESS', result: 'Payment link generated: https://rzp.io/l/rec_4990', executed_at: '2026-08-21T08:23:00Z' }
    ]
  }
];

export const MOCK_AUDIT: AuditEntry[] = [
  {
    id: 'audit_501',
    service_name: 'webhook-receiver',
    action: 'INGEST_WEBHOOK',
    entity_type: 'PAYMENT_EVENT',
    entity_id: 'pay_RZP_0012',
    actor: 'Razorpay Webhook',
    reasoning: 'HMAC SHA256 signature verified (VALID)',
    guardrails_checked: ['HMAC_VERIFICATION'],
    status: 'SUCCESS',
    timestamp: '2026-08-21T08:14:01Z'
  },
  {
    id: 'audit_502',
    service_name: 'failure-detector',
    action: 'DIAGNOSE_FAILURE',
    entity_type: 'FAILURE_RECORD',
    entity_id: 'fail_101',
    actor: 'diagnosis_agent',
    reasoning: 'Error code GATEWAY_ERROR classified as BANK_DECLINE with 0.95 confidence',
    guardrails_checked: ['CONFIDENCE_THRESHOLD'],
    status: 'SUCCESS',
    timestamp: '2026-08-21T08:14:05Z'
  },
  {
    id: 'audit_503',
    service_name: 'recovery-orchestrator',
    action: 'EXECUTE_RETRY',
    entity_type: 'RECOVERY_WORKFLOW',
    entity_id: 'wf_801',
    actor: 'strategy_agent',
    reasoning: 'Retry #1 within limit (max 3), contact window active (10:44 AM IST), cost cap 0% < 20%',
    guardrails_checked: ['MAX_RETRIES', 'CONTACT_WINDOW_IST', 'COST_CAP'],
    status: 'SUCCESS',
    timestamp: '2026-08-21T08:15:20Z'
  },
  {
    id: 'audit_504',
    service_name: 'recovery-orchestrator',
    action: 'GUARDRAIL_BLOCK',
    entity_type: 'RECOVERY_WORKFLOW',
    entity_id: 'wf_809',
    actor: 'guardrails_engine',
    reasoning: 'Attempted outbound contact at 10:15 PM IST. Blocked by CONTACT_WINDOW_IST (9 AM - 9 PM IST)',
    guardrails_checked: ['CONTACT_WINDOW_IST'],
    status: 'BLOCKED',
    timestamp: '2026-08-21T08:18:00Z'
  }
];
