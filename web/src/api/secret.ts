import request from './request'

export interface Secret {
  id: number
  key: string
  value?: string
  desc: string
  created_at: string
  updated_at: string
}

export const secretApi = {
  list: () => request.get('/secrets'),
  create: (data: { key: string; value: string; desc?: string }) => request.post('/secrets', data),
  update: (id: number, data: { value: string; desc?: string }) =>
    request.put(`/secrets/${id}`, data),
  delete: (id: number) => request.delete(`/secrets/${id}`),
}
