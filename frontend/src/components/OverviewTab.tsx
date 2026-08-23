import React from 'react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { ArrowUpRight, TrendingUp, AlertTriangle, RefreshCw, CheckCircle } from 'lucide-react';
import { FailureRecord } from '../types';

export interface TrendPoint {
  day: string;
  atRisk: number;
  recovered: number;
}

interface OverviewTabProps {
  metrics: {
    total_at_risk_paise: number;
    total_recovered_paise: number;
    recovery_rate: number;
    active_workflows: number;
    reconciliation_match: number;
    failed_transactions_count?: number;
    trend_data?: TrendPoint[];
  };
  failures?: FailureRecord[];
}

export const OverviewTab: React.FC<OverviewTabProps> = ({ metrics, failures = [] }) => {
  const daysOfWeek = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

  let trendData: TrendPoint[] = [];

  if (Array.isArray(metrics.trend_data) && metrics.trend_data.length > 0) {
    trendData = metrics.trend_data;
  } else if (Array.isArray(failures) && failures.length > 0) {
    const computedMap: Record<string, { atRisk: number; recovered: number }> = {};
    daysOfWeek.forEach((day) => {
      computedMap[day] = { atRisk: 0, recovered: 0 };
    });

    failures.forEach((f) => {
      if (!f.failed_at) return;
      const dt = new Date(f.failed_at);
      if (isNaN(dt.getTime())) return;
      const dayIdx = (dt.getDay() + 6) % 7;
      const dayName = daysOfWeek[dayIdx];
      const amt = (f.amount_paise || 0) / 100;
      computedMap[dayName].atRisk += amt;
      if (f.recovery_status === 'RECOVERED') {
        computedMap[dayName].recovered += amt;
      }
    });

    trendData = daysOfWeek.map((day) => ({
      day,
      atRisk: computedMap[day].atRisk,
      recovered: computedMap[day].recovered,
    }));
  } else {
    trendData = daysOfWeek.map((day) => ({ day, atRisk: 0, recovered: 0 }));
  }

  const failedCount = metrics.failed_transactions_count ?? (failures ? failures.length : 0);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      {/* Metric Cards Row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px' }}>
        {/* Card 1: Revenue at Risk */}
        <div style={{
          backgroundColor: '#ffffff',
          borderRadius: '8px',
          border: '1px solid #e9ecef',
          padding: '20px',
          boxShadow: '0 1px 2px rgba(0,0,0,0.04)'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span style={{ fontSize: '12px', fontWeight: 600, color: '#8c98a9', textTransform: 'uppercase' }}>
              Revenue at Risk
            </span>
            <AlertTriangle size={18} color="#c92a2a" />
          </div>
          <div style={{ fontSize: '26px', fontWeight: 700, color: '#1a1f2c', marginTop: '8px' }}>
            ₹{(metrics.total_at_risk_paise / 100).toLocaleString('en-IN')}
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '12px', color: '#c92a2a', marginTop: '6px' }}>
            <span>{failedCount} failed transactions detected</span>
          </div>
        </div>

        {/* Card 2: Total Recovered */}
        <div style={{
          backgroundColor: '#ffffff',
          borderRadius: '8px',
          border: '1px solid #e9ecef',
          padding: '20px',
          boxShadow: '0 1px 2px rgba(0,0,0,0.04)'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span style={{ fontSize: '12px', fontWeight: 600, color: '#8c98a9', textTransform: 'uppercase' }}>
              Total Recovered
            </span>
            <TrendingUp size={18} color="#2b8a3e" />
          </div>
          <div style={{ fontSize: '26px', fontWeight: 700, color: '#2b8a3e', marginTop: '8px' }}>
            ₹{(metrics.total_recovered_paise / 100).toLocaleString('en-IN')}
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '12px', color: '#2b8a3e', marginTop: '6px' }}>
            <ArrowUpRight size={14} />
            <span>{(metrics.recovery_rate * 100).toFixed(1)}% recovery rate</span>
          </div>
        </div>

        {/* Card 3: Active Workflows */}
        <div style={{
          backgroundColor: '#ffffff',
          borderRadius: '8px',
          border: '1px solid #e9ecef',
          padding: '20px',
          boxShadow: '0 1px 2px rgba(0,0,0,0.04)'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span style={{ fontSize: '12px', fontWeight: 600, color: '#8c98a9', textTransform: 'uppercase' }}>
              Active Workflows
            </span>
            <RefreshCw size={18} color="#4263eb" />
          </div>
          <div style={{ fontSize: '26px', fontWeight: 700, color: '#1a1f2c', marginTop: '8px' }}>
            {metrics.active_workflows}
          </div>
          <div style={{ fontSize: '12px', color: '#4a5568', marginTop: '6px' }}>
            Bounded DAG state machines
          </div>
        </div>

        {/* Card 4: Reconciliation Match Rate */}
        <div style={{
          backgroundColor: '#ffffff',
          borderRadius: '8px',
          border: '1px solid #e9ecef',
          padding: '20px',
          boxShadow: '0 1px 2px rgba(0,0,0,0.04)'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span style={{ fontSize: '12px', fontWeight: 600, color: '#8c98a9', textTransform: 'uppercase' }}>
              Reconciliation Match
            </span>
            <CheckCircle size={18} color="#1971c2" />
          </div>
          <div style={{ fontSize: '26px', fontWeight: 700, color: '#1a1f2c', marginTop: '8px' }}>
            {(metrics.reconciliation_match * 100).toFixed(1)}%
          </div>
          <div style={{ fontSize: '12px', color: '#1971c2', marginTop: '6px' }}>
            Three-Way Match (Orders ↔ Settlements)
          </div>
        </div>
      </div>

      {/* Chart & Summary Row */}
      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '16px' }}>
        <div style={{
          backgroundColor: '#ffffff',
          borderRadius: '8px',
          border: '1px solid #e9ecef',
          padding: '24px',
          boxShadow: '0 1px 2px rgba(0,0,0,0.04)'
        }}>
          <h3 style={{ fontSize: '15px', fontWeight: 600, color: '#1a1f2c', marginBottom: '16px' }}>
            Weekly Revenue Recovery Performance
          </h3>
          <div style={{ width: '100%', height: '240px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={trendData}>
                <XAxis dataKey="day" stroke="#8c98a9" fontSize={12} tickLine={false} />
                <YAxis stroke="#8c98a9" fontSize={12} tickLine={false} />
                <Tooltip formatter={(value: number) => [`₹${(value).toLocaleString('en-IN')}`, 'Amount']} />
                <Area type="monotone" dataKey="atRisk" stroke="#c92a2a" fill="#fff5f5" name="At Risk" />
                <Area type="monotone" dataKey="recovered" stroke="#2b8a3e" fill="#ebfbee" name="Recovered" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* System Health Panel */}
        <div style={{
          backgroundColor: '#ffffff',
          borderRadius: '8px',
          border: '1px solid #e9ecef',
          padding: '24px',
          boxShadow: '0 1px 2px rgba(0,0,0,0.04)'
        }}>
          <h3 style={{ fontSize: '15px', fontWeight: 600, color: '#1a1f2c', marginBottom: '16px' }}>
            Engine Health & Guardrails
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', fontSize: '13px' }}>
            <div style={{ padding: '10px 12px', borderRadius: '6px', backgroundColor: '#f8f9fa', border: '1px solid #e9ecef' }}>
              <div style={{ fontWeight: 600, color: '#1a1f2c' }}>Circuit Breaker Status</div>
              <div style={{ color: '#2b8a3e', fontSize: '12px', marginTop: '2px' }}>CLOSED (Healthy — 0 failures/min)</div>
            </div>
            <div style={{ padding: '10px 12px', borderRadius: '6px', backgroundColor: '#f8f9fa', border: '1px solid #e9ecef' }}>
              <div style={{ fontWeight: 600, color: '#1a1f2c' }}>Contact Window Guardrail</div>
              <div style={{ color: '#4263eb', fontSize: '12px', marginTop: '2px' }}>ACTIVE (09:00 - 21:00 IST Enforced)</div>
            </div>
            <div style={{ padding: '10px 12px', borderRadius: '6px', backgroundColor: '#f8f9fa', border: '1px solid #e9ecef' }}>
              <div style={{ fontWeight: 600, color: '#1a1f2c' }}>Recovery Cost Cap</div>
              <div style={{ color: '#e67700', fontSize: '12px', marginTop: '2px' }}>HARD CAP (Max 20% of Payment Value)</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
