import React, { useEffect, useState } from 'react';
import { fetchRefunds } from '../lib/api';
import { RotateCcw, ArrowUpRight, CheckCircle2, Zap } from 'lucide-react';

export const RefundsTab: React.FC = () => {
  const [refunds, setRefunds] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchRefunds()
      .then((data) => {
        setRefunds(data.refunds || []);
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to load refunds:', err);
        setLoading(false);
      });
  }, []);

  const totalRefundedPaise = refunds.reduce((acc, item) => acc + (item.amount || 0), 0);
  const processedRefundsCount = refunds.filter((r) => r.status === 'processed').length;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '20px', fontWeight: 700, color: '#1a1f2c' }}>Razorpay Refunds Tracker</h2>
          <p style={{ fontSize: '13px', color: '#6c757d', marginTop: '4px' }}>
            Live customer payment refunds synced from Razorpay (`/v1/refunds`)
          </p>
        </div>
        <div style={{
          backgroundColor: '#e7f5ff',
          color: '#1971c2',
          padding: '8px 14px',
          borderRadius: '6px',
          fontWeight: 600,
          fontSize: '13px',
          display: 'flex',
          alignItems: 'center',
          gap: '6px'
        }}>
          <RotateCcw size={16} />
          {refunds.length} Refunds Synced
        </div>
      </div>

      {/* Summary Cards */}
      <div style={{
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        border: '1px solid #e9ecef',
        padding: '20px',
        display: 'grid',
        gridTemplateColumns: 'repeat(3, 1fr)',
        gap: '16px'
      }}>
        <div>
          <div style={{ fontSize: '12px', color: '#8c98a9', textTransform: 'uppercase', fontWeight: 600 }}>Total Refunded Amount</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#1971c2', marginTop: '4px' }}>
            ₹{(totalRefundedPaise / 100).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
          </div>
        </div>
        <div>
          <div style={{ fontSize: '12px', color: '#8c98a9', textTransform: 'uppercase', fontWeight: 600 }}>Processed Refunds</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#1a1f2c', marginTop: '4px' }}>
            {processedRefundsCount} / {refunds.length}
          </div>
        </div>
        <div>
          <div style={{ fontSize: '12px', color: '#8c98a9', textTransform: 'uppercase', fontWeight: 600 }}>Data Source</div>
          <div style={{ fontSize: '14px', fontWeight: 600, color: '#4263eb', marginTop: '8px', display: 'flex', alignItems: 'center', gap: '4px' }}>
            Live Razorpay REST API
            <ArrowUpRight size={14} />
          </div>
        </div>
      </div>

      {/* Refunds Table */}
      <div style={{
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        border: '1px solid #e9ecef',
        overflow: 'hidden'
      }}>
        {loading ? (
          <div style={{ padding: '32px', textAlign: 'center', color: '#8c98a9' }}>Loading live refunds...</div>
        ) : refunds.length === 0 ? (
          <div style={{ padding: '48px', textAlign: 'center' }}>
            <RotateCcw size={36} color="#adb5bd" style={{ marginBottom: '12px' }} />
            <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#495057' }}>No Refunds Found</h4>
            <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '4px' }}>
              Your Razorpay test account currently has no processed customer refunds.
            </p>
          </div>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13px' }}>
            <thead>
              <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '1px solid #e9ecef', color: '#495057', fontWeight: 600 }}>
                <th style={{ padding: '12px 16px' }}>Refund ID</th>
                <th style={{ padding: '12px 16px' }}>Payment ID</th>
                <th style={{ padding: '12px 16px' }}>Amount (INR)</th>
                <th style={{ padding: '12px 16px' }}>Speed</th>
                <th style={{ padding: '12px 16px' }}>Status</th>
                <th style={{ padding: '12px 16px' }}>Created Date</th>
              </tr>
            </thead>
            <tbody>
              {refunds.map((r, idx) => (
                <tr key={r.id || idx} style={{ borderBottom: '1px solid #f1f3f5' }}>
                  <td style={{ padding: '12px 16px', fontFamily: 'monospace', fontWeight: 600, color: '#4263eb' }}>
                    {r.id}
                  </td>
                  <td style={{ padding: '12px 16px', fontFamily: 'monospace', color: '#495057' }}>
                    {r.payment_id || 'N/A'}
                  </td>
                  <td style={{ padding: '12px 16px', fontWeight: 700, color: '#1971c2' }}>
                    ₹{((r.amount || 0) / 100).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span style={{
                      padding: '3px 6px',
                      borderRadius: '4px',
                      backgroundColor: r.speed === 'optimum' || r.speed === 'instant' ? '#e7f5ff' : '#f8f9fa',
                      color: r.speed === 'optimum' || r.speed === 'instant' ? '#1971c2' : '#495057',
                      fontSize: '11px',
                      fontWeight: 600,
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '3px'
                    }}>
                      <Zap size={10} />
                      {r.speed || 'normal'}
                    </span>
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span style={{
                      padding: '4px 8px',
                      borderRadius: '4px',
                      backgroundColor: r.status === 'processed' ? '#ebfbee' : '#fff9db',
                      color: r.status === 'processed' ? '#2b8a3e' : '#e67700',
                      fontWeight: 600,
                      fontSize: '11px',
                      textTransform: 'uppercase',
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '4px'
                    }}>
                      <CheckCircle2 size={12} />
                      {r.status || 'PROCESSED'}
                    </span>
                  </td>
                  <td style={{ padding: '12px 16px', color: '#8c98a9' }}>
                    {r.created_at ? new Date(r.created_at * 1000).toLocaleString('en-IN') : 'N/A'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
