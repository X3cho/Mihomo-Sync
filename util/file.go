package util

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ParseSize 解析带单位的大小字符串
// 支持：B, KB, MB, GB（不区分大小写）
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)

	if num, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return num, nil
	}

	// 匹配数字和单位（不区分大小写）
	re := regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)(B|KB|MB|GB)?$`)
	match := re.FindStringSubmatch(sizeStr)
	if len(match) == 0 {
		return 0, fmt.Errorf("无效的大小格式：%s", sizeStr)
	}

	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, fmt.Errorf("解析数字失败：%w", err)
	}

	unit := strings.ToUpper(match[2])
	if unit == "" {
		unit = "B"
	}

	var multiplier int64
	switch unit {
	case "B":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	default:
		multiplier = 1
	}

	return int64(value * float64(multiplier)), nil
}

// RotateLog 轮转日志文件
// 当日志文件超过 maxSize 时，保留后一半大小的内容
func RotateLog(logFile string, maxSize int64) error {
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return nil
	}

	info, err := os.Stat(logFile)
	if err != nil {
		return fmt.Errorf("获取文件信息失败：%w", err)
	}

	if info.Size() <= maxSize {
		return nil
	}

	// 读取所有行
	data, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("读取文件失败：%w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}

	// 计算需要删除的行数（保留后 maxSize/2 字节）
	bytesCount := 0
	keepStart := 0
	for i := len(lines) - 1; i >= 0; i-- {
		bytesCount += len(lines[i]) + 1 // +1 for newline
		if int64(bytesCount) >= maxSize/2 {
			keepStart = i
			break
		}
	}

	// 写入文件
	file, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("创建文件失败：%w", err)
	}
	defer file.Close()

	for _, line := range lines[keepStart:] {
		if line != "" {
			fmt.Fprintln(file, line)
		}
	}

	return nil
}
