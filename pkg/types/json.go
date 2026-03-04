package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONRaw 通用 JSON 字段类型，可存储任意 JSON 值（object、array、string 等）
type JSONRaw json.RawMessage

func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "null", nil
	}
	return string(j), nil
}

func (j *JSONRaw) Scan(value any) error {
	if value == nil {
		*j = JSONRaw("null")
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONRaw: %v", value)
	}
	*j = append((*j)[:0], bytes...)
	return nil
}

func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSONRaw) UnmarshalJSON(data []byte) error {
	*j = append((*j)[:0], data...)
	return nil
}

// MustJSONRaw 将任意值序列化为 JSONRaw，用于 GORM map 更新
func MustJSONRaw(v any) JSONRaw {
	b, _ := json.Marshal(v)
	return JSONRaw(b)
}

// JSONMap 通用 JSON 字段类型，用于 GORM 存储 MySQL JSON 列（仅限 object）
type JSONMap map[string]any

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONMap: %v", value)
	}
	return json.Unmarshal(bytes, j)
}
