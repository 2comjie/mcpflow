import request from './request';

export const llmProviderApi = {
  list: () => request.get('/llm-providers'),
  get: (id: string) => request.get(`/llm-providers/${id}`),
  create: (data: any) => request.post('/llm-providers', data),
  update: (id: string, data: any) => request.put(`/llm-providers/${id}`, data),
  delete: (id: string) => request.delete(`/llm-providers/${id}`),
};
