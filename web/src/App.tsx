import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import MainLayout from './layouts/MainLayout'
import Dashboard from './pages/dashboard/Dashboard'
import WorkflowList from './pages/workflow/WorkflowList'
import WorkflowEditor from './pages/workflow/WorkflowEditor'
import MCPServerList from './pages/mcpserver/MCPServerList'

const theme = {
  token: {
    colorPrimary: '#3b5bdb',
    borderRadius: 8,
    fontFamily:
      "'Inter', -apple-system, BlinkMacSystemFont, 'SF Pro Text', 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif",
    colorBgLayout: '#f5f7fa',
    colorBorderSecondary: '#eaecf0',
  },
  components: {
    Button: {
      borderRadius: 8,
    },
    Input: {
      borderRadius: 8,
    },
    Select: {
      borderRadius: 8,
    },
    Modal: {
      borderRadiusLG: 12,
    },
    Card: {
      borderRadiusLG: 12,
    },
    Menu: {
      itemBorderRadius: 8,
    },
  },
}

export default function App() {
  return (
    <ConfigProvider theme={theme}>
      <BrowserRouter>
        <Routes>
          <Route element={<MainLayout />}>
            <Route path="/" element={<Dashboard />} />
            <Route path="/workflows" element={<WorkflowList />} />
            <Route path="/mcp-servers" element={<MCPServerList />} />
          </Route>
          <Route path="/workflows/:id" element={<WorkflowEditor />} />
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}
