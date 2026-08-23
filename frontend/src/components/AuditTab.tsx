import React, { useState } from 'react';
import { AuditEntry } from '../types';
import { Search, ChevronDown, ChevronUp, ShieldCheck } from 'lucide-react';

interface AuditTabProps {
  logs?: AuditEntry[];
}

export const AuditTab: React.FC<AuditTabProps> = ({ logs = [] }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const safeLogs = Array.isArray(logs) ? logs : [];
  const [expandedId, setExpandedId] = useState<string | null>(safeLogs[0]?.id || null);

  const filtered = safeLogs.filter(
    (l) =>
      (l.service_name || '').toLowerCase().includes(searchTerm.toLowerCase()) ||
      (l.action || '').toLowerCase().includes(searchTerm.toLowerCase()) ||
      (l.entity_id || '').toLowerCase().includes(searchTerm.toLowerCase()) ||
      (l.reasoning || '').toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c' }}>
            Centralized System Audit Trail (gRPC Log Store)
          </h3>
          <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '2px' }}>
            Immutable event history from audit_service recording every AI diagnosis, guardrail evaluation, and action execution
          </p>
        </div>

        {/* Search Input */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          padding: '8px 14px',
          backgroundColor: '#ffffff',
          border: '1px solid #e9ecef',
          borderRadius: '6px',
          width: '320px'
        }}>
          <Search size={16} color="#8c98a9" />
          <input
            type="text"
            placeholder="Search service, entity ID, or reasoning..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            style={{ border: 'none', outline: 'none', width: '100%', fontSize: '13px' }}
          />
        </div>
      </div>

      {/* Audit Log Table */}
      <div style={{
        backgroundColor: '#ffffff',
        borderRadius: '8px',
        border: '1px solid #e9ecef',
        boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
        overflow: 'hidden'
      }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px', textAlign: 'left' }}>
          <thead>
            <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '1px solid #e9ecef', color: '#8c98a9', fontWeight: 600 }}>
              <th style={{ padding: '12px 16px' }}>Timestamp (UTC)</th>
              <th style={{ padding: '12px 16px' }}>Service</th>
              <th style={{ padding: '12px 16px' }}>Action</th>
              <th style={{ padding: '12px 16px' }}>Target Entity</th>
              <th style={{ padding: '12px 16px' }}>Status</th>
              <th style={{ padding: '12px 16px' }}>Details</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ padding: '24px', textAlign: 'center', color: '#8c98a9' }}>
                  No audit log entries found.
                </td>
              </tr>
            ) : (
              filtered.map((log, idx) => {
                const isExpanded = expandedId === log.id;
                const isBlocked = log.status === 'BLOCKED';
                const guardrails = Array.isArray(log.guardrails_checked) ? log.guardrails_checked : [];

                return (
                  <React.Fragment key={log.id || idx}>
                    <tr
                      onClick={() => setExpandedId(isExpanded ? null : log.id)}
                      style={{
                        borderBottom: '1px solid #f1f3f5',
                        backgroundColor: isExpanded ? '#edf2ff' : (isBlocked ? '#fff5f5' : '#ffffff'),
                        cursor: 'pointer'
                      }}
                    >
                      <td style={{ padding: '12px 16px', color: '#8c98a9', fontSize: '12px' }} className="mono">
                        {log.timestamp || 'N/A'}
                      </td>
                      <td style={{ padding: '12px 16px', fontWeight: 600, color: '#4263eb' }} className="mono">
                        {log.service_name}
                      </td>
                      <td style={{ padding: '12px 16px', fontWeight: 600, color: '#1a1f2c' }}>
                        {log.action}
                      </td>
                      <td style={{ padding: '12px 16px' }} className="mono">
                        {log.entity_type} ({log.entity_id})
                      </td>
                      <td style={{ padding: '12px 16px' }}>
                        <span style={{
                          padding: '3px 8px',
                          borderRadius: '12px',
                          backgroundColor: isBlocked ? '#fff5f5' : '#ebfbee',
                          color: isBlocked ? '#c92a2a' : '#2b8a3e',
                          fontSize: '11px',
                          fontWeight: 700
                        }}>
                          {log.status}
                        </span>
                      </td>
                      <td style={{ padding: '12px 16px', color: '#8c98a9' }}>
                        {isExpanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
                      </td>
                    </tr>

                    {/* Expanded Detail Panel */}
                    {isExpanded && (
                      <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '1px solid #e9ecef' }}>
                        <td colSpan={6} style={{ padding: '16px 20px' }}>
                          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '12px' }}>
                            <div>
                              <span style={{ fontWeight: 600, color: '#8c98a9' }}>Reasoning & Diagnosis: </span>
                              <span style={{ color: '#1a1f2c', fontWeight: 500 }}>"{log.reasoning || 'N/A'}"</span>
                            </div>
                            <div>
                              <span style={{ fontWeight: 600, color: '#8c98a9' }}>Actor / Agent: </span>
                              <span className="mono" style={{ color: '#4263eb' }}>{log.actor || 'system'}</span>
                            </div>
                            <div style={{ display: 'flex', gap: '6px', alignItems: 'center', marginTop: '4px' }}>
                              <ShieldCheck size={14} color="#2b8a3e" />
                              <span style={{ fontWeight: 600, color: '#8c98a9' }}>Guardrails Verified: </span>
                              {guardrails.length === 0 ? (
                                <span style={{ color: '#8c98a9', fontSize: '11px' }}>None</span>
                              ) : (
                                guardrails.map((g, gIdx) => (
                                  <span key={g + gIdx} style={{ padding: '2px 6px', borderRadius: '4px', backgroundColor: '#ffffff', border: '1px solid #dee2e6', fontSize: '11px' }}>
                                    {g}
                                  </span>
                                ))
                              )}
                            </div>
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
