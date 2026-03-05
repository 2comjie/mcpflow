import request from './request';

export const mcpServerApi = {
  list: () => request.get('/mcp-servers'),
  get: (id: string) => request.get(`/mcp-servers/${id}`),
  create: (data: any) => request.post('/mcp-servers', data),
  update: (id: string, data: any) => request.put(`/mcp-servers/${id}`, data),
  delete: (id: string) => request.delete(`/mcp-servers/${id}`),
  check: (id: string) => request.post(`/mcp-servers/${id}/check`),
};
