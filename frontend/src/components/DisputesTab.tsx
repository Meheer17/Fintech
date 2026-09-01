import React, { useEffect, useState } from 'react';
import { fetchDisputes } from '../lib/api';
import { ShieldAlert, ArrowUpRight, AlertTriangle, CheckCircle2, Clock } from 'lucide-react';

export const DisputesTab: React.FC = () => {
  const [disputes, setDisputes] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDisputes()
      .then((data) => {
        setDisputes(data.disputes || []);
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to load disputes:', err);
        setLoading(false);
      });
  }, []);

  const totalDisputedPaise = disputes.reduce((acc, item) => acc + (item.amount_paise || item.amount || 0), 0);
  const openDisputesCount = disputes.filter((d) => d.status === 'open' || d.status === 'under_review' || d.status === 'needs_response').length;
  const closedDisputesCount = disputes.filter((d) => d.status === 'won' || d.status === 'lost' || d.status === 'closed').length;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '20px', fontWeight: 700, color: '#1a1f2c' }}>Razorpay Chargebacks & Disputes</h2>
          <p style={{ fontSize: '13px', color: '#6c757d', marginTop: '4px' }}>
            Live customer chargebacks and payment disputes synced from Razorpay (`/v1/disputes`)
          </p>
        </div>
        <div style={{
          backgroundColor: disputes.length > 0 ? '#fff5f5' : '#ebfbee',
          color: disputes.length > 0 ? '#c92a2a' : '#2b8a3e',
          padding: '8px 14px',
          borderRadius: '6px',
          fontWeight: 600,
          fontSize: '13px',
          display: 'flex',
          alignItems: 'center',
          gap: '6px'
        }}>
          <ShieldAlert size={16} />
          {disputes.length} Disputes Synced
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
          <div style={{ fontSize: '12px', color: '#8c98a9', textTransform: 'uppercase', fontWeight: 600 }}>Total Disputed Amount</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#c92a2a', marginTop: '4px' }}>
            ₹{(totalDisputedPaise / 100).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
          </div>
        </div>
        <div>
          <div style={{ fontSize: '12px', color: '#8c98a9', textTransform: 'uppercase', fontWeight: 600 }}>Open / Closed Status</div>
          <div style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c', marginTop: '6px', display: 'flex', gap: '12px' }}>
            <span style={{ color: '#e67700' }}>{openDisputesCount} Open</span>
            <span style={{ color: '#8c98a9' }}>|</span>
            <span style={{ color: '#2b8a3e' }}>{closedDisputesCount} Closed</span>
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

      {/* Disputes Table */}
      <div style={{
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        border: '1px solid #e9ecef',
        overflow: 'hidden'
      }}>
        {loading ? (
          <div style={{ padding: '32px', textAlign: 'center', color: '#8c98a9' }}>Loading live disputes...</div>
        ) : disputes.length === 0 ? (
          <div style={{ padding: '48px', textAlign: 'center' }}>
            <ShieldAlert size={36} color="#adb5bd" style={{ marginBottom: '12px' }} />
            <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#495057' }}>No Disputes Found</h4>
            <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '4px' }}>
              Your Razorpay account currently has no active chargebacks or customer disputes.
            </p>
          </div>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13px' }}>
            <thead>
              <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '1px solid #e9ecef', color: '#495057', fontWeight: 600 }}>
                <th style={{ padding: '12px 16px' }}>Dispute ID</th>
                <th style={{ padding: '12px 16px' }}>Payment ID</th>
                <th style={{ padding: '12px 16px' }}>Amount (INR)</th>
                <th style={{ padding: '12px 16px' }}>Reason / Phase</th>
                <th style={{ padding: '12px 16px' }}>Status</th>
                <th style={{ padding: '12px 16px' }}>Respond By</th>
              </tr>
            </thead>
            <tbody>
              {disputes.map((d, idx) => (
                <tr key={d.id || idx} style={{ borderBottom: '1px solid #f1f3f5' }}>
                  <td style={{ padding: '12px 16px', fontFamily: 'monospace', fontWeight: 600, color: '#4263eb' }}>
                    {d.id}
                  </td>
                  <td style={{ padding: '12px 16px', fontFamily: 'monospace', color: '#495057' }}>
                    {d.payment_id || 'N/A'}
                  </td>
                  <td style={{ padding: '12px 16px', fontWeight: 700, color: '#c92a2a' }}>
                    ₹{((d.amount || 0) / 100).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
                  </td>
                  <td style={{ padding: '12px 16px', color: '#495057' }}>
                    {d.reason_code || d.phase || d.reason || 'General Dispute'}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span style={{
                      padding: '4px 8px',
                      borderRadius: '4px',
                      backgroundColor: d.status === 'won' ? '#ebfbee' : (d.status === 'lost' ? '#fff5f5' : '#fff9db'),
                      color: d.status === 'won' ? '#2b8a3e' : (d.status === 'lost' ? '#c92a2a' : '#e67700'),
                      fontWeight: 600,
                      fontSize: '11px',
                      textTransform: 'uppercase',
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '4px'
                    }}>
                      {d.status === 'won' ? <CheckCircle2 size={12} /> : (d.status === 'lost' ? <AlertTriangle size={12} /> : <Clock size={12} />)}
                      {d.status || 'OPEN'}
                    </span>
                  </td>
                  <td style={{ padding: '12px 16px', color: '#8c98a9' }}>
                    {d.respond_by ? new Date(d.respond_by * 1000).toLocaleString('en-IN') : (d.created_at ? new Date(d.created_at * 1000).toLocaleString('en-IN') : 'N/A')}
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
