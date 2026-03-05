import { useEffect, useState } from 'react'
import {
  Button,
  Row,
  Col,
  Space,
  message,
  Popconfirm,
  Modal,
  Form,
  Input,
  Tag,
  Tooltip,
  Drawer,
  Tabs,
  Spin,
  Empty,
  Descriptions,
} from 'antd'
import {
  PlusOutlined,
  MinusCircleOutlined,
  CloudServerOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  LinkOutlined,
  ToolOutlined,
  MessageOutlined,
  DatabaseOutlined,
  ReloadOutlined,
  ShopOutlined,
  CheckOutlined,
} from '@ant-design/icons'
import { mcpServerApi, type MCPServer } from '../../api/mcpserver'
import { marketplaceCatalog, categoryLabels, type MarketplaceMCPServer } from '../../data/mcpMarketplace'

const formatDateTime = (v: string) => {
  if (!v) return '-'
  const d = new Date(v)
  if (isNaN(d.getTime())) return v
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

export default function MCPServerList() {
  const [servers, setServers] = useState<MCPServer[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editId, setEditId] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [form] = Form.useForm()

  // marketplace
  const [marketplaceOpen, setMarketplaceOpen] = useState(false)
  const [marketCategory, setMarketCategory] = useState('all')
  const [marketSearch, setMarketSearch] = useState('')
  const [addingUrl, setAddingUrl] = useState<string | null>(null)

  // tool test
  const [testTool, setTestTool] = useState<any>(null)
  const [testArgs, setTestArgs] = useState<string>('{}')
  const [testResult, setTestResult] = useState<any>(null)
  const [testLoading, setTestLoading] = useState(false)
  const [testDuration, setTestDuration] = useState<number | null>(null)

  // detail drawer
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [detailServer, setDetailServer] = useState<MCPServer | null>(null)
  const [tools, setTools] = useState<any[]>([])
  const [prompts, setPrompts] = useState<any[]>([])
  const [resources, setResources] = useState<any[]>([])
  const [detailLoading, setDetailLoading] = useState(false)

  const fetchList = async () => {
    setLoading(true)
    try {
      const res: any = await mcpServerApi.list()
      setServers(Array.isArray(res) ? res : res.data || [])
    } catch (err: any) {
      message.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleSubmit = async () => {
    const values = await form.validateFields()
    // 把 headersList [{key, value}] 转成 headers map
    const { headersList, ...rest } = values
    const headers: Record<string, string> = {}
    if (headersList) {
      for (const item of headersList) {
        if (item?.key && item?.value) {
          headers[item.key] = item.value
        }
      }
    }
    const payload = { ...rest, headers: Object.keys(headers).length > 0 ? headers : undefined }
    try {
      if (editId) {
        await mcpServerApi.update(editId, payload)
        message.success('Updated')
      } else {
        await mcpServerApi.create(payload)
        message.success('Created')
      }
      setModalOpen(false)
      form.resetFields()
      setEditId(null)
      fetchList()
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const handleEdit = (record: MCPServer) => {
    setEditId(record.id)
    // 把 headers map 转成 headersList [{key, value}]
    const headersList = record.headers
      ? Object.entries(record.headers).map(([key, value]) => ({ key, value }))
      : []
    form.setFieldsValue({ ...record, headersList })
    setModalOpen(true)
  }

  const handleDelete = async (id: string) => {
    await mcpServerApi.delete(id)
    message.success('Deleted')
    fetchList()
  }

  const handleTest = async (id: string) => {
    try {
      const res: any = await mcpServerApi.check(id)
      message.success(`Connected, ${res.tools?.length || 0} tools found`)
      fetchList()
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const openDetail = (server: MCPServer) => {
    setDetailServer(server)
    setDrawerOpen(true)
    setTools(server.tools || [])
    setPrompts(server.prompts || [])
    setResources(server.resources || [])
    setDetailLoading(false)
  }

  const refreshDetail = async () => {
    if (!detailServer) return
    setDetailLoading(true)
    try {
      const res: any = await mcpServerApi.check(detailServer.id)
      setDetailServer(res)
      setTools(res.tools || [])
      setPrompts(res.prompts || [])
      setResources(res.resources || [])
    } catch (err: any) {
      message.error(err.message)
    } finally {
      setDetailLoading(false)
    }
  }

  const handleAddFromMarketplace = async (item: MarketplaceMCPServer) => {
    setAddingUrl(item.url)
    try {
      await mcpServerApi.create({
        name: item.name,
        description: item.description,
        url: item.url,
      })
      message.success(`${item.name} added`)
      fetchList()
    } catch (err: any) {
      message.error(err.message)
    } finally {
      setAddingUrl(null)
    }
  }

  const handleTestTool = async () => {
    if (!detailServer || !testTool) return
    setTestLoading(true)
    setTestResult(null)
    setTestDuration(null)
    try {
      const args = JSON.parse(testArgs)
      const res: any = await mcpServerApi.callTool(detailServer.id, testTool.name, args)
      setTestResult(res.result)
      setTestDuration(res.duration)
    } catch (err: any) {
      setTestResult({ error: err.message })
    } finally {
      setTestLoading(false)
    }
  }

  const openToolTest = (tool: any) => {
    setTestTool(tool)
    setTestResult(null)
    setTestDuration(null)
    // 生成默认参数模板
    const props = tool.inputSchema?.properties || {}
    const defaultArgs: Record<string, any> = {}
    for (const key of Object.keys(props)) {
      const prop = props[key]
      if (prop.type === 'number' || prop.type === 'integer') defaultArgs[key] = 0
      else if (prop.type === 'boolean') defaultArgs[key] = false
      else defaultArgs[key] = ''
    }
    setTestArgs(JSON.stringify(defaultArgs, null, 2))
  }

  const existingUrls = new Set(servers.map((s) => s.url))

  const filteredMarketplace = marketplaceCatalog.filter((item) => {
    if (marketCategory !== 'all' && item.category !== marketCategory) return false
    if (marketSearch) {
      const q = marketSearch.toLowerCase()
      return (
        item.name.toLowerCase().includes(q) ||
        item.description.toLowerCase().includes(q) ||
        item.tags.some((t) => t.includes(q))
      )
    }
    return true
  })

  const filtered = servers.filter(
    (s) =>
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.url.toLowerCase().includes(search.toLowerCase()) ||
      (s.description || '').toLowerCase().includes(search.toLowerCase()),
  )

  const renderToolItem = (tool: any) => (
    <div
      key={tool.name}
      style={{
        padding: '12px 16px',
        background: '#f9fafb',
        borderRadius: 8,
        marginBottom: 8,
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
        <div style={{ fontWeight: 600, fontSize: 13 }}>{tool.name}</div>
        <Button
          type="link"
          size="small"
          icon={<ToolOutlined />}
          onClick={() => openToolTest(tool)}
          style={{ fontSize: 12, padding: 0, height: 'auto' }}
        >
          Test
        </Button>
      </div>
      {tool.description && (
        <div style={{ fontSize: 12, color: '#667085', marginBottom: 8 }}>{tool.description}</div>
      )}
      {tool.inputSchema?.properties && (
        <div style={{ fontSize: 12 }}>
          <span style={{ color: '#98a2b3' }}>Parameters: </span>
          {Object.keys(tool.inputSchema.properties).map((key) => (
            <Tag key={key} style={{ fontSize: 11, borderRadius: 4, marginBottom: 2 }}>
              {key}
              {tool.inputSchema.required?.includes(key) && (
                <span style={{ color: '#f04438' }}>*</span>
              )}
            </Tag>
          ))}
        </div>
      )}
    </div>
  )

  const renderPromptItem = (prompt: any) => (
    <div
      key={prompt.name}
      style={{
        padding: '12px 16px',
        background: '#f9fafb',
        borderRadius: 8,
        marginBottom: 8,
      }}
    >
      <div style={{ fontWeight: 600, fontSize: 13, marginBottom: 4 }}>{prompt.name}</div>
      {prompt.description && (
        <div style={{ fontSize: 12, color: '#667085', marginBottom: 8 }}>{prompt.description}</div>
      )}
      {prompt.arguments && prompt.arguments.length > 0 && (
        <div style={{ fontSize: 12 }}>
          <span style={{ color: '#98a2b3' }}>Arguments: </span>
          {prompt.arguments.map((arg: any) => (
            <Tag key={arg.name} style={{ fontSize: 11, borderRadius: 4, marginBottom: 2 }}>
              {arg.name}
              {arg.required && <span style={{ color: '#f04438' }}>*</span>}
            </Tag>
          ))}
        </div>
      )}
    </div>
  )

  const renderResourceItem = (resource: any) => (
    <div
      key={resource.uri}
      style={{
        padding: '12px 16px',
        background: '#f9fafb',
        borderRadius: 8,
        marginBottom: 8,
      }}
    >
      <div style={{ fontWeight: 600, fontSize: 13, marginBottom: 4 }}>
        {resource.name || resource.uri}
      </div>
      {resource.description && (
        <div style={{ fontSize: 12, color: '#667085', marginBottom: 4 }}>{resource.description}</div>
      )}
      <div style={{ fontSize: 12, color: '#98a2b3', fontFamily: 'monospace' }}>{resource.uri}</div>
      {resource.mimeType && (
        <Tag style={{ fontSize: 11, borderRadius: 4, marginTop: 4 }}>{resource.mimeType}</Tag>
      )}
    </div>
  )

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>MCP Servers</h2>
          <div className="page-header-sub">
            Manage your Model Context Protocol service connections
          </div>
        </div>
        <Space>
          <Button
            icon={<ShopOutlined />}
            onClick={() => setMarketplaceOpen(true)}
            style={{ borderRadius: 10 }}
          >
            Marketplace
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setEditId(null)
              form.resetFields()
              setModalOpen(true)
            }}
            style={{ borderRadius: 10 }}
          >
            Add Server
          </Button>
        </Space>
      </div>

      {servers.length > 0 && (
        <div style={{ marginBottom: 20 }}>
          <Input
            prefix={<SearchOutlined style={{ color: '#98a2b3' }} />}
            placeholder="Search servers..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            allowClear
            style={{
              maxWidth: 360,
              borderRadius: 10,
              height: 40,
            }}
          />
        </div>
      )}

      {loading ? (
        <Row gutter={[16, 16]}>
          {[1, 2, 3].map((i) => (
            <Col key={i} xs={24} sm={12} lg={8}>
              <div className="server-card" style={{ opacity: 0.5, height: 160 }}>
                <div style={{ background: '#f2f4f7', height: 20, width: '50%', borderRadius: 4, marginBottom: 8 }} />
                <div style={{ background: '#f2f4f7', height: 14, width: '70%', borderRadius: 4 }} />
              </div>
            </Col>
          ))}
        </Row>
      ) : filtered.length === 0 ? (
        <div className="empty-state">
          <CloudServerOutlined className="empty-state-icon" />
          <div className="empty-state-title">
            {search ? 'No matching servers' : 'No MCP servers yet'}
          </div>
          <div className="empty-state-desc">
            {search
              ? 'Try different keywords'
              : 'Add your first MCP server to use tools and resources in your workflows'}
          </div>
        </div>
      ) : (
        <Row gutter={[16, 16]}>
          {filtered.map((server) => (
            <Col key={server.id} xs={24} sm={12} lg={8}>
              <div className="server-card">
                <div className="server-card-header">
                  <div>
                    <div
                      className="server-card-name"
                      style={{ cursor: 'pointer' }}
                      onClick={() => openDetail(server)}
                    >
                      <span
                        className={`status-dot ${server.status === 'active' ? 'status-dot-active' : 'status-dot-inactive'}`}
                      />
                      {server.name}
                    </div>
                    <div className="server-card-url">{server.url}</div>
                  </div>
                  <Tag
                    color={server.status === 'active' ? 'green' : 'default'}
                    style={{ borderRadius: 4, fontSize: 11, margin: 0 }}
                  >
                    {server.status === 'active' ? 'Online' : 'Offline'}
                  </Tag>
                </div>

                {server.description && (
                  <div className="server-card-desc">{server.description}</div>
                )}

                <div className="server-card-footer">
                  <Space size={4}>
                    <Tooltip title="Test connection">
                      <Button
                        type="text"
                        size="small"
                        icon={<LinkOutlined />}
                        onClick={() => handleTest(server.id)}
                        style={{ color: '#3b5bdb' }}
                      >
                        Test
                      </Button>
                    </Tooltip>
                    <Tooltip title="Detail">
                      <Button
                        type="text"
                        size="small"
                        icon={<ToolOutlined />}
                        onClick={() => openDetail(server)}
                        style={{ color: '#667085' }}
                      >
                        Detail
                      </Button>
                    </Tooltip>
                    <Tooltip title="Edit">
                      <Button
                        type="text"
                        size="small"
                        icon={<EditOutlined />}
                        onClick={() => handleEdit(server)}
                        style={{ color: '#667085' }}
                      />
                    </Tooltip>
                  </Space>
                  <Popconfirm
                    title="Delete this server?"
                    description="This action cannot be undone"
                    onConfirm={() => handleDelete(server.id)}
                    okText="Delete"
                    cancelText="Cancel"
                  >
                    <Button
                      type="text"
                      size="small"
                      icon={<DeleteOutlined />}
                      danger
                    />
                  </Popconfirm>
                </div>
              </div>
            </Col>
          ))}
        </Row>
      )}

      {/* Detail Drawer */}
      <Drawer
        title={
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <CloudServerOutlined style={{ color: '#3b5bdb' }} />
              <span>{detailServer?.name}</span>
              <Tag
                color={detailServer?.status === 'active' ? 'green' : 'default'}
                style={{ borderRadius: 4, fontSize: 11, margin: 0 }}
              >
                {detailServer?.status === 'active' ? 'Online' : 'Offline'}
              </Tag>
            </div>
            <Tooltip title="Refresh">
              <Button
                type="text"
                size="small"
                icon={<ReloadOutlined spin={detailLoading} />}
                onClick={refreshDetail}
              />
            </Tooltip>
          </div>
        }
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        width={560}
      >
        {detailServer && (
          <>
            <Descriptions column={1} size="small" style={{ marginBottom: 16 }}>
              <Descriptions.Item label="URL">
                <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{detailServer.url}</span>
              </Descriptions.Item>
              {detailServer.description && (
                <Descriptions.Item label="Description">{detailServer.description}</Descriptions.Item>
              )}
              {detailServer.headers && Object.keys(detailServer.headers).length > 0 && (
                <Descriptions.Item label="Headers">
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
                    {Object.keys(detailServer.headers).map((k) => (
                      <Tag key={k} style={{ fontSize: 11, borderRadius: 4 }}>{k}</Tag>
                    ))}
                  </div>
                </Descriptions.Item>
              )}
              {detailServer.checked_at && (
                <Descriptions.Item label="Last Check">{formatDateTime(detailServer.checked_at)}</Descriptions.Item>
              )}
              <Descriptions.Item label="Created">{formatDateTime(detailServer.created_at)}</Descriptions.Item>
            </Descriptions>

            {detailLoading ? (
              <div style={{ textAlign: 'center', padding: 40 }}>
                <Spin tip="Loading capabilities..." />
              </div>
            ) : (
              <Tabs
                defaultActiveKey="tools"
                items={[
                  {
                    key: 'tools',
                    label: (
                      <span>
                        <ToolOutlined /> Tools ({tools.length})
                      </span>
                    ),
                    children:
                      tools.length > 0 ? (
                        tools.map(renderToolItem)
                      ) : (
                        <Empty description="No tools available" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                      ),
                  },
                  {
                    key: 'prompts',
                    label: (
                      <span>
                        <MessageOutlined /> Prompts ({prompts.length})
                      </span>
                    ),
                    children:
                      prompts.length > 0 ? (
                        prompts.map(renderPromptItem)
                      ) : (
                        <Empty description="No prompts available" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                      ),
                  },
                  {
                    key: 'resources',
                    label: (
                      <span>
                        <DatabaseOutlined /> Resources ({resources.length})
                      </span>
                    ),
                    children:
                      resources.length > 0 ? (
                        resources.map(renderResourceItem)
                      ) : (
                        <Empty description="No resources available" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                      ),
                  },
                ]}
              />
            )}
          </>
        )}
      </Drawer>

      {/* Create/Edit Modal */}
      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <CloudServerOutlined style={{ color: '#3b5bdb' }} />
            {editId ? 'Edit Server' : 'Add Server'}
          </div>
        }
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => {
          setModalOpen(false)
          setEditId(null)
        }}
        okText="Save"
        cancelText="Cancel"
        styles={{
          body: { paddingTop: 16 },
        }}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Please enter a server name' }]}
          >
            <Input placeholder="My MCP Server" style={{ borderRadius: 8 }} />
          </Form.Item>
          <Form.Item
            name="url"
            label="URL"
            rules={[{ required: true, message: 'Please enter the server URL' }]}
          >
            <Input placeholder="http://localhost:3001/mcp" style={{ borderRadius: 8 }} />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input.TextArea rows={2} placeholder="Optional description" style={{ borderRadius: 8 }} />
          </Form.Item>
          <Form.Item label="Headers" style={{ marginBottom: 0 }}>
            <Form.List name="headersList">
              {(fields, { add, remove }) => (
                <>
                  {fields.map(({ key, name, ...restField }) => (
                    <div key={key} style={{ display: 'flex', gap: 8, marginBottom: 8 }}>
                      <Form.Item
                        {...restField}
                        name={[name, 'key']}
                        style={{ flex: 1, marginBottom: 0 }}
                        rules={[{ required: true, message: 'Key' }]}
                      >
                        <Input placeholder="Header name" style={{ borderRadius: 8 }} />
                      </Form.Item>
                      <Form.Item
                        {...restField}
                        name={[name, 'value']}
                        style={{ flex: 1, marginBottom: 0 }}
                        rules={[{ required: true, message: 'Value' }]}
                      >
                        <Input placeholder="Header value" style={{ borderRadius: 8 }} />
                      </Form.Item>
                      <Button
                        type="text"
                        danger
                        icon={<MinusCircleOutlined />}
                        onClick={() => remove(name)}
                      />
                    </div>
                  ))}
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                    style={{ borderRadius: 8 }}
                  >
                    Add Header
                  </Button>
                </>
              )}
            </Form.List>
          </Form.Item>
        </Form>
      </Modal>

      {/* Tool Test Modal */}
      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <ToolOutlined style={{ color: '#3b5bdb' }} />
            Test Tool: {testTool?.name}
          </div>
        }
        open={!!testTool}
        onCancel={() => setTestTool(null)}
        footer={null}
        width={600}
        styles={{ body: { paddingTop: 12 } }}
      >
        {testTool && (
          <div>
            {testTool.description && (
              <div style={{ fontSize: 12, color: '#667085', marginBottom: 12 }}>
                {testTool.description}
              </div>
            )}
            <div style={{ marginBottom: 8 }}>
              <div style={{ fontSize: 12, color: '#344054', fontWeight: 500, marginBottom: 4 }}>
                Arguments (JSON)
              </div>
              <Input.TextArea
                rows={6}
                value={testArgs}
                onChange={(e) => setTestArgs(e.target.value)}
                style={{ fontFamily: 'monospace', fontSize: 12, borderRadius: 8 }}
              />
            </div>
            <Button
              type="primary"
              onClick={handleTestTool}
              loading={testLoading}
              style={{ borderRadius: 8, marginBottom: 16 }}
            >
              Run
            </Button>
            {testResult && (
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                  <span style={{ fontSize: 12, color: '#344054', fontWeight: 500 }}>Result</span>
                  {testDuration !== null && (
                    <Tag style={{ fontSize: 10, borderRadius: 4 }}>{testDuration}ms</Tag>
                  )}
                </div>
                <pre
                  style={{
                    background: '#f9fafb',
                    border: '1px solid #eaecf0',
                    borderRadius: 8,
                    padding: 12,
                    maxHeight: 300,
                    overflow: 'auto',
                    fontSize: 11,
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-all',
                  }}
                >
                  {JSON.stringify(testResult, null, 2)}
                </pre>
              </div>
            )}
          </div>
        )}
      </Modal>

      {/* Marketplace Modal */}
      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <ShopOutlined style={{ color: '#3b5bdb' }} />
            MCP Server Marketplace
          </div>
        }
        open={marketplaceOpen}
        onCancel={() => {
          setMarketplaceOpen(false)
          setMarketSearch('')
          setMarketCategory('all')
        }}
        footer={null}
        width={800}
        styles={{ body: { paddingTop: 12 } }}
      >
        <div style={{ marginBottom: 16 }}>
          <Input
            prefix={<SearchOutlined style={{ color: '#98a2b3' }} />}
            placeholder="Search servers..."
            value={marketSearch}
            onChange={(e) => setMarketSearch(e.target.value)}
            allowClear
            style={{ borderRadius: 8, marginBottom: 12 }}
          />
          <Space size={4}>
            {Object.entries(categoryLabels).map(([key, label]) => (
              <Tag
                key={key}
                color={marketCategory === key ? 'blue' : undefined}
                style={{ cursor: 'pointer', borderRadius: 6, padding: '2px 10px' }}
                onClick={() => setMarketCategory(key)}
              >
                {label}
              </Tag>
            ))}
          </Space>
        </div>

        <Row gutter={[12, 12]}>
          {filteredMarketplace.map((item) => {
            const added = existingUrls.has(item.url)
            return (
              <Col key={item.url} xs={24} sm={12}>
                <div
                  style={{
                    padding: '16px',
                    background: '#f9fafb',
                    borderRadius: 10,
                    border: '1px solid #eaecf0',
                  }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 8 }}>
                    <div style={{ fontWeight: 600, fontSize: 14 }}>{item.name}</div>
                    <Button
                      type={added ? 'default' : 'primary'}
                      size="small"
                      icon={added ? <CheckOutlined /> : <PlusOutlined />}
                      disabled={added}
                      loading={addingUrl === item.url}
                      onClick={() => handleAddFromMarketplace(item)}
                      style={{ borderRadius: 6 }}
                    >
                      {added ? 'Added' : 'Add'}
                    </Button>
                  </div>
                  <div style={{ fontSize: 12, color: '#667085', marginBottom: 8, lineHeight: 1.5 }}>
                    {item.description}
                  </div>
                  <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                    {item.tags.map((tag) => (
                      <Tag key={tag} style={{ fontSize: 11, borderRadius: 4, margin: 0 }}>
                        {tag}
                      </Tag>
                    ))}
                  </div>
                </div>
              </Col>
            )
          })}
          {filteredMarketplace.length === 0 && (
            <Col span={24}>
              <Empty description="No matching servers" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            </Col>
          )}
        </Row>
      </Modal>
    </div>
  )
}
