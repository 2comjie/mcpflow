import { useEffect, useState } from 'react'
import { Button, Table, Space, message, Popconfirm } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { workflowApi, type Workflow } from '../../api/workflow'

export default function WorkflowList() {
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const fetchList = async () => {
    setLoading(true)
    try {
      const res: any = await workflowApi.list()
      setWorkflows(res.data || [])
    } catch (err: any) {
      message.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchList() }, [])

  const handleDelete = async (id: number) => {
    try {
      await workflowApi.delete(id)
      message.success('deleted')
      fetchList()
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: 'Name', dataIndex: 'name' },
    { title: 'Status', dataIndex: 'status', width: 100 },
    { title: 'Updated', dataIndex: 'updated_at', width: 200 },
    {
      title: 'Actions',
      width: 200,
      render: (_: any, record: Workflow) => (
        <Space>
          <Button size="small" onClick={() => navigate(`/workflows/${record.id}`)}>Edit</Button>
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
        <h2 style={{ margin: 0 }}>Workflows</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/workflows/new')}>
          New Workflow
        </Button>
      </div>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={workflows}
        loading={loading}
        pagination={{ pageSize: 20 }}
      />
    </div>
  )
}
