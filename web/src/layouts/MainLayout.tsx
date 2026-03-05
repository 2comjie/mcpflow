import { useState } from 'react'
import { Layout, Menu, Tooltip } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  HomeOutlined,
  ApartmentOutlined,
  CloudServerOutlined,
  RobotOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  ThunderboltOutlined,
  ExperimentOutlined,
} from '@ant-design/icons'

const { Content, Sider } = Layout

const menuItems = [
  { key: '/', icon: <HomeOutlined />, label: 'Home' },
  { key: '/workflows', icon: <ApartmentOutlined />, label: 'Workflows' },
  { key: '/executions', icon: <ThunderboltOutlined />, label: 'Executions' },
  { key: '/mcp-servers', icon: <CloudServerOutlined />, label: 'MCP Servers' },
  { key: '/llm-providers', icon: <RobotOutlined />, label: 'LLM Providers' },
  { key: '/agent', icon: <ExperimentOutlined />, label: 'Agent Playground' },
]

export default function MainLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const [collapsed, setCollapsed] = useState(false)

  const getSelectedKey = () => {
    if (location.pathname.startsWith('/agent')) return '/agent'
    if (location.pathname.startsWith('/llm-providers')) return '/llm-providers'
    if (location.pathname.startsWith('/mcp-servers')) return '/mcp-servers'
    if (location.pathname === '/executions') return '/executions'
    if (location.pathname.startsWith('/workflows')) return '/workflows'
    return '/'
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        className="sidebar"
        width={220}
        collapsedWidth={64}
        collapsed={collapsed}
        theme="light"
        style={{
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 100,
          overflow: 'auto',
        }}
      >
        <div
          style={{
            height: 56,
            display: 'flex',
            alignItems: 'center',
            justifyContent: collapsed ? 'center' : 'space-between',
            padding: collapsed ? '0' : '0 16px',
            borderBottom: '1px solid #eaecf0',
          }}
        >
          {collapsed ? (
            <Tooltip title="MCPFlow" placement="right">
              <div
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: 8,
                  background: 'linear-gradient(135deg, #3b5bdb 0%, #5c7cfa 100%)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: '#fff',
                  fontWeight: 700,
                  fontSize: 14,
                  cursor: 'pointer',
                }}
                onClick={() => navigate('/')}
              >
                M
              </div>
            </Tooltip>
          ) : (
            <div
              style={{ display: 'flex', alignItems: 'center', gap: 10, cursor: 'pointer' }}
              onClick={() => navigate('/')}
            >
              <div
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: 8,
                  background: 'linear-gradient(135deg, #3b5bdb 0%, #5c7cfa 100%)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: '#fff',
                  fontWeight: 700,
                  fontSize: 14,
                }}
              >
                M
              </div>
              <span style={{ fontWeight: 700, fontSize: 16, color: '#1a1a2e', letterSpacing: -0.3 }}>
                MCPFlow
              </span>
            </div>
          )}
        </div>

        <Menu
          mode="inline"
          selectedKeys={[getSelectedKey()]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ marginTop: 8 }}
        />

        <div
          style={{
            position: 'absolute',
            bottom: 16,
            left: 0,
            right: 0,
            display: 'flex',
            justifyContent: 'center',
          }}
        >
          <div
            onClick={() => setCollapsed(!collapsed)}
            style={{
              cursor: 'pointer',
              width: 32,
              height: 32,
              borderRadius: 8,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#98a2b3',
              transition: 'all 0.2s',
            }}
          >
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </div>
        </div>
      </Sider>

      <Layout style={{ marginLeft: collapsed ? 64 : 220, transition: 'margin-left 0.2s' }}>
        <Content style={{ padding: 24 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
