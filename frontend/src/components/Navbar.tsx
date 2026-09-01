import React from 'react';
import { UserRole, UserProfile } from '../types';
import { LogIn, UserCircle, Activity, RefreshCw } from 'lucide-react';

interface NavbarProps {
  user: UserProfile;
  onOpenAuth: () => void;
  onSync?: () => void;
  syncing?: boolean;
}

export const Navbar: React.FC<NavbarProps> = ({ user, onOpenAuth, onSync, syncing }) => {
  return (
    <header style={{
      height: '64px',
      backgroundColor: '#ffffff',
      borderBottom: '1px solid #e9ecef',
      padding: '0 32px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      position: 'sticky',
      top: 0,
      zIndex: 9
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
        <h2 style={{ fontSize: '16px', fontWeight: 600, color: '#1a1f2c' }}>
          {user.merchantName} Control Panel
        </h2>
        <span style={{
          padding: '2px 8px',
          borderRadius: '12px',
          backgroundColor: '#ebfbee',
          color: '#2b8a3e',
          fontSize: '12px',
          fontWeight: 600,
          display: 'flex',
          alignItems: 'center',
          gap: '4px'
        }}>
          <Activity size={12} />
          Razorpay Test Mode (Live gRPC)
        </span>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
        <button
          onClick={onSync}
          disabled={syncing}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            padding: '8px 14px',
            borderRadius: '6px',
            backgroundColor: '#e7f5ff',
            color: '#1971c2',
            border: '1px solid #74c0fc',
            fontSize: '13px',
            fontWeight: 600,
            cursor: syncing ? 'not-allowed' : 'pointer',
            transition: 'all 0.15s ease'
          }}
        >
          <RefreshCw size={16} style={{ animation: syncing ? 'spin 1s linear infinite' : 'none' }} />
          <span>Sync Razorpay</span>
        </button>

        <button
          onClick={onOpenAuth}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            padding: '8px 14px',
            borderRadius: '6px',
            backgroundColor: '#f1f3f5',
            color: '#1a1f2c',
            fontSize: '13px',
            fontWeight: 600,
            transition: 'background 0.15s ease'
          }}
        >
          <UserCircle size={18} color="#4263eb" />
          <span>{user.name} ({user.role})</span>
          <LogIn size={14} color="#8c98a9" style={{ marginLeft: '4px' }} />
        </button>
      </div>
    </header>
  );
};
