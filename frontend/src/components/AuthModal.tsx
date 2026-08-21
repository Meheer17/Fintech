import React from 'react';
import { UserRole, UserProfile } from '../types';
import { X, CheckCircle, Sparkles } from 'lucide-react';

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentUser: UserProfile;
  onSelectRole: (user: UserProfile) => void;
}

export const AuthModal: React.FC<AuthModalProps> = ({ isOpen, onClose, currentUser, onSelectRole }) => {
  if (!isOpen) return null;

  const roles: { role: UserRole; title: string; desc: string; user: UserProfile }[] = [
    {
      role: 'razorpay_judge',
      title: '🏆 Razorpay Judge / Admin Persona',
      desc: 'Full system authorization. Access to raw gRPC audit trails, circuit breaker overrides, and system health metrics.',
      user: {
        id: 'usr_judge_01',
        name: 'Razorpay Evaluator',
        email: 'judge@razorpay.com',
        role: 'razorpay_judge',
        merchantName: 'Acme Retail Store (Razorpay Test Merchant)'
      }
    },
    {
      role: 'finance_controller',
      title: '💰 Merchant Finance Controller',
      desc: 'Track 04 Persona. Access to three-way settlement reconciliation, cash flow forecasting, and promise tracking.',
      user: {
        id: 'usr_fin_02',
        name: 'Priya Sharma (CFO)',
        email: 'priya@acmeretail.com',
        role: 'finance_controller',
        merchantName: 'Acme Retail Store'
      }
    },
    {
      role: 'ops_lead',
      title: '🛠️ Merchant Support & Ops Lead',
      desc: 'Track 03 Persona. Access to payment failure classification, manual workflow retries, and customer escalation tickets.',
      user: {
        id: 'usr_ops_03',
        name: 'Rahul Verma (Ops Head)',
        email: 'rahul@acmeretail.com',
        role: 'ops_lead',
        merchantName: 'Acme Retail Store'
      }
    }
  ];

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(0, 0, 0, 0.4)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 100,
      backdropFilter: 'blur(2px)'
    }}>
      <div style={{
        width: '560px',
        backgroundColor: '#ffffff',
        borderRadius: '12px',
        boxShadow: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
        border: '1px solid #e9ecef',
        overflow: 'hidden'
      }}>
        {/* Header */}
        <div style={{
          padding: '20px 24px',
          borderBottom: '1px solid #f1f3f5',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between'
        }}>
          <div>
            <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#1a1f2c' }}>
              Multi-Role Persona Authenticator
            </h3>
            <p style={{ fontSize: '13px', color: '#8c98a9', marginTop: '2px' }}>
              Switch personas to test feature RBAC across Track 03 & Track 04
            </p>
          </div>
          <button onClick={onClose} style={{ color: '#8c98a9', padding: '4px' }}>
            <X size={20} />
          </button>
        </div>

        {/* Persona Cards */}
        <div style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '12px' }}>
          {roles.map((r) => {
            const isSelected = currentUser.role === r.role;
            return (
              <div
                key={r.role}
                onClick={() => {
                  onSelectRole(r.user);
                  onClose();
                }}
                style={{
                  padding: '16px 18px',
                  borderRadius: '8px',
                  border: isSelected ? '2px solid #4263eb' : '1px solid #e9ecef',
                  backgroundColor: isSelected ? '#edf2ff' : '#ffffff',
                  cursor: 'pointer',
                  transition: 'all 0.15s ease'
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <h4 style={{ fontSize: '15px', fontWeight: 600, color: isSelected ? '#4263eb' : '#1a1f2c' }}>
                    {r.title}
                  </h4>
                  {isSelected && <CheckCircle size={18} color="#4263eb" />}
                </div>
                <p style={{ fontSize: '12px', color: '#4a5568', marginTop: '6px', lineHeight: 1.4 }}>
                  {r.desc}
                </p>
                <div style={{ marginTop: '8px', fontSize: '11px', color: '#8c98a9', display: 'flex', gap: '12px' }}>
                  <span>User: <strong>{r.user.name}</strong></span>
                  <span>Email: <strong>{r.user.email}</strong></span>
                </div>
              </div>
            );
          })}
        </div>

        {/* Footer */}
        <div style={{
          padding: '16px 24px',
          backgroundColor: '#f8f9fa',
          borderTop: '1px solid #f1f3f5',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          fontSize: '12px',
          color: '#4a5568'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <Sparkles size={14} color="#4263eb" />
            <span>Built for Razorpay Hackathon Evaluators</span>
          </div>
          <button
            onClick={onClose}
            style={{
              padding: '6px 16px',
              backgroundColor: '#4263eb',
              color: '#ffffff',
              borderRadius: '6px',
              fontWeight: 600,
              fontSize: '13px'
            }}
          >
            Confirm Persona
          </button>
        </div>
      </div>
    </div>
  );
};
