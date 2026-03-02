import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import MainLayout from './layouts/MainLayout'
import WorkflowList from './pages/workflow/WorkflowList'
import WorkflowEditor from './pages/workflow/WorkflowEditor'
import MCPServerList from './pages/mcpserver/MCPServerList'

export default function App() {
  return (
    <ConfigProvider>
      <BrowserRouter>
        <Routes>
          <Route element={<MainLayout />}>
            <Route path="/" element={<Navigate to="/workflows" replace />} />
            <Route path="/workflows" element={<WorkflowList />} />
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
            <Route path="/mcp-servers" element={<MCPServerList />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  )
}
