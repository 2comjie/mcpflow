import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Tag, Typography, Timeline, Collapse, Button, Spin } from 'antd';
import { ArrowLeftOutlined, CheckCircleOutlined, CloseCircleOutlined, ClockCircleOutlined, MinusCircleOutlined } from '@ant-design/icons';
import { executionApi } from '../../api/execution';

const statusConfig: Record<string, { color: string; icon: React.ReactNode }> = {
  completed: { color: 'green', icon: <CheckCircleOutlined style={{ color: '#52c41a' }} /> },
  failed: { color: 'red', icon: <CloseCircleOutlined style={{ color: '#ff4d4f' }} /> },
  running: { color: 'blue', icon: <ClockCircleOutlined style={{ color: '#1677ff' }} /> },
  skipped: { color: 'default', icon: <MinusCircleOutlined style={{ color: '#999' }} /> },
};

export default function ExecutionDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [exec, setExec] = useState<any>(null);
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    Promise.all([
      executionApi.get(id),
      executionApi.getLogs(id),
    ]).then(([execRes, logsRes]: any[]) => {
      setExec(execRes.data);
      setLogs(logsRes.data || []);
    }).finally(() => setLoading(false));
  }, [id]);

  if (loading) return <Spin style={{ display: 'block', margin: '100px auto' }} />;
  if (!exec) return <div>执行记录未找到</div>;

  const duration = exec.started_at && exec.finished_at
    ? ((new Date(exec.finished_at).getTime() - new Date(exec.started_at).getTime()) / 1000).toFixed(2) + 's'
    : '-';

  return (
    <div>
      <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)} style={{ marginBottom: 16 }}>
        返回
      </Button>
      <Typography.Title level={4}>执行详情</Typography.Title>

      <Card style={{ marginBottom: 16 }}>
        <Descriptions column={2} size="small">
          <Descriptions.Item label="执行 ID">{exec.id}</Descriptions.Item>
          <Descriptions.Item label="工作流 ID">{exec.workflow_id}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={statusConfig[exec.status]?.color || 'default'}>{exec.status}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="耗时">{duration}</Descriptions.Item>
          <Descriptions.Item label="开始时间">
            {exec.started_at ? new Date(exec.started_at).toLocaleString() : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="结束时间">
            {exec.finished_at ? new Date(exec.finished_at).toLocaleString() : '-'}
          </Descriptions.Item>
          {exec.error && (
            <Descriptions.Item label="错误" span={2}>
              <Typography.Text type="danger">{exec.error}</Typography.Text>
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>

      <Typography.Title level={5} style={{ marginBottom: 12 }}>节点执行日志</Typography.Title>

      {logs.length > 0 ? (
        <Timeline
          items={logs.map((log: any) => ({
            dot: statusConfig[log.status]?.icon,
            children: (
              <Card size="small" style={{ marginBottom: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <div>
                    <Tag>{log.node_type}</Tag>
                    <strong>{log.node_name || log.node_id}</strong>
                  </div>
                  <div style={{ color: '#999', fontSize: 12 }}>
                    {log.duration}ms
                  </div>
                </div>

                {log.error && (
                  <Typography.Text type="danger" style={{ display: 'block', marginBottom: 8 }}>
                    错误：{log.error}
                  </Typography.Text>
                )}

                {log.output && Object.keys(log.output).length > 0 && (
                  <Collapse
                    size="small"
                    items={[
                      {
                        key: 'output',
                        label: '输出',
                        children: (
                          <pre style={{ fontSize: 12, margin: 0, maxHeight: 300, overflow: 'auto', background: '#f5f5f5', padding: 8, borderRadius: 4 }}>
                            {JSON.stringify(log.output, null, 2)}
                          </pre>
                        ),
                      },
                    ]}
                  />
                )}

                {log.agent_steps && log.agent_steps.length > 0 && (
                  <Collapse
                    size="small"
                    style={{ marginTop: 8 }}
                    items={[
                      {
                        key: 'steps',
                        label: `Agent 步骤 (${log.agent_steps.length})`,
                        children: (
                          <div>
                            {log.agent_steps.map((step: any, i: number) => (
                              <Card key={i} size="small" style={{ marginBottom: 8, background: '#fafafa' }}>
                                <div style={{ marginBottom: 4 }}>
                                  <Tag color={step.type === 'tool_call' ? 'blue' : 'green'}>
                                    {step.type === 'tool_call' ? `调用工具: ${step.tool_name}` : '回复'}
                                  </Tag>
                                  <span style={{ color: '#999', fontSize: 12 }}>迭代 {step.iteration}</span>
                                </div>
                                {step.tool_args && (
                                  <div style={{ fontSize: 12, marginBottom: 4 }}>
                                    <strong>参数：</strong>
                                    <pre style={{ margin: 0, background: '#f0f0f0', padding: 4, borderRadius: 2 }}>
                                      {JSON.stringify(step.tool_args, null, 2)}
                                    </pre>
                                  </div>
                                )}
                                {step.tool_result && (
                                  <div style={{ fontSize: 12, marginBottom: 4 }}>
                                    <strong>结果：</strong>
                                    <pre style={{ margin: 0, background: '#f0f0f0', padding: 4, borderRadius: 2, maxHeight: 200, overflow: 'auto' }}>
                                      {step.tool_result}
                                    </pre>
                                  </div>
                                )}
                                {step.content && (
                                  <div style={{ fontSize: 12 }}>
                                    <strong>内容：</strong> {step.content}
                                  </div>
                                )}
                              </Card>
                            ))}
                          </div>
                        ),
                      },
                    ]}
                  />
                )}
              </Card>
            ),
          }))}
        />
      ) : (
        <Card><Typography.Text type="secondary">暂无执行日志</Typography.Text></Card>
      )}

      {exec.output && (
        <>
          <Typography.Title level={5} style={{ margin: '16px 0 12px' }}>最终输出</Typography.Title>
          <Card>
            <pre style={{ fontSize: 12, margin: 0, maxHeight: 400, overflow: 'auto' }}>
              {JSON.stringify(exec.output, null, 2)}
            </pre>
          </Card>
        </>
      )}
    </div>
  );
}
