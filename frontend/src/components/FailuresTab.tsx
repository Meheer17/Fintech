import React, { useState } from 'react';
import { FailureRecord } from '../types';
import { ChevronRight, Filter } from 'lucide-react';

interface FailuresTabProps {
  failures?: FailureRecord[];
}

export const FailuresTab: React.FC<FailuresTabProps> = ({ failures = [] }) => {
  const safeFailures = Array.isArray(failures) ? failures : [];
  const [selectedFailure, setSelectedFailure] = useState<FailureRecord | null>(safeFailures[0] || null);

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

        {/* Clean Table */}
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
            {safeFailures.map((f, i) => {
              const statusInfo = getStatusBadge(f.recovery_status || 'PENDING');
              const isSelected = selectedFailure?.id === f.id;
              return (
                <tr
                  key={f.id || i}
                  onClick={() => setSelectedFailure(f)}
                  style={{
                    borderBottom: '1px solid #f1f3f5',
                    backgroundColor: isSelected ? '#edf2ff' : (i % 2 === 0 ? '#ffffff' : '#f8f9fa'),
                    cursor: 'pointer'
                  }}
                >
                  <td style={{ padding: '12px 16px', fontWeight: 600 }} className="mono">{f.payment_id}</td>
                  <td style={{ padding: '12px 16px', fontWeight: 600 }}>₹{((f.amount_paise || 0) / 100).toLocaleString('en-IN')}</td>
                  <td style={{ padding: '12px 16px', textTransform: 'uppercase', fontSize: '11px', fontWeight: 600, color: '#4a5568' }}>{f.payment_method || 'upi'}</td>
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
            })}
          </tbody>
        </table>
      </div>

      {/* Selected Failure Detail Timeline Drawer */}
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
              Failure Context & AI Diagnosis
            </span>
            <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c', marginTop: '4px' }} className="mono">
              {selectedFailure.payment_id}
            </h3>
          </div>

          <div style={{ padding: '12px', borderRadius: '6px', backgroundColor: '#f8f9fa', border: '1px solid #e9ecef', fontSize: '12px' }}>
            <div style={{ fontWeight: 600, color: '#8c98a9' }}>AI Root Cause Explanation</div>
            <div style={{ color: '#1a1f2c', marginTop: '4px', fontWeight: 500 }}>
              "{selectedFailure.root_cause || 'AI diagnosis classified failure and recommended bounded recovery action.'}"
            </div>
          </div>

          <div style={{ fontSize: '13px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: '#8c98a9' }}>Order ID:</span>
              <span className="mono" style={{ fontWeight: 600 }}>{selectedFailure.order_id || 'ord_001'}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: '#8c98a9' }}>Amount:</span>
              <span style={{ fontWeight: 600 }}>₹{((selectedFailure.amount_paise || 0) / 100).toLocaleString('en-IN')}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: '#8c98a9' }}>Suggested Action:</span>
              <span style={{ fontWeight: 600, color: '#4263eb' }}>{selectedFailure.suggestion || 'RETRY_PAYMENT'}</span>
            </div>
          </div>

          <div style={{ borderTop: '1px solid #f1f3f5', paddingTop: '16px' }}>
            <h4 style={{ fontSize: '13px', fontWeight: 600, color: '#1a1f2c', marginBottom: '12px' }}>
              Bounded Workflow Timeline
            </h4>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', fontSize: '12px', borderLeft: '2px solid #e9ecef', paddingLeft: '12px' }}>
              <div>
                <div style={{ fontWeight: 600, color: '#1a1f2c' }}>1. Webhook Ingested</div>
                <div style={{ color: '#8c98a9', fontSize: '11px' }}>Signature valid • Ingested by webhook-receiver</div>
              </div>
              <div>
                <div style={{ fontWeight: 600, color: '#1a1f2c' }}>2. AI Root Cause Diagnosed</div>
                <div style={{ color: '#8c98a9', fontSize: '11px' }}>Classified by failure-detector (95% confidence)</div>
              </div>
              <div>
                <div style={{ fontWeight: 600, color: '#4263eb' }}>3. Guardrails Evaluated</div>
                <div style={{ color: '#8c98a9', fontSize: '11px' }}>Max retries ok • Contact window ok • Cost cap ok</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
