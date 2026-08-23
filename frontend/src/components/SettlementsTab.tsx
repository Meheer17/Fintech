import React, { useEffect, useState } from 'react';
import { fetchSettlements } from '../lib/api';
import { Landmark, ArrowUpRight, CheckCircle2, Clock } from 'lucide-react';

export const SettlementsTab: React.FC = () => {
  const [settlements, setSettlements] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchSettlements()
      .then((data) => {
        setSettlements(data.settlements || []);
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to load settlements:', err);
        setLoading(false);
      });
  }, []);

  // Filter out ₹0.00 / blank settlements
  const validSettlements = settlements.filter((s) => {
    const amt = s.amount_paise || s.amount || 0;
    return amt > 0;
  });

  const processedSettlements = validSettlements.filter(
    (s) => s.status === 'processed' || s.status === 'SUCCESS' || s.status === 'PROCESSED'
  );

  const totalSettledPaise = processedSettlements.reduce(
    (acc, item) => acc + (item.amount_paise || item.amount || 0),
    0
  );

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '20px', fontWeight: 700, color: '#1a1f2c' }}>Razorpay Bank Settlements</h2>
          <p style={{ fontSize: '13px', color: '#6c757d', marginTop: '4px' }}>
            Live payouts synced from your Razorpay account (`/v1/settlements`)
          </p>
        </div>
        <div style={{
          backgroundColor: '#ebfbee',
          color: '#2b8a3e',
          padding: '8px 14px',
          borderRadius: '6px',
          fontWeight: 600,
          fontSize: '13px',
          display: 'flex',
          alignItems: 'center',
          gap: '6px'
        }}>
          <Landmark size={16} />
          {validSettlements.length} Valid Payout Records
        </div>
      </div>

      {/* Summary Card */}
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
          <div style={{ fontSize: '12px', color: '#8c98a9', textTransform: 'uppercase', fontWeight: 600 }}>Total Settled Amount</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#2b8a3e', marginTop: '4px' }}>
            ₹{(totalSettledPaise / 100).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </div>
        </div>
        <div>
          <div style={{ fontSize: '12px', color: '#8c98a9', textTransform: 'uppercase', fontWeight: 600 }}>Processed Payouts</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#1a1f2c', marginTop: '4px' }}>
            {processedSettlements.length}
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

      {/* Settlements Table */}
      <div style={{
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        border: '1px solid #e9ecef',
        overflow: 'hidden'
      }}>
        {loading ? (
          <div style={{ padding: '32px', textAlign: 'center', color: '#8c98a9' }}>Loading live settlements...</div>
        ) : validSettlements.length === 0 ? (
          <div style={{ padding: '48px', textAlign: 'center' }}>
            <Landmark size={36} color="#adb5bd" style={{ marginBottom: '12px' }} />
            <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#495057' }}>No Valid Settlements Found</h4>
            <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '4px' }}>
              Your Razorpay test account currently has no active bank settlements.
            </p>
          </div>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13px' }}>
            <thead>
              <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '1px solid #e9ecef', color: '#495057', fontWeight: 600 }}>
                <th style={{ padding: '12px 16px' }}>Settlement ID</th>
                <th style={{ padding: '12px 16px' }}>UTR Number</th>
                <th style={{ padding: '12px 16px' }}>Amount (INR)</th>
                <th style={{ padding: '12px 16px' }}>Status</th>
                <th style={{ padding: '12px 16px' }}>Failure Reason / Details</th>
                <th style={{ padding: '12px 16px' }}>Created Date</th>
              </tr>
            </thead>
            <tbody>
              {validSettlements.map((s, idx) => {
                const amtPaise = s.amount_paise || s.amount || 0;
                const isFailed = s.status === 'failed' || s.status === 'FAILED';
                const isProcessed = s.status === 'processed' || s.status === 'SUCCESS' || s.status === 'PROCESSED';
                const failureReason = s.error_description || s.failure_reason || s.reason || (isFailed ? 'Bank account decline / invalid beneficiary details' : 'N/A');

                return (
                  <tr key={s.id || s.settlement_id || idx} style={{ borderBottom: '1px solid #f1f3f5', backgroundColor: isFailed ? '#fff5f5' : '#ffffff' }}>
                    <td style={{ padding: '12px 16px', fontFamily: 'monospace', fontWeight: 600, color: '#4263eb' }}>
                      {s.id || s.settlement_id || `setl_live_${idx+1}`}
                    </td>
                    <td style={{ padding: '12px 16px', fontFamily: 'monospace', color: '#495057' }}>
                      {s.utr || 'N/A'}
                    </td>
                    <td style={{ padding: '12px 16px', fontWeight: 700, color: isFailed ? '#c92a2a' : '#2b8a3e' }}>
                      ₹{(amtPaise / 100).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                    </td>
                    <td style={{ padding: '12px 16px' }}>
                      <span style={{
                        padding: '4px 8px',
                        borderRadius: '4px',
                        backgroundColor: isProcessed ? '#ebfbee' : (isFailed ? '#fff5f5' : '#fff9db'),
                        color: isProcessed ? '#2b8a3e' : (isFailed ? '#c92a2a' : '#e67700'),
                        fontWeight: 600,
                        fontSize: '11px',
                        textTransform: 'uppercase',
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '4px'
                      }}>
                        {isProcessed ? <CheckCircle2 size={12} /> : <Clock size={12} />}
                        {s.status}
                      </span>
                    </td>
                    <td style={{ padding: '12px 16px', color: isFailed ? '#c92a2a' : '#495057', fontSize: '12px' }}>
                      {isFailed ? (
                        <span style={{ fontWeight: 600 }}>⚠️ {failureReason}</span>
                      ) : (
                        <span style={{ color: '#2b8a3e' }}>✓ Credited via UTR to bank account</span>
                      )}
                    </td>
                    <td style={{ padding: '12px 16px', color: '#8c98a9' }}>
                      {s.created_at ? (typeof s.created_at === 'number' ? new Date(s.created_at * 1000).toLocaleString('en-IN') : s.created_at) : 'N/A'}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
