import { useEffect, useState } from 'react';
import { Button, Card, Table, Space, Modal, Form, Input, message, Popconfirm, Typography, Tag } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { llmProviderApi } from '../../api/llmprovider';

export default function LLMProviderList() {
  const [providers, setProviders] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editId, setEditId] = useState<string | null>(null);
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try {
      const res: any = await llmProviderApi.list();
      setProviders(res.data || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  const handleSave = async () => {
    const values = await form.validateFields();
    // models 输入为逗号分隔字符串，转为数组
    if (typeof values.models === 'string') {
      values.models = values.models.split(',').map((s: string) => s.trim()).filter(Boolean);
    }
    if (editId) {
      await llmProviderApi.update(editId, values);
      message.success('更新成功');
    } else {
      await llmProviderApi.create(values);
      message.success('创建成功');
    }
    setModalOpen(false);
    setEditId(null);
    form.resetFields();
    load();
  };

  const handleEdit = (record: any) => {
    setEditId(record.id);
    form.setFieldsValue({
      ...record,
      models: record.models?.join(', ') || '',
      api_key: record.api_key,
    });
    setModalOpen(true);
  };

  const handleDelete = async (id: string) => {
    await llmProviderApi.delete(id);
    message.success('删除成功');
    load();
  };

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: 'Base URL', dataIndex: 'base_url', key: 'base_url', ellipsis: true },
    {
      title: '模型',
      dataIndex: 'models',
      key: 'models',
      render: (models: string[]) => models?.map((m) => <Tag key={m}>{m}</Tag>) || '-',
    },
    {
      title: 'API Key',
      dataIndex: 'api_key',
      width: 180,
      render: (k: string) => k ? `${k.slice(0, 6)}...${k.slice(-4)}` : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, r: any) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(r)}>
            编辑
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
        <Typography.Title level={4} style={{ margin: 0 }}>LLM 供应商</Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditId(null); form.resetFields(); setModalOpen(true); }}>
          添加供应商
        </Button>
      </div>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={providers} loading={loading} pagination={false} />
      </Card>

      <Modal
        title={editId ? '编辑 LLM 供应商' : '添加 LLM 供应商'}
        open={modalOpen}
        onOk={handleSave}
        onCancel={() => { setModalOpen(false); setEditId(null); form.resetFields(); }}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="例如：DeepSeek" />
          </Form.Item>
          <Form.Item name="base_url" label="Base URL" rules={[{ required: true }]}>
            <Input placeholder="https://api.deepseek.com/v1" />
          </Form.Item>
          <Form.Item name="api_key" label="API Key" rules={[{ required: true }]}>
            <Input.Password placeholder="sk-..." />
          </Form.Item>
          <Form.Item name="models" label="模型（逗号分隔）" rules={[{ required: true }]}>
            <Input placeholder="deepseek-chat, deepseek-reasoner" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
