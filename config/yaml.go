package config

import (
	"io"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML 解析 YAML 数据
func UnmarshalYAML(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

// MarshalYAMLWithOrder 按指定顺序编码 YAML
// 顺序：port, socks-port, allow-lan, mode, log-level, external-controller, secret,
//      geodata-mode, geox-url, dns, proxies, proxy-groups, rules
func MarshalYAMLWithOrder(w io.Writer, config *MihomoConfig) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()

	// 创建一个有序 map 来存储配置
	type OrderedConfig struct {
		Port              *int              `yaml:"port,omitempty"`
		SocksPort         *int              `yaml:"socks-port,omitempty"`
		AllowLan          *bool             `yaml:"allow-lan,omitempty"`
		Mode              string            `yaml:"mode,omitempty"`
		LogLevel          string            `yaml:"log-level,omitempty"`
		ExternalController string           `yaml:"external-controller,omitempty"`
		Secret            string            `yaml:"secret,omitempty"`
		GeodataMode       *bool             `yaml:"geodata-mode,omitempty"`
		GeoxURL           map[string]string `yaml:"geox-url,omitempty"`
		DNS               any               `yaml:"dns,omitempty"`
		Proxies           []any             `yaml:"proxies,omitempty"`
		ProxyGroups       []any             `yaml:"proxy-groups,omitempty"`
		Rules             []any             `yaml:"rules,omitempty"`
	}

	ordered := OrderedConfig{
		Port:              config.Port,
		SocksPort:         config.SocksPort,
		AllowLan:          config.AllowLan,
		Mode:              config.Mode,
		LogLevel:          config.LogLevel,
		ExternalController: config.ExternalController,
		Secret:            config.Secret,
		GeodataMode:       config.GeodataMode,
		GeoxURL:           config.GeoxURL,
		DNS:               config.DNS,
		Proxies:           config.Proxies,
		ProxyGroups:       config.ProxyGroups,
		Rules:             config.Rules,
	}

	return encoder.Encode(ordered)
}
