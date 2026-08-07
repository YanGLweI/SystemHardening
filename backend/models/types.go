package models

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONMap 用于存储 JSON 数据的可序列化 map 类型
type JSONMap map[string]interface{}

// Value 实现 sql/driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// UnmarshalJSON 实现 json.Unmarshaler 接口（让 GORM 能正确处理）
func (j *JSONMap) UnmarshalJSON(data []byte) error {
	if j == nil {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*j = m
	return nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONMap) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]interface{}(j))
}
