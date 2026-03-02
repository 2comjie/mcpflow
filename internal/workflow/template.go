package workflow

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// 匹配 {{xxx.yyy.zzz}} 格式的变量引用
var templatePattern = regexp.MustCompile(`\{\{([^}]+)}}`)

// 渲染模板字符串 把 {{ nodeId.field }} 替换成实际的值
func RenderTemplate(tmpl string, nodeOutputs map[string]map[string]any, input map[string]any) string {
	return templatePattern.ReplaceAllStringFunc(tmpl, func(match string) string {
		// 去掉 {{ 和 }}
		path := strings.TrimSpace(match[2 : len(match)-2])
		val, err := resolvePath(path, nodeOutputs, input)
		if err != nil {
			return match // 解析失败保留原文
		}
		return fmt.Sprintf("%v", val)
	})
}

// 渲染 map 中所有 string 类型的值
func RenderMap(m map[string]any, nodeOutputs map[string]map[string]any, input map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case string:
			result[k] = RenderTemplate(val, nodeOutputs, input)
		case map[string]any:
			result[k] = RenderMap(val, nodeOutputs, input)
		default:
			result[k] = v
		}
	}
	return result
}

// 解析路径，如 "n1.result" 或 "input.query"
func resolvePath(path string, nodeOutputs map[string]map[string]any, input map[string]any) (any, error) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path %s", path)
	}

	source := parts[0] // nodeId 或 "input"
	field := parts[1]  // 字段路径

	var data map[string]any
	if source == "input" {
		data = input
	} else {
		var ok bool
		data, ok = nodeOutputs[source]
		if !ok {
			return nil, fmt.Errorf("node output not found %s", source)
		}
	}

	return getNestedValue(data, field)
}

// renderNodeConfig 渲染节点配置中的模板变量
// 将 Config 序列化为 JSON → map → RenderMap 渲染 → 反序列化回 NodeConfig
func renderNodeConfig(node *Node, nodeOutputs map[string]map[string]any, input map[string]any) *Node {
	b, err := json.Marshal(node.Config)
	if err != nil {
		return node
	}
	var configMap map[string]any
	if err := json.Unmarshal(b, &configMap); err != nil {
		return node
	}
	rendered := RenderMap(configMap, nodeOutputs, input)
	rb, err := json.Marshal(rendered)
	if err != nil {
		return node
	}
	var newConfig NodeConfig
	if err := json.Unmarshal(rb, &newConfig); err != nil {
		return node
	}
	// 返回新节点副本，避免修改原始数据
	newNode := *node
	newNode.Config = newConfig
	return &newNode
}

func getNestedValue(data map[string]any, path string) (any, error) {
	parts := strings.Split(path, ".")
	var current any = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("field not found %s", part)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot access field %s on %T", part, current)
		}
	}

	return current, nil
}
