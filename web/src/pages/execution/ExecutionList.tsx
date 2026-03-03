import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Table, Tag, message, Drawer, Descriptions, Timeline, Space, Tooltip } from 'antd'
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import { executionApi, type Execution, type ExecutionLog } from '../../api/execution'
import { workflowApi } from '../../api/workflow'

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
  const workflowId = Number(id)

  const [workflowName, setWorkflowName] = useState('')
  const [executions, setExecutions] = useState<Execution[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [detail, setDetail] = useState<Execution | null>(null)
  const [logs, setLogs] = useState<ExecutionLog[]>([])
  const [logsLoading, setLogsLoading] = useState(false)

  const loadExecutions = async (p = 1) => {
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
    workflowApi
      .get(workflowId)
      .then((res: any) => setWorkflowName(res.name))
      .catch(() => {})
    loadExecutions()
  }, [workflowId])

  const openDetail = async (execId: number) => {
    try {
      const [exec, logsRes]: any = await Promise.all([
        executionApi.get(execId),
        executionApi.logs(execId),
      ])
      setDetail(exec)
      setLogs(logsRes.data || [])
      setDrawerOpen(true)
    } catch (err: any) {
      message.error(err.message)
    }
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
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
      width: 140,
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
        title={`Execution #${detail?.id}`}
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
                        {log.attempt > 1 && (
                          <Tag style={{ fontSize: 11 }}>attempt #{log.attempt}</Tag>
                        )}
                      </div>
                      <div style={{ fontSize: 12, color: '#667085' }}>
                        Type: {log.node_type} | Duration: {log.duration}ms
                      </div>
                      {log.error && (
                        <div style={{ fontSize: 12, color: '#f04438', marginTop: 4 }}>
                          {log.error}
                        </div>
                      )}
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
