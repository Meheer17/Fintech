import React, { useState } from 'react';
import { FailureRecord } from '../types';
import { ChevronRight, Filter, Sparkles } from 'lucide-react';
import { triggerDiagnosis, triggerWorkflow } from '../lib/api';

interface FailuresTabProps {
  failures?: FailureRecord[];
}

export const FailuresTab: React.FC<FailuresTabProps> = ({ failures = [] }) => {
  const safeFailures = Array.isArray(failures) ? failures : [];
  const [selectedFailure, setSelectedFailure] = useState<FailureRecord | null>(safeFailures[0] || null);
  const [diagnosing, setDiagnosing] = useState(false);
  const [diagnosisResult, setDiagnosisResult] = useState<string | null>(null);
  const [recoveryStep, setRecoveryStep] = useState<number>(0); // 0=idle, 1=diagnosing, 2=orchestrating, 3=done
  const [recoveryResult, setRecoveryResult] = useState<any>(null);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'RECOVERED':
        return { label: '● Recovered', bg: '#ebfbee', color: '#2b8a3e' };
      case 'IN_PROGRESS':
        return { label: '● In Progress', bg: '#e7f5ff', color: '#1971c2' };
      case 'PENDING':
        return { label: '● Pending', bg: '#fff9db', color: '#e67700' };
      default:
        return { label: '● Failed / Abandoned', bg: '#fff5f5', color: '#c92a2a' };
    }
  };

  const handleExecuteRecovery = async () => {
    if (!selectedFailure) return;
    setRecoveryStep(1);
    setDiagnosisResult(null);
    setRecoveryResult(null);
    
    try {
      // Step 1: Diagnose
      const diagRes = await triggerDiagnosis(selectedFailure.payment_id, selectedFailure.category || 'BANK_DECLINE', 'Manual evaluation');
      setDiagnosisResult(diagRes.root_cause || diagRes.reason || 'AI diagnosis completed');
      
      // Step 2: Orchestrate (creates payment link)
      setRecoveryStep(2);
      const orchRes = await triggerWorkflow(
        selectedFailure.payment_id,
        0,
        selectedFailure.amount_paise || 0
      );
      setRecoveryResult(orchRes);
      setRecoveryStep(3);
    } catch (e: any) {
      setRecoveryStep(0);
      setDiagnosisResult('Recovery error: ' + e.message);
    }
  };

  return (
    <div style={{ display: 'grid', gridTemplateColumns: selectedFailure ? '2fr 1fr' : '1fr', gap: '20px' }}>
      {/* Table Container */}
      <div style={{
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        border: '1px solid #e9ecef',
        boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
        overflow: 'hidden'
      }}>
        {/* Table Header Controls */}
        <div style={{
          padding: '16px 20px',
          borderBottom: '1px solid #f1f3f5',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between'
        }}>
          <h3 style={{ fontSize: '15px', fontWeight: 600, color: '#1a1f2c' }}>
            Payment Failures ({safeFailures.length})
          </h3>
          <div style={{ display: 'flex', gap: '8px', fontSize: '12px' }}>
            <span style={{ padding: '6px 12px', borderRadius: '6px', backgroundColor: '#f8f9fa', border: '1px solid #e9ecef', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <Filter size={14} color="#8c98a9" /> All Categories
            </span>
          </div>
        </div>

        {/* Table */}
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px', textAlign: 'left' }}>
          <thead>
            <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '1px solid #e9ecef', color: '#8c98a9', fontWeight: 600 }}>
              <th style={{ padding: '12px 16px' }}>Payment ID</th>
              <th style={{ padding: '12px 16px' }}>Amount</th>
              <th style={{ padding: '12px 16px' }}>Method</th>
              <th style={{ padding: '12px 16px' }}>Category</th>
              <th style={{ padding: '12px 16px' }}>Status</th>
              <th style={{ padding: '12px 16px' }}>Action</th>
            </tr>
          </thead>
          <tbody>
            {safeFailures.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ padding: '24px', textAlign: 'center', color: '#8c98a9' }}>
                  No payment failures recorded.
                </td>
              </tr>
            ) : (
              safeFailures.map((f, i) => {
                const statusInfo = getStatusBadge(f.recovery_status || 'PENDING');
                const isSelected = selectedFailure?.id === f.id || selectedFailure?.payment_id === f.payment_id;
                return (
                  <tr
                    key={f.id || f.payment_id || i}
                    onClick={() => {
                      setSelectedFailure(f);
                      setDiagnosisResult(null);
                      setRecoveryStep(0);
                      setRecoveryResult(null);
                    }}
                    style={{
                      borderBottom: '1px solid #f1f3f5',
                      backgroundColor: isSelected ? '#edf2ff' : (i % 2 === 0 ? '#ffffff' : '#f8f9fa'),
                      cursor: 'pointer'
                    }}
                  >
                    <td style={{ padding: '12px 16px', fontWeight: 600 }} className="mono">{f.payment_id}</td>
                    <td style={{ padding: '12px 16px', fontWeight: 600 }}>₹{((f.amount_paise || 0) / 100).toLocaleString('en-IN')}</td>
                    <td style={{ padding: '12px 16px', textTransform: 'uppercase', fontSize: '11px', fontWeight: 600, color: '#4a5568' }}>{f.payment_method || 'N/A'}</td>
                    <td style={{ padding: '12px 16px', color: '#4a5568' }}>{f.category || 'GATEWAY_ERROR'}</td>
                    <td style={{ padding: '12px 16px' }}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '12px',
                        backgroundColor: statusInfo.bg,
                        color: statusInfo.color,
                        fontSize: '11px',
                        fontWeight: 600
                      }}>
                        {statusInfo.label}
                      </span>
                    </td>
                    <td style={{ padding: '12px 16px', color: '#4263eb' }}>
                      <ChevronRight size={16} />
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Detail Drawer */}
      {selectedFailure && (
        <div style={{
          backgroundColor: '#ffffff',
          borderRadius: '8px',
          border: '1px solid #e9ecef',
          padding: '24px',
          boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
          display: 'flex',
          flexDirection: 'column',
          gap: '16px'
        }}>
          <div>
            <span style={{ fontSize: '11px', textTransform: 'uppercase', color: '#8c98a9', fontWeight: 600 }}>
              Failure Context & Live AI Diagnosis
            </span>
            <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c', marginTop: '4px' }} className="mono">
              {selectedFailure.payment_id}
            </h3>
          </div>

          <div style={{ padding: '12px', borderRadius: '6px', backgroundColor: '#f8f9fa', border: '1px solid #e9ecef', fontSize: '12px' }}>
            <div style={{ fontWeight: 600, color: '#8c98a9' }}>AI Root Cause Explanation</div>
            <div style={{ color: '#1a1f2c', marginTop: '4px', fontWeight: 500 }}>
              "{diagnosisResult || selectedFailure.root_cause || 'No diagnosis available. Click below to run real-time AI diagnosis.'}"
            </div>
          </div>

          {recoveryStep === 0 && (
            <button
              onClick={handleExecuteRecovery}
              style={{
                padding: '10px 16px',
                backgroundColor: '#4263eb',
                color: '#ffffff',
                border: 'none',
                borderRadius: '6px',
                fontWeight: 600,
                fontSize: '13px',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '8px'
              }}
            >
              <Sparkles size={16} /> 🚀 Execute AI Recovery
            </button>
          )}

          {recoveryStep > 0 && recoveryStep < 3 && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '12px', backgroundColor: '#e7f5ff', borderRadius: '6px', color: '#1971c2', fontSize: '13px', fontWeight: 500 }}>
              <div style={{ width: '8px', height: '8px', borderRadius: '50%', backgroundColor: '#1971c2', animation: 'pulse 1.5s infinite' }} />
              {recoveryStep === 1 ? 'Analyzing failure with AI...' : 'Creating recovery workflow & payment link...'}
            </div>
          )}

          {recoveryStep === 3 && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', padding: '16px', backgroundColor: '#ebfbee', borderRadius: '6px', border: '1px solid #b2f2bb' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#2b8a3e', fontSize: '14px', fontWeight: 600 }}>
                <span>✅ Recovery triggered!</span>
              </div>
              <div style={{ fontSize: '13px', color: '#2b8a3e' }}>
                <strong>Action:</strong> {recoveryResult?.action_taken || 'CREATED PAYMENT LINK'}<br />
                <strong>Guardrails verified:</strong> {recoveryResult?.guardrails_passed ? 'Yes' : 'N/A'}
              </div>
              {recoveryResult?.payment_link_url && (
                <a href={recoveryResult.payment_link_url} target="_blank" rel="noopener noreferrer"
                   style={{ display: 'block', textAlign: 'center', padding: '10px 16px', backgroundColor: '#2b8a3e', color: '#fff', borderRadius: '6px', textDecoration: 'none', fontWeight: 600, fontSize: '13px' }}>
                  Open Razorpay Payment Link →
                </a>
              )}
            </div>
          )}

          <div style={{ fontSize: '13px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: '#8c98a9' }}>Order ID:</span>
              <span className="mono" style={{ fontWeight: 600 }}>{selectedFailure.order_id || 'N/A'}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: '#8c98a9' }}>Amount:</span>
              <span style={{ fontWeight: 600 }}>₹{((selectedFailure.amount_paise || 0) / 100).toLocaleString('en-IN')}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: '#8c98a9' }}>Suggested Action:</span>
              <span style={{ fontWeight: 600, color: '#4263eb' }}>{selectedFailure.suggestion || 'N/A'}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
