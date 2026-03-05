import request from './request'

export interface MCPServer {
  id: string
  name: string
  description: string
  url: string
  headers: Record<string, string>
  status: string
  tools: any[]
  prompts: any[]
  resources: any[]
  checked_at: string
  created_at: string
  updated_at: string
}

export const mcpServerApi = {
  list: () => request.get('/mcp-servers'),
  create: (data: Partial<MCPServer>) => request.post('/mcp-servers', data),
  update: (id: string, data: Partial<MCPServer>) => request.put(`/mcp-servers/${id}`, data),
  delete: (id: string) => request.delete(`/mcp-servers/${id}`),
  check: (id: string) => request.post(`/mcp-servers/${id}/check`),
  callTool: (id: string, toolName: string, args: Record<string, any>) =>
    request.post(`/mcp-servers/${id}/tools/call`, { tool_name: toolName, arguments: args }),
}
