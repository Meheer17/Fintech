import React, { useState, useEffect } from 'react';
import { Sparkles, RefreshCw } from 'lucide-react';
import { fetchReconciliation, triggerReconciliationBatch } from '../lib/api';

export const ReconciliationTab: React.FC = () => {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [reconciling, setReconciling] = useState(false);

  const loadReconciliationData = async () => {
    setLoading(true);
    try {
      const res = await fetchReconciliation();
      setData(res);
    } catch (e) {
      console.error('Reconciliation load error:', e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadReconciliationData();
  }, []);

  const handleRunReconciliation = async () => {
    setReconciling(true);
    try {
      await triggerReconciliationBatch([]);
      await loadReconciliationData();
    } catch (e) {
      console.error('Reconciliation error:', e);
    } finally {
      setReconciling(false);
    }
  };

  const total = data?.total_records ?? 0;
  const exact = data?.exact_matches ?? 0;
  const fuzzy = data?.fuzzy_matches ?? 0;
  const ai = data?.ai_matches ?? 0;
  const unmatched = data?.unmatched ?? 0;
  const matchRatePct = (((data?.match_rate ?? 0)) * 100).toFixed(1);
  const exceptions = data?.exceptions || [];

  const exactPct = total > 0 ? (exact / total) * 100 : 0;
  const fuzzyPct = total > 0 ? (fuzzy / total) * 100 : 0;
  const aiPct = total > 0 ? (ai / total) * 100 : 0;
  const unmatchedPct = total > 0 ? (unmatched / total) * 100 : 0;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c' }}>
            Settlement Reconciliation Engine (Track 04)
          </h3>
          <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '2px' }}>
            Three-way automated matching across Merchant Orders ↔ Razorpay Payments ↔ Bank Settlements
          </p>
        </div>
        <button
          onClick={handleRunReconciliation}
          disabled={reconciling}
          style={{
            padding: '10px 16px',
            backgroundColor: '#2b8a3e',
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
          <RefreshCw size={16} /> {reconciling ? 'Running Batch Reconciler...' : 'Run Live Batch Reconciliation'}
        </button>
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
            <div style={{ fontSize: '28px', fontWeight: 700, color: '#2b8a3e', marginTop: '2px' }}>
              {loading ? 'Loading...' : `${matchRatePct}% Auto-Matched`}
            </div>
          </div>
          <div style={{ fontSize: '13px', textAlign: 'right' }}>
            <div>Total Records: <strong>{total} Records</strong></div>
            <div style={{ color: '#8c98a9', fontSize: '12px' }}>Live Engine Batch Run</div>
          </div>
        </div>

        {/* Visual Progress Bar */}
        <div style={{ height: '12px', borderRadius: '6px', backgroundColor: '#e9ecef', overflow: 'hidden', display: 'flex' }}>
          <div style={{ width: `${exactPct}%`, backgroundColor: '#2b8a3e' }} title={`Exact Matches (${exact})`} />
          <div style={{ width: `${fuzzyPct}%`, backgroundColor: '#1971c2' }} title={`Fuzzy Matches (${fuzzy})`} />
          <div style={{ width: `${aiPct}%`, backgroundColor: '#e67700' }} title={`AI Matches (${ai})`} />
          <div style={{ width: `${unmatchedPct}%`, backgroundColor: '#c92a2a' }} title={`Unmatched Exceptions (${unmatched})`} />
        </div>

        <div style={{ display: 'flex', gap: '24px', fontSize: '12px', color: '#4a5568' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ width: '10px', height: '10px', borderRadius: '2px', backgroundColor: '#2b8a3e' }} />
            <span>Exact Matches: <strong>{exact} ({exactPct.toFixed(0)}%)</strong></span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ width: '10px', height: '10px', borderRadius: '2px', backgroundColor: '#1971c2' }} />
            <span>Fuzzy Matches: <strong>{fuzzy} ({fuzzyPct.toFixed(0)}%)</strong></span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ width: '10px', height: '10px', borderRadius: '2px', backgroundColor: '#e67700' }} />
            <span>AI Resolved: <strong>{ai} ({aiPct.toFixed(0)}%)</strong></span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <span style={{ width: '10px', height: '10px', borderRadius: '2px', backgroundColor: '#c92a2a' }} />
            <span>Exceptions: <strong>{unmatched} ({unmatchedPct.toFixed(0)}%)</strong></span>
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
            {exceptions.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ padding: '24px', textAlign: 'center', color: '#8c98a9' }}>
                  No exceptions reported in current reconciliation batch.
                </td>
              </tr>
            ) : (
              exceptions.map((exc: any) => (
                <tr key={exc.id} style={{ borderBottom: '1px solid #f1f3f5' }}>
                  <td style={{ padding: '12px 16px', fontWeight: 600 }} className="mono">{exc.id}</td>
                  <td style={{ padding: '12px 16px', fontWeight: 600, color: '#e67700' }}>{exc.type}</td>
                  <td style={{ padding: '12px 16px' }} className="mono">{exc.order_id}</td>
                  <td style={{ padding: '12px 16px' }}>
                    ₹{((exc.expected || 0) / 100).toFixed(2)} / ₹{((exc.actual || 0) / 100).toFixed(2)}
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
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
