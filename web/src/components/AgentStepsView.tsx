import { useState } from 'react'
import { Tag, Collapse } from 'antd'
import {
  SearchOutlined,
  RobotOutlined,
  ToolOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'

interface AgentStep {
  step?: number
  iteration?: number
  type: string
  server_url?: string
  tools?: string[]
  messages_count?: number
  tools_count?: number
  has_tool_calls?: boolean
  tool_calls_count?: number
  content_preview?: string
  content?: string
  tool_name?: string
  tool_args?: any
  tool_result?: string
  tool_error?: string
  duration?: number
}

const stepIcon: Record<string, { icon: React.ReactNode; color: string }> = {
  tool_discovery: { icon: <SearchOutlined />, color: '#3b5bdb' },
  llm_call: { icon: <RobotOutlined />, color: '#7c3aed' },
  tool_call: { icon: <ToolOutlined />, color: '#0891b2' },
  final_response: { icon: <CheckCircleOutlined />, color: '#12b76a' },
  response: { icon: <CheckCircleOutlined />, color: '#12b76a' },
}

export default function AgentStepsView({ steps }: { steps: AgentStep[] }) {
  if (!steps || steps.length === 0) return null

  return (
    <div style={{ marginTop: 8 }}>
      <div style={{ fontSize: 11, color: '#98a2b3', marginBottom: 6 }}>Agent Execution Steps:</div>
      {steps.map((s, i) => {
        const cfg = stepIcon[s.type] || { icon: null, color: '#667085' }
        return (
          <div
            key={i}
            style={{
              display: 'flex',
              gap: 8,
              padding: '6px 0',
              borderBottom: i < steps.length - 1 ? '1px solid #f2f4f7' : undefined,
              alignItems: 'flex-start',
            }}
          >
            <div style={{ color: cfg.color, fontSize: 13, marginTop: 1, flexShrink: 0 }}>
              {cfg.icon}
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <StepContent step={s} />
            </div>
            <div style={{ fontSize: 11, color: '#98a2b3', flexShrink: 0 }}>
              {(s.duration ?? 0) > 0 ? `${s.duration}ms` : ''}
            </div>
          </div>
        )
      })}
    </div>
  )
}

function StepContent({ step }: { step: AgentStep }) {
  switch (step.type) {
    case 'tool_discovery':
      return (
        <div style={{ fontSize: 12 }}>
          <span style={{ color: '#344054' }}>
            Discovered {step.tools?.length || 0} tools
          </span>
          {step.server_url && (
            <span style={{ color: '#98a2b3', marginLeft: 4 }}>from {step.server_url}</span>
          )}
          {step.tools && step.tools.length > 0 && (
            <div style={{ marginTop: 4, display: 'flex', gap: 4, flexWrap: 'wrap' }}>
              {step.tools.map((t) => (
                <Tag key={t} style={{ fontSize: 10, borderRadius: 4, margin: 0 }}>{t}</Tag>
              ))}
            </div>
          )}
        </div>
      )

    case 'llm_call':
      return (
        <div style={{ fontSize: 12 }}>
          <span style={{ color: '#344054' }}>LLM Call</span>
          <span style={{ color: '#98a2b3', marginLeft: 6 }}>
            {step.messages_count} messages, {step.tools_count} tools
          </span>
          {step.has_tool_calls && (
            <Tag color="purple" style={{ fontSize: 10, borderRadius: 4, marginLeft: 6 }}>
              {step.tool_calls_count} tool calls
            </Tag>
          )}
          {step.content_preview && !step.has_tool_calls && (
            <div style={{ color: '#667085', marginTop: 4, fontSize: 11, lineHeight: 1.4 }}>
              {step.content_preview}
            </div>
          )}
        </div>
      )

    case 'tool_call':
      return (
        <div style={{ fontSize: 12 }}>
          <span style={{ color: '#344054', fontWeight: 500 }}>{step.tool_name}</span>
          {step.tool_error && (
            <Tag color="error" style={{ fontSize: 10, borderRadius: 4, marginLeft: 6 }}>
              <CloseCircleOutlined /> error
            </Tag>
          )}
          <ToolCallDetail step={step} />
        </div>
      )

    case 'final_response':
    case 'response':
      return (
        <div style={{ fontSize: 12 }}>
          <span style={{ color: '#12b76a', fontWeight: 500 }}>Final Response</span>
          {(step.content_preview || step.content) && (
            <div style={{ color: '#667085', marginTop: 4, fontSize: 11, lineHeight: 1.4 }}>
              {(() => { const text = step.content_preview || step.content || ''; return text.length > 200 ? text.slice(0, 200) + '...' : text })()}
            </div>
          )}
        </div>
      )

    default:
      return <div style={{ fontSize: 12, color: '#667085' }}>{step.type}</div>
  }
}

function ToolCallDetail({ step }: { step: AgentStep }) {
  const [expanded, setExpanded] = useState(false)
  const hasDetail = step.tool_args || step.tool_result || step.tool_error

  if (!hasDetail) return null

  return (
    <Collapse
      ghost
      size="small"
      activeKey={expanded ? ['1'] : []}
      onChange={() => setExpanded(!expanded)}
      style={{ marginTop: 4 }}
      items={[
        {
          key: '1',
          label: <span style={{ fontSize: 11, color: '#98a2b3' }}>Details</span>,
          children: (
            <div style={{ fontSize: 11 }}>
              {step.tool_args && (
                <div style={{ marginBottom: 4 }}>
                  <span style={{ color: '#98a2b3' }}>Args: </span>
                  <code style={{ background: '#f2f4f7', padding: '1px 4px', borderRadius: 3 }}>
                    {JSON.stringify(step.tool_args)}
                  </code>
                </div>
              )}
              {step.tool_result && (
                <div style={{ marginBottom: 4 }}>
                  <span style={{ color: '#98a2b3' }}>Result: </span>
                  <pre
                    style={{
                      background: '#f9fafb',
                      border: '1px solid #e5e7eb',
                      borderRadius: 4,
                      padding: '4px 8px',
                      marginTop: 2,
                      maxHeight: 120,
                      overflow: 'auto',
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-all',
                      fontSize: 10,
                    }}
                  >
                    {step.tool_result}
                  </pre>
                </div>
              )}
              {step.tool_error && (
                <div style={{ color: '#f04438' }}>
                  <span style={{ color: '#98a2b3' }}>Error: </span>
                  {step.tool_error}
                </div>
              )}
            </div>
          ),
        },
      ]}
    />
  )
}
