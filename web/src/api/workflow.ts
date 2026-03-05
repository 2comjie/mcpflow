import request from './request';

export const workflowApi = {
  list: (page = 1, pageSize = 20) =>
    request.get('/workflows', { params: { page, page_size: pageSize } }),
  get: (id: string) => request.get(`/workflows/${id}`),
  create: (data: any) => request.post('/workflows', data),
  update: (id: string, data: any) => request.put(`/workflows/${id}`, data),
  delete: (id: string) => request.delete(`/workflows/${id}`),
  execute: (id: string, input?: any) =>
    request.post(`/workflows/${id}/execute`, input || {}),
  listExecutions: (id: string, page = 1, pageSize = 20) =>
    request.get(`/workflows/${id}/executions`, { params: { page, page_size: pageSize } }),
};
