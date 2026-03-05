import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import AppLayout from './components/Layout';
import Dashboard from './pages/dashboard/Dashboard';
import WorkflowList from './pages/workflow/WorkflowList';
import WorkflowEditor from './pages/workflow/WorkflowEditor';
import ExecutionList from './pages/execution/ExecutionList';
import ExecutionDetail from './pages/execution/ExecutionDetail';
import MCPServerList from './pages/mcpserver/MCPServerList';
import LLMProviderList from './pages/llmprovider/LLMProviderList';

export default function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <BrowserRouter>
        <Routes>
          <Route element={<AppLayout />}>
            <Route path="/" element={<Dashboard />} />
            <Route path="/workflows" element={<WorkflowList />} />
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
            <Route path="/executions" element={<ExecutionList />} />
            <Route path="/executions/:id" element={<ExecutionDetail />} />
            <Route path="/mcp-servers" element={<MCPServerList />} />
            <Route path="/llm-providers" element={<LLMProviderList />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  );
}
