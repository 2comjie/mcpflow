import { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, message, Drawer, Form, Input, Select, Space, InputNumber, Spin, Typography, Card } from 'antd';
import { SaveOutlined, PlayCircleOutlined, ArrowLeftOutlined, PlusOutlined } from '@ant-design/icons';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  type Connection,
  type Node,
  type Edge,
  MarkerType,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { workflowApi } from '../../api/workflow';

const nodeTypeOptions = [
  { label: '开始', value: 'start' },
  { label: '结束', value: 'end' },
  { label: 'LLM', value: 'llm' },
  { label: 'Agent', value: 'agent' },
  { label: '条件', value: 'condition' },
  { label: '代码', value: 'code' },
  { label: 'HTTP', value: 'http' },
  { label: '邮件', value: 'email' },
];

const nodeColors: Record<string, string> = {
  start: '#52c41a',
  end: '#ff4d4f',
  llm: '#1677ff',
  agent: '#722ed1',
  condition: '#fa8c16',
  code: '#13c2c2',
  http: '#eb2f96',
  email: '#faad14',
};

export default function WorkflowEditor() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [workflow, setWorkflow] = useState<any>(null);
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([] as Node[]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([] as Edge[]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // 节点编辑抽屉
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editingNode, setEditingNode] = useState<any>(null);
  const [form] = Form.useForm();
  const [nodeType, setNodeType] = useState<string>('start');

  // 添加节点面板
  const [addDrawerOpen, setAddDrawerOpen] = useState(false);
  const [addForm] = Form.useForm();

  useEffect(() => {
    if (!id) return;
    workflowApi.get(id).then((res: any) => {
      const wf = res.data;
      setWorkflow(wf);
      // 转换为 React Flow 格式
      const rfNodes: Node[] = (wf.nodes || []).map((n: any) => ({
        id: n.id,
        type: 'default',
        position: n.position || { x: 0, y: 0 },
        data: {
          label: (
            <div style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 10, color: '#999' }}>{n.type}</div>
              <div>{n.name}</div>
            </div>
          ),
          nodeData: n,
        },
        style: {
          background: nodeColors[n.type] || '#ddd',
          color: '#fff',
          border: 'none',
          borderRadius: 8,
          padding: '8px 16px',
          fontSize: 13,
          minWidth: 120,
        },
      }));
      const rfEdges: Edge[] = (wf.edges || []).map((e: any) => ({
        id: e.id,
        source: e.source,
        target: e.target,
        label: e.condition || undefined,
        animated: true,
        markerEnd: { type: MarkerType.ArrowClosed },
        style: { strokeWidth: 2 },
      }));
      setNodes(rfNodes);
      setEdges(rfEdges);
    }).finally(() => setLoading(false));
  }, [id]);

  const onConnect = useCallback((conn: Connection) => {
    const newEdge: Edge = {
      id: `e-${Date.now()}`,
      source: conn.source,
      target: conn.target,
      sourceHandle: conn.sourceHandle ?? null,
      targetHandle: conn.targetHandle ?? null,
      animated: true,
      markerEnd: { type: MarkerType.ArrowClosed },
      style: { strokeWidth: 2 },
    };
    setEdges((eds) => addEdge(newEdge, eds) as Edge[]);
  }, [setEdges]);

  const onNodeDoubleClick = useCallback((_: any, node: Node) => {
    const nd = (node.data as any).nodeData;
    setEditingNode(nd);
    setNodeType(nd.type);
    form.setFieldsValue({
      name: nd.name,
      type: nd.type,
      ...flattenConfig(nd.config, nd.type),
    });
    setDrawerOpen(true);
  }, [form]);

  const handleSaveNode = async () => {
    const values = await form.validateFields();
    const config = buildConfig(values, nodeType);
    const updatedNodeData = {
      ...editingNode,
      name: values.name,
      type: nodeType,
      config,
    };

    setNodes((nds) =>
      nds.map((n) => {
        if (n.id === editingNode.id) {
          return {
            ...n,
            data: {
              label: (
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: 10, color: '#999' }}>{nodeType}</div>
                  <div>{values.name}</div>
                </div>
              ),
              nodeData: updatedNodeData,
            },
            style: {
              ...n.style,
              background: nodeColors[nodeType] || '#ddd',
            },
          };
        }
        return n;
      })
    );
    setDrawerOpen(false);
  };

  const handleAddNode = async () => {
    const values = await addForm.validateFields();
    const nodeId = `node-${Date.now()}`;
    const newNode: Node = {
      id: nodeId,
      type: 'default',
      position: { x: 200 + Math.random() * 200, y: 100 + Math.random() * 200 },
      data: {
        label: (
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 10, color: '#999' }}>{values.type}</div>
            <div>{values.name}</div>
          </div>
        ),
        nodeData: {
          id: nodeId,
          type: values.type,
          name: values.name,
          config: {},
          position: { x: 200, y: 100 },
        },
      },
      style: {
        background: nodeColors[values.type] || '#ddd',
        color: '#fff',
        border: 'none',
        borderRadius: 8,
        padding: '8px 16px',
        fontSize: 13,
        minWidth: 120,
      },
    };
    setNodes((nds) => [...nds, newNode]);
    setAddDrawerOpen(false);
    addForm.resetFields();
  };

  const handleSave = async () => {
    if (!id) return;
    setSaving(true);
    try {
      const wfNodes = nodes.map((n: any) => ({
        ...(n.data as any).nodeData,
        position: n.position,
      }));
      const wfEdges = edges.map((e: Edge) => ({
        id: e.id,
        source: e.source,
        target: e.target,
        condition: e.label || '',
      }));
      await workflowApi.update(id, { nodes: wfNodes, edges: wfEdges });
      message.success('保存成功');
    } finally {
      setSaving(false);
    }
  };

  const handleExecute = async () => {
    if (!id) return;
    await handleSave();
    const res: any = await workflowApi.execute(id);
    const exec = res.data;
    if (exec) {
      message.success('执行完成');
      navigate(`/executions/${exec.id}`);
    }
  };

  if (loading) return <Spin style={{ display: 'block', margin: '100px auto' }} />;

  return (
    <div style={{ height: 'calc(100vh - 120px)' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/workflows')}>返回</Button>
          <Typography.Title level={5} style={{ margin: 0 }}>{workflow?.name}</Typography.Title>
        </Space>
        <Space>
          <Button icon={<PlusOutlined />} onClick={() => { addForm.resetFields(); setAddDrawerOpen(true); }}>
            添加节点
          </Button>
          <Button icon={<SaveOutlined />} onClick={handleSave} loading={saving}>
            保存
          </Button>
          <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleExecute}>
            保存并执行
          </Button>
        </Space>
      </div>

      <Card bodyStyle={{ padding: 0 }} style={{ height: 'calc(100% - 48px)' }}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeDoubleClick={onNodeDoubleClick}
          fitView
          deleteKeyCode="Delete"
        >
          <Background />
          <Controls />
          <MiniMap />
        </ReactFlow>
      </Card>

      {/* 添加节点抽屉 */}
      <Drawer title="添加节点" open={addDrawerOpen} onClose={() => setAddDrawerOpen(false)} width={360}>
        <Form form={addForm} layout="vertical">
          <Form.Item name="type" label="节点类型" rules={[{ required: true }]}>
            <Select options={nodeTypeOptions} />
          </Form.Item>
          <Form.Item name="name" label="节点名称" rules={[{ required: true }]}>
            <Input placeholder="输入节点名称" />
          </Form.Item>
        </Form>
        <Button type="primary" block onClick={handleAddNode}>添加</Button>
      </Drawer>

      {/* 编辑节点抽屉 */}
      <Drawer
        title="编辑节点"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        width={480}
        extra={<Button type="primary" onClick={handleSaveNode}>确定</Button>}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="节点名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="节点类型">
            <Select options={nodeTypeOptions} onChange={(v) => setNodeType(v)} />
          </Form.Item>

          {/* LLM 配置 */}
          {nodeType === 'llm' && (
            <>
              <Form.Item name="llm_base_url" label="Base URL"><Input placeholder="https://api.deepseek.com/v1" /></Form.Item>
              <Form.Item name="llm_api_key" label="API Key"><Input.Password /></Form.Item>
              <Form.Item name="llm_model" label="模型"><Input placeholder="deepseek-chat" /></Form.Item>
              <Form.Item name="llm_system_msg" label="System Prompt"><Input.TextArea rows={2} /></Form.Item>
              <Form.Item name="llm_prompt" label="Prompt"><Input.TextArea rows={4} placeholder="支持 {{input.xxx}} {{nodes.node_id.xxx}} 模板变量" /></Form.Item>
              <Form.Item name="llm_temperature" label="Temperature"><InputNumber min={0} max={2} step={0.1} /></Form.Item>
              <Form.Item name="llm_max_tokens" label="Max Tokens"><InputNumber min={1} max={128000} /></Form.Item>
              <Form.Item name="llm_output_schema" label="Output Schema (JSON)">
                <Input.TextArea rows={4} placeholder='{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}' />
              </Form.Item>
            </>
          )}

          {/* Agent 配置 */}
          {nodeType === 'agent' && (
            <>
              <Form.Item name="agent_base_url" label="Base URL"><Input placeholder="https://api.deepseek.com/v1" /></Form.Item>
              <Form.Item name="agent_api_key" label="API Key"><Input.Password /></Form.Item>
              <Form.Item name="agent_model" label="模型"><Input placeholder="deepseek-chat" /></Form.Item>
              <Form.Item name="agent_system_msg" label="System Prompt"><Input.TextArea rows={2} /></Form.Item>
              <Form.Item name="agent_prompt" label="Prompt"><Input.TextArea rows={4} placeholder="支持模板变量" /></Form.Item>
              <Form.Item name="agent_mcp_servers" label="MCP 服务器 URL（每行一个）">
                <Input.TextArea rows={3} placeholder="http://localhost:3001/sse" />
              </Form.Item>
              <Form.Item name="agent_max_iterations" label="最大迭代次数"><InputNumber min={1} max={50} /></Form.Item>
              <Form.Item name="agent_temperature" label="Temperature"><InputNumber min={0} max={2} step={0.1} /></Form.Item>
            </>
          )}

          {/* 条件配置 */}
          {nodeType === 'condition' && (
            <Form.Item name="condition_expression" label="条件表达式">
              <Input.TextArea rows={3} placeholder='nodes["node-1"]["status"] == 200' />
            </Form.Item>
          )}

          {/* 代码配置 */}
          {nodeType === 'code' && (
            <>
              <Form.Item name="code_language" label="语言">
                <Select options={[{ label: 'JavaScript', value: 'javascript' }]} />
              </Form.Item>
              <Form.Item name="code_code" label="代码">
                <Input.TextArea rows={8} placeholder="// 可使用 input, nodes 变量&#10;var result = input;&#10;result;" style={{ fontFamily: 'monospace' }} />
              </Form.Item>
            </>
          )}

          {/* HTTP 配置 */}
          {nodeType === 'http' && (
            <>
              <Form.Item name="http_method" label="方法">
                <Select options={['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map((m) => ({ label: m, value: m }))} />
              </Form.Item>
              <Form.Item name="http_url" label="URL"><Input placeholder="https://api.example.com/data" /></Form.Item>
              <Form.Item name="http_headers" label="Headers (JSON)"><Input.TextArea rows={2} placeholder='{"Content-Type":"application/json"}' /></Form.Item>
              <Form.Item name="http_body" label="Body"><Input.TextArea rows={4} placeholder="支持模板变量" /></Form.Item>
            </>
          )}

          {/* 邮件配置 */}
          {nodeType === 'email' && (
            <>
              <Form.Item name="email_smtp_host" label="SMTP Host"><Input placeholder="smtp.qq.com" /></Form.Item>
              <Form.Item name="email_smtp_port" label="SMTP Port"><InputNumber placeholder="587" /></Form.Item>
              <Form.Item name="email_username" label="用户名"><Input /></Form.Item>
              <Form.Item name="email_password" label="密码"><Input.Password /></Form.Item>
              <Form.Item name="email_from" label="发件人"><Input /></Form.Item>
              <Form.Item name="email_to" label="收件人"><Input placeholder="支持模板变量" /></Form.Item>
              <Form.Item name="email_subject" label="主题"><Input placeholder="支持模板变量" /></Form.Item>
              <Form.Item name="email_body" label="正文"><Input.TextArea rows={4} placeholder="支持模板变量" /></Form.Item>
              <Form.Item name="email_content_type" label="内容类型">
                <Select options={[{ label: 'text/plain', value: 'text/plain' }, { label: 'text/html', value: 'text/html' }]} />
              </Form.Item>
            </>
          )}
        </Form>
      </Drawer>
    </div>
  );
}

// 将嵌套 config 展平为 form 字段
function flattenConfig(config: any, type: string): Record<string, any> {
  if (!config) return {};
  const result: Record<string, any> = {};
  const cfg = config[type] || config.llm || config.agent || config.condition || config.code || config.http || config.email;
  if (!cfg) return {};

  switch (type) {
    case 'llm':
      result.llm_base_url = cfg.base_url;
      result.llm_api_key = cfg.api_key;
      result.llm_model = cfg.model;
      result.llm_prompt = cfg.prompt;
      result.llm_system_msg = cfg.system_msg;
      result.llm_temperature = cfg.temperature;
      result.llm_max_tokens = cfg.max_tokens;
      result.llm_output_schema = cfg.output_schema ? JSON.stringify(cfg.output_schema, null, 2) : '';
      break;
    case 'agent':
      result.agent_base_url = cfg.base_url;
      result.agent_api_key = cfg.api_key;
      result.agent_model = cfg.model;
      result.agent_prompt = cfg.prompt;
      result.agent_system_msg = cfg.system_msg;
      result.agent_mcp_servers = cfg.mcp_servers?.map((s: any) => s.url).join('\n') || '';
      result.agent_max_iterations = cfg.max_iterations;
      result.agent_temperature = cfg.temperature;
      break;
    case 'condition':
      result.condition_expression = cfg.expression;
      break;
    case 'code':
      result.code_language = cfg.language;
      result.code_code = cfg.code;
      break;
    case 'http':
      result.http_method = cfg.method;
      result.http_url = cfg.url;
      result.http_headers = cfg.headers ? JSON.stringify(cfg.headers) : '';
      result.http_body = cfg.body;
      break;
    case 'email':
      result.email_smtp_host = cfg.smtp_host;
      result.email_smtp_port = cfg.smtp_port;
      result.email_username = cfg.username;
      result.email_password = cfg.password;
      result.email_from = cfg.from;
      result.email_to = cfg.to;
      result.email_subject = cfg.subject;
      result.email_body = cfg.body;
      result.email_content_type = cfg.content_type;
      break;
  }
  return result;
}

// 从 form 字段构建嵌套 config
function buildConfig(values: any, type: string): any {
  switch (type) {
    case 'start':
    case 'end':
      return {};
    case 'llm': {
      let outputSchema;
      try { outputSchema = values.llm_output_schema ? JSON.parse(values.llm_output_schema) : undefined; } catch { outputSchema = undefined; }
      return {
        llm: {
          base_url: values.llm_base_url,
          api_key: values.llm_api_key,
          model: values.llm_model,
          prompt: values.llm_prompt,
          system_msg: values.llm_system_msg,
          temperature: values.llm_temperature,
          max_tokens: values.llm_max_tokens,
          output_schema: outputSchema,
        },
      };
    }
    case 'agent': {
      const urls = (values.agent_mcp_servers || '').split('\n').filter(Boolean);
      return {
        agent: {
          base_url: values.agent_base_url,
          api_key: values.agent_api_key,
          model: values.agent_model,
          prompt: values.agent_prompt,
          system_msg: values.agent_system_msg,
          mcp_servers: urls.map((url: string) => ({ url: url.trim() })),
          max_iterations: values.agent_max_iterations,
          temperature: values.agent_temperature,
        },
      };
    }
    case 'condition':
      return { condition: { expression: values.condition_expression } };
    case 'code':
      return { code: { language: values.code_language, code: values.code_code } };
    case 'http': {
      let headers;
      try { headers = values.http_headers ? JSON.parse(values.http_headers) : undefined; } catch { headers = undefined; }
      return {
        http: {
          method: values.http_method,
          url: values.http_url,
          headers,
          body: values.http_body,
        },
      };
    }
    case 'email':
      return {
        email: {
          smtp_host: values.email_smtp_host,
          smtp_port: values.email_smtp_port,
          username: values.email_username,
          password: values.email_password,
          from: values.email_from,
          to: values.email_to,
          subject: values.email_subject,
          body: values.email_body,
          content_type: values.email_content_type,
        },
      };
    default:
      return {};
  }
}
