// 测试用 MCP Server，实现 JSON-RPC 2.0 协议
// 包含示例 tools, prompts, resources
// 启动: go run ./cmd/testmcp
// 默认监听: http://localhost:3001/mcp
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// --- 示例 Tools ---

var sampleTools = []map[string]any{
	{
		"name":        "get_weather",
		"description": "获取指定城市的天气信息",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"city": map[string]any{
					"type":        "string",
					"description": "城市名称，例如 Beijing, Shanghai",
				},
			},
			"required": []string{"city"},
		},
	},
	{
		"name":        "calculate",
		"description": "执行简单数学计算",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "数学表达式，例如 2+3*4",
				},
			},
			"required": []string{"expression"},
		},
	},
	{
		"name":        "translate",
		"description": "将文本翻译为指定语言",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{
					"type":        "string",
					"description": "要翻译的文本",
				},
				"target_lang": map[string]any{
					"type":        "string",
					"description": "目标语言 (en, zh, ja, ko)",
				},
			},
			"required": []string{"text", "target_lang"},
		},
	},
}

// --- 示例 Prompts ---

var samplePrompts = []map[string]any{
	{
		"name":        "code_review",
		"description": "代码审查提示词，帮助审查代码质量和安全性",
		"arguments": []map[string]any{
			{"name": "code", "description": "要审查的代码", "required": true},
			{"name": "language", "description": "编程语言", "required": false},
		},
	},
	{
		"name":        "summarize",
		"description": "文本摘要提示词，将长文本压缩为要点",
		"arguments": []map[string]any{
			{"name": "text", "description": "要总结的文本", "required": true},
			{"name": "max_length", "description": "摘要最大长度", "required": false},
		},
	},
}

// --- 示例 Resources ---

var sampleResources = []map[string]any{
	{
		"uri":         "file:///config/app.yaml",
		"name":        "应用配置",
		"description": "应用程序的主要配置文件",
		"mimeType":    "application/yaml",
	},
	{
		"uri":         "file:///docs/api.md",
		"name":        "API 文档",
		"description": "REST API 接口文档",
		"mimeType":    "text/markdown",
	},
	{
		"uri":         "db://users/schema",
		"name":        "用户表结构",
		"description": "users 表的数据库 schema 定义",
		"mimeType":    "application/json",
	},
}

func handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req jsonRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, nil, -32700, "Parse error")
		return
	}

	log.Printf("[MCP] %s (id=%v)", req.Method, req.ID)

	var result any
	var rpcErr *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	switch req.Method {
	case "initialize":
		result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"prompts":   map[string]any{},
				"resources": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "test-mcp-server",
				"version": "1.0.0",
			},
		}

	case "tools/list":
		result = map[string]any{"tools": sampleTools}

	case "tools/call":
		result = handleToolCall(req.Params)

	case "prompts/list":
		result = map[string]any{"prompts": samplePrompts}

	case "prompts/get":
		result = handlePromptGet(req.Params)

	case "resources/list":
		result = map[string]any{"resources": sampleResources}

	case "resources/read":
		result = handleResourceRead(req.Params)

	default:
		rpcErr = &struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)}
	}

	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
		Error:   rpcErr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleToolCall(params json.RawMessage) map[string]any {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	json.Unmarshal(params, &p)

	switch p.Name {
	case "get_weather":
		city, _ := p.Arguments["city"].(string)
		if city == "" {
			return map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": "error: city parameter is required"},
				},
				"isError": true,
			}
		}
		weatherText := fetchWeather(city)
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": weatherText,
				},
			},
		}

	case "calculate":
		expr, _ := p.Arguments["expression"].(string)
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": fmt.Sprintf("计算 %s = 42 (示例结果)", expr),
				},
			},
		}

	case "translate":
		text, _ := p.Arguments["text"].(string)
		lang, _ := p.Arguments["target_lang"].(string)
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": fmt.Sprintf("[%s] %s (示例翻译结果)", strings.ToUpper(lang), text),
				},
			},
		}

	default:
		return map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": fmt.Sprintf("Unknown tool: %s", p.Name)},
			},
			"isError": true,
		}
	}
}

