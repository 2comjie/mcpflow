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
  // ==================== Basic ====================
  {
    id: 'simple-llm',
    name: 'Simple LLM Chat',
    description: 'LLM single-turn conversation: input a question, get an answer',
    category: 'basic',
    nodes: [
      {
        id: 'node_1', type: 'start', name: 'Start',
        config: {
          start: {
            input_defs: [
              { name: 'query', type: 'text', required: true, description: 'Your question' },
            ],
          },
        },
        position: { x: 300, y: 50 },
      },
      {
        id: 'node_2', type: 'llm', name: 'LLM Chat',
        config: {
          llm: {
            base_url: '', api_key: '', model: '',
            prompt: '{{input.query}}',
            system_msg: 'You are a helpful assistant.',
            temperature: 0.7, max_tokens: 1024,
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
    id: 'llm-translate',
    name: 'Translation Pipeline',
    description: 'LLM translates text, then Code node extracts a word count',
    category: 'basic',
    nodes: [
      {
        id: 'node_1', type: 'start', name: 'Start',
        config: {
          start: {
            input_defs: [
              { name: 'text', type: 'text', required: true, description: 'Text to translate' },
              { name: 'target_lang', type: 'string', required: false, description: 'Target language', default: 'English' },
            ],
          },
        },
        position: { x: 300, y: 50 },
      },
      {
        id: 'node_2', type: 'llm', name: 'Translate',
        config: {
          llm: {
            base_url: '', api_key: '', model: '',
            prompt: 'Translate the following text to {{input.target_lang}}:\n\n{{input.text}}',
            system_msg: 'You are a professional translator. Only output the translated text, nothing else.',
            temperature: 0.3, max_tokens: 2048,
          },
        },
        position: { x: 300, y: 200 },
      },
      {
        id: 'node_3', type: 'code', name: 'Word Count',
        config: {
          code: {
            language: 'javascript',
            code: 'var text = input.content || ""; var words = text.trim().split(/\\s+/).length; return { word_count: words, content: input.content };',
          },
        },
        position: { x: 300, y: 350 },
      },
      { id: 'node_4', type: 'end', name: 'End', config: {}, position: { x: 300, y: 500 } },
    ],
    edges: [
      { id: 'e1-2', source: 'node_1', target: 'node_2' },
      { id: 'e2-3', source: 'node_2', target: 'node_3' },
      { id: 'e3-4', source: 'node_3', target: 'node_4' },
    ],
  },
  {
    id: 'http-llm-summary',
    name: 'HTTP Fetch + LLM Summary',
    description: 'Fetch data from an HTTP API, then use LLM to summarize the result',
    category: 'basic',
    nodes: [
      {
        id: 'node_1', type: 'start', name: 'Start',
        config: {
          start: {
            input_defs: [
              { name: 'url', type: 'string', required: true, description: 'HTTP URL to fetch', default: 'https://httpbin.org/get' },
            ],
          },
        },
        position: { x: 300, y: 50 },
      },
      {
        id: 'node_2', type: 'http', name: 'HTTP Request',
        config: {
          http: {
            method: 'GET',
            url: '{{input.url}}',
            headers: {},
            body: '',
          },
        },
        position: { x: 300, y: 200 },
      },
      {
        id: 'node_3', type: 'llm', name: 'Summarize',
        config: {
          llm: {
            base_url: '', api_key: '', model: '',
            prompt: 'Summarize the following API response data in a concise paragraph:\n\n{{nodes.node_2}}',
            system_msg: 'You are a data analyst. Summarize the key information.',
            temperature: 0.5, max_tokens: 512,
          },
        },
        position: { x: 300, y: 350 },
      },
      { id: 'node_4', type: 'end', name: 'End', config: {}, position: { x: 300, y: 500 } },
    ],
    edges: [
      { id: 'e1-2', source: 'node_1', target: 'node_2' },
      { id: 'e2-3', source: 'node_2', target: 'node_3' },
      { id: 'e3-4', source: 'node_3', target: 'node_4' },
    ],
  },

  // ==================== Agent ====================
  {
    id: 'agent-mcp',
    name: 'Agent + MCP Tools',
    description: 'Agent automatically discovers and uses MCP tools to complete tasks',
    category: 'agent',
    nodes: [
      {
        id: 'node_1', type: 'start', name: 'Start',
        config: {
          start: {
            input_defs: [
              { name: 'query', type: 'text', required: true, description: 'Task for the agent' },
            ],
          },
        },
        position: { x: 300, y: 50 },
      },
      {
        id: 'node_2', type: 'agent', name: 'Agent',
        config: {
          agent: {
            base_url: '', api_key: '', model: '',
            prompt: '{{input.query}}',
            system_msg: 'You are an intelligent assistant. Use available tools to complete the task.',
            mcp_servers: [],
            max_iterations: 10, temperature: 0.3, max_tokens: 1024,
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
    id: 'multi-agent',
    name: 'Multi-Agent Collaboration',
    description: 'Agent 1 researches information, Agent 2 analyzes and summarizes',
    category: 'agent',
    nodes: [
      {
        id: 'node_1', type: 'start', name: 'Start',
        config: {
          start: {
            input_defs: [
              { name: 'topic', type: 'string', required: true, description: 'Research topic' },
            ],
          },
        },
        position: { x: 300, y: 50 },
      },
      {
        id: 'node_2', type: 'agent', name: 'Research Agent',
        config: {
          agent: {
            base_url: '', api_key: '', model: '',
            prompt: 'Research the following topic using available tools: {{input.topic}}',
            system_msg: 'You are a research assistant. Use tools to gather comprehensive information.',
            mcp_servers: [],
            max_iterations: 10, temperature: 0.3, max_tokens: 2048,
          },
        },
        position: { x: 300, y: 200 },
      },
      {
        id: 'node_3', type: 'agent', name: 'Analysis Agent',
        config: {
          agent: {
            base_url: '', api_key: '', model: '',
            prompt: 'Based on the research results below, write a professional analysis report:\n\n{{nodes.node_2.content}}',
            system_msg: 'You are a professional analyst. Organize raw information into a structured report.',
            mcp_servers: [],
            max_iterations: 3, temperature: 0.7, max_tokens: 2048,
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
    id: 'agent-deepwiki',
    name: 'Agent + DeepWiki (Public MCP)',
    description: 'Agent uses DeepWiki MCP to query GitHub repo documentation and answer questions',
    category: 'agent',
    nodes: [
      {
        id: 'node_1', type: 'start', name: 'Start',
        config: {
          start: {
            input_defs: [
              { name: 'repo', type: 'string', required: true, description: 'GitHub repo (owner/repo)', default: 'facebook/react' },
              { name: 'question', type: 'text', required: true, description: 'Question about the repo', default: 'What are the main features?' },
            ],
          },
        },
        position: { x: 300, y: 50 },
      },
      {
        id: 'node_2', type: 'agent', name: 'DeepWiki Agent',
        config: {
          agent: {
            base_url: '', api_key: '', model: '',
            prompt: 'Answer the following question about the GitHub repository {{input.repo}}:\n\n{{input.question}}',
            system_msg: 'You are a code research assistant. Use the available MCP tools to read documentation and answer questions about GitHub repositories. Always use tools to get accurate information.',
            mcp_servers: [{ url: 'https://mcp.deepwiki.com/mcp' }],
            max_iterations: 5, temperature: 0.3, max_tokens: 2048,
          },
        },
        position: { x: 300, y: 220 },
      },
      { id: 'node_3', type: 'end', name: 'End', config: {}, position: { x: 300, y: 390 } },
    ],
    edges: [
      { id: 'e1-2', source: 'node_1', target: 'node_2' },
      { id: 'e2-3', source: 'node_2', target: 'node_3' },
    ],
  },

  // ==================== Advanced ====================
  {
    id: 'conditional-branch',
    name: 'Conditional Branching',
    description: 'LLM scoring + Code extraction + Condition branching to different paths',
    category: 'advanced',
    nodes: [
      {
        id: 'node_1', type: 'start', name: 'Start',
        config: {
          start: {
            input_defs: [
              { name: 'content', type: 'text', required: true, description: 'Content to analyze' },
            ],
          },
        },
        position: { x: 300, y: 50 },
      },
      {
        id: 'node_2', type: 'llm', name: 'Analyze & Score',
        config: {
          llm: {
            base_url: '', api_key: '', model: '',
            prompt: 'Analyze the following content, return a JSON with a "score" field (0-100):\n\n{{input.content}}',
            system_msg: 'You are a scoring assistant. Return JSON format only, e.g. {"score": 85}',
            temperature: 0.3, max_tokens: 256,
          },
        },
        position: { x: 300, y: 180 },
      },
      {
        id: 'node_3', type: 'code', name: 'Parse Score',
        config: {
          code: {
            language: 'javascript',
            code: 'var text = input.content || ""; var m = text.match(/"score"\\s*:\\s*(\\d+)/); var score = m ? parseInt(m[1], 10) : 0; return { score: score };',
          },
        },
        position: { x: 300, y: 310 },
      },
      {
        id: 'node_4', type: 'condition', name: 'Score >= 80?',
        config: { condition: { expression: 'score >= 80' } },
        position: { x: 300, y: 440 },
      },
      {
        id: 'node_5', type: 'llm', name: 'High Score Path',
        config: {
          llm: {
            base_url: '', api_key: '', model: '',
            prompt: 'Generate a congratulation message. The score is {{nodes.node_3.score}}.',
            temperature: 0.7, max_tokens: 512,
          },
        },
        position: { x: 100, y: 580 },
      },
      {
        id: 'node_6', type: 'llm', name: 'Low Score Path',
        config: {
          llm: {
            base_url: '', api_key: '', model: '',
            prompt: 'Generate improvement suggestions. Current score is {{nodes.node_3.score}}.',
            temperature: 0.7, max_tokens: 512,
          },
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
    id: 'agent-email-report',
    name: 'Agent Research + Email Report',
    description: 'Agent researches a topic, then sends the report via email',
    category: 'advanced',
    nodes: [
      {
        id: 'node_1', type: 'start', name: 'Start',
        config: {
          start: {
            input_defs: [
              { name: 'topic', type: 'string', required: true, description: 'Research topic' },
              { name: 'email_to', type: 'string', required: true, description: 'Recipient email', default: '3217998214@qq.com' },
            ],
          },
        },
        position: { x: 300, y: 50 },
      },
      {
        id: 'node_2', type: 'agent', name: 'Research Agent',
        config: {
          agent: {
            base_url: '', api_key: '', model: '',
            prompt: 'Research the following topic and write a detailed report: {{input.topic}}',
            system_msg: 'You are a research assistant. Use available tools to gather information and generate a report.',
            mcp_servers: [],
            max_iterations: 10, temperature: 0.5, max_tokens: 2048,
          },
        },
        position: { x: 300, y: 200 },
      },
      {
        id: 'node_3', type: 'email', name: 'Send Report',
        config: {
          email: {
            smtp_host: 'smtp.qq.com',
            smtp_port: 465,
            username: '3217998214@qq.com',
            password: 'wbzcxscbfwgbdhac',
            from: '3217998214@qq.com',
            to: '{{input.email_to}}',
            subject: 'Research Report: {{input.topic}}',
            body: '<h2>Research Report: {{input.topic}}</h2><div>{{nodes.node_2.content}}</div>',
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
  {
    id: 'llm-code-http',
    name: 'LLM + Code + HTTP Pipeline',
    description: 'LLM generates data, Code transforms it, HTTP sends to external API',
    category: 'advanced',
    nodes: [
      {
        id: 'node_1', type: 'start', name: 'Start',
        config: {
          start: {
            input_defs: [
              { name: 'prompt', type: 'text', required: true, description: 'What to generate' },
              { name: 'webhook_url', type: 'string', required: false, description: 'Webhook URL to send result', default: 'https://httpbin.org/post' },
            ],
          },
        },
        position: { x: 300, y: 50 },
      },
      {
        id: 'node_2', type: 'llm', name: 'Generate',
        config: {
          llm: {
            base_url: '', api_key: '', model: '',
            prompt: '{{input.prompt}}\n\nReturn the result as JSON.',
            system_msg: 'You are a data generator. Always return valid JSON.',
            temperature: 0.5, max_tokens: 1024,
          },
        },
        position: { x: 300, y: 200 },
      },
      {
        id: 'node_3', type: 'code', name: 'Transform',
        config: {
          code: {
            language: 'javascript',
            code: 'var content = input.content || "{}"; try { var data = JSON.parse(content); return { payload: JSON.stringify({ source: "mcpflow", data: data }) }; } catch(e) { return { payload: JSON.stringify({ source: "mcpflow", raw: content }) }; }',
          },
        },
        position: { x: 300, y: 350 },
      },
      {
        id: 'node_4', type: 'http', name: 'Send to Webhook',
        config: {
          http: {
            method: 'POST',
            url: '{{input.webhook_url}}',
            headers: { 'Content-Type': 'application/json' },
            body: '{{nodes.node_3.payload}}',
          },
        },
        position: { x: 300, y: 500 },
      },
      { id: 'node_5', type: 'end', name: 'End', config: {}, position: { x: 300, y: 650 } },
    ],
    edges: [
      { id: 'e1-2', source: 'node_1', target: 'node_2' },
      { id: 'e2-3', source: 'node_2', target: 'node_3' },
      { id: 'e3-4', source: 'node_3', target: 'node_4' },
      { id: 'e4-5', source: 'node_4', target: 'node_5' },
    ],
  },
]

export const templateCategoryLabels: Record<string, string> = {
  all: 'All',
  basic: 'Basic',
  agent: 'Agent',
  advanced: 'Advanced',
}
