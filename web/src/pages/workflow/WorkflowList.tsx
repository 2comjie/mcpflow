import { useEffect, useState } from 'react'
import { Button, Row, Col, Tag, Dropdown, message, Input } from 'antd'
import {
  PlusOutlined,
  SearchOutlined,
  ApartmentOutlined,
  ClockCircleOutlined,
  NodeIndexOutlined,
  MoreOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { workflowApi, type Workflow } from '../../api/workflow'

export default function WorkflowList() {
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [loading, setLoading] = useState(false)
  const [search, setSearch] = useState('')
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

  useEffect(() => {
    fetchList()
  }, [])

  const handleDelete = async (id: number) => {
    try {
      await workflowApi.delete(id)
      message.success('Deleted')
      fetchList()
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const handleExecute = async (id: number) => {
    try {
      const res: any = await workflowApi.execute(id)
      message.success(`Executed, status: ${res.status}`)
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const filtered = workflows.filter(
    (w) =>
      w.name.toLowerCase().includes(search.toLowerCase()) ||
      (w.description || '').toLowerCase().includes(search.toLowerCase()),
  )

  const formatTime = (t: string) => {
    if (!t) return '-'
    const d = new Date(t)
    const now = new Date()
    const diff = now.getTime() - d.getTime()
    if (diff < 60000) return 'just now'
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  }

  return (
    <div>
      <div className="page-header">
        <div>
          <h2>Workflows</h2>
          <div className="page-header-sub">
            Create and manage your automation workflows
          </div>
        </div>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate('/workflows/new')}
          style={{ borderRadius: 10 }}
        >
          New Workflow
        </Button>
      </div>

      {workflows.length > 0 && (
        <div style={{ marginBottom: 20 }}>
          <Input
            prefix={<SearchOutlined style={{ color: '#98a2b3' }} />}
            placeholder="Search workflows..."
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
            <Col key={i} xs={24} sm={12} lg={8} xl={6}>
              <div className="workflow-card" style={{ opacity: 0.5, height: 180 }}>
                <div style={{ background: '#f2f4f7', height: 20, width: '60%', borderRadius: 4, marginBottom: 8 }} />
                <div style={{ background: '#f2f4f7', height: 14, width: '80%', borderRadius: 4 }} />
              </div>
            </Col>
          ))}
        </Row>
      ) : filtered.length === 0 ? (
        <div className="empty-state">
          <ApartmentOutlined className="empty-state-icon" />
          <div className="empty-state-title">
            {search ? 'No matching workflows' : 'No workflows yet'}
          </div>
          <div className="empty-state-desc">
            {search
              ? 'Try different keywords'
              : 'Create your first workflow to start building automation pipelines'}
          </div>
          {!search && (
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => navigate('/workflows/new')}
              style={{ borderRadius: 10 }}
            >
              New Workflow
            </Button>
          )}
        </div>
      ) : (
        <Row gutter={[16, 16]}>
          {filtered.map((w) => (
            <Col key={w.id} xs={24} sm={12} lg={8} xl={6}>
              <div
                className="workflow-card"
                onClick={() => navigate(`/workflows/${w.id}`)}
              >
                <div className="workflow-card-title">
                  <ApartmentOutlined style={{ color: '#3b5bdb', fontSize: 16 }} />
                  <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {w.name || 'Untitled'}
                  </span>
                  <Tag
                    color={w.status === 'active' ? 'green' : 'default'}
                    style={{ margin: 0, borderRadius: 4, fontSize: 11 }}
                  >
                    {w.status === 'active' ? 'Active' : 'Draft'}
                  </Tag>
                </div>

                <div className="workflow-card-desc">
                  {w.description || 'No description'}
                </div>

                <div className="workflow-card-footer">
                  <div className="workflow-card-meta">
                    <span>
                      <NodeIndexOutlined />
                      {w.nodes?.length || 0} nodes
                    </span>
                    <span>
                      <ClockCircleOutlined />
                      {formatTime(w.updated_at)}
                    </span>
                  </div>
                  <Dropdown
                    trigger={['click']}
                    menu={{
                      items: [
                        {
                          key: 'edit',
                          icon: <EditOutlined />,
                          label: 'Edit',
                          onClick: (e) => {
                            e.domEvent.stopPropagation()
                            navigate(`/workflows/${w.id}`)
                          },
                        },
                        {
                          key: 'run',
                          icon: <PlayCircleOutlined />,
                          label: 'Execute',
                          onClick: (e) => {
                            e.domEvent.stopPropagation()
                            handleExecute(w.id)
                          },
                        },
                        {
                          key: 'executions',
                          icon: <UnorderedListOutlined />,
                          label: 'Executions',
                          onClick: (e) => {
                            e.domEvent.stopPropagation()
                            navigate(`/workflows/${w.id}/executions`)
                          },
                        },
                        { type: 'divider' },
                        {
                          key: 'delete',
                          icon: <DeleteOutlined />,
                          label: 'Delete',
                          danger: true,
                          onClick: (e) => {
                            e.domEvent.stopPropagation()
                            handleDelete(w.id)
                          },
                        },
                      ],
                    }}
                  >
                    <Button
                      type="text"
                      size="small"
                      icon={<MoreOutlined />}
                      onClick={(e) => e.stopPropagation()}
                      style={{ color: '#98a2b3' }}
                    />
                  </Dropdown>
                </div>
              </div>
            </Col>
          ))}
        </Row>
      )}
    </div>
  )
}
