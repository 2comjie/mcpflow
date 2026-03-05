// MongoDB 初始化脚本：清空旧数据 + 插入测试数据
// 用法: mongosh mongodb://127.0.0.1:27017/mcpflow deploy/mongo-init.js

db = db.getSiblingDB("mcpflow");

// 清空所有集合
db.workflows.drop();
db.executions.drop();
db.execution_logs.drop();
db.mcp_servers.drop();
db.llm_providers.drop();
db.counters.drop();

print("--- 已清空所有集合 ---");

// 初始化计数器（workflow 自增 ID 从 0 开始，插入时 +1）
db.counters.insertOne({ _id: "workflows", seq: NumberLong(0) });

// 插入 LLM Provider 测试数据
db.llm_providers.insertMany([
  {
    name: "OpenAI",
    base_url: "https://api.openai.com/v1",
    api_key: "sk-xxx",
    models: ["gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"],
    created_at: new Date(),
    updated_at: new Date(),
  },
  {
    name: "DeepSeek",
    base_url: "https://api.deepseek.com",
    api_key: "sk-853775c33f6e4570817830bc86e62151",
    models: ["deepseek-chat", "deepseek-reasoner"],
    created_at: new Date(),
    updated_at: new Date(),
  },
]);

print("--- 已插入 LLM Provider 测试数据 ---");

// 插入 MCP Server 测试数据
db.mcp_servers.insertMany([
  {
    name: "Weather MCP",
    url: "https://mcp.example.com/weather/sse",
    headers: {},
    status: "inactive",
    tools: [],
    created_at: new Date(),
    updated_at: new Date(),
  },
  {
    name: "Search MCP",
    url: "https://mcp.example.com/search/sse",
    headers: {},
    status: "inactive",
    tools: [],
    created_at: new Date(),
    updated_at: new Date(),
  },
]);

print("--- 已插入 MCP Server 测试数据 ---");

// 插入示例工作流（使用自增 ID）
var wfCounter = db.counters.findOneAndUpdate(
  { _id: "workflows" },
  { $inc: { seq: NumberLong(1) } },
  { returnDocument: "after" }
);

db.workflows.insertOne({
  _id: NumberLong(wfCounter.seq),
  name: "Hello World 工作流",
  description: "一个简单的 LLM 调用示例工作流",
  nodes: [
    {
      id: "start_1",
      type: "start",
      name: "开始",
      config: {
        start: {
          input_defs: [
            {
              name: "topic",
              type: "string",
              required: true,
              description: "对话主题",
              default: "人工智能",
            },
          ],
        },
      },
      position: { x: 250, y: 50 },
    },
    {
      id: "llm_1",
      type: "llm",
      name: "LLM 生成",
      config: {
        llm: {
          base_url: "https://api.deepseek.com",
          api_key: "sk-853775c33f6e4570817830bc86e62151",
          model: "deepseek-chat",
          prompt: "请用简短的语言介绍一下 {{input.topic}}",
          system_msg: "你是一个知识渊博的助手",
          temperature: 0.7,
          max_tokens: 1024,
        },
      },
      position: { x: 250, y: 200 },
    },
    {
      id: "end_1",
      type: "end",
      name: "结束",
      config: {},
      position: { x: 250, y: 350 },
    },
  ],
  edges: [
    { id: "e1", source: "start_1", target: "llm_1" },
    { id: "e2", source: "llm_1", target: "end_1" },
  ],
  created_at: new Date(),
  updated_at: new Date(),
});

print("--- 已插入示例工作流 ---");
print("=== MongoDB 初始化完成 ===");
