import React, { useState } from 'react';
import { FailureRecord } from '../types';
import { ChevronRight, Filter, Sparkles } from 'lucide-react';
import { triggerDiagnosis } from '../lib/api';

interface FailuresTabProps {
  failures?: FailureRecord[];
}

export const FailuresTab: React.FC<FailuresTabProps> = ({ failures = [] }) => {
  const safeFailures = Array.isArray(failures) ? failures : [];
  const [selectedFailure, setSelectedFailure] = useState<FailureRecord | null>(safeFailures[0] || null);
  const [diagnosing, setDiagnosing] = useState(false);
  const [diagnosisResult, setDiagnosisResult] = useState<string | null>(null);

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

  const handleRunDiagnosis = async () => {
    if (!selectedFailure) return;
    setDiagnosing(true);
    try {
      const res = await triggerDiagnosis(selectedFailure.payment_id, selectedFailure.category || 'BANK_DECLINE', 'Manual evaluation');
      setDiagnosisResult(res.root_cause || res.reason || 'AI diagnosis completed.');
    } catch (e: any) {
      setDiagnosisResult(`Diagnosis error: ${e.message}`);
    } finally {
      setDiagnosing(false);
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

          <button
            onClick={handleRunDiagnosis}
            disabled={diagnosing}
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
            <Sparkles size={16} /> {diagnosing ? 'Running AI Diagnosis...' : 'Execute Live AI Diagnosis'}
          </button>

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
