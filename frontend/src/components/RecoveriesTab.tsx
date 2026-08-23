import React, { useState } from 'react';
import { WorkflowRecord } from '../types';
import { RefreshCw, CheckCircle2, Shield, Play } from 'lucide-react';
import { triggerWorkflow } from '../lib/api';

interface RecoveriesTabProps {
  workflows?: WorkflowRecord[];
}

export const RecoveriesTab: React.FC<RecoveriesTabProps> = ({ workflows = [] }) => {
  const safeWorkflows = Array.isArray(workflows) && workflows.length > 0 ? workflows : [];
  const [executingId, setExecutingId] = useState<string | null>(null);
  const [statusMsg, setStatusMsg] = useState<string | null>(null);

  const handleRunWorkflow = async (paymentId: string, currentRetries: number, amount: number) => {
    setExecutingId(paymentId);
    try {
      const res = await triggerWorkflow(paymentId, currentRetries, amount);
      setStatusMsg(`Triggered live workflow for ${paymentId}: ${res.action || 'SUCCESS'} (${res.reason || 'Executed step'})`);
    } catch (e: any) {
      setStatusMsg(`Workflow error: ${e.message}`);
    } finally {
      setExecutingId(null);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c' }}>
            Bounded Recovery Workflows (GraphBuilder DAGs)
          </h3>
          <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '2px' }}>
            Multi-agent state machines executing bounded recovery actions under hard compliance guardrails
          </p>
        </div>
      </div>

      {statusMsg && (
        <div style={{ padding: '12px', borderRadius: '6px', backgroundColor: '#e7f5ff', border: '1px solid #74c0fc', color: '#1971c2', fontSize: '13px', fontWeight: 600 }}>
          {statusMsg}
        </div>
      )}

      {safeWorkflows.length === 0 ? (
        <div style={{
          backgroundColor: '#ffffff',
          borderRadius: '8px',
          border: '1px solid #e9ecef',
          padding: '40px',
          textAlign: 'center',
          color: '#8c98a9'
        }}>
          No active recovery workflows found.
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '16px' }}>
          {safeWorkflows.map((wf, idx) => {
            const guardrails = Array.isArray(wf.guardrails_checked) ? wf.guardrails_checked : [];
            const steps = Array.isArray(wf.steps) ? wf.steps : [];

            return (
              <div key={wf.workflow_id || `wf_${idx}`} style={{
                backgroundColor: '#ffffff',
                borderRadius: '8px',
                border: '1px solid #e9ecef',
                padding: '20px',
                boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
                display: 'flex',
                flexDirection: 'column',
                gap: '16px'
              }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <div>
                    <span className="mono" style={{ fontSize: '14px', fontWeight: 700, color: '#4263eb' }}>
                      {wf.workflow_id}
                    </span>
                    <div style={{ fontSize: '12px', color: '#8c98a9', marginTop: '2px' }}>
                      Payment: <strong className="mono">{wf.payment_id}</strong> (₹{((wf.amount_paise || 0) / 100).toLocaleString('en-IN')})
                    </div>
                  </div>
                  <span style={{
                    padding: '4px 10px',
                    borderRadius: '12px',
                    backgroundColor: wf.workflow_status === 'WF_RECOVERED' ? '#ebfbee' : '#e7f5ff',
                    color: wf.workflow_status === 'WF_RECOVERED' ? '#2b8a3e' : '#1971c2',
                    fontSize: '11px',
                    fontWeight: 700
                  }}>
                    {wf.workflow_status || 'IN_PROGRESS'}
                  </span>
                </div>

                {/* Guardrails Checked */}
                <div style={{ padding: '10px 12px', borderRadius: '6px', backgroundColor: '#f8f9fa', border: '1px solid #e9ecef', fontSize: '12px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontWeight: 600, color: '#1a1f2c', marginBottom: '6px' }}>
                    <Shield size={14} color="#2b8a3e" />
                    Hard Compliance Guardrails Passed
                  </div>
                  <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                    {guardrails.length === 0 ? (
                      <span style={{ color: '#8c98a9', fontSize: '11px' }}>None</span>
                    ) : (
                      guardrails.map((g, gIdx) => (
                        <span key={g + gIdx} style={{ padding: '2px 6px', borderRadius: '4px', backgroundColor: '#ffffff', border: '1px solid #dee2e6', fontSize: '10px', fontWeight: 600, color: '#4a5568' }}>
                          ✓ {g}
                        </span>
                      ))
                    )}
                  </div>
                </div>

                {/* Trigger Button */}
                <button
                  onClick={() => handleRunWorkflow(wf.payment_id, wf.retry_count || 0, wf.amount_paise || 0)}
                  disabled={executingId === wf.payment_id}
                  style={{
                    padding: '8px 12px',
                    backgroundColor: '#4263eb',
                    color: '#ffffff',
                    border: 'none',
                    borderRadius: '6px',
                    fontWeight: 600,
                    fontSize: '12px',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: '6px'
                  }}
                >
                  <Play size={14} /> {executingId === wf.payment_id ? 'Orchestrating...' : 'Trigger Live Recovery Step'}
                </button>

                {/* Steps Execution Pipeline */}
                <div>
                  <div style={{ fontSize: '12px', fontWeight: 600, color: '#8c98a9', marginBottom: '8px' }}>
                    Workflow Step Pipeline
                  </div>
                  {steps.length === 0 ? (
                    <div style={{ color: '#8c98a9', fontSize: '12px' }}>No workflow steps executed yet</div>
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                      {steps.map((st, sIdx) => (
                        <div key={st.step_id || sIdx} style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '12px' }}>
                          <CheckCircle2 size={16} color="#2b8a3e" />
                          <div style={{ flex: 1 }}>
                            <span style={{ fontWeight: 600, color: '#1a1f2c' }}>{st.action}</span>
                            <div style={{ color: '#8c98a9', fontSize: '11px' }}>{st.result}</div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
