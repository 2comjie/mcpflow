import { useCallback, useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Input, message, Tooltip, Divider, Form, Select, Tag } from 'antd'
import {
  SaveOutlined,
  PlayCircleOutlined,
  ArrowLeftOutlined,
  BranchesOutlined,
  CodeOutlined,
  ApiOutlined,
  RobotOutlined,
  FlagOutlined,
  CheckCircleOutlined,
  CloudOutlined,
  FileTextOutlined,
  GlobalOutlined,
  SettingOutlined,
  CloseOutlined,
} from '@ant-design/icons'
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

const nodeGroups = [
  {
    title: 'CONTROL FLOW',
    items: [
      { type: 'start', label: 'Start', color: '#12b76a', icon: <FlagOutlined /> },
      { type: 'end', label: 'End', color: '#f04438', icon: <CheckCircleOutlined /> },
      { type: 'condition', label: 'Condition', color: '#f79009', icon: <BranchesOutlined /> },
    ],
  },
  {
    title: 'MCP',
    items: [
      { type: 'mcp_tool', label: 'MCP Tool', color: '#3b5bdb', icon: <ApiOutlined /> },
      { type: 'mcp_prompt', label: 'MCP Prompt', color: '#7c3aed', icon: <FileTextOutlined /> },
      { type: 'mcp_resource', label: 'MCP Resource', color: '#0891b2', icon: <CloudOutlined /> },
    ],
  },
  {
    title: 'PROCESSING',
    items: [
      { type: 'llm', label: 'LLM', color: '#ea580c', icon: <RobotOutlined /> },
      { type: 'code', label: 'Code', color: '#4f46e5', icon: <CodeOutlined /> },
      { type: 'http', label: 'HTTP', color: '#db2777', icon: <GlobalOutlined /> },
    ],
  },
]

