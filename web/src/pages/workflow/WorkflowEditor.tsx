import { useCallback, useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Input, message, Tooltip, Divider, Form, Select, Tag, Spin } from 'antd'
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
  GlobalOutlined,
  MailOutlined,
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
import { mcpServerApi, type MCPServer } from '../../api/mcpserver'
import { llmProviderApi, type LLMProvider } from '../../api/llm_provider'

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
    title: 'PROCESSING',
    items: [
      { type: 'llm', label: 'LLM', color: '#ea580c', icon: <RobotOutlined /> },
      { type: 'mcp', label: 'MCP', color: '#3b5bdb', icon: <ApiOutlined /> },
      { type: 'code', label: 'Code', color: '#4f46e5', icon: <CodeOutlined /> },
      { type: 'http', label: 'HTTP', color: '#db2777', icon: <GlobalOutlined /> },
      { type: 'email', label: 'Email', color: '#0891b2', icon: <MailOutlined /> },
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

  // MCP servers & capabilities
  const [mcpServers, setMcpServers] = useState<MCPServer[]>([])
  const [mcpTools, setMcpTools] = useState<any[]>([])
  const [mcpPrompts, setMcpPrompts] = useState<any[]>([])
  const [mcpResources, setMcpResources] = useState<any[]>([])
  const [mcpLoading, setMcpLoading] = useState(false)

  // LLM Providers
  const [llmProviders, setLlmProviders] = useState<LLMProvider[]>([])

  useEffect(() => {
    mcpServerApi.list().then((res: any) => setMcpServers(Array.isArray(res) ? res : res.data || [])).catch(() => {})
    llmProviderApi.list().then((res: any) => setLlmProviders(Array.isArray(res) ? res : res.data || [])).catch(() => {})
  }, [])

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
          nodeId = flowNodes.length
          requestAnimationFrame(() => {
            setEdges(flowEdges)
          })
        })
        .catch((err: any) => message.error(err.message))
    }
  }, [id])

  // 加载 MCP Server 能力
  const loadServerCapabilities = async (serverId: number) => {
    setMcpLoading(true)
    try {
      const [toolsRes, promptsRes, resourcesRes]: any[] = await Promise.allSettled([
        mcpServerApi.tools(serverId),
        mcpServerApi.prompts(serverId),
        mcpServerApi.resources(serverId),
      ])
      if (toolsRes.status === 'fulfilled') setMcpTools(Array.isArray(toolsRes.value) ? toolsRes.value : toolsRes.value?.tools || [])
      if (promptsRes.status === 'fulfilled') setMcpPrompts(Array.isArray(promptsRes.value) ? promptsRes.value : promptsRes.value?.prompts || [])
      if (resourcesRes.status === 'fulfilled') setMcpResources(Array.isArray(resourcesRes.value) ? resourcesRes.value : resourcesRes.value?.resources || [])
    } catch {
      // ignore
    } finally {
      setMcpLoading(false)
    }
  }

  // 当选中 MCP 节点且已选了 server，加载能力
  useEffect(() => {
    if (!selectedNode) return
    const nodeType = (selectedNode.data as any).nodeType
    if (nodeType !== 'mcp') return
    const config = (selectedNode.data as any).config?.mcp || {}
    if (config.server_url) {
      const server = mcpServers.find((s) => s.url === config.server_url)
      if (server) loadServerCapabilities(server.id)
    } else {
      setMcpTools([])
      setMcpPrompts([])
      setMcpResources([])
    }
  }, [selectedNode?.id])

  const isValidConnection = useCallback(
    (connection: Connection | FlowEdge) => {
      const sourceNode = nodes.find((n) => n.id === connection.source)
      const targetNode = nodes.find((n) => n.id === connection.target)
      if (!sourceNode || !targetNode) return false

      const sourceType = (sourceNode.data as any).nodeType
      const targetType = (targetNode.data as any).nodeType

      if (connection.source === connection.target) return false
      if (sourceType === 'end') return false
      if (targetType === 'start') return false

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

      if (sourceType === 'condition') {
        edge.label = outEdges.length === 0 ? 'true' : 'false'
      }

      setEdges((eds) => addEdge(edge, eds))
    },
    [nodes, edges, setEdges],
  )

  const onAddNode = (type: string, label: string, color: string) => {
    if (type === 'start' && nodes.some((n) => (n.data as any).nodeType === 'start')) {
      message.warning('Only one Start node is allowed')
      return
    }
    if (type === 'end' && nodes.some((n) => (n.data as any).nodeType === 'end')) {
      message.warning('Only one End node is allowed')
      return
    }

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
      message.success(`Execution started: #${res.execution_id}`)
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

  const serverOptions = mcpServers
    .filter((s) => s.status === 'active')
    .map((s) => ({ value: s.url, label: `${s.name} (${s.url})` }))

  const handleMCPServerChange = (serverUrl: string) => {
    if (!selectedNode) return
    const config = JSON.parse(JSON.stringify((selectedNode.data as any).config || {}))
    if (!config.mcp) config.mcp = {}
    config.mcp.server_url = serverUrl

    // 自动填充 headers
    const server = mcpServers.find((s) => s.url === serverUrl)
    if (server?.headers && Object.keys(server.headers).length > 0) {
      config.mcp.headers = server.headers
    } else {
      delete config.mcp.headers
    }

    // 清除之前的选择
    config.mcp.tool_name = ''
    config.mcp.prompt_name = ''
    config.mcp.resource_uri = ''
    setMcpTools([])
    setMcpPrompts([])
    setMcpResources([])

    setNodes((nds) =>
      nds.map((n) => (n.id === selectedNode.id ? { ...n, data: { ...n.data, config } } : n)),
    )
    setSelectedNode({ ...selectedNode, data: { ...selectedNode.data, config } })

    if (server) loadServerCapabilities(server.id)
  }

  const handleLLMProviderChange = (providerId: number) => {
    const provider = llmProviders.find((p) => p.id === providerId)
    if (!provider || !selectedNode) return
    const config = JSON.parse(JSON.stringify((selectedNode.data as any).config || {}))
    if (!config.llm) config.llm = {}
    config.llm.base_url = provider.base_url
    config.llm.api_key = provider.api_key

    setNodes((nds) =>
      nds.map((n) => (n.id === selectedNode.id ? { ...n, data: { ...n.data, config } } : n)),
    )
    setSelectedNode({ ...selectedNode, data: { ...selectedNode.data, config } })
  }

  // MCP 统一面板
  const renderMCPPanel = () => {
    const config = (selectedNode?.data as any)?.config?.mcp || {}
    return (
      <>
        <Form.Item label="Action">
          <Select
            value={config.action || undefined}
            onChange={(v) => updateNodeConfig('mcp.action', v)}
            placeholder="Select action"
            style={{ borderRadius: 8 }}
            options={[
              { value: 'call_tool', label: 'Call Tool' },
              { value: 'get_prompt', label: 'Get Prompt' },
              { value: 'read_resource', label: 'Read Resource' },
            ]}
          />
        </Form.Item>
        <Form.Item label="MCP Server">
          <Select
            value={config.server_url || undefined}
            onChange={handleMCPServerChange}
            options={serverOptions}
            placeholder="Select MCP Server"
            style={{ borderRadius: 8 }}
            showSearch
            optionFilterProp="label"
          />
        </Form.Item>

        {config.action === 'call_tool' && (
          <Form.Item label="Tool">
            {mcpLoading ? (
              <Spin size="small" />
            ) : (
              <Select
                value={config.tool_name || undefined}
                onChange={(v) => updateNodeConfig('mcp.tool_name', v)}
                placeholder={config.server_url ? 'Select tool' : 'Select a server first'}
                disabled={!config.server_url}
                style={{ borderRadius: 8 }}
                showSearch
                optionFilterProp="label"
                options={mcpTools.map((t: any) => ({
                  value: t.name,
                  label: t.name,
                  title: t.description,
                }))}
              />
            )}
          </Form.Item>
        )}

        {config.action === 'get_prompt' && (
          <Form.Item label="Prompt">
            {mcpLoading ? (
              <Spin size="small" />
            ) : (
              <Select
                value={config.prompt_name || undefined}
                onChange={(v) => updateNodeConfig('mcp.prompt_name', v)}
                placeholder={config.server_url ? 'Select prompt' : 'Select a server first'}
                disabled={!config.server_url}
                style={{ borderRadius: 8 }}
                showSearch
                optionFilterProp="label"
                options={mcpPrompts.map((p: any) => ({
                  value: p.name,
                  label: p.name,
                  title: p.description,
                }))}
              />
            )}
          </Form.Item>
        )}

        {config.action === 'read_resource' && (
          <Form.Item label="Resource URI">
            {mcpLoading ? (
              <Spin size="small" />
            ) : (
              <Select
                value={config.resource_uri || undefined}
                onChange={(v) => updateNodeConfig('mcp.resource_uri', v)}
                placeholder={config.server_url ? 'Select resource' : 'Select a server first'}
                disabled={!config.server_url}
                style={{ borderRadius: 8 }}
                showSearch
                optionFilterProp="label"
                options={mcpResources.map((r: any) => ({
                  value: r.uri,
                  label: r.name || r.uri,
                  title: r.description,
                }))}
              />
            )}
          </Form.Item>
        )}

        {/* 显示选中项的描述 */}
        {config.tool_name && mcpTools.length > 0 && (() => {
          const tool = mcpTools.find((t: any) => t.name === config.tool_name)
          if (!tool?.description) return null
          return (
            <div style={{ fontSize: 12, color: '#667085', marginBottom: 12, padding: '8px 12px', background: '#f9fafb', borderRadius: 8 }}>
              {tool.description}
              {tool.inputSchema?.properties && (
                <div style={{ marginTop: 4 }}>
                  <span style={{ color: '#98a2b3' }}>Params: </span>
                  {Object.keys(tool.inputSchema.properties).map((key) => (
                    <Tag key={key} style={{ fontSize: 11, borderRadius: 4, marginBottom: 2 }}>
                      {key}
                      {tool.inputSchema.required?.includes(key) && <span style={{ color: '#f04438' }}>*</span>}
                    </Tag>
                  ))}
                </div>
              )}
            </div>
          )
        })()}
      </>
    )
  }

  // LLM 面板
  const renderLLMPanel = () => {
    const config = (selectedNode?.data as any)?.config?.llm || {}
    return (
      <>
        <Form.Item label="Provider (Quick Fill)">
          <Select
            value={undefined}
            onChange={handleLLMProviderChange}
            placeholder="Select to auto-fill base_url & api_key"
            style={{ borderRadius: 8 }}
            allowClear
            options={llmProviders.map((p) => ({
              value: p.id,
              label: p.name,
            }))}
          />
        </Form.Item>
        <Form.Item label="Base URL">
          <Input
            value={config.base_url}
            onChange={(e) => updateNodeConfig('llm.base_url', e.target.value)}
            placeholder="https://api.deepseek.com/v1"
            style={{ borderRadius: 8, fontFamily: 'monospace', fontSize: 12 }}
          />
        </Form.Item>
        <Form.Item label="API Key">
          <Input.Password
            value={config.api_key}
            onChange={(e) => updateNodeConfig('llm.api_key', e.target.value)}
            placeholder="sk-..."
            style={{ borderRadius: 8 }}
          />
        </Form.Item>
        <Form.Item label="Model">
          <Input
            value={config.model}
            onChange={(e) => updateNodeConfig('llm.model', e.target.value)}
            placeholder="deepseek-chat"
            style={{ borderRadius: 8 }}
          />
        </Form.Item>
        <Form.Item label="System Message">
          <Input.TextArea
            value={config.system_msg}
            onChange={(e) => updateNodeConfig('llm.system_msg', e.target.value)}
            rows={2}
            placeholder="Optional system message"
            style={{ borderRadius: 8 }}
          />
        </Form.Item>
        <Form.Item label="Prompt">
          <Input.TextArea
            value={config.prompt}
            onChange={(e) => updateNodeConfig('llm.prompt', e.target.value)}
            rows={3}
            placeholder={'Supports {{.input.query}} {{.node_1.content}}'}
            style={{ borderRadius: 8 }}
          />
        </Form.Item>
      </>
    )
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

                {(selectedNode.data as any).nodeType === 'llm' && renderLLMPanel()}
                {(selectedNode.data as any).nodeType === 'mcp' && renderMCPPanel()}

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

                {(selectedNode.data as any).nodeType === 'email' && (
                  <>
                    <Form.Item label="SMTP Host">
                      <Input
                        value={(selectedNode.data as any).config?.email?.smtp_host}
                        onChange={(e) => updateNodeConfig('email.smtp_host', e.target.value)}
                        placeholder="smtp.example.com"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="SMTP Port">
                      <Input
                        type="number"
                        value={(selectedNode.data as any).config?.email?.smtp_port}
                        onChange={(e) => updateNodeConfig('email.smtp_port', Number(e.target.value))}
                        placeholder="587"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Username">
                      <Input
                        value={(selectedNode.data as any).config?.email?.username}
                        onChange={(e) => updateNodeConfig('email.username', e.target.value)}
                        placeholder="user@example.com"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Password">
                      <Input.Password
                        value={(selectedNode.data as any).config?.email?.password}
                        onChange={(e) => updateNodeConfig('email.password', e.target.value)}
                        placeholder="SMTP password or app password"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="From">
                      <Input
                        value={(selectedNode.data as any).config?.email?.from}
                        onChange={(e) => updateNodeConfig('email.from', e.target.value)}
                        placeholder="sender@example.com"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="To">
                      <Input
                        value={(selectedNode.data as any).config?.email?.to}
                        onChange={(e) => updateNodeConfig('email.to', e.target.value)}
                        placeholder="a@example.com, b@example.com"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Cc">
                      <Input
                        value={(selectedNode.data as any).config?.email?.cc}
                        onChange={(e) => updateNodeConfig('email.cc', e.target.value)}
                        placeholder="Optional, comma separated"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Subject">
                      <Input
                        value={(selectedNode.data as any).config?.email?.subject}
                        onChange={(e) => updateNodeConfig('email.subject', e.target.value)}
                        placeholder="支持模板变量 {{.content}}"
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Content Type">
                      <Select
                        value={(selectedNode.data as any).config?.email?.content_type || 'text/html'}
                        onChange={(v) => updateNodeConfig('email.content_type', v)}
                        options={[
                          { value: 'text/html', label: 'HTML' },
                          { value: 'text/plain', label: 'Plain Text' },
                        ]}
                        style={{ borderRadius: 8 }}
                      />
                    </Form.Item>
                    <Form.Item label="Body">
                      <Input.TextArea
                        value={(selectedNode.data as any).config?.email?.body}
                        onChange={(e) => updateNodeConfig('email.body', e.target.value)}
                        rows={8}
                        placeholder={'<h1>执行结果</h1>\n<p>{{.content}}</p>'}
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
                      placeholder="status == 'ok'"
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
              </Form>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
