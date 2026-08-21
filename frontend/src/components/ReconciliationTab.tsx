import React from 'react';
import { Sparkles } from 'lucide-react';

export const ReconciliationTab: React.FC = () => {
  const exceptions = [
    {
      id: 'exc_001',
      type: 'FEE_DISCREPANCY',
      order_id: 'order_RZP_0012',
      expected: 250000,
      actual: 245000,
      suggestion: 'Accept 2% processing fee deduction (₹50.00)',
      status: 'OPEN'
    },
    {
      id: 'exc_002',
      type: 'TIMING_MISMATCH',
      order_id: 'order_RZP_0018',
      expected: 499000,
      actual: 499000,
      suggestion: 'Bank settlement delayed by 1 day due to weekend holiday',
      status: 'RESOLVED'
    }
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      <div>
        <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c' }}>
          Settlement Reconciliation Engine (Track 04)
        </h3>
        <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '2px' }}>
          Three-way automated matching across Merchant Orders ↔ Razorpay Payments ↔ Bank Settlements
        </p>
      </div>

      {/* Match Breakdown Panel */}
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
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <div style={{ fontSize: '12px', fontWeight: 600, color: '#8c98a9', textTransform: 'uppercase' }}>Overall Match Rate</div>
            <div style={{ fontSize: '28px', fontWeight: 700, color: '#2b8a3e', marginTop: '2px' }}>94.2% Auto-Matched</div>
          </div>
          <div style={{ fontSize: '13px', textAlign: 'right' }}>
            <div>Total Records: <strong>50 Records</strong></div>
            <div style={{ color: '#8c98a9', fontSize: '12px' }}>Batch Run: 2026-08-21 08:30 IST</div>
          </div>
        </div>

        {/* Visual Progress Bar */}
        <div style={{ height: '12px', borderRadius: '6px', backgroundColor: '#e9ecef', overflow: 'hidden', display: 'flex' }}>
          <div style={{ width: '84%', backgroundColor: '#2b8a3e' }} title="Exact Matches (84%)" />
          <div style={{ width: '10%', backgroundColor: '#1971c2' }} title="Fuzzy Matches (10%)" />
          <div style={{ width: '4%', backgroundColor: '#e67700' }} title="AI Matches (4%)" />
          <div style={{ width: '2%', backgroundColor: '#c92a2a' }} title="Unmatched Exceptions (2%)" />
        </div>

        <div style={{ display: 'flex', gap: '24px', fontSize: '12px', color: '#4a5568' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ width: '10px', height: '10px', borderRadius: '2px', backgroundColor: '#2b8a3e' }} />
            <span>Exact Matches: <strong>42 (84%)</strong></span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ width: '10px', height: '10px', borderRadius: '2px', backgroundColor: '#1971c2' }} />
            <span>Fuzzy Matches: <strong>5 (10%)</strong></span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ width: '10px', height: '10px', borderRadius: '2px', backgroundColor: '#e67700' }} />
            <span>AI Resolved: <strong>2 (4%)</strong></span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ width: '10px', height: '10px', borderRadius: '2px', backgroundColor: '#c92a2a' }} />
            <span>Exceptions: <strong>1 (2%)</strong></span>
          </div>
        </div>
      </div>

      {/* Exception Resolution Queue */}
      <div style={{
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        border: '1px solid #e9ecef',
        boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
        overflow: 'hidden'
      }}>
        <div style={{ padding: '16px 20px', borderBottom: '1px solid #f1f3f5', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <h4 style={{ fontSize: '15px', fontWeight: 600, color: '#1a1f2c' }}>
            Honest Exception Resolution Queue ({exceptions.length})
          </h4>
        </div>

        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px', textAlign: 'left' }}>
          <thead>
            <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '1px solid #e9ecef', color: '#8c98a9', fontWeight: 600 }}>
              <th style={{ padding: '12px 16px' }}>Exception ID</th>
              <th style={{ padding: '12px 16px' }}>Type</th>
              <th style={{ padding: '12px 16px' }}>Order ID</th>
              <th style={{ padding: '12px 16px' }}>Expected / Actual</th>
              <th style={{ padding: '12px 16px' }}>AI Resolution Suggestion</th>
              <th style={{ padding: '12px 16px' }}>Status</th>
            </tr>
          </thead>
          <tbody>
            {exceptions.map((exc) => (
              <tr key={exc.id} style={{ borderBottom: '1px solid #f1f3f5' }}>
                <td style={{ padding: '12px 16px', fontWeight: 600 }} className="mono">{exc.id}</td>
                <td style={{ padding: '12px 16px', fontWeight: 600, color: '#e67700' }}>{exc.type}</td>
                <td style={{ padding: '12px 16px' }} className="mono">{exc.order_id}</td>
                <td style={{ padding: '12px 16px' }}>
                  ₹{(exc.expected / 100).toFixed(2)} / ₹{(exc.actual / 100).toFixed(2)}
                </td>
                <td style={{ padding: '12px 16px', color: '#4263eb' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <Sparkles size={14} />
                    {exc.suggestion}
                  </div>
                </td>
                <td style={{ padding: '12px 16px' }}>
                  <span style={{
                    padding: '4px 8px',
                    borderRadius: '12px',
                    backgroundColor: exc.status === 'RESOLVED' ? '#ebfbee' : '#fff9db',
                    color: exc.status === 'RESOLVED' ? '#2b8a3e' : '#e67700',
                    fontSize: '11px',
                    fontWeight: 600
                  }}>
                    {exc.status}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
