package handler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"mihomo-sync/config"
	"mihomo-sync/util"
)

// Logger 类型别名
type Logger = *util.Logger

// downloadConfig 从远程 URL 下载配置
func downloadConfig(ctx context.Context, url string, insecureSkipVerify bool) (string, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败：%w", err)
	}

	// 设置 User-Agent
	req.Header.Set("User-Agent", "clashmeta")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载配置失败：%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP 错误：%s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败：%w", err)
	}

	return string(body), nil
}

// downloadConfigWithRetry 带重试的下载配置
func downloadConfigWithRetry(ctx context.Context, url string, insecureSkipVerify bool, retry int, timeout int, logger Logger) (string, error) {
	var lastErr error
	_ = timeout // timeout 通过 context 传递

	for i := 0; i <= retry; i++ {
		if i > 0 {
			logger.Debug("重试下载 (%d/%d): %s", i, retry, url)
			// 重试前等待 5 秒
			select {
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		text, err := downloadConfig(ctx, url, insecureSkipVerify)
		if err == nil {
			return text, nil
		}
		lastErr = err

		// 如果是上下文取消或超时，且还有重试机会，继续重试
		if err == context.DeadlineExceeded || err == context.Canceled {
			logger.Debug("下载超时或取消，准备重试...")
		}
	}

	return "", lastErr
}

// UpdateSubscription 更新订阅配置
func UpdateSubscription(sub *config.Subscription, templateFile, outputPath string, logger Logger, cfg *config.Config) bool {
	alias := sub.Alias
	logger.Info("开始更新：%s", alias)
	logger.Debug("订阅 URL: %s", sub.URL)
	logger.Debug("模板文件：%s", templateFile)
	logger.Debug("输出路径：%s", outputPath)

	// 获取重试次数和超时时间（默认值）
	retry := cfg.Retry
	if retry <= 0 {
		retry = 3
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	// 下载远程配置（带重试）
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	remoteText, err := downloadConfigWithRetry(ctx, sub.URL, sub.Insecure, retry, timeout, logger)
	if err != nil {
		logger.Error("更新失败：%s - 下载配置失败：%v", alias, err)
		return false
	}
	logger.Debug("下载配置成功，大小：%d 字节", len(remoteText))

	// 解析远程配置
	var remoteConfig config.MihomoConfig
	if err := config.UnmarshalYAML([]byte(remoteText), &remoteConfig); err != nil {
		logger.Error("更新失败：%s - 解析远程配置失败：%v", alias, err)
		return false
	}

	// 读取本地模板
	localConfig, err := config.LoadMihomoConfig(templateFile)
	if err != nil {
		logger.Error("更新失败：%s - 读取模板失败：%v", alias, err)
		return false
	}
	logger.Debug("读取模板成功")

	// 更新指定字段
	updatedCount := 0
	if len(remoteConfig.Proxies) > 0 {
		localConfig.Proxies = remoteConfig.Proxies
		logger.Debug("更新 proxies: %d 项", len(remoteConfig.Proxies))
		updatedCount++
	}
	if len(remoteConfig.ProxyGroups) > 0 {
		localConfig.ProxyGroups = remoteConfig.ProxyGroups
		logger.Debug("更新 proxy-groups: %d 项", len(remoteConfig.ProxyGroups))
		updatedCount++
	}
	if len(remoteConfig.Rules) > 0 {
		localConfig.Rules = remoteConfig.Rules
		logger.Debug("更新 rules: %d 项", len(remoteConfig.Rules))
		updatedCount++
	}

	if updatedCount == 0 {
		logger.Warning("未找到可更新的字段 (proxies, proxy-groups, rules)")
	}

	// 设置更新时间
	localConfig.UpdateTime = time.Now().Format("2006-01-02 15:04:05")

	// 保存配置
	if err := config.SaveMihomoConfig(localConfig, outputPath); err != nil {
		logger.Error("更新失败：%s - 保存配置失败：%v", alias, err)
		return false
	}

	logger.Info("更新成功：%s", alias)
	logger.Debug("配置文件已保存：%s", outputPath)

	return true
}

// UpdateAllSubscriptions 更新所有订阅
func UpdateAllSubscriptions(cfg *config.Config, logger Logger) {
	subs := cfg.Subscriptions
	if len(subs) == 0 {
		logger.Info("没有订阅需要更新")
		return
	}

	logger.Debug("输出目录：%s", cfg.OutputDir)
	for _, sub := range subs {
		alias := sub.Alias
		outputPath := filepath.Join(cfg.OutputDir, alias+".yaml")
		UpdateSubscription(&sub, cfg.Template, outputPath, logger, cfg)
	}
}
