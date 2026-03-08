package config

import "gopkg.in/yaml.v3"

// UnmarshalYAML 解析 YAML 数据
func UnmarshalYAML(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
