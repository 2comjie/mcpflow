import { useEffect, useState } from 'react'
import { Button, Row, Col } from 'antd'
import {
  PlusOutlined,
  ApartmentOutlined,
  CloudServerOutlined,
  ArrowRightOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { workflowApi, type Workflow } from '../../api/workflow'
import { mcpServerApi, type MCPServer } from '../../api/mcpserver'

const getGreeting = () => {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good Morning'
  if (hour < 18) return 'Good Afternoon'
  return 'Good Evening'
}

const formatDate = () => {
  return new Date().toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

const getWeatherEmoji = () => {
  const hour = new Date().getHours()
  if (hour >= 6 && hour < 18) return { icon: '\u2600\uFE0F', text: 'Sunny' }
  return { icon: '\uD83C\uDF19', text: 'Clear Night' }
}

export default function Dashboard() {
  const navigate = useNavigate()
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [servers, setServers] = useState<MCPServer[]>([])

  useEffect(() => {
    workflowApi.list().then((res: any) => setWorkflows(res.data || [])).catch(() => {})
    mcpServerApi.list().then((res: any) => setServers(res.data || [])).catch(() => {})
  }, [])

  const weather = getWeatherEmoji()
  const activeServers = servers.filter((s) => s.status === 'active').length
  const totalNodes = workflows.reduce((sum, w) => sum + (w.nodes?.length || 0), 0)

  return (
    <div style={{ maxWidth: 960, margin: '0 auto' }}>
      {/* Hero greeting */}
      <div style={{ marginBottom: 48, marginTop: 20 }}>
        <div
          style={{
            fontSize: 13,
            color: '#98a2b3',
            fontWeight: 500,
            letterSpacing: 0.5,
            marginBottom: 8,
            textTransform: 'uppercase',
          }}
        >
          {formatDate()} &nbsp;&middot;&nbsp; {weather.icon} {weather.text}
        </div>
        <h1
          style={{
            fontSize: 36,
            fontWeight: 300,
            color: '#1a1a2e',
            margin: 0,
            letterSpacing: -0.5,
            lineHeight: 1.2,
          }}
        >
          {getGreeting()}, <span style={{ fontWeight: 600 }}>welcome back</span>
        </h1>
        <p
          style={{
            fontSize: 16,
            color: '#667085',
            marginTop: 12,
            fontWeight: 400,
            lineHeight: 1.6,
          }}
        >
          Build and orchestrate your automation workflows with MCP tools.
        </p>
      </div>

      {/* Stats */}
      <Row gutter={[16, 16]} style={{ marginBottom: 40 }}>
        <Col xs={24} sm={8}>
          <div
            style={{
              background: '#fff',
              borderRadius: 14,
              padding: '24px 24px',
              border: '1px solid #eaecf0',
            }}
          >
            <div style={{ fontSize: 12, color: '#98a2b3', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Workflows
            </div>
            <div style={{ fontSize: 32, fontWeight: 600, color: '#1a1a2e', marginTop: 4 }}>
              {workflows.length}
            </div>
            <div style={{ fontSize: 13, color: '#667085', marginTop: 2 }}>
              {totalNodes} nodes total
            </div>
          </div>
        </Col>
        <Col xs={24} sm={8}>
          <div
            style={{
              background: '#fff',
              borderRadius: 14,
              padding: '24px 24px',
              border: '1px solid #eaecf0',
            }}
          >
            <div style={{ fontSize: 12, color: '#98a2b3', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              MCP Servers
            </div>
            <div style={{ fontSize: 32, fontWeight: 600, color: '#1a1a2e', marginTop: 4 }}>
              {servers.length}
            </div>
            <div style={{ fontSize: 13, color: '#667085', marginTop: 2 }}>
              {activeServers} online
            </div>
          </div>
        </Col>
        <Col xs={24} sm={8}>
          <div
            style={{
              background: '#fff',
              borderRadius: 14,
              padding: '24px 24px',
              border: '1px solid #eaecf0',
            }}
          >
            <div style={{ fontSize: 12, color: '#98a2b3', fontWeight: 500, textTransform: 'uppercase', letterSpacing: 0.5 }}>
              Status
            </div>
            <div style={{ fontSize: 32, fontWeight: 600, color: '#12b76a', marginTop: 4 }}>
              Healthy
            </div>
            <div style={{ fontSize: 13, color: '#667085', marginTop: 2 }}>
              All systems operational
            </div>
          </div>
        </Col>
      </Row>

      {/* Quick actions */}
      <div style={{ marginBottom: 40 }}>
        <div
          style={{
            fontSize: 13,
            color: '#98a2b3',
            fontWeight: 500,
            textTransform: 'uppercase',
            letterSpacing: 0.5,
            marginBottom: 16,
          }}
        >
          Quick Actions
        </div>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12}>
            <div
              onClick={() => navigate('/workflows/new')}
              style={{
                background: 'linear-gradient(135deg, #3b5bdb 0%, #5c7cfa 100%)',
                borderRadius: 14,
                padding: '28px 28px',
                cursor: 'pointer',
                transition: 'all 0.2s',
                color: '#fff',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 12 }}>
                <div
                  style={{
                    width: 40,
                    height: 40,
                    borderRadius: 10,
                    background: 'rgba(255,255,255,0.2)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 18,
                  }}
                >
                  <PlusOutlined />
                </div>
                <div>
                  <div style={{ fontSize: 16, fontWeight: 600 }}>New Workflow</div>
                  <div style={{ fontSize: 13, opacity: 0.8 }}>
                    Create a new automation pipeline
                  </div>
                </div>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={12}>
            <div
              onClick={() => navigate('/mcp-servers')}
              style={{
                background: '#fff',
                borderRadius: 14,
                padding: '28px 28px',
                cursor: 'pointer',
                transition: 'all 0.2s',
                border: '1px solid #eaecf0',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 12 }}>
                <div
                  style={{
                    width: 40,
                    height: 40,
                    borderRadius: 10,
                    background: '#f0f5ff',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 18,
                    color: '#3b5bdb',
                  }}
                >
                  <CloudServerOutlined />
                </div>
                <div>
                  <div style={{ fontSize: 16, fontWeight: 600, color: '#1a1a2e' }}>
                    Manage Servers
                  </div>
                  <div style={{ fontSize: 13, color: '#667085' }}>
                    Connect and configure MCP services
                  </div>
                </div>
              </div>
            </div>
          </Col>
        </Row>
      </div>

      {/* Recent workflows */}
      {workflows.length > 0 && (
        <div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              marginBottom: 16,
            }}
          >
            <div
              style={{
                fontSize: 13,
                color: '#98a2b3',
                fontWeight: 500,
                textTransform: 'uppercase',
                letterSpacing: 0.5,
              }}
            >
              Recent Workflows
            </div>
            <Button
              type="link"
              size="small"
              onClick={() => navigate('/workflows')}
              style={{ color: '#3b5bdb', fontSize: 13 }}
            >
              View All <ArrowRightOutlined />
            </Button>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {workflows.slice(0, 5).map((w) => (
              <div
                key={w.id}
                onClick={() => navigate(`/workflows/${w.id}`)}
                style={{
                  background: '#fff',
                  borderRadius: 10,
                  padding: '14px 20px',
                  border: '1px solid #eaecf0',
                  cursor: 'pointer',
                  transition: 'all 0.15s',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <ApartmentOutlined style={{ color: '#3b5bdb', fontSize: 16 }} />
                  <div>
                    <div style={{ fontWeight: 500, fontSize: 14, color: '#1a1a2e' }}>
                      {w.name || 'Untitled'}
                    </div>
                    {w.description && (
                      <div style={{ fontSize: 12, color: '#98a2b3', marginTop: 2 }}>
                        {w.description}
                      </div>
                    )}
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <span style={{ fontSize: 12, color: '#98a2b3' }}>
                    {w.nodes?.length || 0} nodes
                  </span>
                  <ArrowRightOutlined style={{ color: '#d0d5dd', fontSize: 12 }} />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
