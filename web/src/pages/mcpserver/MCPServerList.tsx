import { useEffect, useState } from 'react';
import { Button, Card, Table, Space, Modal, Form, Input, message, Popconfirm, Typography, Tag, Collapse } from 'antd';
import { PlusOutlined, DeleteOutlined, SyncOutlined, EyeOutlined } from '@ant-design/icons';
import { mcpServerApi } from '../../api/mcpserver';

export default function MCPServerList() {
  const [servers, setServers] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [detailModal, setDetailModal] = useState<any>(null);
  const [checkingId, setCheckingId] = useState<string | null>(null);
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try {
      const res: any = await mcpServerApi.list();
      setServers(res.data || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  const handleCreate = async () => {
    const values = await form.validateFields();
    await mcpServerApi.create(values);
    message.success('创建成功');
    setModalOpen(false);
    form.resetFields();
    load();
  };

  const handleDelete = async (id: string) => {
    await mcpServerApi.delete(id);
    message.success('删除成功');
    load();
  };

  const handleCheck = async (id: string) => {
    setCheckingId(id);
    try {
      await mcpServerApi.check(id);
      message.success('连接成功');
      load();
    } catch {
      load();
    } finally {
      setCheckingId(null);
    }
  };

  const statusColors: Record<string, string> = {
    connected: 'green',
    error: 'red',
    unknown: 'default',
  };

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: 'URL', dataIndex: 'url', key: 'url', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (s: string) => <Tag color={statusColors[s] || 'default'}>{s}</Tag>,
    },
    {
      title: '工具数',
      key: 'tools',
      width: 80,
      render: (_: any, r: any) => Array.isArray(r.tools) ? r.tools.length : 0,
    },
    {
      title: '检测时间',
      dataIndex: 'checked_at',
      width: 180,
      render: (t: string) => t ? new Date(t).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 260,
      render: (_: any, r: any) => (
        <Space>
          <Button
            size="small"
            icon={<SyncOutlined spin={checkingId === r.id} />}
            onClick={() => handleCheck(r.id)}
            loading={checkingId === r.id}
          >
            检测
          </Button>
          <Button size="small" icon={<EyeOutlined />} onClick={() => setDetailModal(r)}>
            详情
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
        <Typography.Title level={4} style={{ margin: 0 }}>MCP 服务器</Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          添加服务器
        </Button>
      </div>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={servers} loading={loading} pagination={false} />
      </Card>

      <Modal
        title="添加 MCP 服务器"
        open={modalOpen}
        onOk={handleCreate}
        onCancel={() => { setModalOpen(false); form.resetFields(); }}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="例如：my-mcp-server" />
          </Form.Item>
          <Form.Item name="url" label="SSE URL" rules={[{ required: true }]}>
            <Input placeholder="http://localhost:3001/sse" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={detailModal?.name || 'MCP 服务器详情'}
        open={!!detailModal}
        onCancel={() => setDetailModal(null)}
        footer={null}
        width={700}
      >
        {detailModal && (
          <div>
            <p><strong>URL:</strong> {detailModal.url}</p>
            <p><strong>状态:</strong> <Tag color={statusColors[detailModal.status]}>{detailModal.status}</Tag></p>
            {detailModal.description && <p><strong>描述:</strong> {detailModal.description}</p>}

            <Collapse
              style={{ marginTop: 16 }}
              items={[
                {
                  key: 'tools',
                  label: `工具 (${Array.isArray(detailModal.tools) ? detailModal.tools.length : 0})`,
                  children: Array.isArray(detailModal.tools) ? (
                    <div>
                      {detailModal.tools.map((tool: any, i: number) => (
                        <Card key={i} size="small" style={{ marginBottom: 8 }}>
                          <strong>{tool.name}</strong>
                          <div style={{ color: '#666', fontSize: 12 }}>{tool.description}</div>
                          {tool.inputSchema && (
                            <pre style={{ fontSize: 11, marginTop: 4, background: '#f5f5f5', padding: 4, borderRadius: 2, maxHeight: 150, overflow: 'auto' }}>
                              {JSON.stringify(tool.inputSchema, null, 2)}
                            </pre>
                          )}
                        </Card>
                      ))}
                    </div>
                  ) : <span>无工具</span>,
                },
                {
                  key: 'prompts',
                  label: `提示 (${Array.isArray(detailModal.prompts) ? detailModal.prompts.length : 0})`,
                  children: (
                    <pre style={{ fontSize: 12 }}>
                      {JSON.stringify(detailModal.prompts, null, 2)}
                    </pre>
                  ),
                },
                {
                  key: 'resources',
                  label: `资源 (${Array.isArray(detailModal.resources) ? detailModal.resources.length : 0})`,
                  children: (
                    <pre style={{ fontSize: 12 }}>
                      {JSON.stringify(detailModal.resources, null, 2)}
                    </pre>
                  ),
                },
              ]}
            />
          </div>
        )}
      </Modal>
    </div>
  );
}
