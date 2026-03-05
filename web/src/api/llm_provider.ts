import request from './request'

export interface LLMProvider {
  id: string
  name: string
  base_url: string
  api_key: string
  models: string[]
  created_at: string
  updated_at: string
}

export const llmProviderApi = {
  list: () => request.get('/llm-providers'),
  get: (id: string) => request.get(`/llm-providers/${id}`),
  create: (data: Partial<LLMProvider>) => request.post('/llm-providers', data),
  update: (id: string, data: Partial<LLMProvider>) => request.put(`/llm-providers/${id}`, data),
  delete: (id: string) => request.delete(`/llm-providers/${id}`),
}
