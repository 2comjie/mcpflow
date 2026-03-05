#!/usr/bin/env python3
"""一键初始化 MCPFlow 默认数据

包含:
- 默认 MCP Server（本地测试服务器）
- 默认 LLM Provider（需替换为实际值）
- 多个工作流模板
"""

import json
import urllib.request

API_BASE = "http://localhost:8080/api/v1"

# ── LLM 配置（请替换为实际值）──────────────────────────
LLM_BASE_URL = "YOUR_LLM_BASE_URL"
LLM_API_KEY = "YOUR_LLM_API_KEY"
LLM_MODEL = "YOUR_MODEL"

# ── MCP Server 地址 ─────────────────────────────────
MCP_SERVER_URL = "http://localhost:3002/mcp"


def api_post(path, data):
    """POST 请求到 API"""
    body = json.dumps(data, ensure_ascii=False).encode("utf-8")
    req = urllib.request.Request(
        f"{API_BASE}{path}",
        data=body,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read().decode())


def create_mcp_server():
    """创建默认 MCP Server"""
    try:
        result = api_post("/mcp-servers", {
            "name": "Test MCP Server",
            "description": "本地测试 MCP 服务器，提供天气查询、计算、翻译等工具",
            "url": MCP_SERVER_URL,
        })
        print(f"  MCP Server created: #{result.get('id')}")
        return result.get("id")
    except Exception as e:
        print(f"  MCP Server creation failed: {e}")
        return None


def create_llm_provider():
    """创建默认 LLM Provider"""
    try:
        result = api_post("/llm-providers", {
            "name": "Default LLM",
            "base_url": LLM_BASE_URL,
            "api_key": LLM_API_KEY,
            "models": [LLM_MODEL],
        })
        print(f"  LLM Provider created: #{result.get('id')}")
        return result.get("id")
    except Exception as e:
        print(f"  LLM Provider creation failed: {e}")
        return None


# ── 工作流模板定义 ──────────────────────────────────

