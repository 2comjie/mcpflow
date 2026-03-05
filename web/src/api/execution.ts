import request from './request';

export const executionApi = {
  list: (page = 1, pageSize = 20) =>
    request.get('/executions', { params: { page, page_size: pageSize } }),
  get: (id: string) => request.get(`/executions/${id}`),
  getLogs: (id: string) => request.get(`/executions/${id}/logs`),
  delete: (id: string) => request.delete(`/executions/${id}`),
};
