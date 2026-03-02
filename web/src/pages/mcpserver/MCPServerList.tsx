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
} from 'antd'
import {
  PlusOutlined,
  CloudServerOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  LinkOutlined,
} from '@ant-design/icons'
import { mcpServerApi, type MCPServer } from '../../api/mcpserver'

export default function MCPServerList() {
  const [servers, setServers] = useState<MCPServer[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [search, setSearch] = useState('')
  const [form] = Form.useForm()

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

  const filtered = servers.filter(
    (s) =>
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.url.toLowerCase().includes(search.toLowerCase()) ||
      (s.description || '').toLowerCase().includes(search.toLowerCase()),
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
                    <div className="server-card-name">
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
