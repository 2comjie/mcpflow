package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONMap 通用 JSON 字段类型，用于 GORM 存储 MySQL JSON 列
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
