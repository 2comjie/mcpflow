import { useEffect, useState, useRef } from 'react'
import { Modal, Input, InputNumber, Switch, Radio, message, Tag } from 'antd'
import {
  LoadingOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { workflowApi } from '../api/workflow'

interface InputDef {
  name: string
  type: string
  required: boolean
  description?: string
  default?: string
}

interface Props {
  open: boolean
  workflowId: string | null
  nodes?: any[]
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

/** 从 Start 节点的 config.start.input_defs 提取输入定义 */
function extractInputDefs(nodes: any[]): InputDef[] {
  const startNode = (nodes || []).find(
    (n: any) => n.type === 'start' || (n.data?.nodeType) === 'start',
  )
  if (!startNode) return []
  const config = startNode.config || startNode.data?.config || {}
  return config.start?.input_defs || []
}

/** 从节点配置模板中提取 {{input.xxx}} 变量名（fallback） */
function extractTemplateVars(nodes: any[]): string[] {
  const vars = new Set<string>()
  const configStr = JSON.stringify((nodes || []).map((n: any) => n.config || (n.data?.config) || {}))
  const re = /\{\{input\.(\w+)\}\}/g
  let m
  while ((m = re.exec(configStr)) !== null) {
    vars.add(m[1])
  }
  return Array.from(vars)
}

export default function ExecuteWorkflowModal({ open, workflowId, nodes: propNodes, onClose, onSuccess }: Props) {
  const [inputDefs, setInputDefs] = useState<InputDef[]>([])
  const [formValues, setFormValues] = useState<Record<string, any>>({})
  const [jsonInput, setJsonInput] = useState('{}')
  const [mode, setMode] = useState<'form' | 'json'>('form')
  const [loading, setLoading] = useState(false)

  // SSE streaming state
  const [streaming, setStreaming] = useState(false)
  const [nodeEvents, setNodeEvents] = useState<NodeEventData[]>([])
  const [executionId, setExecutionId] = useState<string | null>(null)
  const [streamError, setStreamError] = useState<string | null>(null)
  const [streamDone, setStreamDone] = useState(false)
  const abortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    if (!open || !workflowId) return

    const processNodes = (nodes: any[]) => {
      // 优先用 Start 节点的 input_defs
      let defs = extractInputDefs(nodes)
      if (defs.length === 0) {
        // fallback: 从模板变量提取
        const vars = extractTemplateVars(nodes)
        defs = vars.map((name) => ({ name, type: 'string', required: false }))
      }
      setInputDefs(defs)
      setMode(defs.length > 0 ? 'form' : 'json')
      // 设置默认值
      const defaults: Record<string, any> = {}
      for (const def of defs) {
        if (def.default) {
          defaults[def.name] = def.type === 'number' ? Number(def.default) :
            def.type === 'boolean' ? def.default === 'true' : def.default
        }
      }
      setFormValues(defaults)
    }

    if (propNodes) {
      processNodes(propNodes)
    } else {
      workflowApi
        .get(workflowId)
        .then((res: any) => {
          const wf = res.data || res
          processNodes(wf.nodes || [])
        })
        .catch((err: any) => message.error(err.message))
    }

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
      // 按类型转换
      for (const def of inputDefs) {
        const val = formValues[def.name]
        if (def.required && (val === undefined || val === '')) {
          message.warning(`请填写必填参数: ${def.name}`)
          return
        }
        if (val !== undefined && val !== '') {
          input[def.name] = val
        }
      }
    }

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

  const renderFormField = (def: InputDef) => {
    const label = (
      <span>
        {def.name}
        {def.required && <span style={{ color: '#f04438', marginLeft: 2 }}>*</span>}
        {def.description && <span style={{ color: '#98a2b3', fontWeight: 400, marginLeft: 6 }}>{def.description}</span>}
      </span>
    )

    switch (def.type) {
      case 'number':
        return (
          <div key={def.name} style={{ marginBottom: 12 }}>
            <div style={{ fontSize: 12, fontWeight: 500, marginBottom: 4, color: '#344054' }}>{label}</div>
            <InputNumber
              value={formValues[def.name]}
              onChange={(v) => setFormValues((prev) => ({ ...prev, [def.name]: v }))}
              placeholder={def.default || `输入 ${def.name}`}
              style={{ width: '100%', borderRadius: 8 }}
            />
          </div>
        )
      case 'boolean':
        return (
          <div key={def.name} style={{ marginBottom: 12, display: 'flex', alignItems: 'center', gap: 8 }}>
            <div style={{ fontSize: 12, fontWeight: 500, color: '#344054', flex: 1 }}>{label}</div>
            <Switch
              checked={!!formValues[def.name]}
              onChange={(v) => setFormValues((prev) => ({ ...prev, [def.name]: v }))}
            />
          </div>
        )
      case 'text':
        return (
          <div key={def.name} style={{ marginBottom: 12 }}>
            <div style={{ fontSize: 12, fontWeight: 500, marginBottom: 4, color: '#344054' }}>{label}</div>
            <Input.TextArea
              value={formValues[def.name] || ''}
              onChange={(e) => setFormValues((prev) => ({ ...prev, [def.name]: e.target.value }))}
              placeholder={def.default || `输入 ${def.name}`}
              rows={4}
              style={{ borderRadius: 8 }}
            />
          </div>
        )
      default: // string
        return (
          <div key={def.name} style={{ marginBottom: 12 }}>
            <div style={{ fontSize: 12, fontWeight: 500, marginBottom: 4, color: '#344054' }}>{label}</div>
            <Input
              value={formValues[def.name] || ''}
              onChange={(e) => setFormValues((prev) => ({ ...prev, [def.name]: e.target.value }))}
              placeholder={def.default || `输入 ${def.name}`}
              style={{ borderRadius: 8 }}
            />
          </div>
        )
    }
  }

  return (
    <Modal
      title="执行工作流"
      open={open}
      onCancel={handleClose}
      onOk={streaming ? handleClose : handleExecute}
      okText={streaming ? (streamDone ? '关闭' : '执行中...') : '执行'}
      confirmLoading={loading && !streaming}
      okButtonProps={streaming && !streamDone ? { disabled: true } : undefined}
    >
      {!streaming ? (
        <>
          {inputDefs.length > 0 && (
            <Radio.Group
              value={mode}
              onChange={(e) => setMode(e.target.value)}
              size="small"
              style={{ marginBottom: 12 }}
            >
              <Radio.Button value="form">表单</Radio.Button>
              <Radio.Button value="json">JSON</Radio.Button>
            </Radio.Group>
          )}

          {mode === 'form' && inputDefs.length > 0 ? (
            <div>
              <div style={{ marginBottom: 8, color: '#667085', fontSize: 12 }}>
                {inputDefs.length} 个输入参数
              </div>
              {inputDefs.map((def) => renderFormField(def))}
            </div>
          ) : (
            <>
              <div style={{ marginBottom: 8, color: '#667085', fontSize: 13 }}>
                输入参数 (JSON):
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
              执行记录 #{executionId}
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
              执行完成
            </div>
          )}
        </div>
      )}
    </Modal>
  )
}
