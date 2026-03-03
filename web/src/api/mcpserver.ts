import request from './request'

export interface MCPServer {
  id: number
  name: string
  description: string
  url: string
  headers: Record<string, string>
  status: string
  tools: Record<string, any>
  prompts: Record<string, any>
  resources: Record<string, any>
  checked_at: string
  created_at: string
  updated_at: string
}

export const mcpServerApi = {
  list: () => request.get('/mcp-servers'),
  get: (id: number) => request.get(`/mcp-servers/${id}`),
  create: (data: Partial<MCPServer>) => request.post('/mcp-servers', data),
  update: (id: number, data: Partial<MCPServer>) => request.put(`/mcp-servers/${id}`, data),
  delete: (id: number) => request.delete(`/mcp-servers/${id}`),
  test: (id: number) => request.post(`/mcp-servers/${id}/test`),
  ping: (id: number) => request.post(`/mcp-servers/${id}/ping`),
  healthCheckAll: () => request.post('/mcp-servers/health-check'),
  tools: (id: number) => request.get(`/mcp-servers/${id}/tools`),
  prompts: (id: number) => request.get(`/mcp-servers/${id}/prompts`),
  resources: (id: number) => request.get(`/mcp-servers/${id}/resources`),
}
