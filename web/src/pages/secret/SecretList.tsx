import { useEffect, useState } from 'react'
import { Button, Table, Modal, Form, Input, message, Popconfirm, Space } from 'antd'
import { PlusOutlined, DeleteOutlined, EditOutlined, LockOutlined } from '@ant-design/icons'
import { secretApi, type Secret } from '../../api/secret'

export default function SecretList() {
  const [secrets, setSecrets] = useState<Secret[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const res: any = await secretApi.list()
      setSecrets(res.data || [])
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
      if (editingId) {
        await secretApi.update(editingId, { value: values.value, desc: values.desc })
        message.success('Updated')
      } else {
        await secretApi.create(values)
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
      await secretApi.delete(id)
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

  const openEdit = (record: Secret) => {
    setEditingId(record.id)
    form.setFieldsValue({ key: record.key, value: '', desc: record.desc })
    setModalOpen(true)
  }

  const columns = [
    {
      title: 'Key',
      dataIndex: 'key',
      render: (text: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 500 }}>
          <LockOutlined style={{ marginRight: 6, color: '#f79009' }} />
          {text}
        </span>
      ),
    },
    { title: 'Description', dataIndex: 'desc', ellipsis: true },
    { title: 'Created', dataIndex: 'created_at', width: 180 },
    { title: 'Updated', dataIndex: 'updated_at', width: 180 },
    {
      title: 'Actions',
      width: 120,
      render: (_: any, record: Secret) => (
        <Space>
          <Button type="text" size="small" icon={<EditOutlined />} onClick={() => openEdit(record)} />
          <Popconfirm title="Delete this secret?" onConfirm={() => handleDelete(record.id)}>
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
          <h2>Secrets</h2>
          <div className="page-header-sub">
            Manage global secrets for workflow templates. Use {'{{secret.key}}'} to reference.
          </div>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          Add Secret
        </Button>
      </div>

      <Table
        dataSource={secrets}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={false}
        style={{ background: '#fff', borderRadius: 12 }}
      />

      <Modal
        title={editingId ? 'Edit Secret' : 'Add Secret'}
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
            name="key"
            label="Key"
            rules={[{ required: true, message: 'Key is required' }]}
          >
            <Input
              placeholder="e.g. deepseek_api_key"
              disabled={!!editingId}
              style={{ fontFamily: 'monospace' }}
            />
          </Form.Item>
          <Form.Item
            name="value"
            label="Value"
            rules={[{ required: true, message: 'Value is required' }]}
          >
            <Input.Password placeholder="Enter secret value" />
          </Form.Item>
          <Form.Item name="desc" label="Description">
            <Input placeholder="Optional description" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
