import { useEffect, useState } from 'react';
import { Button, Card, Table, Space, Modal, Form, Input, message, Popconfirm, Typography } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { workflowApi } from '../../api/workflow';

export default function WorkflowList() {
  const [workflows, setWorkflows] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const load = async (p = page) => {
    setLoading(true);
    try {
      const res: any = await workflowApi.list(p, 20);
      setWorkflows(res.data || []);
      setTotal(res.total || 0);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [page]);

  const handleCreate = async () => {
    const values = await form.validateFields();
    await workflowApi.create({
      ...values,
      nodes: [],
      edges: [],
    });
    message.success('创建成功');
    setModalOpen(false);
    form.resetFields();
    load();
  };

  const handleDelete = async (id: string) => {
    await workflowApi.delete(id);
    message.success('删除成功');
    load();
  };

  const handleExecute = async (id: string) => {
    const res: any = await workflowApi.execute(id);
    const exec = res.data;
    if (exec) {
      message.success('执行完成');
      navigate(`/executions/${exec.id}`);
    }
  };

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
    {
      title: '节点数',
      key: 'nodes',
      render: (_: any, r: any) => r.nodes?.length || 0,
      width: 80,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (t: string) => t ? new Date(t).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 220,
      render: (_: any, r: any) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => navigate(`/workflows/${r.id}`)}>
            编辑
          </Button>
          <Button size="small" type="primary" icon={<PlayCircleOutlined />} onClick={() => handleExecute(r.id)}>
            执行
          </Button>
          <Popconfirm title="确认删除?" onConfirm={() => handleDelete(r.id)}>
            <Button size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>工作流管理</Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          新建工作流
        </Button>
      </div>
      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={workflows}
          loading={loading}
          pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
        />
      </Card>

      <Modal
        title="新建工作流"
        open={modalOpen}
        onOk={handleCreate}
        onCancel={() => { setModalOpen(false); form.resetFields(); }}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="输入工作流名称" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="描述工作流用途" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
