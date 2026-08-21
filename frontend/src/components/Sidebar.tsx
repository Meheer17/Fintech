import React from 'react';
import { LayoutDashboard, AlertOctagon, RefreshCw, Handshake, CheckCircle2, FileText, MessageSquare, ShieldCheck, UserCheck } from 'lucide-react';
import { UserRole } from '../types';

interface SidebarProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  role: UserRole;
}

export const Sidebar: React.FC<SidebarProps> = ({ activeTab, setActiveTab, role }) => {
  const getRoleBadge = () => {
    switch (role) {
      case 'razorpay_judge':
        return { label: 'Razorpay Judge / Admin', bg: '#edf2ff', color: '#4263eb' };
      case 'finance_controller':
        return { label: 'Finance Controller', bg: '#ebfbee', color: '#2b8a3e' };
      case 'ops_lead':
        return { label: 'Ops / Support Lead', bg: '#fff9db', color: '#e67700' };
    }
  };

  const roleInfo = getRoleBadge();

  const navItems = [
    { id: 'overview', label: 'Overview', icon: LayoutDashboard },
    { id: 'failures', label: 'Payment Failures', icon: AlertOctagon },
    { id: 'recoveries', label: 'Recovery Workflows', icon: RefreshCw },
    { id: 'promises', label: 'Promise Tracker', icon: Handshake },
    { id: 'reconciliation', label: 'Reconciliation', icon: CheckCircle2 },
    { id: 'audit', label: 'Audit Trail', icon: FileText },
    { id: 'chat', label: 'AI Chat Copilot', icon: MessageSquare },
  ];

  return (
    <aside style={{
      width: '260px',
      height: '100vh',
      backgroundColor: '#ffffff',
      borderRight: '1px solid #e9ecef',
      display: 'flex',
      flexDirection: 'column',
      position: 'fixed',
      left: 0,
      top: 0,
      zIndex: 10
    }}>
      {/* Brand Header */}
      <div style={{ padding: '24px 20px', borderBottom: '1px solid #f1f3f5' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <div style={{
            width: '32px',
            height: '32px',
            borderRadius: '8px',
            backgroundColor: '#4263eb',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
            fontWeight: 700,
            fontSize: '16px'
          }}>
            R
          </div>
          <div>
            <h1 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c', lineHeight: 1.2 }}>RevenueIQ</h1>
            <span style={{ fontSize: '11px', color: '#8c98a9', fontWeight: 500 }}>Razorpay Merchant AI</span>
          </div>
        </div>
      </div>

      {/* Role Badge */}
      <div style={{ padding: '16px 20px', borderBottom: '1px solid #f1f3f5' }}>
        <div style={{ fontSize: '11px', textTransform: 'uppercase', color: '#8c98a9', fontWeight: 600, marginBottom: '6px' }}>
          Active Persona Role
        </div>
        <div style={{
          padding: '6px 10px',
          borderRadius: '6px',
          backgroundColor: roleInfo.bg,
          color: roleInfo.color,
          fontSize: '12px',
          fontWeight: 600,
          display: 'flex',
          alignItems: 'center',
          gap: '6px'
        }}>
          <UserCheck size={14} />
          {roleInfo.label}
        </div>
      </div>

      {/* Nav List */}
      <nav style={{ padding: '16px 12px', flex: 1, overflowY: 'auto' }}>
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = activeTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => setActiveTab(item.id)}
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                gap: '12px',
                padding: '10px 12px',
                borderRadius: '6px',
                marginBottom: '4px',
                fontSize: '14px',
                fontWeight: isActive ? 600 : 500,
                color: isActive ? '#4263eb' : '#4a5568',
                backgroundColor: isActive ? '#edf2ff' : 'transparent',
                textAlign: 'left',
                transition: 'all 0.15s ease'
              }}
            >
              <Icon size={18} color={isActive ? '#4263eb' : '#8c98a9'} />
              {item.label}
            </button>
          );
        })}
      </nav>

      {/* Footer System Status */}
      <div style={{ padding: '16px 20px', borderTop: '1px solid #f1f3f5', fontSize: '12px', color: '#8c98a9' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#2b8a3e', fontWeight: 600 }}>
          <ShieldCheck size={16} />
          <span>Guardrails Enforced</span>
        </div>
        <div style={{ marginTop: '4px', fontSize: '11px' }}>
          Mesh: 13 Microservices Active
        </div>
      </div>
    </aside>
  );
};
