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
  CloudServerOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  LinkOutlined,
  ToolOutlined,
  MessageOutlined,
  DatabaseOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import { mcpServerApi, type MCPServer } from '../../api/mcpserver'

export default function MCPServerList() {
  const [servers, setServers] = useState<MCPServer[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [search, setSearch] = useState('')
  const [form] = Form.useForm()

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
      setServers(res.data || [])
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
    try {
      if (editId) {
        await mcpServerApi.update(editId, values)
        message.success('Updated')
      } else {
        await mcpServerApi.create(values)
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
    form.setFieldsValue(record)
    setModalOpen(true)
  }

  const handleDelete = async (id: number) => {
    await mcpServerApi.delete(id)
    message.success('Deleted')
    fetchList()
  }

  const handleTest = async (id: number) => {
    try {
      const res: any = await mcpServerApi.test(id)
      message.success(`Connected, ${res.tools?.length || 0} tools found`)
      fetchList()
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const openDetail = async (server: MCPServer) => {
    setDetailServer(server)
    setDrawerOpen(true)
    setDetailLoading(true)
    setTools([])
    setPrompts([])
    setResources([])

    try {
      const [toolsRes, promptsRes, resourcesRes]: any[] = await Promise.allSettled([
        mcpServerApi.tools(server.id),
        mcpServerApi.prompts(server.id),
        mcpServerApi.resources(server.id),
      ])
      if (toolsRes.status === 'fulfilled') setTools(toolsRes.value?.tools || [])
      if (promptsRes.status === 'fulfilled') setPrompts(promptsRes.value?.prompts || [])
      if (resourcesRes.status === 'fulfilled') setResources(resourcesRes.value?.resources || [])
    } catch {
      // ignore
    } finally {
      setDetailLoading(false)
    }
  }

  const refreshDetail = () => {
    if (detailServer) openDetail(detailServer)
  }

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
      <div style={{ fontWeight: 600, fontSize: 13, marginBottom: 4 }}>{tool.name}</div>
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
          {!search && (
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
          )}
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
              <Descriptions.Item label="Created">{detailServer.created_at}</Descriptions.Item>
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
            <Input placeholder="http://localhost:3001/sse" style={{ borderRadius: 8 }} />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input.TextArea rows={2} placeholder="Optional description" style={{ borderRadius: 8 }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
