import type { Node, Edge } from '../api/workflow'

export interface WorkflowTemplate {
  id: string
  name: string
  description: string
  category: 'basic' | 'agent' | 'advanced'
  nodes: Node[]
  edges: Edge[]
}

export const workflowTemplates: WorkflowTemplate[] = [
  {
    id: 'simple-llm',
    name: 'Simple LLM Chat',
    description: '最基础的 LLM 调用：接收输入 → LLM 处理 → 返回结果',
    category: 'basic',
    nodes: [
      { id: 'node_1', type: 'start', name: 'Start', config: {}, position: { x: 300, y: 50 } },
      {
        id: 'node_2',
        type: 'llm',
        name: 'LLM Chat',
        config: {
          llm: {
            base_url: '',
            api_key: '',
            model: '',
            prompt: '{{.input.query}}',
            system_msg: 'You are a helpful assistant.',
            temperature: 0.7,
            max_tokens: 1024,
          },
        },
        position: { x: 300, y: 200 },
      },
      { id: 'node_3', type: 'end', name: 'End', config: {}, position: { x: 300, y: 350 } },
    ],
    edges: [
      { id: 'e1-2', source: 'node_1', target: 'node_2' },
      { id: 'e2-3', source: 'node_2', target: 'node_3' },
    ],
  },
  {
    id: 'agent-mcp',
    name: 'Agent + MCP Tools',
    description: 'Agent 自动发现并使用 MCP 工具完成任务',
    category: 'agent',
    nodes: [
      { id: 'node_1', type: 'start', name: 'Start', config: {}, position: { x: 300, y: 50 } },
      {
        id: 'node_2',
        type: 'agent',
        name: 'Agent',
        config: {
          agent: {
            base_url: '',
            api_key: '',
            model: '',
            prompt: '{{.input.query}}',
            system_msg: '你是一个智能助手，请使用可用的工具来完成用户的任务。',
            mcp_servers: [],
            max_iterations: 10,
            temperature: 0.3,
            max_tokens: 1024,
          },
        },
        position: { x: 300, y: 200 },
      },
      { id: 'node_3', type: 'end', name: 'End', config: {}, position: { x: 300, y: 350 } },
    ],
    edges: [
      { id: 'e1-2', source: 'node_1', target: 'node_2' },
      { id: 'e2-3', source: 'node_2', target: 'node_3' },
    ],
  },
  {
    id: 'conditional-branch',
    name: 'Conditional Branching',
    description: 'LLM 分析 → 代码提取 → 条件判断 → 不同分支处理',
    category: 'advanced',
    nodes: [
      { id: 'node_1', type: 'start', name: 'Start', config: {}, position: { x: 300, y: 50 } },
      {
        id: 'node_2',
        type: 'llm',
        name: 'Analyze',
        config: {
          llm: {
            base_url: '',
            api_key: '',
            model: '',
            prompt: '分析以下内容并返回 JSON 格式结果，包含 score 字段（0-100）：\n\n{{.input.content}}',
            system_msg: '你是一个评分助手，请返回 JSON 格式 {"score": 85}',
            temperature: 0.3,
            max_tokens: 256,
          },
        },
        position: { x: 300, y: 180 },
      },
      {
        id: 'node_3',
        type: 'code',
        name: 'Parse Score',
        config: {
          code: {
            language: 'javascript',
            code: 'var text = input.content || ""; var m = text.match(/"score"\\s*:\\s*(\\d+)/); var score = m ? parseInt(m[1], 10) : 0; return {score: score};',
          },
        },
        position: { x: 300, y: 310 },
      },
      {
        id: 'node_4',
        type: 'condition',
        name: 'Score Check',
        config: { condition: { expression: 'score >= 80' } },
        position: { x: 300, y: 440 },
      },
      {
        id: 'node_5',
        type: 'llm',
        name: 'High Score',
        config: {
          llm: { base_url: '', api_key: '', model: '', prompt: '生成一段祝贺消息，得分为 {{.node_3.score}} 分', temperature: 0.7, max_tokens: 512 },
        },
        position: { x: 100, y: 580 },
      },
      {
        id: 'node_6',
        type: 'llm',
        name: 'Low Score',
        config: {
          llm: { base_url: '', api_key: '', model: '', prompt: '生成改进建议，当前得分为 {{.node_3.score}} 分', temperature: 0.7, max_tokens: 512 },
        },
        position: { x: 500, y: 580 },
      },
      { id: 'node_7', type: 'end', name: 'End', config: {}, position: { x: 300, y: 720 } },
    ],
    edges: [
      { id: 'e1-2', source: 'node_1', target: 'node_2' },
      { id: 'e2-3', source: 'node_2', target: 'node_3' },
      { id: 'e3-4', source: 'node_3', target: 'node_4' },
      { id: 'e4-5', source: 'node_4', target: 'node_5', condition: 'true' },
      { id: 'e4-6', source: 'node_4', target: 'node_6', condition: 'false' },
      { id: 'e5-7', source: 'node_5', target: 'node_7' },
      { id: 'e6-7', source: 'node_6', target: 'node_7' },
    ],
  },
  {
    id: 'multi-agent',
    name: 'Multi-Agent Collaboration',
    description: '多智能体协作：Agent1 查询信息 → Agent2 分析总结',
    category: 'agent',
    nodes: [
      { id: 'node_1', type: 'start', name: 'Start', config: {}, position: { x: 300, y: 50 } },
      {
        id: 'node_2',
        type: 'agent',
        name: 'Research Agent',
        config: {
          agent: {
            base_url: '',
            api_key: '',
            model: '',
            prompt: '请使用可用的工具，查询关于 {{.input.topic}} 的相关信息。',
            system_msg: '你是一个信息查询助手，擅长使用工具收集信息。',
            mcp_servers: [],
            max_iterations: 10,
            temperature: 0.3,
            max_tokens: 2048,
          },
        },
        position: { x: 300, y: 200 },
      },
      {
        id: 'node_3',
        type: 'agent',
        name: 'Analysis Agent',
        config: {
          agent: {
            base_url: '',
            api_key: '',
            model: '',
            prompt: '根据以下调研结果，撰写一份专业的分析报告：\n\n{{.node_2.content}}',
            system_msg: '你是一个专业分析师，擅长将原始信息整理为结构化报告。',
            mcp_servers: [],
            max_iterations: 3,
            temperature: 0.7,
            max_tokens: 2048,
          },
        },
        position: { x: 300, y: 380 },
      },
      { id: 'node_4', type: 'end', name: 'End', config: {}, position: { x: 300, y: 530 } },
    ],
    edges: [
      { id: 'e1-2', source: 'node_1', target: 'node_2' },
      { id: 'e2-3', source: 'node_2', target: 'node_3' },
      { id: 'e3-4', source: 'node_3', target: 'node_4' },
    ],
  },
  {
    id: 'agent-email-report',
    name: 'Agent Research + Email',
    description: 'Agent 调研主题 → 生成报告 → 邮件发送',
    category: 'advanced',
    nodes: [
      { id: 'node_1', type: 'start', name: 'Start', config: {}, position: { x: 300, y: 50 } },
      {
        id: 'node_2',
        type: 'agent',
        name: 'Research Agent',
        config: {
          agent: {
            base_url: '',
            api_key: '',
            model: '',
            prompt: '研究以下主题并写一份详细报告：{{.input.topic}}',
            system_msg: '你是一个研究助手，请使用可用工具收集信息并生成报告。',
            mcp_servers: [],
            max_iterations: 10,
            temperature: 0.5,
            max_tokens: 2048,
          },
        },
        position: { x: 300, y: 200 },
      },
      {
        id: 'node_3',
        type: 'email',
        name: 'Send Report',
        config: {
          email: {
            smtp_host: 'smtp.qq.com',
            smtp_port: 465,
            username: '3217998214@qq.com',
            password: '',
            from: '3217998214@qq.com',
            to: '3217998214@qq.com',
            subject: '研究报告: {{.input.topic}}',
            body: '<h2>研究报告</h2><div>{{.node_2.content}}</div>',
            content_type: 'text/html',
          },
        },
        position: { x: 300, y: 380 },
      },
      { id: 'node_4', type: 'end', name: 'End', config: {}, position: { x: 300, y: 530 } },
    ],
    edges: [
      { id: 'e1-2', source: 'node_1', target: 'node_2' },
      { id: 'e2-3', source: 'node_2', target: 'node_3' },
      { id: 'e3-4', source: 'node_3', target: 'node_4' },
    ],
  },
]

export const templateCategoryLabels: Record<string, string> = {
  all: 'All',
  basic: 'Basic',
  agent: 'Agent',
  advanced: 'Advanced',
}
