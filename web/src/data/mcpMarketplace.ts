export interface MarketplaceMCPServer {
  name: string
  description: string
  url: string
  category: 'search' | 'tools' | 'data' | 'dev'
  tags: string[]
}

export const marketplaceCatalog: MarketplaceMCPServer[] = [
  {
    name: 'Weather Service',
    description: '获取全球城市天气信息，支持当前天气和天气预报查询',
    url: 'http://localhost:3002/mcp',
    category: 'data',
    tags: ['weather', 'forecast'],
  },
  {
    name: 'Web Search',
    description: '互联网搜索引擎，支持网页、新闻、图片等多种搜索',
    url: 'https://mcp-search.example.com/mcp',
    category: 'search',
    tags: ['search', 'web'],
  },
  {
    name: 'GitHub',
    description: 'GitHub 仓库管理，支持 Issue、PR、代码搜索等操作',
    url: 'https://mcp-github.example.com/mcp',
    category: 'dev',
    tags: ['github', 'git', 'code'],
  },
  {
    name: 'Database Query',
    description: '数据库查询工具，支持 MySQL、PostgreSQL 等关系型数据库',
    url: 'https://mcp-database.example.com/mcp',
    category: 'data',
    tags: ['database', 'sql', 'query'],
  },
  {
    name: 'File System',
    description: '文件系统操作工具，支持文件读写、目录管理等',
    url: 'https://mcp-filesystem.example.com/mcp',
    category: 'tools',
    tags: ['file', 'filesystem'],
  },
  {
    name: 'Translation',
    description: '多语言翻译服务，支持中英日韩等主流语言互译',
    url: 'https://mcp-translate.example.com/mcp',
    category: 'tools',
    tags: ['translate', 'language'],
  },
  {
    name: 'Code Interpreter',
    description: '代码执行沙箱，支持 Python、JavaScript 等语言的安全执行',
    url: 'https://mcp-code.example.com/mcp',
    category: 'dev',
    tags: ['code', 'sandbox', 'python'],
  },
  {
    name: 'Knowledge Base',
    description: '知识库检索，支持文档向量搜索和语义匹配',
    url: 'https://mcp-knowledge.example.com/mcp',
    category: 'search',
    tags: ['rag', 'knowledge', 'embedding'],
  },
]

export const categoryLabels: Record<string, string> = {
  all: 'All',
  search: 'Search',
  tools: 'Tools',
  data: 'Data',
  dev: 'Dev',
}
