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
    description: 'Get global city weather info, supports current weather and forecast queries',
    url: 'http://localhost:3002/mcp',
    category: 'data',
    tags: ['weather', 'forecast'],
  },
  {
    name: 'Web Search',
    description: 'Internet search engine supporting web, news, image and more',
    url: 'https://mcp-search.example.com/mcp',
    category: 'search',
    tags: ['search', 'web'],
  },
  {
    name: 'GitHub',
    description: 'GitHub repository management: issues, PRs, code search and more',
    url: 'https://mcp-github.example.com/mcp',
    category: 'dev',
    tags: ['github', 'git', 'code'],
  },
  {
    name: 'Database Query',
    description: 'Database query tool supporting MySQL, PostgreSQL and other RDBMS',
    url: 'https://mcp-database.example.com/mcp',
    category: 'data',
    tags: ['database', 'sql', 'query'],
  },
  {
    name: 'File System',
    description: 'File system operations: read, write, and directory management',
    url: 'https://mcp-filesystem.example.com/mcp',
    category: 'tools',
    tags: ['file', 'filesystem'],
  },
  {
    name: 'Translation',
    description: 'Multi-language translation service supporting major languages',
    url: 'https://mcp-translate.example.com/mcp',
    category: 'tools',
    tags: ['translate', 'language'],
  },
  {
    name: 'Code Interpreter',
    description: 'Code execution sandbox supporting Python, JavaScript and more',
    url: 'https://mcp-code.example.com/mcp',
    category: 'dev',
    tags: ['code', 'sandbox', 'python'],
  },
  {
    name: 'Knowledge Base',
    description: 'Knowledge base retrieval with vector search and semantic matching',
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
