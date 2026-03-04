package engine

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/2comjie/mcpflow/internal/model"
)

// RenderString 使用 text/template 渲染字符串
func RenderString(s string, data map[string]any) string {
	if s == "" {
		return s
	}
	funcMap := template.FuncMap{
		"json": func(v any) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	}
	t, err := template.New("").Funcs(funcMap).Option("missingkey=zero").Parse(s)
	if err != nil {
		return s
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return s
	}
	// missingkey=zero 对 interface{} 会渲染为 "<no value>"，替换为空字符串
	result := buf.String()
	result = strings.ReplaceAll(result, "<no value>", "")
	return result
}

// RenderMap 递归渲染 map 中所有 string 值
func RenderMap(m map[string]any, data map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case string:
			result[k] = RenderString(val, data)
		case map[string]any:
			result[k] = RenderMap(val, data)
		default:
			result[k] = v
		}
	}
	return result
}

// RenderNodeConfig 渲染节点配置中的模板变量
func RenderNodeConfig(node *model.Node, data map[string]any) model.NodeConfig {
	b, err := json.Marshal(node.Config)
	if err != nil {
		return node.Config
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return node.Config
	}

	rendered := RenderMap(m, data)

	b, err = json.Marshal(rendered)
	if err != nil {
		return node.Config
	}
	var cfg model.NodeConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return node.Config
	}
	return cfg
}
