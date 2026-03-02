import { useEffect, useState } from 'react'
import { Button, Table, Space, message, Popconfirm, Modal, Form, Input, Tag } from 'antd'
import { PlusOutlined, ApiOutlined } from '@ant-design/icons'
import { mcpServerApi, type MCPServer } from '../../api/mcpserver'

export default function MCPServerList() {
  const [servers, setServers] = useState<MCPServer[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
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

  useEffect(() => { fetchList() }, [])

  const handleSubmit = async () => {
    const values = await form.validateFields()
    try {
      if (editId) {
        await mcpServerApi.update(editId, values)
        message.success('updated')
      } else {
        await mcpServerApi.create(values)
        message.success('created')
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
    message.success('deleted')
    fetchList()
  }

  const handleTest = async (id: number) => {
    try {
      const res: any = await mcpServerApi.test(id)
      message.success(`connected, ${res.tools?.length || 0} tools found`)
      fetchList()
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: 'Name', dataIndex: 'name' },
    { title: 'URL', dataIndex: 'url' },
    {
      title: 'Status', dataIndex: 'status', width: 100,
      render: (s: string) => <Tag color={s === 'active' ? 'green' : 'red'}>{s}</Tag>,
    },
    {
      title: 'Actions', width: 280,
      render: (_: any, record: MCPServer) => (
        <Space>
          <Button size="small" icon={<ApiOutlined />} onClick={() => handleTest(record.id)}>Test</Button>
          <Button size="small" onClick={() => handleEdit(record)}>Edit</Button>
          <Popconfirm title="Delete?" onConfirm={() => handleDelete(record.id)}>
            <Button size="small" danger>Delete</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>MCP Servers</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditId(null); form.resetFields(); setModalOpen(true) }}>
          Add Server
        </Button>
      </div>
      <Table rowKey="id" columns={columns} dataSource={servers} loading={loading} pagination={{ pageSize: 20 }} />

      <Modal
        title={editId ? 'Edit Server' : 'Add Server'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => { setModalOpen(false); setEditId(null) }}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="url" label="URL" rules={[{ required: true }]}>
            <Input placeholder="http://localhost:3001" />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input.TextArea rows={2} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
