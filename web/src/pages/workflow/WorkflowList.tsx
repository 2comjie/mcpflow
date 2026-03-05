import { useEffect, useState } from 'react'
import { Button, Row, Col, Dropdown, message, Input, Modal, Tag, Space, Empty } from 'antd'
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
  AppstoreOutlined,
  RocketOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { workflowApi, type Workflow } from '../../api/workflow'
import { workflowTemplates, templateCategoryLabels, type WorkflowTemplate } from '../../data/workflowTemplates'
import ExecuteWorkflowModal from '../../components/ExecuteWorkflowModal'

export default function WorkflowList() {
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [loading, setLoading] = useState(false)
  const [search, setSearch] = useState('')
  const [templateModalOpen, setTemplateModalOpen] = useState(false)
  const [templateCategory, setTemplateCategory] = useState('all')
  const [creatingTemplate, setCreatingTemplate] = useState<string | null>(null)
  const [executeWorkflowId, setExecuteWorkflowId] = useState<string | null>(null)
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

  const handleDelete = async (id: string) => {
    try {
      await workflowApi.delete(id)
      message.success('Deleted')
      fetchList()
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const handleExecute = (id: string) => {
    setExecuteWorkflowId(id)
  }

  const handleUseTemplate = async (template: WorkflowTemplate) => {
    setCreatingTemplate(template.id)
    try {
      const res: any = await workflowApi.create({
        name: template.name,
        description: template.description,
        nodes: template.nodes,
        edges: template.edges,
      })
      message.success('Workflow created from template')
      setTemplateModalOpen(false)
      const wf = res.data || res
      navigate(`/workflows/${wf.id}`)
    } catch (err: any) {
      message.error(err.message)
    } finally {
      setCreatingTemplate(null)
    }
  }

  const filteredTemplates = workflowTemplates.filter(
    (t) => templateCategory === 'all' || t.category === templateCategory,
  )

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
        <Space>
          <Button
            icon={<AppstoreOutlined />}
            onClick={() => setTemplateModalOpen(true)}
            style={{ borderRadius: 10 }}
          >
            Templates
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/workflows/new')}
            style={{ borderRadius: 10 }}
          >
            New Workflow
          </Button>
        </Space>
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

      {/* Template Modal */}
      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <AppstoreOutlined style={{ color: '#3b5bdb' }} />
            Workflow Templates
          </div>
        }
        open={templateModalOpen}
        onCancel={() => {
          setTemplateModalOpen(false)
          setTemplateCategory('all')
        }}
        footer={null}
        width={720}
        styles={{ body: { paddingTop: 12 } }}
      >
        <div style={{ marginBottom: 16 }}>
          <Space size={4}>
            {Object.entries(templateCategoryLabels).map(([key, label]) => (
              <Tag
                key={key}
                color={templateCategory === key ? 'blue' : undefined}
                style={{ cursor: 'pointer', borderRadius: 6, padding: '2px 10px' }}
                onClick={() => setTemplateCategory(key)}
              >
                {label}
              </Tag>
            ))}
          </Space>
        </div>

        <Row gutter={[12, 12]}>
          {filteredTemplates.map((template) => (
            <Col key={template.id} xs={24} sm={12}>
              <div
                style={{
                  padding: '16px',
                  background: '#f9fafb',
                  borderRadius: 10,
                  border: '1px solid #eaecf0',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 8 }}>
                  <div style={{ fontWeight: 600, fontSize: 14 }}>{template.name}</div>
                  <Button
                    type="primary"
                    size="small"
                    icon={<RocketOutlined />}
                    loading={creatingTemplate === template.id}
                    onClick={() => handleUseTemplate(template)}
                    style={{ borderRadius: 6 }}
                  >
                    Use
                  </Button>
                </div>
                <div style={{ fontSize: 12, color: '#667085', marginBottom: 8, lineHeight: 1.5 }}>
                  {template.description}
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 11, color: '#98a2b3' }}>
                  <span><NodeIndexOutlined /> {template.nodes.length} nodes</span>
                  <Tag style={{ fontSize: 11, borderRadius: 4, margin: 0 }}>{template.category}</Tag>
                </div>
              </div>
            </Col>
          ))}
          {filteredTemplates.length === 0 && (
            <Col span={24}>
              <Empty description="No templates in this category" image={Empty.PRESENTED_IMAGE_SIMPLE} />
            </Col>
          )}
        </Row>
      </Modal>

      <ExecuteWorkflowModal
        open={executeWorkflowId !== null}
        workflowId={executeWorkflowId}
        onClose={() => setExecuteWorkflowId(null)}
      />
    </div>
  )
}
