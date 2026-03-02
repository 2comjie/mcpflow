import { useCallback, useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Input, Space, message } from 'antd'
import { SaveOutlined, PlayCircleOutlined } from '@ant-design/icons'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  type Connection,
  type Node as FlowNode,
  type Edge as FlowEdge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { workflowApi } from '../../api/workflow'

const nodeTypes = [
  { type: 'start', label: 'Start', color: '#52c41a' },
  { type: 'end', label: 'End', color: '#ff4d4f' },
  { type: 'mcp_tool', label: 'MCP Tool', color: '#1890ff' },
  { type: 'mcp_prompt', label: 'MCP Prompt', color: '#722ed1' },
  { type: 'mcp_resource', label: 'MCP Resource', color: '#13c2c2' },
  { type: 'llm', label: 'LLM', color: '#fa8c16' },
  { type: 'condition', label: 'Condition', color: '#faad14' },
  { type: 'code', label: 'Code', color: '#2f54eb' },
  { type: 'http', label: 'HTTP', color: '#eb2f96' },
]

let nodeId = 0
const getId = () => `node_${++nodeId}`

export default function WorkflowEditor() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new'

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [nodes, setNodes, onNodesChange] = useNodesState<FlowNode>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<FlowEdge>([])

  useEffect(() => {
    if (!isNew && id) {
      workflowApi.get(Number(id)).then((res: any) => {
        setName(res.name)
        setDescription(res.description || '')
        const flowNodes = (res.nodes || []).map((n: any) => ({
          id: n.id,
          type: 'default',
          position: n.position || { x: 0, y: 0 },
          data: { label: `${n.name} (${n.type})`, nodeType: n.type, config: n.config },
          style: { background: nodeTypes.find(t => t.type === n.type)?.color || '#ddd', color: '#fff', borderRadius: 8 },
        }))
        const flowEdges = (res.edges || []).map((e: any) => ({
          id: e.id,
          source: e.source,
          target: e.target,
          label: e.condition || '',
          animated: true,
        }))
        setNodes(flowNodes)
        setEdges(flowEdges)
        nodeId = flowNodes.length
      }).catch((err: any) => message.error(err.message))
    }
  }, [id])

  const onConnect = useCallback((params: Connection) => {
    setEdges((eds) => addEdge({ ...params, animated: true }, eds))
  }, [setEdges])

  const onAddNode = (type: string, label: string, color: string) => {
    const newNode: FlowNode = {
      id: getId(),
      type: 'default',
      position: { x: 250, y: nodes.length * 100 + 50 },
      data: { label: `${label} (${type})`, nodeType: type, config: {} },
      style: { background: color, color: '#fff', borderRadius: 8 },
    }
    setNodes((nds) => [...nds, newNode])
  }

  const handleSave = async () => {
    const wfNodes = nodes.map((n) => ({
      id: n.id,
      type: (n.data as any).nodeType,
      name: String((n.data as any).label).split(' (')[0],
      config: (n.data as any).config || {},
      position: n.position,
    }))
    const wfEdges = edges.map((e) => ({
      id: e.id,
      source: e.source,
      target: e.target,
      condition: (e.label as string) || '',
    }))

    try {
      if (isNew) {
        await workflowApi.create({ name, description, nodes: wfNodes, edges: wfEdges })
        message.success('created')
        navigate('/workflows')
      } else {
        await workflowApi.update(Number(id), { name, description, nodes: wfNodes, edges: wfEdges })
        message.success('saved')
      }
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const handleExecute = async () => {
    if (isNew) return
    try {
      const res: any = await workflowApi.execute(Number(id))
      message.success(`executed, status: ${res.status}`)
    } catch (err: any) {
      message.error(err.message)
    }
  }

  return (
    <div style={{ height: 'calc(100vh - 112px)', display: 'flex', flexDirection: 'column' }}>
      <div style={{ marginBottom: 12, display: 'flex', gap: 12, alignItems: 'center' }}>
        <Input
          placeholder="Workflow name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          style={{ width: 240 }}
        />
        <Input
          placeholder="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          style={{ width: 320 }}
        />
        <Space>
          <Button type="primary" icon={<SaveOutlined />} onClick={handleSave}>Save</Button>
          {!isNew && (
            <Button icon={<PlayCircleOutlined />} onClick={handleExecute}>Execute</Button>
          )}
        </Space>
      </div>

      <div style={{ display: 'flex', flex: 1, gap: 12 }}>
        <div style={{ width: 140, background: '#fafafa', borderRadius: 8, padding: 8 }}>
          <div style={{ fontWeight: 600, marginBottom: 8 }}>Nodes</div>
          {nodeTypes.map((nt) => (
            <Button
              key={nt.type}
              size="small"
              block
              style={{ marginBottom: 4 }}
              onClick={() => onAddNode(nt.type, nt.label, nt.color)}
            >
              {nt.label}
            </Button>
          ))}
        </div>

        <div style={{ flex: 1, borderRadius: 8, overflow: 'hidden', border: '1px solid #f0f0f0' }}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            fitView
          >
            <Background />
            <Controls />
            <MiniMap />
          </ReactFlow>
        </div>
      </div>
    </div>
  )
}