TEMPLATES = [
    {
        "name": "Simple LLM Chat",
        "description": "最基础的 LLM 对话工作流：接收输入 → LLM 处理 → 返回结果",
        "nodes": [
            {"id": "node_1", "type": "start", "name": "Start", "config": {}, "position": {"x": 300, "y": 50}},
            {
                "id": "node_2", "type": "llm", "name": "LLM Chat",
                "config": {"llm": {
                    "base_url": LLM_BASE_URL, "api_key": LLM_API_KEY, "model": LLM_MODEL,
                    "prompt": "{{.input.query}}", "system_msg": "You are a helpful assistant.",
                    "temperature": 0.7, "max_tokens": 1024,
                }},
                "position": {"x": 300, "y": 200},
            },
            {"id": "node_3", "type": "end", "name": "End", "config": {}, "position": {"x": 300, "y": 350}},
        ],
        "edges": [
            {"id": "e1-2", "source": "node_1", "target": "node_2"},
            {"id": "e2-3", "source": "node_2", "target": "node_3"},
        ],
    },
    {
        "name": "Agent + MCP Tools",
        "description": "Agent 自动发现并使用 MCP 工具完成任务，展示 MCP 协议核心能力",
        "nodes": [
            {"id": "node_1", "type": "start", "name": "Start", "config": {}, "position": {"x": 300, "y": 50}},
            {
                "id": "node_2", "type": "agent", "name": "Agent",
                "config": {"agent": {
                    "base_url": LLM_BASE_URL, "api_key": LLM_API_KEY, "model": LLM_MODEL,
                    "prompt": "{{.input.query}}",
                    "system_msg": "你是一个智能助手，请使用可用的工具来完成用户的任务。",
                    "mcp_servers": [{"url": MCP_SERVER_URL}],
                    "max_iterations": 10, "temperature": 0.3, "max_tokens": 1024,
                }},
                "position": {"x": 300, "y": 200},
            },
            {"id": "node_3", "type": "end", "name": "End", "config": {}, "position": {"x": 300, "y": 350}},
        ],
        "edges": [
            {"id": "e1-2", "source": "node_1", "target": "node_2"},
            {"id": "e2-3", "source": "node_2", "target": "node_3"},
        ],
    },
    {
        "name": "天气预警（多智能体协作）",
        "description": "Agent查询天气 → 提取温度 → 条件判断 → Agent生成报告/天气简报",
        "nodes": [
            {"id": "node_1", "type": "start", "name": "Start", "config": {}, "position": {"x": 250, "y": 50}},
            {
                "id": "node_2", "type": "agent", "name": "天气查询 Agent",
                "config": {"agent": {
                    "base_url": LLM_BASE_URL, "api_key": LLM_API_KEY, "model": LLM_MODEL,
                    "system_msg": "你是一个天气查询助手。请使用可用的工具查询天气信息，并以JSON格式返回结果。",
                    "prompt": "请查询 {{.city}} 的当前天气情况。",
                    "mcp_servers": [{"url": MCP_SERVER_URL}],
                    "max_iterations": 5, "temperature": 0.3, "max_tokens": 1024,
                }},
                "position": {"x": 250, "y": 200},
            },
            {
                "id": "node_3", "type": "code", "name": "提取温度",
                "config": {"code": {
                    "language": "javascript",
                    "code": 'var content = input.content || ""; var temp = 0; var m = content.match(/"temperature"\\s*[:：]\\s*(\\d+)/); if (m) { temp = parseInt(m[1], 10); } else { m = content.match(/温度\\s*[:：]?\\s*(\\d+)/); if (m) temp = parseInt(m[1], 10); } return {content: content, temp: temp};',
                }},
                "position": {"x": 250, "y": 350},
            },
            {
                "id": "node_4", "type": "condition", "name": "高温判断",
                "config": {"condition": {"expression": "temp > 35"}},
                "position": {"x": 250, "y": 500},
            },
            {
                "id": "node_5", "type": "agent", "name": "报告生成 Agent",
                "config": {"agent": {
                    "base_url": LLM_BASE_URL, "api_key": LLM_API_KEY, "model": LLM_MODEL,
                    "system_msg": "你是一个专业的气象分析师。",
                    "prompt": "根据以下天气信息，生成一份高温预警报告：\n\n{{.node_3.content}}",
                    "mcp_servers": [], "max_iterations": 3, "temperature": 0.7, "max_tokens": 1024,
                }},
                "position": {"x": 80, "y": 650},
            },
            {
                "id": "node_6", "type": "llm", "name": "生成天气简报",
                "config": {"llm": {
                    "base_url": LLM_BASE_URL, "api_key": LLM_API_KEY, "model": LLM_MODEL,
                    "prompt": "根据以下天气信息，生成一段简洁的天气播报：\n\n{{.node_3.content}}",
                    "system_msg": "你是一个天气播报主持人",
                    "temperature": 0.5, "max_tokens": 512,
                }},
                "position": {"x": 420, "y": 650},
            },
            {"id": "node_7", "type": "end", "name": "End", "config": {}, "position": {"x": 250, "y": 800}},
        ],
        "edges": [
            {"id": "e1-2", "source": "node_1", "target": "node_2"},
            {"id": "e2-3", "source": "node_2", "target": "node_3"},
            {"id": "e3-4", "source": "node_3", "target": "node_4"},
            {"id": "e4-5", "source": "node_4", "target": "node_5", "condition": "true"},
            {"id": "e4-6", "source": "node_4", "target": "node_6", "condition": "false"},
            {"id": "e5-7", "source": "node_5", "target": "node_7"},
            {"id": "e6-7", "source": "node_6", "target": "node_7"},
        ],
    },
    {
        "name": "Multi-Agent Research",
        "description": "多智能体协作：Research Agent 收集信息 → Analysis Agent 分析总结",
        "nodes": [
            {"id": "node_1", "type": "start", "name": "Start", "config": {}, "position": {"x": 300, "y": 50}},
            {
                "id": "node_2", "type": "agent", "name": "Research Agent",
                "config": {"agent": {
                    "base_url": LLM_BASE_URL, "api_key": LLM_API_KEY, "model": LLM_MODEL,
                    "prompt": "请使用可用工具，查询关于 {{.input.topic}} 的相关信息。",
                    "system_msg": "你是一个信息查询助手，擅长使用工具收集信息。",
                    "mcp_servers": [{"url": MCP_SERVER_URL}],
                    "max_iterations": 10, "temperature": 0.3, "max_tokens": 2048,
                }},
                "position": {"x": 300, "y": 200},
            },
            {
                "id": "node_3", "type": "agent", "name": "Analysis Agent",
                "config": {"agent": {
                    "base_url": LLM_BASE_URL, "api_key": LLM_API_KEY, "model": LLM_MODEL,
                    "prompt": "根据以下调研结果，撰写一份专业分析报告：\n\n{{.node_2.content}}",
                    "system_msg": "你是一个专业分析师，擅长整理信息为结构化报告。",
                    "mcp_servers": [], "max_iterations": 3, "temperature": 0.7, "max_tokens": 2048,
                }},
                "position": {"x": 300, "y": 380},
            },
            {"id": "node_4", "type": "end", "name": "End", "config": {}, "position": {"x": 300, "y": 530}},
        ],
        "edges": [
            {"id": "e1-2", "source": "node_1", "target": "node_2"},
            {"id": "e2-3", "source": "node_2", "target": "node_3"},
            {"id": "e3-4", "source": "node_3", "target": "node_4"},
        ],
    },
]


def main():
    print("=== MCPFlow Default Data Setup ===\n")

    print("[1/3] Creating MCP Server...")
    create_mcp_server()

    print("[2/3] Creating LLM Provider...")
    create_llm_provider()

    print(f"[3/3] Creating {len(TEMPLATES)} workflow templates...")
    for tmpl in TEMPLATES:
        try:
            result = api_post("/workflows", tmpl)
            wf_id = result.get("id")
            print(f"  Workflow '{tmpl['name']}' created: #{wf_id}")
        except Exception as e:
            print(f"  Workflow '{tmpl['name']}' failed: {e}")

    print("\nDone! Visit http://localhost:5173 to view your workflows.")
    print("Note: Please update LLM_BASE_URL, LLM_API_KEY, LLM_MODEL in this script before running workflows.")


if __name__ == "__main__":
    try:
        main()
    except urllib.error.URLError as e:
        print(f"Connection failed: {e.reason} (please ensure server is running at localhost:8080)")
