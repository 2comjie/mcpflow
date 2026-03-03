import request from './request'

export interface LLMProvider {
  id: number
  name: string
  base_url: string
  api_key: string
  models: string[]
  created_at: string
  updated_at: string
}

export const llmProviderApi = {
  list: () => request.get('/llm-providers'),
  create: (data: Partial<LLMProvider>) => request.post('/llm-providers', data),
  update: (id: number, data: Partial<LLMProvider>) => request.put(`/llm-providers/${id}`, data),
  delete: (id: number) => request.delete(`/llm-providers/${id}`),
}
