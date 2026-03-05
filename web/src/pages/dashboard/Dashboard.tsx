import { useEffect, useState } from 'react'
import {
  PlusOutlined,
  ApartmentOutlined,
  CloudServerOutlined,
  ArrowRightOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ThunderboltOutlined,
  ExperimentOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { workflowApi, type Workflow } from '../../api/workflow'
import { mcpServerApi, type MCPServer } from '../../api/mcpserver'
import { executionApi } from '../../api/execution'

const getGreeting = () => {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good Morning'
  if (hour < 18) return 'Good Afternoon'
  return 'Good Evening'
}

const formatDate = () => {
  return new Date().toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'short',
    day: 'numeric',
  })
}

// Shared card style
const card = (extra?: React.CSSProperties): React.CSSProperties => ({
  background: '#fff',
  borderRadius: 14,
  border: '1px solid #eaecf0',
  boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
  ...extra,
})

export default function Dashboard() {
  const navigate = useNavigate()
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [servers, setServers] = useState<MCPServer[]>([])
  const [stats, setStats] = useState<any>({})

  useEffect(() => {
    workflowApi.list().then((res: any) => setWorkflows(res.data || [])).catch(() => {})
    mcpServerApi.list().then((res: any) => setServers(res.data || [])).catch(() => {})
    executionApi.stats().then((res: any) => setStats(res || {})).catch(() => {})
  }, [])

  const activeServers = servers.filter((s) => s.status === 'active').length
  const totalNodes = workflows.reduce((s, w) => s + (w.nodes?.length || 0), 0)

  return (
    <div style={{ maxWidth: 960, margin: '0 auto' }}>
      {/* ---- Bento grid ---- */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(12, 1fr)',
          gridAutoRows: 'minmax(0, auto)',
          gap: 14,
        }}
      >
        {/* ===== Row 1: Hero greeting (spans 8) + date card (spans 4) ===== */}
        <div
          style={{
            gridColumn: 'span 8',
            ...card({ padding: '28px 30px' }),
            background: 'linear-gradient(135deg, #3b5bdb 0%, #5c7cfa 100%)',
            border: 'none',
            color: '#fff',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          {/* Decorative circles */}
          <div style={{ position: 'absolute', top: -30, right: -30, width: 120, height: 120, borderRadius: '50%', background: 'rgba(255,255,255,0.07)' }} />
          <div style={{ position: 'absolute', bottom: -20, right: 60, width: 80, height: 80, borderRadius: '50%', background: 'rgba(255,255,255,0.05)' }} />
          <div style={{ position: 'relative' }}>
            <div style={{ fontSize: 11, fontWeight: 500, letterSpacing: 0.8, textTransform: 'uppercase', opacity: 0.7, marginBottom: 6 }}>
              {formatDate()}
            </div>
            <h1 style={{ fontSize: 22, fontWeight: 300, margin: 0, lineHeight: 1.35 }}>
              {getGreeting()}, <span style={{ fontWeight: 600 }}>welcome back</span>
            </h1>
            <p style={{ fontSize: 13, opacity: 0.7, marginTop: 8, lineHeight: 1.5 }}>
              Build and orchestrate automation workflows with MCP tools.
            </p>
          </div>
        </div>

        {/* Stats stacked (spans 4) */}
        <div style={{ gridColumn: 'span 4', display: 'flex', flexDirection: 'column', gap: 14 }}>
          <div style={{ ...card({ padding: '16px 20px' }), flex: 1 }}>
            <div style={{ fontSize: 11, color: '#98a2b3', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Workflows
            </div>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, marginTop: 6 }}>
              <span style={{ fontSize: 26, fontWeight: 700, color: '#1a1a2e', lineHeight: 1 }}>{workflows.length}</span>
              <span style={{ fontSize: 12, color: '#98a2b3' }}>{totalNodes} nodes</span>
            </div>
          </div>
          <div style={{ ...card({ padding: '16px 20px' }), flex: 1 }}>
            <div style={{ fontSize: 11, color: '#98a2b3', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              MCP Servers
            </div>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, marginTop: 6 }}>
              <span style={{ fontSize: 26, fontWeight: 700, color: '#1a1a2e', lineHeight: 1 }}>{servers.length}</span>
              <span style={{ fontSize: 12, color: activeServers > 0 ? '#12b76a' : '#98a2b3' }}>
                {activeServers > 0 && <span style={{ display: 'inline-block', width: 6, height: 6, borderRadius: 3, background: '#12b76a', marginRight: 4, verticalAlign: 'middle' }} />}
                {activeServers} online
              </span>
            </div>
          </div>
        </div>

        {/* ===== Row 1.5: Execution Stats (4 cards, span 3 each) ===== */}
        <div style={{ gridColumn: 'span 3', ...card({ padding: '16px 20px' }) }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <ThunderboltOutlined style={{ color: '#3b5bdb', fontSize: 14 }} />
            <span style={{ fontSize: 11, color: '#98a2b3', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Executions
            </span>
          </div>
          <span style={{ fontSize: 26, fontWeight: 700, color: '#1a1a2e', lineHeight: 1 }}>
            {stats.total_executions || 0}
          </span>
        </div>

        <div style={{ gridColumn: 'span 3', ...card({ padding: '16px 20px' }) }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <CheckCircleOutlined style={{ color: '#12b76a', fontSize: 14 }} />
            <span style={{ fontSize: 11, color: '#98a2b3', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Success
            </span>
          </div>
          <span style={{ fontSize: 26, fontWeight: 700, color: '#12b76a', lineHeight: 1 }}>
            {stats.success_count || 0}
          </span>
        </div>

        <div style={{ gridColumn: 'span 3', ...card({ padding: '16px 20px' }) }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <CloseCircleOutlined style={{ color: '#f04438', fontSize: 14 }} />
            <span style={{ fontSize: 11, color: '#98a2b3', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Failed
            </span>
          </div>
          <span style={{ fontSize: 26, fontWeight: 700, color: '#f04438', lineHeight: 1 }}>
            {stats.failed_count || 0}
          </span>
        </div>

        <div style={{ gridColumn: 'span 3', ...card({ padding: '16px 20px' }) }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <ExperimentOutlined style={{ color: '#7c3aed', fontSize: 14 }} />
            <span style={{ fontSize: 11, color: '#98a2b3', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Success Rate
            </span>
          </div>
          <span style={{ fontSize: 26, fontWeight: 700, color: '#1a1a2e', lineHeight: 1 }}>
            {stats.total_executions > 0 ? `${(stats.success_rate || 0).toFixed(0)}%` : '-'}
          </span>
        </div>

        {/* ===== Row 2: Quick actions (spans 5) + Recent list (spans 7) ===== */}
        <div style={{ gridColumn: 'span 5', display: 'flex', flexDirection: 'column', gap: 14 }}>
          {/* New Workflow */}
          <div
            onClick={() => navigate('/workflows/new')}
            style={{
              ...card({ padding: '20px 22px' }),
              cursor: 'pointer',
              transition: 'transform 0.15s, box-shadow 0.15s',
              flex: 1,
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-2px)'
              e.currentTarget.style.boxShadow = '0 6px 16px rgba(59,91,219,0.12)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)'
              e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.04)'
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{
                width: 36, height: 36, borderRadius: 10,
                background: 'linear-gradient(135deg, #3b5bdb, #5c7cfa)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: '#fff', fontSize: 15, flexShrink: 0,
              }}>
                <PlusOutlined />
              </div>
              <div>
                <div style={{ fontSize: 14, fontWeight: 600, color: '#1a1a2e' }}>New Workflow</div>
                <div style={{ fontSize: 12, color: '#98a2b3', marginTop: 1 }}>Create a new automation pipeline</div>
              </div>
            </div>
          </div>

          {/* Manage Servers */}
          <div
            onClick={() => navigate('/mcp-servers')}
            style={{
              ...card({ padding: '20px 22px' }),
              cursor: 'pointer',
              transition: 'transform 0.15s, box-shadow 0.15s',
              flex: 1,
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-2px)'
              e.currentTarget.style.boxShadow = '0 6px 16px rgba(0,0,0,0.06)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)'
              e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.04)'
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{
                width: 36, height: 36, borderRadius: 10,
                background: '#f0f5ff',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: '#3b5bdb', fontSize: 15, flexShrink: 0,
              }}>
                <CloudServerOutlined />
              </div>
              <div>
                <div style={{ fontSize: 14, fontWeight: 600, color: '#1a1a2e' }}>MCP Servers</div>
                <div style={{ fontSize: 12, color: '#98a2b3', marginTop: 1 }}>Manage service connections</div>
              </div>
            </div>
          </div>

          {/* View all workflows */}
          <div
            onClick={() => navigate('/workflows')}
            style={{
              ...card({ padding: '20px 22px' }),
              cursor: 'pointer',
              transition: 'transform 0.15s, box-shadow 0.15s',
              flex: 1,
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-2px)'
              e.currentTarget.style.boxShadow = '0 6px 16px rgba(0,0,0,0.06)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)'
              e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.04)'
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{
                width: 36, height: 36, borderRadius: 10,
                background: '#f0fdf4',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: '#12b76a', fontSize: 15, flexShrink: 0,
              }}>
                <ApartmentOutlined />
              </div>
              <div>
                <div style={{ fontSize: 14, fontWeight: 600, color: '#1a1a2e' }}>All Workflows</div>
                <div style={{ fontSize: 12, color: '#98a2b3', marginTop: 1 }}>Browse & search workflows</div>
              </div>
            </div>
          </div>
        </div>

        {/* Recent workflows panel (spans 7) */}
        <div
          style={{
            gridColumn: 'span 7',
            ...card({ padding: 0, overflow: 'hidden' }),
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          {/* Panel header */}
          <div style={{
            display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            padding: '14px 20px',
            borderBottom: '1px solid #f2f4f7',
          }}>
            <span style={{ fontSize: 13, fontWeight: 600, color: '#1a1a2e' }}>Recent Workflows</span>
            <span
              onClick={() => navigate('/workflows')}
              style={{ fontSize: 12, color: '#3b5bdb', cursor: 'pointer', fontWeight: 500 }}
            >
              View all <ArrowRightOutlined style={{ fontSize: 10 }} />
            </span>
          </div>

          {/* List */}
          <div style={{ flex: 1 }}>
            {workflows.length === 0 ? (
              <div style={{
                padding: '40px 20px', textAlign: 'center', color: '#98a2b3', fontSize: 13,
              }}>
                <ApartmentOutlined style={{ fontSize: 28, display: 'block', marginBottom: 8, color: '#d0d5dd' }} />
                No workflows yet
              </div>
            ) : (
              workflows.slice(0, 5).map((w, i) => (
                <div
                  key={w.id}
                  onClick={() => navigate(`/workflows/${w.id}`)}
                  style={{
                    padding: '12px 20px',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    borderBottom: i < Math.min(workflows.length, 5) - 1 ? '1px solid #f8f9fb' : 'none',
                    transition: 'background 0.1s',
                  }}
                  onMouseEnter={(e) => { e.currentTarget.style.background = '#fafbfc' }}
                  onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0, flex: 1 }}>
                    <ApartmentOutlined style={{ color: '#3b5bdb', fontSize: 13, flexShrink: 0 }} />
                    <div style={{ minWidth: 0 }}>
                      <div style={{ fontWeight: 500, fontSize: 13, color: '#1a1a2e', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {w.name || 'Untitled'}
                      </div>
                      {w.description && (
                        <div style={{ fontSize: 11, color: '#b0b5c0', marginTop: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {w.description}
                        </div>
                      )}
                    </div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexShrink: 0, marginLeft: 12 }}>
                    <span style={{ fontSize: 11, color: '#c0c5d0' }}>{w.nodes?.length || 0} nodes</span>
                    <ArrowRightOutlined style={{ color: '#d0d5dd', fontSize: 10 }} />
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
