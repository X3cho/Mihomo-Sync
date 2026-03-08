package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CreateSymlink 创建软链接：target -> source
func CreateSymlink(source, targetDir string) error {
	if targetDir == "" {
		return nil
	}

	target := filepath.Join(targetDir, "config.yaml")
	sourcePath, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败：%w", err)
	}

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("源文件不存在：%s", sourcePath)
	}

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	// 检查目标是否已存在
	if targetInfo, err := os.Lstat(target); err == nil {
		if targetInfo.Mode()&os.ModeSymlink != 0 {
			// 已是软链接
			currentTarget, err := os.Readlink(target)
			if err != nil {
				return fmt.Errorf("读取软链接失败：%w", err)
			}
			if currentTarget == sourcePath {
				// 软链接已存在，指向相同目标
				return nil
			}
			// 删除旧软链接
			if err := os.Remove(target); err != nil {
				return fmt.Errorf("删除旧软链接失败：%w", err)
			}
		} else {
			// 是普通文件，备份它
			backup := fmt.Sprintf("%s.bak.%s", target, time.Now().Format("20060102150405"))
			if err := os.Rename(target, backup); err != nil {
				return fmt.Errorf("备份文件失败：%w", err)
			}
		}
	}

	// 创建软链接
	if err := os.Symlink(sourcePath, target); err != nil {
		return fmt.Errorf("创建软链接失败：%w", err)
	}

	return nil
}

// SetupCrontab 设置定时任务
func SetupCrontab(alias string, interval int, scriptPath, configFile, baseDir, logFile string, enable bool) error {
	if !enable {
		// 删除定时任务
		cmd := exec.Command("sh", "-c",
			fmt.Sprintf("crontab -l 2>/dev/null | grep -v '%s.*-s %s ' | crontab -", scriptPath, alias))
		_ = cmd.Run()
		return nil
	}

	// 构建 cron 表达式：每 interval 小时执行一次
	cronExpr := fmt.Sprintf("0 */%d * * *", interval)

	// 构建命令
	script := fmt.Sprintf("%s -s %s --no-crontab -c %s", scriptPath, alias, configFile)
	cronLine := fmt.Sprintf("%s %s >> %s 2>&1", cronExpr, script, logFile)

	// 先删除旧的定时任务，再添加新的
	removeCmd := exec.Command("sh", "-c",
		fmt.Sprintf("crontab -l 2>/dev/null | grep -v '%s.*-s %s ' | crontab - || true", scriptPath, alias))
	_ = removeCmd.Run()

	// 添加新的定时任务
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("(crontab -l 2>/dev/null; echo '%s') | crontab -", cronLine))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("设置 crontab 失败：%w", err)
	}

	return nil
}

// RemoveCrontab 删除指定订阅的定时任务
func RemoveCrontab(alias string) error {
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("crontab -l 2>/dev/null | grep -v '%s' | crontab -", regexp.QuoteMeta(alias)))
	return cmd.Run()
}

// ExecuteCommand 执行重启命令
func ExecuteCommand(cmd string) (bool, error) {
	if cmd == "" {
		return false, nil
	}

	command := exec.Command("sh", "-c", cmd)
	output, err := command.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("执行命令失败：%w, 输出：%s", err, string(output))
	}

	return true, nil
}

// GetCrontab 获取当前 crontab 内容
func GetCrontab() (string, error) {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil // 没有 crontab
		}
		return "", fmt.Errorf("获取 crontab 失败：%w", err)
	}
	return string(output), nil
}

// ListSubscriptionsWithCrontab 列出订阅及其 crontab 状态
func ListSubscriptionsWithCrontab(subs []struct {
	Alias      string
	URL        string
	AutoUpdate bool
	Interval   int
}) (string, error) {
	crontab, err := GetCrontab()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for i, sub := range subs {
		status := "✗"
		if sub.AutoUpdate {
			if strings.Contains(crontab, sub.Alias) {
				status = "✓"
			}
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, sub.Alias))
		sb.WriteString(fmt.Sprintf("   地址：%s\n", sub.URL))
		sb.WriteString(fmt.Sprintf("   自动更新：%s (每%d小时)\n", status, sub.Interval))
	}
	return sb.String(), nil
}
