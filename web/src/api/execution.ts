import request from './request'

export interface Execution {
  id: number
  workflow_id: number
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
  id: number
  execution_id: number
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
  get: (id: number) => request.get(`/executions/${id}`),
  logs: (id: number) => request.get(`/executions/${id}/logs`),
  cancel: (id: number) => request.post(`/executions/${id}/cancel`),
  listByWorkflow: (workflowId: number, page = 1, pageSize = 20) =>
    request.get(`/workflows/${workflowId}/executions`, { params: { page, page_size: pageSize } }),
}
