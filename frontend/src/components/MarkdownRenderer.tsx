import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

interface MarkdownRendererProps {
  content: string;
  isUser?: boolean;
}

export const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ content, isUser }) => {
  if (isUser) {
    return <span style={{ whiteSpace: 'pre-wrap' }}>{content}</span>;
  }

  return (
    <div className="markdown-body" style={{ fontSize: '14px', lineHeight: 1.6 }}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ children }) => (
            <h1 style={{ fontSize: '18px', fontWeight: 700, margin: '14px 0 8px 0', color: '#0f172a', borderBottom: '1px solid #e2e8f0', paddingBottom: '4px' }}>
              {children}
            </h1>
          ),
          h2: ({ children }) => (
            <h2 style={{ fontSize: '16px', fontWeight: 700, margin: '12px 0 6px 0', color: '#0f172a' }}>
              {children}
            </h2>
          ),
          h3: ({ children }) => (
            <h3 style={{ fontSize: '14px', fontWeight: 700, margin: '10px 0 4px 0', color: '#1e293b' }}>
              {children}
            </h3>
          ),
          h4: ({ children }) => (
            <h4 style={{ fontSize: '13px', fontWeight: 700, margin: '8px 0 4px 0', color: '#334155' }}>
              {children}
            </h4>
          ),
          p: ({ children }) => (
            <p style={{ margin: '0 0 8px 0', color: '#334155' }}>
              {children}
            </p>
          ),
          ul: ({ children }) => (
            <ul style={{ margin: '4px 0 10px 0', paddingLeft: '20px', color: '#334155' }}>
              {children}
            </ul>
          ),
          ol: ({ children }) => (
            <ol style={{ margin: '4px 0 10px 0', paddingLeft: '20px', color: '#334155' }}>
              {children}
            </ol>
          ),
          li: ({ children }) => (
            <li style={{ margin: '3px 0' }}>
              {children}
            </li>
          ),
          strong: ({ children }) => (
            <strong style={{ fontWeight: 700, color: '#0f172a' }}>
              {children}
            </strong>
          ),
          em: ({ children }) => (
            <em style={{ fontStyle: 'italic', color: '#475569' }}>
              {children}
            </em>
          ),
          a: ({ href, children }) => (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: '#2563eb', textDecoration: 'underline', fontWeight: 600 }}
            >
              {children}
            </a>
          ),
          code: ({ inline, className, children, ...props }: any) => {
            if (inline) {
              return (
                <code
                  style={{
                    backgroundColor: '#f1f5f9',
                    color: '#0f172a',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    fontSize: '12px',
                    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
                    border: '1px solid #e2e8f0'
                  }}
                  {...props}
                >
                  {children}
                </code>
              );
            }
            return (
              <pre
                style={{
                  backgroundColor: '#0f172a',
                  color: '#f8fafc',
                  padding: '12px 16px',
                  borderRadius: '8px',
                  overflowX: 'auto',
                  fontSize: '13px',
                  margin: '10px 0',
                  fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace'
                }}
              >
                <code {...props}>{children}</code>
              </pre>
            );
          },
          blockquote: ({ children }) => (
            <blockquote
              style={{
                borderLeft: '4px solid #3b82f6',
                backgroundColor: '#eff6ff',
                padding: '8px 14px',
                margin: '10px 0',
                borderRadius: '0 6px 6px 0',
                color: '#1e40af',
                fontSize: '13px'
              }}
            >
              {children}
            </blockquote>
          ),
          table: ({ children }) => (
            <div style={{ overflowX: 'auto', margin: '12px 0' }}>
              <table
                style={{
                  width: '100%',
                  borderCollapse: 'collapse',
                  fontSize: '13px',
                  border: '1px solid #e2e8f0'
                }}
              >
                {children}
              </table>
            </div>
          ),
          thead: ({ children }) => (
            <thead style={{ backgroundColor: '#f8fafc', borderBottom: '2px solid #cbd5e1' }}>
              {children}
            </thead>
          ),
          th: ({ children }) => (
            <th
              style={{
                padding: '8px 12px',
                fontWeight: 700,
                textAlign: 'left',
                color: '#0f172a',
                border: '1px solid #e2e8f0'
              }}
            >
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td
              style={{
                padding: '8px 12px',
                color: '#334155',
                border: '1px solid #e2e8f0'
              }}
            >
              {children}
            </td>
          ),
          hr: () => (
            <hr style={{ border: 'none', borderTop: '1px solid #e2e8f0', margin: '14px 0' }} />
          )
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};
