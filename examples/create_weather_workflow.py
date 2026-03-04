#!/usr/bin/env python3
"""一键创建「天气预警与邮件通知」测试工作流"""

import json
import urllib.request

API_BASE = "http://localhost:8080/api/v1"

workflow = {
    "name": "天气预警与邮件通知",
    "description": "获取城市天气 → 提取信息 → 条件分支(高温/正常) → LLM生成报告 → 邮件发送预警",
    "nodes": [
        {
            "id": "node_1",
            "type": "start",
            "name": "Start",
            "config": {},
            "position": {"x": 250, "y": 50},
        },
        {
            "id": "node_2",
            "type": "mcp",
            "name": "获取天气",
            "config": {
                "mcp": {
                    "action": "call_tool",
                    "server_url": "http://localhost:3002/mcp",
                    "tool_name": "get_weather",
                    "arguments": {"city": "{{.city}}"},
                }
            },
            "position": {"x": 250, "y": 180},
        },
        {
            "id": "node_3",
            "type": "code",
            "name": "提取天气文本",
            "config": {
                "code": {
                    "language": "javascript",
                    "code": 'var text = ""; if (input.content) { for (var i = 0; i < input.content.length; i++) { if (input.content[i].text) text += input.content[i].text; } } var temp = 0; var m = text.match(/温度: (\\d+)/); if (m) { temp = parseInt(m[1], 10); } return {weather_text: text, temp: temp};',
                }
            },
            "position": {"x": 250, "y": 310},
        },
        {
            "id": "node_4",
            "type": "condition",
            "name": "温度判断",
            "config": {"condition": {"expression": "temp > 35"}},
            "position": {"x": 250, "y": 440},
        },
        {
            "id": "node_5",
            "type": "llm",
            "name": "生成高温预警报告",
            "config": {
                "llm": {
                    "base_url": "YOUR_LLM_BASE_URL",
                    "api_key": "YOUR_LLM_API_KEY",
                    "model": "YOUR_MODEL",
                    "prompt": "根据以下天气信息生成一份高温预警报告，要求包含预警等级、防护建议：\n\n{{.node_3.weather_text}}",
                    "system_msg": "你是一个专业的气象分析师",
                    "temperature": 0.7,
                    "max_tokens": 1024,
                }
            },
            "position": {"x": 100, "y": 570},
        },
        {
            "id": "node_6",
            "type": "llm",
            "name": "生成天气简报",
            "config": {
                "llm": {
                    "base_url": "YOUR_LLM_BASE_URL",
                    "api_key": "YOUR_LLM_API_KEY",
                    "model": "YOUR_MODEL",
                    "prompt": "根据以下天气信息，生成一段简洁的天气播报：\n\n{{.node_3.weather_text}}",
                    "system_msg": "你是一个天气播报主持人",
                    "temperature": 0.5,
                    "max_tokens": 512,
                }
            },
            "position": {"x": 400, "y": 570},
        },
        {
            "id": "node_7",
            "type": "email",
            "name": "发送预警邮件",
            "config": {
                "email": {
                    "smtp_host": "smtp.qq.com",
                    "smtp_port": 465,
                    "username": "3217998214@qq.com",
                    "password": "dyotgyfxoffudddc",
                    "from": "3217998214@qq.com",
                    "to": "3217998214@qq.com",
                    "subject": "高温预警报告 - {{.city}}",
                    "body": "<h2>高温预警报告</h2><p>城市: {{.city}}</p><hr/><div>{{.node_5.content}}</div>",
                    "content_type": "text/html",
                }
            },
            "position": {"x": 100, "y": 700},
        },
        {
            "id": "node_8",
            "type": "end",
            "name": "End",
            "config": {},
            "position": {"x": 250, "y": 830},
        },
    ],
    "edges": [
        {"id": "e1-2", "source": "node_1", "target": "node_2"},
        {"id": "e2-3", "source": "node_2", "target": "node_3"},
        {"id": "e3-4", "source": "node_3", "target": "node_4"},
        {"id": "e4-5", "source": "node_4", "target": "node_5", "condition": "true"},
        {"id": "e4-6", "source": "node_4", "target": "node_6", "condition": "false"},
        {"id": "e5-7", "source": "node_5", "target": "node_7"},
        {"id": "e6-8", "source": "node_6", "target": "node_8"},
        {"id": "e7-8", "source": "node_7", "target": "node_8"},
    ],
}

data = json.dumps(workflow, ensure_ascii=False).encode("utf-8")
req = urllib.request.Request(
    f"{API_BASE}/workflows",
    data=data,
    headers={"Content-Type": "application/json"},
    method="POST",
)

try:
    with urllib.request.urlopen(req) as resp:
        result = json.loads(resp.read().decode())
        print(f"工作流创建成功! ID: {result.get('id')}")
        print(f"访问: http://localhost:5173/workflows/{result.get('id')}")
except urllib.error.HTTPError as e:
    print(f"创建失败: {e.code} {e.read().decode()}")
except urllib.error.URLError as e:
    print(f"连接失败: {e.reason} (请确认服务已启动在 localhost:8080)")
