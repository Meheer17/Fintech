import React, { useState } from 'react';
import { sendAIChatMessage } from '../lib/api';
import { Send, Bot, User, Sparkles } from 'lucide-react';
import { MarkdownRenderer } from './MarkdownRenderer';

export const ChatTab: React.FC = () => {
  const [messages, setMessages] = useState<{ role: 'user' | 'assistant'; text: string }[]>([
    { role: 'assistant', text: "Hello! I am the RevenueIQ Master Agent. Ask me about your settlement discrepancies, forward cash position forecasts, or active recovery workflows." }
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSend = async () => {
    if (!input.trim() || loading) return;
    const userMsg = input.trim();
    setInput('');
    setMessages((prev) => [...prev, { role: 'user', text: userMsg }]);
    setLoading(true);

    try {
      const reply = await sendAIChatMessage(userMsg);
      setMessages((prev) => [...prev, { role: 'assistant', text: reply }]);
    } catch (err: any) {
      setMessages((prev) => [...prev, { role: 'assistant', text: `Error communicating with AI Gateway: ${err.message || err}` }]);
    } finally {
      setLoading(false);
    }
  };

  const samplePrompts = [
    "Why was yesterday's settlement short?",
    "Show forward 7-day cash forecast",
    "What is our active recovery rate across UPI payments?"
  ];

  return (
    <div style={{
      backgroundColor: '#ffffff',
      borderRadius: '8px',
      border: '1px solid #e9ecef',
      height: 'calc(100vh - 140px)',
      display: 'flex',
      flexDirection: 'column',
      boxShadow: '0 1px 2px rgba(0,0,0,0.04)'
    }}>
      {/* Chat Header */}
      <div style={{ padding: '16px 24px', borderBottom: '1px solid #f1f3f5', display: 'flex', alignItems: 'center', gap: '10px' }}>
        <Bot size={22} color="#4263eb" />
        <div>
          <h3 style={{ fontSize: '16px', fontWeight: 700, color: '#1a1f2c' }}>
            RevenueIQ Master Agent Copilot
          </h3>
          <span style={{ fontSize: '12px', color: '#8c98a9' }}>
            Natural Language Settlement Q&A & Cash Forecasting Engine
          </span>
        </div>
      </div>

      {/* Message List */}
      <div style={{ flex: 1, padding: '24px', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '16px' }}>
        {messages.map((m, i) => (
          <div
            key={i}
            style={{
              display: 'flex',
              gap: '12px',
              alignSelf: m.role === 'user' ? 'flex-end' : 'flex-start',
              maxWidth: '80%'
            }}
          >
            {m.role === 'assistant' && (
              <div style={{ width: '32px', height: '32px', borderRadius: '50%', backgroundColor: '#edf2ff', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Bot size={18} color="#4263eb" />
              </div>
            )}
            <div style={{
              padding: '12px 16px',
              borderRadius: '12px',
              backgroundColor: m.role === 'user' ? '#4263eb' : '#f8f9fa',
              color: m.role === 'user' ? '#ffffff' : '#1a1f2c',
              border: m.role === 'assistant' ? '1px solid #e9ecef' : 'none',
              fontSize: '14px',
              lineHeight: 1.5,
              wordBreak: 'break-word'
            }}>
              <MarkdownRenderer content={m.text} isUser={m.role === 'user'} />
            </div>
            {m.role === 'user' && (
              <div style={{ width: '32px', height: '32px', borderRadius: '50%', backgroundColor: '#f1f3f5', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <User size={18} color="#4a5568" />
              </div>
            )}
          </div>
        ))}
        {loading && (
          <div style={{ color: '#8c98a9', fontSize: '13px', display: 'flex', alignItems: 'center', gap: '6px' }}>
            <Sparkles size={14} color="#4263eb" /> AI Agent thinking...
          </div>
        )}
      </div>

      {/* Suggested Prompts */}
      <div style={{ padding: '8px 24px', display: 'flex', gap: '8px', overflowX: 'auto' }}>
        {samplePrompts.map((sp) => (
          <button
            key={sp}
            onClick={() => setInput(sp)}
            style={{
              padding: '6px 12px',
              borderRadius: '16px',
              backgroundColor: '#f8f9fa',
              border: '1px solid #e9ecef',
              fontSize: '12px',
              color: '#4a5568',
              whiteSpace: 'nowrap'
            }}
          >
            💡 {sp}
          </button>
        ))}
      </div>

      {/* Input Form */}
      <div style={{ padding: '16px 24px', borderTop: '1px solid #f1f3f5', display: 'flex', gap: '12px' }}>
        <input
          type="text"
          placeholder="Ask Master Agent a question about settlements or recovery..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          style={{
            flex: 1,
            padding: '12px 16px',
            borderRadius: '8px',
            border: '1px solid #dee2e6',
            fontSize: '14px',
            outline: 'none'
          }}
        />
        <button
          onClick={handleSend}
          style={{
            padding: '0 20px',
            borderRadius: '8px',
            backgroundColor: '#4263eb',
            color: '#ffffff',
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: '8px'
          }}
        >
          <Send size={16} /> Send
        </button>
      </div>
    </div>
  );
};
