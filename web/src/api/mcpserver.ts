import request from './request'

export interface MCPServer {
  id: number
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
  update: (id: number, data: Partial<MCPServer>) => request.put(`/mcp-servers/${id}`, data),
  delete: (id: number) => request.delete(`/mcp-servers/${id}`),
  test: (id: number) => request.post(`/mcp-servers/${id}/test`),
  tools: (id: number) => request.get(`/mcp-servers/${id}/tools`),
  prompts: (id: number) => request.get(`/mcp-servers/${id}/prompts`),
  resources: (id: number) => request.get(`/mcp-servers/${id}/resources`),
}
