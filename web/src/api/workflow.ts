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

export interface InputDef {
  name: string
  type: string // string, number, boolean, text
  required: boolean
  description?: string
  default?: string
}

export interface StartConfig {
  input_defs?: InputDef[]
}

export interface NodeConfig {
  start?: StartConfig
  llm?: LLMConfig
  agent?: AgentConfig
  condition?: ConditionConfig
  code?: CodeConfig
  http?: HTTPConfig
  email?: EmailConfig
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

export interface AgentMCPServer {
  url: string
  headers?: Record<string, string>
}

export interface AgentConfig {
  base_url: string
  api_key: string
  model: string
  prompt: string
  system_msg?: string
  mcp_servers: AgentMCPServer[]
  max_iterations?: number
  temperature?: number
  max_tokens?: number
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

export interface EmailConfig {
  smtp_host: string
  smtp_port: number
  username: string
  password: string
  from: string
  to: string
  cc?: string
  subject: string
  body: string
  content_type?: string
}

export interface Edge {
  id: string
  source: string
  target: string
  condition?: string
}

export const workflowApi = {
  list: (page = 1, page_size = 20) => request.get('/workflows', { params: { page, page_size } }),
  get: (id: string) => request.get(`/workflows/${id}`),
  create: (data: Partial<Workflow>) => request.post('/workflows', data),
  update: (id: string, data: Partial<Workflow>) => request.put(`/workflows/${id}`, data),
  delete: (id: string) => request.delete(`/workflows/${id}`),
  execute: (id: string, input?: Record<string, any>) =>
    request.post(`/workflows/${id}/execute`, input || {}),
  listExecutions: (id: string, page = 1, page_size = 20) =>
    request.get(`/workflows/${id}/executions`, { params: { page, page_size } }),
}
