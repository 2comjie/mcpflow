import { Layout, Menu } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { ApartmentOutlined, CloudServerOutlined } from '@ant-design/icons'

const { Header, Content, Sider } = Layout

const menuItems = [
  { key: '/workflows', icon: <ApartmentOutlined />, label: 'Workflows' },
  { key: '/mcp-servers', icon: <CloudServerOutlined />, label: 'MCP Servers' },
]

export default function MainLayout() {
  const navigate = useNavigate()
  const location = useLocation()

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider theme="light" width={200}>
        <div style={{ height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, fontSize: 18 }}>
          MCPFlow
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname.startsWith('/mcp-servers') ? '/mcp-servers' : '/workflows']}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', borderBottom: '1px solid #f0f0f0' }} />
        <Content style={{ margin: 16 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
