import request from './request'

export interface Execution {
  id: string
  workflow_id: string
  status: string
  input: Record<string, any>
  output: Record<string, any>
  node_states: Record<string, NodeState>
  error: string
  started_at: string
  finished_at: string
  created_at: string
}

export interface NodeState {
  node_id: string
  status: string
  input: any
  output: any
  error: string
  duration: number
}

export interface ExecutionLog {
  id: string
  execution_id: string
  node_id: string
  node_name: string
  node_type: string
  attempt: number
  status: string
  input: Record<string, any>
  output: Record<string, any>
  error: string
  duration: number
  created_at: string
}

export const executionApi = {
  list: (page = 1, size = 20) =>
    request.get('/executions', { params: { page, size } }),
  get: (id: string) => request.get(`/executions/${id}`),
  logs: (id: string) => request.get(`/executions/${id}/logs`),
  delete: (id: string) => request.delete(`/executions/${id}`),
  listByWorkflow: (workflowId: string, page = 1, size = 20) =>
    request.get(`/workflows/${workflowId}/executions`, { params: { page, size } }),
  stats: () => request.get('/stats'),
}
