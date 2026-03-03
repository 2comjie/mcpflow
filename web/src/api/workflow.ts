import request from './request'

export interface Workflow {
  id: number
  name: string
  description: string
  nodes: Node[]
  edges: Edge[]
  created_at: string
  updated_at: string
}

export interface Node {
  id: string
  type: string
  name: string
  config: NodeConfig
  position: { x: number; y: number }
  timeout?: number
  retry?: { max: number; interval: number }
}

export interface NodeConfig {
  llm?: LLMConfig
  mcp?: MCPConfig
  condition?: ConditionConfig
  code?: CodeConfig
  http?: HTTPConfig
}

export interface LLMConfig {
  base_url: string
  api_key: string
  model: string
  prompt: string
  system_msg?: string
  temperature?: number
  max_tokens?: number
}

export interface MCPConfig {
  action: string // call_tool, get_prompt, read_resource
  server_url: string
  headers?: Record<string, string>
  tool_name?: string
  arguments?: Record<string, any>
  prompt_name?: string
  prompt_args?: Record<string, any>
  resource_uri?: string
}

export interface ConditionConfig {
  expression: string
}

export interface CodeConfig {
  language: string
  code: string
}

export interface HTTPConfig {
  method: string
  url: string
  headers?: Record<string, string>
  body?: string
}

export interface Edge {
  id: string
  source: string
  target: string
  condition?: string
}

export const workflowApi = {
  list: (page = 1, size = 20) => request.get('/workflows', { params: { page, size } }),
  get: (id: number) => request.get(`/workflows/${id}`),
  create: (data: Partial<Workflow>) => request.post('/workflows', data),
  update: (id: number, data: Partial<Workflow>) => request.put(`/workflows/${id}`, data),
  delete: (id: number) => request.delete(`/workflows/${id}`),
  execute: (id: number, input?: Record<string, any>) =>
    request.post(`/workflows/${id}/execute`, input || {}),
  listExecutions: (id: number, page = 1, size = 20) =>
    request.get(`/workflows/${id}/executions`, { params: { page, size } }),
}
