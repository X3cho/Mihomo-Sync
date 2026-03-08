package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Subscription 订阅配置
type Subscription struct {
	Alias      string `yaml:"alias" json:"alias"`
	URL        string `yaml:"url" json:"url"`
	AutoUpdate bool   `yaml:"auto_update" json:"auto_update"`
	Interval   int    `yaml:"interval" json:"interval"`                                             // 小时
	Insecure   bool   `yaml:"insecure_skip_verify,omitempty" json:"insecure_skip_verify,omitempty"` // 忽略证书错误
}

// Config 主配置文件结构
type Config struct {
	OutputDir     string         `yaml:"output_dir" json:"output_dir"`
	Template      string         `yaml:"template" json:"template"`
	LogFile       string         `yaml:"log_file" json:"log_file"`
	LogLevel      string         `yaml:"log_level" json:"log_level"`
	MaxLogSize    string         `yaml:"max_log_size" json:"max_log_size"`
	Retry         int            `yaml:"retry,omitempty" json:"retry,omitempty"`     // 重试次数，默认 3 次
	Timeout       int            `yaml:"timeout,omitempty" json:"timeout,omitempty"` // 超时时间 (秒)，默认 30 秒
	RestartCmd    string         `yaml:"restart_command,omitempty" json:"restart_command,omitempty"`
	Subscriptions []Subscription `yaml:"subscriptions" json:"subscriptions"`
}

// MihomoConfig Mihomo 配置文件结构
type MihomoConfig struct {
	Proxies            []any             `yaml:"proxies" json:"proxies"`
	ProxyGroups        []any             `yaml:"proxy-groups" json:"proxy-groups"`
	Rules              []any             `yaml:"rules" json:"rules"`
	Port               *int              `yaml:"port,omitempty" json:"port,omitempty"`
	SocksPort          *int              `yaml:"socks-port,omitempty" json:"socks-port,omitempty"`
	AllowLan           *bool             `yaml:"allow-lan,omitempty" json:"allow-lan,omitempty"`
	Mode               string            `yaml:"mode,omitempty" json:"mode,omitempty"`
	LogLevel           string            `yaml:"log-level,omitempty" json:"log-level,omitempty"`
	ExternalController string            `yaml:"external-controller,omitempty" json:"external-controller,omitempty"`
	Secret             string            `yaml:"secret,omitempty" json:"secret,omitempty"`
	GeodataMode        *bool             `yaml:"geodata-mode,omitempty" json:"geodata-mode,omitempty"`
	GeoxURL            map[string]string `yaml:"geox-url,omitempty" json:"geox-url,omitempty"`
	DNS                any               `yaml:"dns,omitempty" json:"dns,omitempty"`
	UpdateTime         string            `yaml:"-" json:"-"`
}

// Validate 验证订阅配置是否有效
func (s *Subscription) Validate() error {
	// 验证别名：不能为空，只能包含字母、数字、下划线、中划线
	if s.Alias == "" {
		return fmt.Errorf("别名不能为空")
	}
	if len(s.Alias) > 32 {
		return fmt.Errorf("别名长度不能超过 32 个字符")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(s.Alias) {
		return fmt.Errorf("别名只能包含字母、数字、下划线、中划线")
	}

	// 验证 URL：不能为空，必须是有效的 URL
	if s.URL == "" {
		return fmt.Errorf("URL 不能为空")
	}
	if _, err := url.ParseRequestURI(s.URL); err != nil {
		return fmt.Errorf("URL 格式无效：%v", err)
	}

	// 验证更新间隔：必须是正数，范围 1-720 小时
	if s.Interval <= 0 {
		return fmt.Errorf("更新间隔必须是正数 (当前：%d)", s.Interval)
	}
	if s.Interval > 720 {
		return fmt.Errorf("更新间隔不能超过 720 小时 (30 天)")
	}

	return nil
}

// Validate 验证主配置是否有效
func (c *Config) Validate() error {
	// 验证日志等级
	validLevels := map[string]bool{
		"DEBUG": true, "debug": true,
		"INFO": true, "info": true,
		"WARN": true, "warning": true,
		"ERROR": true, "error": true,
		"CRITICAL": true, "critical": true,
	}
	if c.LogLevel != "" && !validLevels[c.LogLevel] {
		return fmt.Errorf("无效的日志等级：%s (有效值：DEBUG, INFO, WARN, ERROR, CRITICAL)", c.LogLevel)
	}

	// 验证日志大小格式
	if c.MaxLogSize != "" {
		if _, err := ParseSize(c.MaxLogSize); err != nil {
			return fmt.Errorf("无效的日志大小格式：%s (示例：10MB, 1GB)", c.MaxLogSize)
		}
	}

	// 验证所有订阅
	if len(c.Subscriptions) == 0 {
		return fmt.Errorf("至少需要一个订阅")
	}

	aliases := make(map[string]bool)
	for i, sub := range c.Subscriptions {
		if err := sub.Validate(); err != nil {
			return fmt.Errorf("第 %d 个订阅配置无效：%w", i+1, err)
		}
		// 检查别名重复
		if aliases[sub.Alias] {
			return fmt.Errorf("订阅别名重复：%s", sub.Alias)
		}
		aliases[sub.Alias] = true
	}

	return nil
}

// ParseSize 解析带单位的大小字符串
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)

	// 纯数字，直接返回
	if regexp.MustCompile(`^\d+$`).MatchString(sizeStr) {
		return strconv.ParseInt(sizeStr, 10, 64)
	}

	// 匹配数字和单位
	re := regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)(B|KB|MB|GB)$`)
	match := re.FindStringSubmatch(sizeStr)
	if len(match) == 0 {
		return 0, fmt.Errorf("无效的大小格式：%s", sizeStr)
	}

	num, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}

	unit := strings.ToUpper(match[2])
	multiplier := int64(1)
	switch unit {
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	}

	return int64(num * float64(multiplier)), nil
}

// DefaultConfig 创建默认配置
func DefaultConfig(baseDir string) *Config {
	return &Config{
		OutputDir:  baseDir,
		Template:   filepath.Join(baseDir, "template.yaml"),
		LogFile:    filepath.Join(baseDir, "update.log"),
		LogLevel:   "INFO",
		MaxLogSize: "10MB",
		Subscriptions: []Subscription{
			{
				Alias:      "main",
				URL:        "https://example.com/sub?token=xxx",
				AutoUpdate: true,
				Interval:   6,
			},
		},
	}
}

// GetIntervalDuration 获取更新间隔的 duration
func (s *Subscription) GetIntervalDuration() time.Duration {
	if s.Interval <= 0 {
		s.Interval = 6
	}
	return time.Duration(s.Interval) * time.Hour
}
