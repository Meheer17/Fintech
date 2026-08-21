import React, { useState, useEffect } from 'react';
import { Handshake, Calendar, Clock } from 'lucide-react';
import { fetchPromises } from '../lib/api';

export const PromisesTab: React.FC = () => {
  const [promises, setPromises] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchPromises()
      .then((data) => setPromises(data))
      .catch((err) => console.error('Promises API error:', err))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      <div>
        <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c' }}>
          Promise-to-Pay Tracker (Customer Commitments)
        </h3>
        <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '2px' }}>
          Track customer payment commitments, promised dates, and follow-up compliance schedule
        </p>
      </div>

      {loading ? (
        <div style={{ padding: '24px', textAlign: 'center', color: '#8c98a9' }}>Loading live promises...</div>
      ) : promises.length === 0 ? (
        <div style={{ padding: '24px', textAlign: 'center', color: '#8c98a9' }}>No active promises found.</div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '16px' }}>
          {promises.map((p) => (
            <div key={p.id} style={{
              backgroundColor: '#ffffff',
              borderRadius: '8px',
              border: '1px solid #e9ecef',
              padding: '20px',
              boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
              display: 'flex',
              flexDirection: 'column',
              gap: '12px'
            }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <Handshake size={20} color="#4263eb" />
                  <span style={{ fontSize: '15px', fontWeight: 600, color: '#1a1f2c' }}>{p.customer}</span>
                </div>
                <span style={{
                  padding: '4px 10px',
                  borderRadius: '12px',
                  backgroundColor: p.status === 'KEPT' ? '#ebfbee' : '#fff9db',
                  color: p.status === 'KEPT' ? '#2b8a3e' : '#e67700',
                  fontSize: '11px',
                  fontWeight: 700
                }}>
                  ● {p.status}
                </span>
              </div>

              <div style={{ fontSize: '20px', fontWeight: 700, color: '#1a1f2c' }}>
                ₹{((p.amount_paise || 0) / 100).toLocaleString('en-IN')}
              </div>

              <div style={{ display: 'flex', alignItems: 'center', gap: '16px', fontSize: '12px', color: '#4a5568' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <Calendar size={14} color="#8c98a9" />
                  <span>Promised Date: <strong>{p.promised_date}</strong></span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <Clock size={14} color="#8c98a9" />
                  <span>Workflow: <strong className="mono">{p.workflow_id}</strong></span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
