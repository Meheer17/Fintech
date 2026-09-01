import React, { useState, useEffect } from 'react';
import { Handshake, Calendar, Clock, Plus, CheckCircle2, AlertCircle, XCircle, Filter } from 'lucide-react';
import { fetchPromises, savePromise } from '../lib/api';

export const PromisesTab: React.FC = () => {
  const [promises, setPromises] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'ALL' | 'PENDING' | 'KEPT' | 'BROKEN'>('ALL');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [saving, setSaving] = useState(false);

  // Form state
  const [customerName, setCustomerName] = useState('');
  const [amountINR, setAmountINR] = useState('');
  const [promisedDate, setPromisedDate] = useState('');
  const [workflowId, setWorkflowId] = useState('');
  const [status, setStatus] = useState('PENDING');

  const loadPromises = async () => {
    setLoading(true);
    try {
      const data = await fetchPromises();
      setPromises(data || []);
    } catch (err) {
      console.error('Promises API error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPromises();
  }, []);

  const handleUpdateStatus = async (promiseItem: any, newStatus: string) => {
    try {
      await savePromise({
        id: promiseItem.id,
        customer: promiseItem.customer,
        amount_paise: promiseItem.amount_paise,
        promised_date: promiseItem.promised_date,
        status: newStatus,
        workflow_id: promiseItem.workflow_id
      });
      loadPromises();
    } catch (err) {
      console.error('Failed to update status:', err);
    }
  };

  const handleCreatePromise = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!customerName || !amountINR || !promisedDate) return;
    setSaving(true);
    try {
      await savePromise({
        customer: customerName,
        amount_paise: Math.round(parseFloat(amountINR) * 100),
        promised_date: promisedDate,
        status: status,
        workflow_id: workflowId || `wf_manual_${Date.now().toString().slice(-4)}`
      });
      setIsModalOpen(false);
      setCustomerName('');
      setAmountINR('');
      setPromisedDate('');
      setWorkflowId('');
      setStatus('PENDING');
      loadPromises();
    } catch (err) {
      console.error('Failed to create promise:', err);
    } finally {
      setSaving(false);
    }
  };

  const filteredPromises = promises.filter((p) => {
    if (filter === 'ALL') return true;
    return p.status === filter;
  });

  const totalPromisedPaise = promises.reduce((sum, p) => sum + (p.amount_paise || 0), 0);
  const keptPaise = promises.filter((p) => p.status === 'KEPT').reduce((sum, p) => sum + (p.amount_paise || 0), 0);
  const keptCount = promises.filter((p) => p.status === 'KEPT').length;
  const keptRate = promises.length > 0 ? ((keptCount / promises.length) * 100).toFixed(1) : '0.0';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      {/* Header & Main Action */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c' }}>
            Promise-to-Pay Tracker (Customer Commitments)
          </h3>
          <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '2px' }}>
            Track customer payment commitments negotiated during recovery outreach & Hinglish IVR calls
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
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
            gap: '8px'
          }}
        >
          <Plus size={16} /> Log Customer Promise
        </button>
      </div>

      {/* KPI Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '16px' }}>
        <div style={{ backgroundColor: '#ffffff', padding: '20px', borderRadius: '8px', border: '1px solid #e9ecef' }}>
          <div style={{ fontSize: '12px', fontWeight: 600, color: '#8c98a9', textTransform: 'uppercase' }}>Total Promised Receivables</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#1a1f2c', marginTop: '4px' }}>
            ₹{(totalPromisedPaise / 100).toLocaleString('en-IN')}
          </div>
          <div style={{ fontSize: '12px', color: '#8c98a9', marginTop: '4px' }}>{promises.length} commitments total</div>
        </div>

        <div style={{ backgroundColor: '#ffffff', padding: '20px', borderRadius: '8px', border: '1px solid #e9ecef' }}>
          <div style={{ fontSize: '12px', fontWeight: 600, color: '#8c98a9', textTransform: 'uppercase' }}>Fulfillment / Kept Rate</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#2b8a3e', marginTop: '4px' }}>
            {keptRate}% Kept
          </div>
          <div style={{ fontSize: '12px', color: '#8c98a9', marginTop: '4px' }}>₹{(keptPaise / 100).toLocaleString('en-IN')} recovered</div>
        </div>

        <div style={{ backgroundColor: '#ffffff', padding: '20px', borderRadius: '8px', border: '1px solid #e9ecef' }}>
          <div style={{ fontSize: '12px', fontWeight: 600, color: '#8c98a9', textTransform: 'uppercase' }}>Pending Follow-ups</div>
          <div style={{ fontSize: '24px', fontWeight: 700, color: '#e67700', marginTop: '4px' }}>
            {promises.filter((p) => p.status === 'PENDING').length} Commitments
          </div>
          <div style={{ fontSize: '12px', color: '#8c98a9', marginTop: '4px' }}>Enforcing IST contact window</div>
        </div>
      </div>

      {/* Filter Tabs */}
      <div style={{ display: 'flex', gap: '8px', borderBottom: '1px solid #e9ecef', paddingBottom: '12px' }}>
        {(['ALL', 'PENDING', 'KEPT', 'BROKEN'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setFilter(t)}
            style={{
              padding: '6px 14px',
              borderRadius: '16px',
              fontSize: '12px',
              fontWeight: 600,
              border: '1px solid',
              borderColor: filter === t ? '#4263eb' : '#dee2e6',
              backgroundColor: filter === t ? '#edf2ff' : '#ffffff',
              color: filter === t ? '#4263eb' : '#4a5568'
            }}
          >
            {t === 'ALL' ? 'All Promises' : t}
          </button>
        ))}
      </div>

      {/* Promises Cards Grid */}
      {loading ? (
        <div style={{ padding: '24px', textAlign: 'center', color: '#8c98a9' }}>Loading live promises...</div>
      ) : filteredPromises.length === 0 ? (
        <div style={{ padding: '40px', textAlign: 'center', color: '#8c98a9', backgroundColor: '#ffffff', borderRadius: '8px', border: '1px solid #e9ecef' }}>
          No promises matching active filter.
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '16px' }}>
          {filteredPromises.map((p) => (
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
                  backgroundColor: p.status === 'KEPT' ? '#ebfbee' : (p.status === 'PENDING' ? '#fff9db' : '#fff5f5'),
                  color: p.status === 'KEPT' ? '#2b8a3e' : (p.status === 'PENDING' ? '#e67700' : '#c92a2a'),
                  fontSize: '11px',
                  fontWeight: 700
                }}>
                  ● {p.status}
                </span>
              </div>

              <div style={{ fontSize: '20px', fontWeight: 700, color: '#1a1f2c' }}>
                ₹{((p.amount_paise || 0) / 100).toLocaleString('en-IN')}
              </div>

              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontSize: '12px', color: '#4a5568' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <Calendar size={14} color="#8c98a9" />
                  <span>Promised Date: <strong>{p.promised_date}</strong></span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <Clock size={14} color="#8c98a9" />
                  <span>Workflow: <strong className="mono">{p.workflow_id}</strong></span>
                </div>
              </div>

              {/* Status Action Buttons */}
              <div style={{ display: 'flex', gap: '8px', marginTop: '8px', paddingTop: '12px', borderTop: '1px solid #f1f3f5' }}>
                <button
                  onClick={() => handleUpdateStatus(p, 'KEPT')}
                  disabled={p.status === 'KEPT'}
                  style={{
                    flex: 1,
                    padding: '6px',
                    borderRadius: '4px',
                    border: '1px solid #b2f2bb',
                    backgroundColor: p.status === 'KEPT' ? '#ebfbee' : '#ffffff',
                    color: '#2b8a3e',
                    fontSize: '11px',
                    fontWeight: 600,
                    cursor: p.status === 'KEPT' ? 'default' : 'pointer'
                  }}
                >
                  ✓ Mark Kept
                </button>

                <button
                  onClick={() => handleUpdateStatus(p, 'PENDING')}
                  disabled={p.status === 'PENDING'}
                  style={{
                    flex: 1,
                    padding: '6px',
                    borderRadius: '4px',
                    border: '1px solid #ffe066',
                    backgroundColor: p.status === 'PENDING' ? '#fff9db' : '#ffffff',
                    color: '#e67700',
                    fontSize: '11px',
                    fontWeight: 600,
                    cursor: p.status === 'PENDING' ? 'default' : 'pointer'
                  }}
                >
                  ⏳ Mark Pending
                </button>

                <button
                  onClick={() => handleUpdateStatus(p, 'BROKEN')}
                  disabled={p.status === 'BROKEN'}
                  style={{
                    flex: 1,
                    padding: '6px',
                    borderRadius: '4px',
                    border: '1px solid #ffc9c9',
                    backgroundColor: p.status === 'BROKEN' ? '#fff5f5' : '#ffffff',
                    color: '#c92a2a',
                    fontSize: '11px',
                    fontWeight: 600,
                    cursor: p.status === 'BROKEN' ? 'default' : 'pointer'
                  }}
                >
                  ✕ Mark Broken
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Log Promise Modal */}
      {isModalOpen && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000
        }}>
          <div style={{
            backgroundColor: '#ffffff',
            borderRadius: '8px',
            width: '440px',
            padding: '24px',
            boxShadow: '0 10px 25px rgba(0,0,0,0.15)',
            display: 'flex',
            flexDirection: 'column',
            gap: '16px'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <h3 style={{ fontSize: '16px', fontWeight: 700, color: '#1a1f2c' }}>Log Customer Promise</h3>
              <button onClick={() => setIsModalOpen(false)} style={{ fontSize: '18px', color: '#8c98a9' }}>✕</button>
            </div>

            <form onSubmit={handleCreatePromise} style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              <div>
                <label style={{ fontSize: '12px', fontWeight: 600, color: '#4a5568', display: 'block', marginBottom: '4px' }}>Customer Name</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Vikram Mehta"
                  value={customerName}
                  onChange={(e) => setCustomerName(e.target.value)}
                  style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', border: '1px solid #dee2e6', fontSize: '13px' }}
                />
              </div>

              <div>
                <label style={{ fontSize: '12px', fontWeight: 600, color: '#4a5568', display: 'block', marginBottom: '4px' }}>Promised Amount (INR ₹)</label>
                <input
                  type="number"
                  step="0.01"
                  required
                  placeholder="e.g. 3500"
                  value={amountINR}
                  onChange={(e) => setAmountINR(e.target.value)}
                  style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', border: '1px solid #dee2e6', fontSize: '13px' }}
                />
              </div>

              <div>
                <label style={{ fontSize: '12px', fontWeight: 600, color: '#4a5568', display: 'block', marginBottom: '4px' }}>Promised Date</label>
                <input
                  type="date"
                  required
                  value={promisedDate}
                  onChange={(e) => setPromisedDate(e.target.value)}
                  style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', border: '1px solid #dee2e6', fontSize: '13px' }}
                />
              </div>

              <div>
                <label style={{ fontSize: '12px', fontWeight: 600, color: '#4a5568', display: 'block', marginBottom: '4px' }}>Linked Workflow ID (Optional)</label>
                <input
                  type="text"
                  placeholder="e.g. wf_801"
                  value={workflowId}
                  onChange={(e) => setWorkflowId(e.target.value)}
                  style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', border: '1px solid #dee2e6', fontSize: '13px' }}
                />
              </div>

              <div>
                <label style={{ fontSize: '12px', fontWeight: 600, color: '#4a5568', display: 'block', marginBottom: '4px' }}>Status</label>
                <select
                  value={status}
                  onChange={(e) => setStatus(e.target.value)}
                  style={{ width: '100%', padding: '8px 12px', borderRadius: '6px', border: '1px solid #dee2e6', fontSize: '13px' }}
                >
                  <option value="PENDING">PENDING</option>
                  <option value="KEPT">KEPT</option>
                  <option value="BROKEN">BROKEN</option>
                </select>
              </div>

              <div style={{ display: 'flex', gap: '8px', marginTop: '12px' }}>
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  style={{ flex: 1, padding: '10px', borderRadius: '6px', border: '1px solid #dee2e6', backgroundColor: '#f8f9fa', fontWeight: 600, fontSize: '13px' }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  style={{ flex: 1, padding: '10px', borderRadius: '6px', border: 'none', backgroundColor: '#4263eb', color: '#ffffff', fontWeight: 600, fontSize: '13px' }}
                >
                  {saving ? 'Saving...' : 'Save Promise'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
