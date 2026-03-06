import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Table, Tag, message, Drawer, Descriptions, Timeline, Space, Tooltip, Popconfirm } from 'antd'
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  FileTextOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { executionApi, type Execution, type ExecutionLog } from '../../api/execution'
import { workflowApi } from '../../api/workflow'
import AgentStepsView from '../../components/AgentStepsView'

const formatDateTime = (v: string) => {
  if (!v) return '-'
  const d = new Date(v)
  if (isNaN(d.getTime())) return v
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const statusTag = (status: string) => {
  const map: Record<string, { color: string; icon: React.ReactNode }> = {
    completed: { color: 'success', icon: <CheckCircleOutlined /> },
    failed: { color: 'error', icon: <CloseCircleOutlined /> },
    running: { color: 'processing', icon: <SyncOutlined spin /> },
    pending: { color: 'warning', icon: <ClockCircleOutlined /> },
  }
  const cfg = map[status] || { color: 'default', icon: null }
  return (
    <Tag color={cfg.color} icon={cfg.icon}>
      {status}
    </Tag>
  )
}

export default function ExecutionList() {
  const { id } = useParams()
  const navigate = useNavigate()
  const workflowId = id

  const [workflowName, setWorkflowName] = useState('')
  const [executions, setExecutions] = useState<Execution[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [detail, setDetail] = useState<Execution | null>(null)
  const [logs, setLogs] = useState<ExecutionLog[]>([])
  const loadExecutions = async (p = 1) => {
    if (!workflowId) return
    setLoading(true)
    try {
      const res: any = await executionApi.listByWorkflow(workflowId, p)
      setExecutions(res.data || [])
      setTotal(res.total || 0)
    } catch (err: any) {
      message.error(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!workflowId) return
    workflowApi
      .get(workflowId)
      .then((res: any) => setWorkflowName((res.data || res).name))
      .catch(() => {})
    loadExecutions()
  }, [workflowId])

  const handleDelete = async (execId: string) => {
    try {
      await executionApi.delete(execId)
      message.success('Deleted')
      loadExecutions(page)
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const openDetail = async (execId: string) => {
    try {
      const [exec, logsRes]: any = await Promise.all([
        executionApi.get(execId),
        executionApi.logs(execId),
      ])
      setDetail(exec.data || exec)
      setLogs(Array.isArray(logsRes) ? logsRes : logsRes.data || [])
      setDrawerOpen(true)
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 100,
      render: (v: string) => (
        <Tooltip title={v}>
          <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{v?.slice(0, 8)}</span>
        </Tooltip>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      width: 120,
      render: (s: string) => statusTag(s),
    },
    {
      title: 'Started',
      dataIndex: 'started_at',
      width: 180,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: 'Finished',
      dataIndex: 'finished_at',
      width: 180,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: 'Error',
      dataIndex: 'error',
      ellipsis: true,
      render: (v: string) =>
        v ? (
          <Tooltip title={v}>
            <span style={{ color: '#f04438' }}>{v}</span>
          </Tooltip>
        ) : (
          '-'
        ),
    },
    {
      title: 'Actions',
      width: 180,
      render: (_: any, record: Execution) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => openDetail(record.id)}
          >
            Detail
          </Button>
          {record.status !== 'running' && (
            <Popconfirm
              title="Delete this execution?"
              onConfirm={() => handleDelete(record.id)}
            >
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                Delete
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div className="page-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <Button
            type="text"
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/workflows')}
            style={{ color: '#667085' }}
          />
          <div>
            <h2>Executions</h2>
            <div className="page-header-sub">{workflowName || `Workflow #${workflowId}`}</div>
          </div>
        </div>
      </div>

      <Table
        dataSource={executions}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          total,
          pageSize: 20,
          onChange: (p) => {
            setPage(p)
            loadExecutions(p)
          },
        }}
        style={{ background: '#fff', borderRadius: 12 }}
      />

      <Drawer
        title={`Execution #${detail?.id?.slice(0, 8)}`}
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        width={560}
      >
        {detail && (
          <>
            <Descriptions column={1} size="small" style={{ marginBottom: 24 }}>
              <Descriptions.Item label="Status">{statusTag(detail.status)}</Descriptions.Item>
              <Descriptions.Item label="Started">{formatDateTime(detail.started_at)}</Descriptions.Item>
              <Descriptions.Item label="Finished">{formatDateTime(detail.finished_at)}</Descriptions.Item>
              {detail.error && (
                <Descriptions.Item label="Error">
                  <span style={{ color: '#f04438' }}>{detail.error}</span>
                </Descriptions.Item>
              )}
            </Descriptions>

            <h4 style={{ marginBottom: 12, fontWeight: 600 }}>Node Execution Logs</h4>
            {logs.length === 0 ? (
              <div style={{ color: '#98a2b3', textAlign: 'center', padding: 24 }}>No logs</div>
            ) : (
              <Timeline
                items={logs.map((log) => ({
                  color: log.status === 'completed' ? 'green' : 'red',
                  children: (
                    <div>
                      <div style={{ fontWeight: 500, marginBottom: 4 }}>
                        {log.node_name || log.node_id}
                        <Tag
                          style={{ marginLeft: 8, fontSize: 11 }}
                          color={log.status === 'completed' ? 'success' : 'error'}
                        >
                          {log.status}
                        </Tag>
                      </div>
                      <div style={{ fontSize: 12, color: '#667085' }}>
                        Type: {log.node_type} | Duration: {log.duration}ms
                      </div>
                      {log.error && (
                        <div style={{ fontSize: 12, color: '#f04438', marginTop: 4 }}>
                          {log.error}
                        </div>
                      )}
                      {log.node_type === 'agent' && ((log.agent_steps?.length ?? 0) > 0 || log.output?.agent_steps) ? (
                        <>
                          {log.output?.content && (
                            <div style={{ fontSize: 12, color: '#344054', marginTop: 4, padding: '4px 8px', background: '#f0fdf4', borderRadius: 4, border: '1px solid #bbf7d0' }}>
                              {log.output.content.length > 300 ? log.output.content.slice(0, 300) + '...' : log.output.content}
                            </div>
                          )}
                          <AgentStepsView steps={log.agent_steps || log.output?.agent_steps} />
                        </>
                      ) : log.output && Object.keys(log.output).length > 0 ? (
                        <pre
                          style={{
                            fontSize: 11,
                            background: '#f9fafb',
                            border: '1px solid #e5e7eb',
                            borderRadius: 6,
                            padding: '6px 10px',
                            marginTop: 6,
                            maxHeight: 160,
                            overflow: 'auto',
                            whiteSpace: 'pre-wrap',
                            wordBreak: 'break-all',
                          }}
                        >
                          {JSON.stringify(log.output, null, 2)}
                        </pre>
                      ) : null}
                    </div>
                  ),
                }))}
              />
            )}
          </>
        )}
      </Drawer>
    </div>
  )
}
