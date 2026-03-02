import request from './request'

export interface Workflow {
  id: number
  name: string
  description: string
  status: string
  nodes: Node[]
  edges: Edge[]
  variables: Record<string, any>
  created_at: string
  updated_at: string
}

export interface Node {
  id: string
  type: string
  name: string
  config: Record<string, any>
  position: { x: number; y: number }
}

export interface Edge {
  id: string
  source: string
  target: string
  source_port?: string
  target_port?: string
  condition?: string
}

export const workflowApi = {
  list: () => request.get('/workflows'),
  get: (id: number) => request.get(`/workflows/${id}`),
  create: (data: Partial<Workflow>) => request.post('/workflows', data),
  update: (id: number, data: Partial<Workflow>) => request.put(`/workflows/${id}`, data),
  delete: (id: number) => request.delete(`/workflows/${id}`),
  execute: (id: number, input?: Record<string, any>) =>
    request.post(`/workflows/${id}/execute`, { input }),
}
