import { useEffect, useState, useRef } from 'react'
import { Modal, Input, Radio, message, Tag } from 'antd'
import {
  LoadingOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { workflowApi } from '../api/workflow'

interface Props {
  open: boolean
  workflowId: string | null
  nodes?: any[] // 如果已有节点数据可直接传入，省去一次请求
  onClose: () => void
  onSuccess?: (executionId: string) => void
}

interface NodeEventData {
  node_id: string
  node_name: string
  node_type: string
  status: string
  error?: string
  duration?: number
}

/** 从节点配置中提取 {{.input.xxx}} 和 {{.xxx}} 变量名 */
export function extractInputVars(nodes: any[]): string[] {
  const vars = new Set<string>()
  const configStr = JSON.stringify((nodes || []).map((n: any) => n.config || (n.data?.config) || {}))
  const re = /\{\{\s*\.(?:input\.)?(\w+)\s*\}\}/g
  let m
  while ((m = re.exec(configStr)) !== null) {
    const name = m[1]
    if (!name.startsWith('node_') && name !== 'input') {
      vars.add(name)
    }
  }
  return Array.from(vars)
}

export default function ExecuteWorkflowModal({ open, workflowId, nodes: propNodes, onClose, onSuccess }: Props) {
  const [inputVars, setInputVars] = useState<string[]>([])
  const [formValues, setFormValues] = useState<Record<string, string>>({})
  const [jsonInput, setJsonInput] = useState('{}')
  const [mode, setMode] = useState<'form' | 'json'>('form')
  const [loading, setLoading] = useState(false)
  const [_fetching, setFetching] = useState(false)

  // SSE streaming state
  const [streaming, setStreaming] = useState(false)
  const [nodeEvents, setNodeEvents] = useState<NodeEventData[]>([])
  const [executionId, setExecutionId] = useState<string | null>(null)
  const [streamError, setStreamError] = useState<string | null>(null)
  const [streamDone, setStreamDone] = useState(false)
  const abortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    if (!open || !workflowId) return

    if (propNodes) {
      const vars = extractInputVars(propNodes)
      setInputVars(vars)
      setMode(vars.length > 0 ? 'form' : 'json')
    } else {
      setFetching(true)
      workflowApi
        .get(workflowId)
        .then((wf: any) => {
          const vars = extractInputVars(wf.nodes || [])
          setInputVars(vars)
          setMode(vars.length > 0 ? 'form' : 'json')
        })
        .catch((err: any) => message.error(err.message))
        .finally(() => setFetching(false))
    }

    setFormValues({})
    setJsonInput('{}')
    setStreaming(false)
    setNodeEvents([])
    setExecutionId(null)
    setStreamError(null)
    setStreamDone(false)
  }, [open, workflowId])

  const handleExecute = async () => {
    if (!workflowId) return
    let input: Record<string, any> = {}
    if (mode === 'json') {
      try {
        input = JSON.parse(jsonInput)
      } catch {
        message.error('Invalid JSON')
        return
      }
    } else {
      input = { ...formValues }
    }

    // Use SSE streaming
    setStreaming(true)
    setNodeEvents([])
    setExecutionId(null)
    setStreamError(null)
    setStreamDone(false)
    setLoading(true)

    const abort = new AbortController()
    abortRef.current = abort

    try {
      const resp = await fetch(`/api/v1/workflows/${workflowId}/execute/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
        signal: abort.signal,
      })

      if (!resp.ok || !resp.body) {
        const text = await resp.text()
        throw new Error(text || 'Stream request failed')
      }

      const reader = resp.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        let currentEvent = ''
        for (const line of lines) {
          if (line.startsWith('event: ')) {
            currentEvent = line.slice(7).trim()
          } else if (line.startsWith('data: ')) {
            const dataStr = line.slice(6)
            try {
              const data = JSON.parse(dataStr)
              if (currentEvent === 'execution_id') {
                setExecutionId(data.execution_id)
              } else if (currentEvent === 'node_event') {
                setNodeEvents((prev) => {
                  // Update existing node or add new
                  const existing = prev.findIndex((e) => e.node_id === data.node_id && e.status === 'running')
                  if (existing >= 0 && data.status !== 'running') {
                    const updated = [...prev]
                    updated[existing] = data
                    return updated
                  }
                  return [...prev, data]
                })
              } else if (currentEvent === 'error') {
                setStreamError(data.error)
              } else if (currentEvent === 'done') {
                setStreamDone(true)
              }
            } catch {
              // ignore parse errors
            }
            currentEvent = ''
          }
        }
      }
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        setStreamError(err.message)
      }
    } finally {
      setLoading(false)
      setStreamDone(true)
    }
  }

  const handleClose = () => {
    abortRef.current?.abort()
    onClose()
    if (executionId) {
      onSuccess?.(executionId)
    }
  }

  const statusIcon = (status: string) => {
    if (status === 'running') return <LoadingOutlined style={{ color: '#3b5bdb' }} />
    if (status === 'completed') return <CheckCircleOutlined style={{ color: '#12b76a' }} />
    if (status === 'failed') return <CloseCircleOutlined style={{ color: '#f04438' }} />
    return null
  }

  return (
    <Modal
      title="Execute Workflow"
      open={open}
      onCancel={handleClose}
      onOk={streaming ? handleClose : handleExecute}
      okText={streaming ? (streamDone ? 'Close' : 'Running...') : 'Execute'}
      confirmLoading={loading && !streaming}
      okButtonProps={streaming && !streamDone ? { disabled: true } : undefined}
    >
      {!streaming ? (
        <>
          {inputVars.length > 0 && (
            <Radio.Group
              value={mode}
              onChange={(e) => setMode(e.target.value)}
              size="small"
              style={{ marginBottom: 12 }}
            >
              <Radio.Button value="form">Form</Radio.Button>
              <Radio.Button value="json">JSON</Radio.Button>
            </Radio.Group>
          )}

          {mode === 'form' && inputVars.length > 0 ? (
            <div>
              <div style={{ marginBottom: 8, color: '#667085', fontSize: 12 }}>
                Detected {inputVars.length} input variable{inputVars.length > 1 ? 's' : ''}:
              </div>
              {inputVars.map((varName) => (
                <div key={varName} style={{ marginBottom: 12 }}>
                  <div style={{ fontSize: 12, fontWeight: 500, marginBottom: 4, color: '#344054' }}>
                    {varName}
                  </div>
                  <Input
                    value={formValues[varName] || ''}
                    onChange={(e) =>
                      setFormValues((prev) => ({ ...prev, [varName]: e.target.value }))
                    }
                    placeholder={`Enter ${varName}...`}
                    style={{ borderRadius: 8 }}
                  />
                </div>
              ))}
            </div>
          ) : (
            <>
              <div style={{ marginBottom: 8, color: '#667085', fontSize: 13 }}>
                Input parameters (JSON):
              </div>
              <Input.TextArea
                value={jsonInput}
                onChange={(e) => setJsonInput(e.target.value)}
                rows={6}
                placeholder='{"city": "Beijing"}'
                style={{ fontFamily: 'monospace', fontSize: 12, borderRadius: 8 }}
              />
            </>
          )}
        </>
      ) : (
        <div>
          {executionId && (
            <div style={{ marginBottom: 12, fontSize: 12, color: '#98a2b3' }}>
              Execution #{executionId}
            </div>
          )}

          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {nodeEvents.map((evt, i) => (
              <div
                key={`${evt.node_id}-${evt.status}-${i}`}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  padding: '8px 12px',
                  background: '#f9fafb',
                  borderRadius: 8,
                  border: '1px solid #eaecf0',
                }}
              >
                {statusIcon(evt.status)}
                <div style={{ flex: 1, minWidth: 0 }}>
                  <span style={{ fontSize: 13, fontWeight: 500, color: '#1a1a2e' }}>
                    {evt.node_name || evt.node_id}
                  </span>
                  <Tag
                    style={{ marginLeft: 8, fontSize: 10, borderRadius: 4 }}
                    color={
                      evt.node_type === 'agent' ? 'purple' :
                      evt.node_type === 'llm' ? 'blue' :
                      evt.node_type === 'condition' ? 'orange' :
                      undefined
                    }
                  >
                    {evt.node_type}
                  </Tag>
                </div>
                {evt.duration !== undefined && evt.duration > 0 && (
                  <span style={{ fontSize: 11, color: '#98a2b3' }}>{evt.duration}ms</span>
                )}
                {evt.error && (
                  <span style={{ fontSize: 11, color: '#f04438' }}>{evt.error}</span>
                )}
              </div>
            ))}
          </div>

          {streamError && (
            <div style={{ marginTop: 12, padding: 8, background: '#fef3f2', borderRadius: 8, fontSize: 12, color: '#f04438' }}>
              {streamError}
            </div>
          )}

          {streamDone && !streamError && (
            <div style={{ marginTop: 12, padding: 8, background: '#f0fdf4', borderRadius: 8, fontSize: 12, color: '#12b76a' }}>
              Execution completed successfully
            </div>
          )}
        </div>
      )}
    </Modal>
  )
}
