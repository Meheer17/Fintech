export type UserRole = 'razorpay_judge' | 'finance_controller' | 'ops_lead';

export interface UserProfile {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  merchantName: string;
}

export interface FailureRecord {
  id: string;
  payment_id: string;
  order_id: string;
  amount_paise: number;
  payment_method: string;
  category: string;
  root_cause: string;
  recovery_status: 'PENDING' | 'IN_PROGRESS' | 'RECOVERED' | 'FAILED' | 'ABANDONED';
  failed_at: string;
  suggestion: string;
}

export interface WorkflowRecord {
  workflow_id: string;
  failure_id?: string;
  payment_id: string;
  amount_paise: number;
  workflow_status: string;
  retry_count?: number;
  recovered_id?: string;
  contact_count?: number;
  guardrails_checked?: string[];
  steps?: {
    step_id: string;
    action: string;
    status: string;
    result: string;
    executed_at: string;
  }[];
}

export interface AuditEntry {
  id: string;
  service_name: string;
  action: string;
  entity_type: string;
  entity_id: string;
  actor: string;
  reasoning: string;
  guardrails_checked: string[];
  status: 'SUCCESS' | 'FAILED' | 'BLOCKED';
  timestamp: string;
}
