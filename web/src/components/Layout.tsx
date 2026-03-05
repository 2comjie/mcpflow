import { Layout, Menu } from 'antd';
import {
  DashboardOutlined,
  ApartmentOutlined,
  HistoryOutlined,
  CloudServerOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';

const { Sider, Content, Header } = Layout;

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: 'Dashboard' },
  { key: '/workflows', icon: <ApartmentOutlined />, label: '工作流' },
  { key: '/executions', icon: <HistoryOutlined />, label: '执行记录' },
  { key: '/mcp-servers', icon: <CloudServerOutlined />, label: 'MCP 服务器' },
  { key: '/llm-providers', icon: <ThunderboltOutlined />, label: 'LLM 供应商' },
];

export default function AppLayout() {
  const navigate = useNavigate();
  const location = useLocation();

  const selectedKey = menuItems.find(
    (item) => item.key !== '/' && location.pathname.startsWith(item.key)
  )?.key || '/';

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={200} theme="light" style={{ borderRight: '1px solid #f0f0f0' }}>
        <div style={{ height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, fontSize: 18, borderBottom: '1px solid #f0f0f0' }}>
          MCPFlow
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ borderRight: 0 }}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', borderBottom: '1px solid #f0f0f0', height: 48, lineHeight: '48px', fontSize: 14, color: '#666' }}>
          MCP 协议驱动的智能工作流引擎
        </Header>
        <Content style={{ margin: 24, overflow: 'auto' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