const allNodeTypes = nodeGroups.flatMap((g) => g.items)

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
  const [selectedNode, setSelectedNode] = useState<FlowNode | null>(null)

  useEffect(() => {
    if (!isNew && id) {
      workflowApi
        .get(Number(id))
        .then((res: any) => {
          setName(res.name)
          setDescription(res.description || '')
          const flowNodes = (res.nodes || []).map((n: any) => {
            const nt = allNodeTypes.find((t) => t.type === n.type)
            return {
              id: n.id,
              type: 'default',
              position: n.position || { x: 0, y: 0 },
              data: { label: n.name || nt?.label || n.type, nodeType: n.type, config: n.config },
              style: {
                background: nt?.color || '#667085',
                color: '#fff',
                borderRadius: 10,
                fontWeight: 500,
                fontSize: 13,
                padding: '8px 16px',
              },
            }
          })
          const flowEdges = (res.edges || []).map((e: any) => ({
            id: e.id,
            source: e.source,
            target: e.target,
            label: e.condition || '',
            animated: true,
            style: { stroke: '#98a2b3', strokeWidth: 2 },
          }))
          setNodes(flowNodes)
          setEdges(flowEdges)
          nodeId = flowNodes.length
        })
        .catch((err: any) => message.error(err.message))
    }
  }, [id])

  const isValidConnection = useCallback(
    (connection: Connection) => {
      const sourceNode = nodes.find((n) => n.id === connection.source)
      const targetNode = nodes.find((n) => n.id === connection.target)
      if (!sourceNode || !targetNode) return false

      const sourceType = (sourceNode.data as any).nodeType
      const targetType = (targetNode.data as any).nodeType

      // 不能自连
      if (connection.source === connection.target) return false
      // End 不能有出边
      if (sourceType === 'end') return false
      // Start 不能有入边
      if (targetType === 'start') return false

      // 出度校验
      const outCount = edges.filter((e) => e.source === connection.source).length
      if (sourceType === 'condition' && outCount >= 2) return false
      if (sourceType !== 'condition' && outCount >= 1) return false

      return true
    },
    [nodes, edges],
  )

  const onConnect = useCallback(
    (params: Connection) => {
      const sourceNode = nodes.find((n) => n.id === params.source)
      if (!sourceNode) return

      const sourceType = (sourceNode.data as any).nodeType
      const outEdges = edges.filter((e) => e.source === params.source)

      const edge: any = { ...params, animated: true, style: { stroke: '#98a2b3', strokeWidth: 2 } }

      // Condition 节点自动标记 true/false
      if (sourceType === 'condition') {
        edge.label = outEdges.length === 0 ? 'true' : 'false'
      }

      setEdges((eds) => addEdge(edge, eds))
    },
    [nodes, edges, setEdges],
  )

  const onAddNode = (type: string, label: string, color: string) => {
    const newNode: FlowNode = {
      id: getId(),
      type: 'default',
      position: { x: 300 + Math.random() * 100, y: nodes.length * 120 + 80 },
      data: { label, nodeType: type, config: {} },
      style: {
        background: color,
        color: '#fff',
        borderRadius: 10,
        fontWeight: 500,
        fontSize: 13,
        padding: '8px 16px',
      },
    }
    setNodes((nds) => [...nds, newNode])
  }

  const onNodeClick = useCallback((_: any, node: FlowNode) => {
    setSelectedNode(node)
  }, [])

  const onPaneClick = useCallback(() => {
    setSelectedNode(null)
  }, [])

  const handleSave = async () => {
    if (!name.trim()) {
      message.warning('Please enter a workflow name')
      return
    }
    const wfNodes = nodes.map((n) => ({
      id: n.id,
      type: (n.data as any).nodeType,
      name: String((n.data as any).label),
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
        message.success('Created')
        navigate('/workflows')
      } else {
        await workflowApi.update(Number(id), { name, description, nodes: wfNodes, edges: wfEdges })
        message.success('Saved')
      }
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const handleExecute = async () => {
    if (isNew) return
    try {
      const res: any = await workflowApi.execute(Number(id))
      message.success(`Executed, status: ${res.status}`)
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const selectedNodeType = selectedNode
    ? allNodeTypes.find((t) => t.type === (selectedNode.data as any).nodeType)
    : null

  const updateNodeConfig = (path: string, value: any) => {
    if (!selectedNode) return
    const keys = path.split('.')
    const config = JSON.parse(JSON.stringify((selectedNode.data as any).config || {}))
    let obj = config
    for (let i = 0; i < keys.length - 1; i++) {
      if (!obj[keys[i]]) obj[keys[i]] = {}
      obj = obj[keys[i]]
    }
    obj[keys[keys.length - 1]] = value
    setNodes((nds) =>
      nds.map((n) => (n.id === selectedNode.id ? { ...n, data: { ...n.data, config } } : n)),
    )
    setSelectedNode({ ...selectedNode, data: { ...selectedNode.data, config } })
  }

  return (
    <div className="editor-container">
      {/* Toolbar */}
      <div className="editor-toolbar">
        <div className="editor-toolbar-left">
          <Tooltip title="Back to list">
            <Button
              type="text"
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/workflows')}
              style={{ color: '#667085' }}
            />
          </Tooltip>
          <Divider type="vertical" style={{ height: 24, margin: 0 }} />
          <Input
            variant="borderless"
            placeholder="Workflow name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            style={{ width: 200, fontWeight: 600, fontSize: 15 }}
          />
          <Input
            variant="borderless"
            placeholder="Add description..."
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            style={{ width: 200, color: '#667085', fontSize: 13 }}
          />
        </div>
        <div className="editor-toolbar-right">
          {!isNew && (
            <Button icon={<PlayCircleOutlined />} onClick={handleExecute} style={{ borderRadius: 8 }}>
              Run
            </Button>
          )}
          <Button type="primary" icon={<SaveOutlined />} onClick={handleSave} style={{ borderRadius: 8 }}>
            Save
          </Button>
        </div>
      </div>

      {/* Main body */}
      <div className="editor-body">
        {/* Left: Node palette */}
        <div className="node-palette">
          {nodeGroups.map((group) => (
            <div key={group.title} className="node-palette-group">
              <div className="node-palette-group-title">{group.title}</div>
              {group.items.map((item) => (
                <div
                  key={item.type}
                  className="node-palette-item"
                  onClick={() => onAddNode(item.type, item.label, item.color)}
                >
                  <div className="node-palette-icon" style={{ background: item.color }}>
                    {item.icon}
                  </div>
                  <span className="node-palette-label">{item.label}</span>
                </div>
              ))}
            </div>
          ))}
        </div>

        {/* Center: Canvas */}
        <div className="canvas-area">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            isValidConnection={isValidConnection}
            onNodeClick={onNodeClick}
            onPaneClick={onPaneClick}
            fitView
            proOptions={{ hideAttribution: true }}
          >
            <Background gap={20} size={1} color="#e5e7eb" />
            <Controls position="bottom-left" style={{ marginBottom: 16, marginLeft: 16 }} />
            <MiniMap
              position="bottom-right"
              style={{ marginBottom: 16, marginRight: 16 }}
              maskColor="rgba(0,0,0,0.05)"
              nodeColor={(n) => {
                const nt = allNodeTypes.find((t) => t.type === (n.data as any)?.nodeType)
                return nt?.color || '#667085'
              }}
            />
          </ReactFlow>
        </div>

        {/* Right: Properties panel */}
        {selectedNode && (
          <div className="node-props-panel">
            <div className="node-props-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <div
                  style={{
                    width: 24,
                    height: 24,
                    borderRadius: 6,
                    background: selectedNodeType?.color || '#667085',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: '#fff',
                    fontSize: 12,
                  }}
                >
                  {selectedNodeType?.icon || <SettingOutlined />}
                </div>
                <h4>{(selectedNode.data as any).label}</h4>
              </div>
              <Button
                type="text"
                size="small"
                icon={<CloseOutlined />}
                onClick={() => setSelectedNode(null)}
                style={{ color: '#98a2b3' }}
              />
            </div>
            <div className="node-props-body">
              <Form layout="vertical" size="small">
                <Form.Item label="Name">
                  <Input
                    value={(selectedNode.data as any).label}
                    onChange={(e) => {
                      setNodes((nds) =>
                        nds.map((n) =>
                          n.id === selectedNode.id
                            ? { ...n, data: { ...n.data, label: e.target.value } }
                            : n,
                        ),
                      )
                      setSelectedNode({
                        ...selectedNode,
                        data: { ...selectedNode.data, label: e.target.value },
                      })
                    }}
                    style={{ borderRadius: 8 }}
                  />
                </Form.Item>
                <Form.Item label="Type">
                  <Tag color={selectedNodeType?.color} style={{ borderRadius: 4 }}>
                    {selectedNodeType?.label || (selectedNode.data as any).nodeType}
                  </Tag>
                </Form.Item>
                <Form.Item label="ID">
                  <Input
                    value={selectedNode.id}
                    disabled
                    style={{ borderRadius: 8, fontFamily: 'monospace', fontSize: 12 }}
                  />
                </Form.Item>

                <Divider style={{ margin: '12px 0' }} />

                {(selectedNode.data as any).nodeType === 'llm' && (
                  <>
                    <Form.Item label="Provider">
                      <Select
                        value={(selectedNode.data as any).config?.llm?.provider}
                        onChange={(v) => updateNodeConfig('llm.provider', v)}
                        options={[
                          { value: 'openai', label: 'OpenAI' },
                          { value: 'anthropic', label: 'Anthropic' },
                        ]}
                        placeholder="Select provider"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Model">
                      <Input
                        value={(selectedNode.data as any).config?.llm?.model}
                        onChange={(e) => updateNodeConfig('llm.model', e.target.value)}
                        placeholder="gpt-4o"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Prompt">
                      <Input.TextArea
                        value={(selectedNode.data as any).config?.llm?.prompt}
                        onChange={(e) => updateNodeConfig('llm.prompt', e.target.value)}
                        rows={3}
                        placeholder="Enter prompt..."
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                  </>
                )}

                {(selectedNode.data as any).nodeType === 'mcp_tool' && (
                  <>
                    <Form.Item label="Server URL">
                      <Input
                        value={(selectedNode.data as any).config?.mcp_tool?.server_url}
                        onChange={(e) => updateNodeConfig('mcp_tool.server_url', e.target.value)}
                        placeholder="http://localhost:3001"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Tool Name">
                      <Input
                        value={(selectedNode.data as any).config?.mcp_tool?.tool_name}
                        onChange={(e) => updateNodeConfig('mcp_tool.tool_name', e.target.value)}
                        placeholder="Tool name"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                  </>
                )}

                {(selectedNode.data as any).nodeType === 'http' && (
                  <>
                    <Form.Item label="Method">
                      <Select
                        value={(selectedNode.data as any).config?.http?.method}
                        onChange={(v) => updateNodeConfig('http.method', v)}
                        options={[
                          { value: 'GET', label: 'GET' },
                          { value: 'POST', label: 'POST' },
                          { value: 'PUT', label: 'PUT' },
                          { value: 'DELETE', label: 'DELETE' },
                        ]}
                        placeholder="Select method"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="URL">
                      <Input
                        value={(selectedNode.data as any).config?.http?.url}
                        onChange={(e) => updateNodeConfig('http.url', e.target.value)}
                        placeholder="https://api.example.com"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Body">
                      <Input.TextArea
                        value={(selectedNode.data as any).config?.http?.body}
                        onChange={(e) => updateNodeConfig('http.body', e.target.value)}
                        rows={3}
                        placeholder='{"key": "value"}'
                        style={{ borderRadius: 8, fontFamily: 'monospace', fontSize: 12 }}
                      />
                    </Form.Item>
                  </>
                )}

                {(selectedNode.data as any).nodeType === 'condition' && (
                  <Form.Item label="Expression">
                    <Input.TextArea
                      value={(selectedNode.data as any).config?.condition?.expression}
                      onChange={(e) => updateNodeConfig('condition.expression', e.target.value)}
                      rows={2}
                      placeholder="output.status == 'ok'"
                      style={{ borderRadius: 8, fontFamily: 'monospace', fontSize: 12 }}
                    />
                  </Form.Item>
                )}

                {(selectedNode.data as any).nodeType === 'code' && (
                  <>
                    <Form.Item label="Language">
                      <Select
                        value={(selectedNode.data as any).config?.code?.language}
                        onChange={(v) => updateNodeConfig('code.language', v)}
                        options={[
                          { value: 'javascript', label: 'JavaScript' },
                          { value: 'lua', label: 'Lua' },
                        ]}
                        placeholder="Select language"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Code">
                      <Input.TextArea
                        value={(selectedNode.data as any).config?.code?.code}
                        onChange={(e) => updateNodeConfig('code.code', e.target.value)}
                        rows={6}
                        placeholder="// your code here"
                        style={{ borderRadius: 8, fontFamily: 'monospace', fontSize: 12 }}
                      />
                    </Form.Item>
                  </>
                )}

                {(selectedNode.data as any).nodeType === 'mcp_prompt' && (
                  <>
                    <Form.Item label="Server URL">
                      <Input
                        value={(selectedNode.data as any).config?.mcp_prompt?.server_url}
                        onChange={(e) => updateNodeConfig('mcp_prompt.server_url', e.target.value)}
                        placeholder="http://localhost:3001"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Prompt Name">
                      <Input
                        value={(selectedNode.data as any).config?.mcp_prompt?.prompt_name}
                        onChange={(e) => updateNodeConfig('mcp_prompt.prompt_name', e.target.value)}
                        placeholder="Prompt name"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                  </>
                )}

                {(selectedNode.data as any).nodeType === 'mcp_resource' && (
                  <>
                    <Form.Item label="Server URL">
                      <Input
                        value={(selectedNode.data as any).config?.mcp_resource?.server_url}
                        onChange={(e) => updateNodeConfig('mcp_resource.server_url', e.target.value)}
                        placeholder="http://localhost:3001"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Resource URI">
                      <Input
                        value={(selectedNode.data as any).config?.mcp_resource?.uri}
                        onChange={(e) => updateNodeConfig('mcp_resource.uri', e.target.value)}
                        placeholder="resource://example"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                  </>
                )}
              </Form>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
