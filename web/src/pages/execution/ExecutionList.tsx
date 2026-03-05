import { useEffect, useState } from 'react';
import { Card, Table, Tag, Typography, Button, Popconfirm, message } from 'antd';
import { EyeOutlined, DeleteOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { executionApi } from '../../api/execution';

const statusColors: Record<string, string> = {
  completed: 'green',
  failed: 'red',
  running: 'blue',
  pending: 'default',
};

export default function ExecutionList() {
  const [executions, setExecutions] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const load = async (p = page) => {
    setLoading(true);
    try {
      const res: any = await executionApi.list(p, 20);
      setExecutions(res.data || []);
      setTotal(res.total || 0);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [page]);

  const handleDelete = async (id: string) => {
    await executionApi.delete(id);
    message.success('删除成功');
    load();
  };

  const columns = [
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (s: string) => <Tag color={statusColors[s] || 'default'}>{s}</Tag>,
    },
    {
      title: '工作流 ID',
      dataIndex: 'workflow_id',
      ellipsis: true,
      width: 220,
    },
    {
      title: '开始时间',
      dataIndex: 'started_at',
      width: 180,
      render: (t: string) => t ? new Date(t).toLocaleString() : '-',
    },
    {
      title: '结束时间',
      dataIndex: 'finished_at',
      width: 180,
      render: (t: string) => t ? new Date(t).toLocaleString() : '-',
    },
    {
      title: '错误',
      dataIndex: 'error',
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      width: 140,
      render: (_: any, r: any) => (
        <>
          <Button size="small" icon={<EyeOutlined />} onClick={() => navigate(`/executions/${r.id}`)} style={{ marginRight: 8 }}>
            详情
          </Button>
          <Popconfirm title="确认删除?" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </>
      ),
    },
  ];

  return (
    <div>
      <Typography.Title level={4} style={{ marginBottom: 16 }}>执行记录</Typography.Title>
      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={executions}
          loading={loading}
          pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
        />
      </Card>
    </div>
  );
}
