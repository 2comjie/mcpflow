import { useEffect, useState } from 'react'
import { Button, Table, Modal, Form, Input, Select, message, Popconfirm, Space, Tag } from 'antd'
import { PlusOutlined, DeleteOutlined, EditOutlined, RobotOutlined } from '@ant-design/icons'
import { llmProviderApi, type LLMProvider } from '../../api/llm_provider'

export default function LLMProviderList() {
  const [providers, setProviders] = useState<LLMProvider[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const res: any = await llmProviderApi.list()
      setProviders(Array.isArray(res) ? res : res.data || [])
    } catch (err: any) {
      message.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      // 处理 models：逗号分隔字符串转数组
      if (typeof values.models === 'string') {
        values.models = values.models
          .split(',')
          .map((s: string) => s.trim())
          .filter(Boolean)
      }
      if (editingId) {
        await llmProviderApi.update(editingId, values)
        message.success('Updated')
      } else {
        await llmProviderApi.create(values)
        message.success('Created')
      }
      setModalOpen(false)
      form.resetFields()
      setEditingId(null)
      load()
    } catch (err: any) {
      if (err.message) message.error(err.message)
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await llmProviderApi.delete(id)
      message.success('Deleted')
      load()
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const openCreate = () => {
    setEditingId(null)
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = (record: LLMProvider) => {
    setEditingId(record.id)
    form.setFieldsValue({
      ...record,
      api_key: '', // 编辑时不回填脱敏的 key
      models: Array.isArray(record.models) ? record.models.join(', ') : '',
    })
    setModalOpen(true)
  }

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      render: (text: string) => (
        <span style={{ fontWeight: 500 }}>
          <RobotOutlined style={{ marginRight: 6, color: '#3b5bdb' }} />
          {text}
        </span>
      ),
    },
    {
      title: 'Base URL',
      dataIndex: 'base_url',
      ellipsis: true,
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{text}</span>
      ),
    },
    {
      title: 'API Key',
      dataIndex: 'api_key',
      width: 160,
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontSize: 12, color: '#98a2b3' }}>{text}</span>
      ),
    },
    {
      title: 'Models',
      dataIndex: 'models',
      render: (models: string[]) =>
        models && models.length > 0 ? (
          <Space size={4} wrap>
            {models.map((m) => (
              <Tag key={m} style={{ borderRadius: 4, fontSize: 11 }}>
                {m}
              </Tag>
            ))}
          </Space>
        ) : (
          '-'
        ),
    },
    {
      title: 'Actions',
      width: 120,
      render: (_: any, record: LLMProvider) => (
        <Space>
          <Button type="text" size="small" icon={<EditOutlined />} onClick={() => openEdit(record)} />
          <Popconfirm title="Delete this provider?" onConfirm={() => handleDelete(record.id)}>
            <Button type="text" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>LLM Providers</h2>
          <div className="page-header-sub">
            Manage LLM provider configurations for workflow nodes
          </div>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          Add Provider
        </Button>
      </div>

      <Table
        dataSource={providers}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={false}
        style={{ background: '#fff', borderRadius: 12 }}
      />

      <Modal
        title={editingId ? 'Edit LLM Provider' : 'Add LLM Provider'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => {
          setModalOpen(false)
          setEditingId(null)
        }}
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Name is required' }]}
          >
            <Input placeholder="e.g. DeepSeek, OpenAI" />
          </Form.Item>
          <Form.Item
            name="base_url"
            label="Base URL"
            rules={[{ required: true, message: 'Base URL is required' }]}
          >
            <Input placeholder="https://api.deepseek.com/v1" style={{ fontFamily: 'monospace' }} />
          </Form.Item>
          <Form.Item
            name="api_key"
            label="API Key"
            rules={editingId ? [] : [{ required: true, message: 'API Key is required' }]}
          >
            <Input.Password placeholder={editingId ? 'Leave empty to keep current' : 'Enter API key'} />
          </Form.Item>
          <Form.Item name="models" label="Models">
            <Input placeholder="deepseek-chat, deepseek-coder (comma separated)" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
