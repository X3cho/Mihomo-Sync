package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig 加载订阅配置文件
func LoadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err // 直接返回原始错误，让调用者判断
		}
		return nil, fmt.Errorf("读取配置文件失败：%w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败：%w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败：%w", err)
	}

	return &config, nil
}

// SaveConfig 保存订阅配置文件
func SaveConfig(config *Config, configFile string) error {
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败：%w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败：%w", err)
	}

	return nil
}

// CreateExampleConfig 创建示例配置文件
func CreateExampleConfig(configFile string) error {
	baseDir := filepath.Dir(configFile)
	config := DefaultConfig(baseDir)
	return SaveConfig(config, configFile)
}

// LoadMihomoConfig 加载 Mihomo 配置文件
func LoadMihomoConfig(templateFile string) (*MihomoConfig, error) {
	data, err := os.ReadFile(templateFile)
	if err != nil {
		return nil, fmt.Errorf("读取模板文件失败：%w", err)
	}

	var config MihomoConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析模板文件失败：%w", err)
	}

	return &config, nil
}

// SaveMihomoConfig 保存 Mihomo 配置文件
func SaveMihomoConfig(config *MihomoConfig, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败：%w", err)
	}
	defer file.Close()

	// 写入更新时间注释
	if config.UpdateTime != "" {
		if _, err := fmt.Fprintf(file, "# 更新时间：%s\n", config.UpdateTime); err != nil {
			return fmt.Errorf("写入注释失败：%w", err)
		}
	}

	// 按顺序编码 YAML（先其他配置，最后 proxies, proxy-groups, rules）
	if err := MarshalYAMLWithOrder(file, config); err != nil {
		return fmt.Errorf("编码 YAML 失败：%w", err)
	}

	return nil
}
