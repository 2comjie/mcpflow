import request from './request'

export interface AgentChatRequest {
  llm_provider_id: number
  mcp_server_ids: number[]
  message: string
  system_msg?: string
  max_iterations?: number
  temperature?: number
  max_tokens?: number
}

export interface AgentChatResponse {
  content: string
  agent_steps: any[]
  tool_calls_count: number
  iterations: number
  total_tokens: number
}

export const agentApi = {
  chat: (data: AgentChatRequest) => request.post('/agent/chat', data),
}
