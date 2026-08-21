import React, { useState, useEffect } from 'react';
import { Sidebar } from './components/Sidebar';
import { Navbar } from './components/Navbar';
import { AuthModal } from './components/AuthModal';
import { OverviewTab } from './components/OverviewTab';
import { FailuresTab } from './components/FailuresTab';
import { RecoveriesTab } from './components/RecoveriesTab';
import { PromisesTab } from './components/PromisesTab';
import { ReconciliationTab } from './components/ReconciliationTab';
import { AuditTab } from './components/AuditTab';
import { ChatTab } from './components/ChatTab';
import { UserProfile, FailureRecord, WorkflowRecord, AuditEntry } from './types';
import { fetchOverviewMetrics, fetchFailures, fetchWorkflows, fetchAuditLogs } from './lib/api';

interface ErrorBoundaryProps {
  children: React.ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('UI Render Error caught by ErrorBoundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{
          backgroundColor: '#ffffff',
          borderRadius: '8px',
          border: '1px solid #ffc9c9',
          padding: '32px',
          textAlign: 'center',
          boxShadow: '0 2px 4px rgba(0,0,0,0.05)'
        }}>
          <h3 style={{ fontSize: '18px', fontWeight: 700, color: '#c92a2a' }}>
            View Render Discrepancy Caught
          </h3>
          <p style={{ fontSize: '13px', color: '#4a5568', marginTop: '8px' }}>
            {this.state.error?.message || 'An unexpected rendering error occurred.'}
          </p>
          <button
            onClick={() => this.setState({ hasError: false })}
            style={{
              marginTop: '16px',
              padding: '8px 16px',
              backgroundColor: '#4263eb',
              color: '#ffffff',
              border: 'none',
              borderRadius: '6px',
              fontWeight: 600,
              fontSize: '13px',
              cursor: 'pointer'
            }}
          >
            Reset View State
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview');
  const [isAuthOpen, setIsAuthOpen] = useState(false);
  const [currentUser, setCurrentUser] = useState<UserProfile>({
    id: 'usr_judge_01',
    name: 'Razorpay Evaluator',
    email: 'judge@razorpay.com',
    role: 'razorpay_judge',
    merchantName: 'Acme Retail Store (Razorpay Test Merchant)'
  });

  const [metrics, setMetrics] = useState({
    total_at_risk_paise: 23450000,
    total_recovered_paise: 17200000,
    recovery_rate: 0.733,
    active_workflows: 14,
    reconciliation_match: 0.942
  });

  const [failures, setFailures] = useState<FailureRecord[]>([]);
  const [workflows, setWorkflows] = useState<WorkflowRecord[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditEntry[]>([]);

  useEffect(() => {
    fetchOverviewMetrics().then(setMetrics);
    fetchFailures().then(setFailures);
    fetchWorkflows().then(setWorkflows);
    fetchAuditLogs().then(setAuditLogs);
  }, []);

  return (
    <div style={{ display: 'flex', minHeight: '100vh', backgroundColor: '#f8f9fa' }}>
      <Sidebar activeTab={activeTab} setActiveTab={setActiveTab} role={currentUser.role} />

      <div style={{ marginLeft: '260px', flex: 1, display: 'flex', flexDirection: 'column' }}>
        <Navbar user={currentUser} onOpenAuth={() => setIsAuthOpen(true)} />

        <main style={{ padding: '32px', flex: 1 }}>
          <ErrorBoundary>
            {activeTab === 'overview' && <OverviewTab metrics={metrics} />}
            {activeTab === 'failures' && <FailuresTab failures={failures} />}
            {(activeTab === 'recoveries' || activeTab === 'workflows') && <RecoveriesTab workflows={workflows} />}
            {activeTab === 'promises' && <PromisesTab />}
            {activeTab === 'reconciliation' && <ReconciliationTab />}
            {activeTab === 'audit' && <AuditTab logs={auditLogs} />}
            {activeTab === 'chat' && <ChatTab />}
          </ErrorBoundary>
        </main>
      </div>

      <AuthModal
        isOpen={isAuthOpen}
        onClose={() => setIsAuthOpen(false)}
        currentUser={currentUser}
        onSelectRole={setCurrentUser}
      />
    </div>
  );
};
