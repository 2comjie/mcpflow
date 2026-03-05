export interface MarketplaceMCPServer {
  name: string
  description: string
  url: string
  category: 'search' | 'tools' | 'data' | 'dev'
  tags: string[]
}

export const marketplaceCatalog: MarketplaceMCPServer[] = [
  {
    name: 'DeepWiki',
    description: 'AI-powered documentation for GitHub repositories. Read wiki structure, contents, and ask questions about any repo.',
    url: 'https://mcp.deepwiki.com/mcp',
    category: 'dev',
    tags: ['github', 'documentation', 'wiki', 'code'],
  },
  {
    name: '123elec Store',
    description: 'E-commerce MCP server: search products, check stock, get recommendations via AI model.',
    url: 'https://mcp.123elec.com/mcp',
    category: 'data',
    tags: ['ecommerce', 'products', 'search'],
  },
  {
    name: 'A2ABench Q&A',
    description: 'Developer Q&A platform: search questions, fetch threads, and get AI-grounded answers with citations.',
    url: 'https://a2abench-mcp.web.app/mcp',
    category: 'search',
    tags: ['qa', 'developer', 'search'],
  },
  {
    name: '1stDibs Marketplace',
    description: 'Browse and search luxury furniture, art, jewelry and more on 1stDibs marketplace.',
    url: 'https://www.1stdibs.com/soa/mcp/',
    category: 'data',
    tags: ['marketplace', 'luxury', 'search'],
  },
  {
    name: 'Weather Service',
    description: 'Get global city weather info, supports current weather and forecast queries.',
    url: 'http://localhost:3002/mcp',
    category: 'data',
    tags: ['weather', 'forecast'],
  },
]

export const categoryLabels: Record<string, string> = {
  all: 'All',
  search: 'Search',
  tools: 'Tools',
  data: 'Data',
  dev: 'Dev',
}
