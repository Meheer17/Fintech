import React, { useEffect, useState } from 'react';
import { fetchSubscriptions } from '../lib/api';
import { Layers, CheckCircle2 } from 'lucide-react';

export const SubscriptionsTab: React.FC = () => {
  const [subscriptions, setSubscriptions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchSubscriptions()
      .then((data) => {
        setSubscriptions(data.subscriptions || []);
        setLoading(false);
      })
      .catch((err) => {
        console.error('Failed to load subscriptions:', err);
        setLoading(false);
      });
  }, []);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h2 style={{ fontSize: '20px', fontWeight: 700, color: '#1a1f2c' }}>Razorpay Subscriptions</h2>
          <p style={{ fontSize: '13px', color: '#6c757d', marginTop: '4px' }}>
            Live subscription billing profiles synced from your Razorpay test account (`/v1/subscriptions`)
          </p>
        </div>
        <div style={{
          backgroundColor: '#edf2ff',
          color: '#4263eb',
          padding: '8px 14px',
          borderRadius: '6px',
          fontWeight: 600,
          fontSize: '13px',
          display: 'flex',
          alignItems: 'center',
          gap: '6px'
        }}>
          <Layers size={16} />
          {subscriptions.length} Subscriptions Active
        </div>
      </div>

      {/* Subscriptions Content */}
      <div style={{
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        border: '1px solid #e9ecef',
        overflow: 'hidden'
      }}>
        {loading ? (
          <div style={{ padding: '32px', textAlign: 'center', color: '#8c98a9' }}>Loading live subscriptions...</div>
        ) : subscriptions.length === 0 ? (
          <div style={{ padding: '48px', textAlign: 'center' }}>
            <Layers size={36} color="#adb5bd" style={{ marginBottom: '12px' }} />
            <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#495057' }}>No Subscriptions Found</h4>
            <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '4px', maxWidth: '400px', margin: '4px auto 0' }}>
              Your Razorpay test account currently has 0 subscriptions. Zero fake data is generated.
            </p>
          </div>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '13px' }}>
            <thead>
              <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '1px solid #e9ecef', color: '#495057', fontWeight: 600 }}>
                <th style={{ padding: '12px 16px' }}>Subscription ID</th>
                <th style={{ padding: '12px 16px' }}>Plan ID</th>
                <th style={{ padding: '12px 16px' }}>Customer ID</th>
                <th style={{ padding: '12px 16px' }}>Status</th>
                <th style={{ padding: '12px 16px' }}>Created Date</th>
              </tr>
            </thead>
            <tbody>
              {subscriptions.map((s, idx) => (
                <tr key={s.id || idx} style={{ borderBottom: '1px solid #f1f3f5' }}>
                  <td style={{ padding: '12px 16px', fontFamily: 'monospace', fontWeight: 600, color: '#4263eb' }}>
                    {s.id}
                  </td>
                  <td style={{ padding: '12px 16px', fontFamily: 'monospace', color: '#495057' }}>
                    {s.plan_id || 'N/A'}
                  </td>
                  <td style={{ padding: '12px 16px', color: '#495057' }}>
                    {s.customer_id || 'N/A'}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span style={{
                      padding: '4px 8px',
                      borderRadius: '4px',
                      backgroundColor: '#ebfbee',
                      color: '#2b8a3e',
                      fontWeight: 600,
                      fontSize: '11px',
                      textTransform: 'uppercase'
                    }}>
                      {s.status || 'active'}
                    </span>
                  </td>
                  <td style={{ padding: '12px 16px', color: '#8c98a9' }}>
                    {s.created_at ? new Date(s.created_at * 1000).toLocaleString('en-IN') : 'N/A'}
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