func handlePromptGet(params json.RawMessage) map[string]any {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	json.Unmarshal(params, &p)

	switch p.Name {
	case "code_review":
		code, _ := p.Arguments["code"].(string)
		lang, _ := p.Arguments["language"].(string)
		if lang == "" {
			lang = "unknown"
		}
		return map[string]any{
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": fmt.Sprintf("请审查以下 %s 代码，关注代码质量、安全性和最佳实践:\n\n```%s\n%s\n```", lang, lang, code),
					},
				},
			},
		}

	case "summarize":
		text, _ := p.Arguments["text"].(string)
		return map[string]any{
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": fmt.Sprintf("请将以下文本压缩为关键要点摘要:\n\n%s", text),
					},
				},
			},
		}

	default:
		return map[string]any{
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": fmt.Sprintf("Unknown prompt: %s", p.Name),
					},
				},
			},
		}
	}
}

func handleResourceRead(params json.RawMessage) map[string]any {
	var p struct {
		URI string `json:"uri"`
	}
	json.Unmarshal(params, &p)

	switch p.URI {
	case "file:///config/app.yaml":
		return map[string]any{
			"contents": []map[string]any{
				{
					"uri":      p.URI,
					"mimeType": "application/yaml",
					"text":     "server:\n  port: 8080\n  host: 0.0.0.0\ndatabase:\n  driver: mysql\n  dsn: root:password@tcp(127.0.0.1:3306)/mcpflow\n",
				},
			},
		}

	case "file:///docs/api.md":
		return map[string]any{
			"contents": []map[string]any{
				{
					"uri":      p.URI,
					"mimeType": "text/markdown",
					"text":     "# MCPFlow API\n\n## Workflows\n- GET /api/v1/workflows\n- POST /api/v1/workflows\n- GET /api/v1/workflows/:id\n",
				},
			},
		}

	case "db://users/schema":
		return map[string]any{
			"contents": []map[string]any{
				{
					"uri":      p.URI,
					"mimeType": "application/json",
					"text":     `{"table":"users","columns":[{"name":"id","type":"bigint","primary":true},{"name":"name","type":"varchar(255)"},{"name":"email","type":"varchar(255)","unique":true},{"name":"created_at","type":"datetime"}]}`,
				},
			},
		}

	default:
		return map[string]any{
			"contents": []map[string]any{
				{
					"uri":  p.URI,
					"text": fmt.Sprintf("Resource not found: %s", p.URI),
				},
			},
		}
	}
}

// fetchWeather 通过 wttr.in 获取真实天气数据
func fetchWeather(city string) string {
	apiURL := fmt.Sprintf("https://wttr.in/%s?format=j1&lang=zh", url.QueryEscape(city))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		log.Printf("[weather] fetch error: %v", err)
		return fmt.Sprintf("%s: 天气数据获取失败 (%v)", city, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("%s: 读取响应失败", city)
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		// wttr.in 可能返回纯文本错误
		return fmt.Sprintf("%s: %s", city, string(body))
	}

	// 解析 current_condition
	conditions, ok := data["current_condition"].([]any)
	if !ok || len(conditions) == 0 {
		return fmt.Sprintf("%s: 未找到天气数据", city)
	}
	cur, ok := conditions[0].(map[string]any)
	if !ok {
		return fmt.Sprintf("%s: 天气数据格式异常", city)
	}

	tempC, _ := cur["temp_C"].(string)
	humidity, _ := cur["humidity"].(string)
	windSpeed, _ := cur["windspeedKmph"].(string)
	pressure, _ := cur["pressure"].(string)
	visibility, _ := cur["visibility"].(string)

	// 中文天气描述
	desc := ""
	if descArr, ok := cur["lang_zh"].([]any); ok && len(descArr) > 0 {
		if d, ok := descArr[0].(map[string]any); ok {
			desc, _ = d["value"].(string)
		}
	}
	if desc == "" {
		if descArr, ok := cur["weatherDesc"].([]any); ok && len(descArr) > 0 {
			if d, ok := descArr[0].(map[string]any); ok {
				desc, _ = d["value"].(string)
			}
		}
	}

	return fmt.Sprintf(
		"%s 天气: %s, 温度: %s°C, 湿度: %s%%, 风速: %skm/h, 气压: %shPa, 能见度: %skm",
		city, desc, tempC, humidity, windSpeed, pressure, visibility,
	)
}

func writeError(w http.ResponseWriter, id any, code int, msg string) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{Code: code, Message: msg},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	addr := ":3002"

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", handleRPC)

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	log.Printf("Test MCP Server starting on %s", addr)
	log.Printf("Endpoint: http://localhost%s/mcp  (change port if conflicts)", addr)
	log.Printf("Tools: %d, Prompts: %d, Resources: %d",
		len(sampleTools), len(samplePrompts), len(sampleResources))

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
