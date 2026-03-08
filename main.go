package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"mihomo-sync/config"
	"mihomo-sync/handler"
	"mihomo-sync/util"
	"mihomo-sync/version"
)

// 默认配置
const (
	DefaultBaseDir  = "/data/Mihomo-sync-go"
	DefaultConfig   = "subs.conf"
	DefaultTemplate = "template.yaml"
	ConfigsDir      = "configs"
)

// Args 命令行参数
type Args struct {
	List      bool
	Select    string
	Target    string
	Config    string
	BaseDir   string
	NoCrontab bool
	Template  string
	All       bool
}

func main() {
	args := parseArgs()

	// 确定基础目录
	baseDir := args.BaseDir
	if baseDir == "" {
		baseDir = DefaultBaseDir
	}

	// 确定配置文件路径
	configFile := args.Config
	if configFile == "" {
		configFile = filepath.Join(baseDir, DefaultConfig)
	}

	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		// 配置文件不存在，创建示例配置
		if os.IsNotExist(err) {
			fmt.Printf("配置文件不存在：%s\n", configFile)
			fmt.Println("正在创建示例配置...")

			if err := config.CreateExampleConfig(configFile); err != nil {
				fmt.Printf("创建示例配置失败：%v\n", err)
				os.Exit(1)
			}

			fmt.Printf("已创建示例配置：%s\n", configFile)
			fmt.Println("请编辑配置文件后重新运行")
			os.Exit(0)
		}
		// 配置验证失败或其他错误
		fmt.Printf("配置加载失败：%v\n", err)
		os.Exit(1)
	}

	// 获取日志配置
	logFile := cfg.LogFile
	if logFile == "" {
		logFile = filepath.Join(baseDir, "update.log")
	}
	logLevel := cfg.LogLevel
	if logLevel == "" {
		logLevel = "INFO"
	}
	maxLogSizeStr := cfg.MaxLogSize
	if maxLogSizeStr == "" {
		maxLogSizeStr = "10MB"
	}

	// 解析日志大小
	maxLogSize, err := util.ParseSize(maxLogSizeStr)
	if err != nil {
		fmt.Printf("解析日志大小失败：%v\n", err)
		maxLogSize = 10 * 1024 * 1024 // 默认 10MB
	}

	// 轮转日志
	if err := util.RotateLog(logFile, maxLogSize); err != nil {
		fmt.Printf("轮转日志失败：%v\n", err)
	}

	// 创建日志记录器
	logger, err := util.NewLogger(util.LoggerConfig{
		File:  logFile,
		Level: logLevel,
	})
	if err != nil {
		fmt.Printf("创建日志记录器失败：%v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("配置文件：%s", configFile)
	logger.Debug("日志文件：%s", logFile)
	logger.Debug("日志等级：%s", logLevel)

	// 确定模板路径
	template := args.Template
	if template == "" {
		template = cfg.Template
	}
	if template == "" {
		template = filepath.Join(baseDir, DefaultTemplate)
	}

	// 确定软链接目标目录
	targetDir := args.Target
	if targetDir == "" {
		targetDir = cfg.OutputDir
	}
	if targetDir == "" {
		targetDir = baseDir
	}

	// 确定 configs 目录
	configsDir := filepath.Join(baseDir, ConfigsDir)

	// 列出订阅
	if args.List {
		listSubscriptions(cfg, logger)
		return
	}

	// 更新所有订阅
	if args.All {
		logger.Info("=== 更新所有订阅 ===")
		// 临时修改输出目录为 configs 目录
		origOutputDir := cfg.OutputDir
		cfg.OutputDir = configsDir
		handler.UpdateAllSubscriptions(cfg, logger)
		cfg.OutputDir = origOutputDir
		logger.Info("所有订阅已更新到 configs 目录")
		if cfg.RestartCmd != "" {
			if success, err := util.ExecuteCommand(cfg.RestartCmd); err != nil || !success {
				logger.Error("重启命令执行失败：%v", err)
			} else {
				logger.Info("已执行重启命令")
			}
		}
		return
	}

	// 不带参数时，显示订阅列表并让用户选择
	if args.Select == "" {
		sub, ok := listSubscriptionsInteractive(cfg, logger)
		if !ok || sub == nil {
			return
		}
		// 用户选择了订阅，继续执行更新逻辑
		// 确定输出路径：configs/<别名>.yaml
		outputPath := filepath.Join(configsDir, sub.Alias+".yaml")

		// 更新指定订阅
		logger.Info("更新订阅：%s", sub.Alias)
		success := handler.UpdateSubscription(sub, template, outputPath, logger, cfg)

		if !success {
			logger.Error("更新失败，跳过后续操作")
			os.Exit(1)
		}

		// 创建软链接
		if err := util.CreateSymlink(outputPath, targetDir); err != nil {
			logger.Error("创建软链接失败：%v", err)
		} else {
			logger.Info("已创建软链接：%s/config.yaml -> %s", targetDir, outputPath)
		}

		// 设置定时任务
		if !args.NoCrontab && sub.AutoUpdate {
			scriptPath := filepath.Join(baseDir, "mihomo-sync")
			if err := util.SetupCrontab(sub.Alias, sub.Interval, scriptPath, configFile, baseDir, logFile, true); err != nil {
				logger.Error("设置定时任务失败：%v", err)
			} else {
				logger.Info("已设置定时任务：%s (每%d小时)", sub.Alias, sub.Interval)
			}
		}

		logger.Info("更新完成")
		return
	}

	// 选择订阅（只更新指定的订阅）
	var sub *config.Subscription
	if args.Select != "" {
		sub = findSubscription(cfg, args.Select)
		if sub == nil {
			logger.Error("未找到订阅：%s", args.Select)
			os.Exit(1)
		}
	}

	// 确定输出路径：configs/<别名>.yaml
	outputPath := filepath.Join(configsDir, sub.Alias+".yaml")

	// 更新指定订阅
	logger.Info("更新订阅：%s", sub.Alias)
	success := handler.UpdateSubscription(sub, template, outputPath, logger, cfg)

	if !success {
		logger.Error("更新失败，跳过后续操作")
		os.Exit(1)
	}

	// 创建软链接
	if err := util.CreateSymlink(outputPath, targetDir); err != nil {
		logger.Error("创建软链接失败：%v", err)
	} else {
		logger.Info("已创建软链接：%s/config.yaml -> %s", targetDir, outputPath)
	}

	// 设置定时任务
	if !args.NoCrontab && sub.AutoUpdate {
		scriptPath := filepath.Join(baseDir, "mihomo-sync")
		if err := util.SetupCrontab(sub.Alias, sub.Interval, scriptPath, configFile, baseDir, logFile, true); err != nil {
			logger.Error("设置定时任务失败：%v", err)
		} else {
			logger.Info("已设置定时任务：%s (每%d小时)", sub.Alias, sub.Interval)
		}
	}

	logger.Info("更新完成")
}

// parseArgs 解析命令行参数
func parseArgs() Args {
	var args Args
	argsList := os.Args[1:]

	for i := 0; i < len(argsList); i++ {
		switch argsList[i] {
		case "-l", "--list":
			args.List = true
		case "-v", "--version":
			fmt.Printf("mihomo-sync version %s\n", version.Version)
			os.Exit(0)
		case "-s", "--select":
			if i+1 < len(argsList) {
				args.Select = argsList[i+1]
				i++
			}
		case "-t", "--target":
			if i+1 < len(argsList) {
				args.Target = argsList[i+1]
				i++
			}
		case "-c", "--config":
			if i+1 < len(argsList) {
				args.Config = argsList[i+1]
				i++
			}
		case "-d", "--base-dir":
			if i+1 < len(argsList) {
				args.BaseDir = argsList[i+1]
				i++
			}
		case "--no-crontab":
			args.NoCrontab = true
		case "--template":
			if i+1 < len(argsList) {
				args.Template = argsList[i+1]
				i++
			}
		case "--all":
			args.All = true
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		}
	}

	return args
}

// printUsage 打印使用说明
// printUsage 打印使用说明
func printUsage() {
	fmt.Println(`Mihomo 订阅管理

用法：mihomo-sync [选项]

选项:
  -l, --list              列出所有订阅
  -s, --select ALIAS      指定订阅别名
  -t, --target PATH       软链接目标目录
  -c, --config PATH       配置文件路径
  -d, --base-dir PATH     脚本所在目录
  --no-crontab            不设置定时任务
  --template PATH         配置模板路径
  --all                   更新所有订阅
  -v, --version           显示版本号
  -h, --help              显示帮助信息

示例:
  mihomo-sync --list                    # 列出所有订阅
  mihomo-sync --all                     # 更新所有订阅
  mihomo-sync -s main                   # 更新 main 订阅并切换
  mihomo-sync -s main -t /etc/mihomo    # 指定软链接目标目录`)
}

// listSubscriptions 列出所有订阅
func listSubscriptions(cfg *config.Config, logger *util.Logger) {
	logger.Info("=== 订阅列表 ===")
	for i, sub := range cfg.Subscriptions {
		status := "✗"
		if sub.AutoUpdate {
			status = "✓"
		}
		logger.Info("%d.%-8s 每%-2d小时   自动更新：%s", i+1, sub.Alias, sub.Interval, status)
	}
}

// listSubscriptionsInteractive 交互式列出订阅并让用户选择
// 返回选择的订阅和是否成功选择
func listSubscriptionsInteractive(cfg *config.Config, logger *util.Logger) (*config.Subscription, bool) {
	logger.Info("=== 订阅列表 ===")
	for i, sub := range cfg.Subscriptions {
		status := "✗"
		if sub.AutoUpdate {
			status = "✓"
		}
		logger.Info("%d. %-8s 每%-2d小时   自动更新：%s", i+1, sub.Alias, sub.Interval, status)
	}

	// 检测 stdin 是否是终端
	info, err := os.Stdin.Stat()
	isTerminal := err == nil && (info.Mode()&os.ModeCharDevice) != 0

	if isTerminal {
		fmt.Print("\n请选择订阅（输入序号或别名，直接回车取消）：")
	} else {
		// 非交互式环境，等待一小段时间看是否有输入
		fmt.Println()
	}

	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		// 读取失败（可能是 EOF）
		if !isTerminal {
			logger.Info("非交互式环境，请使用 -s 参数指定订阅")
		}
		return nil, false
	}
	choice = strings.TrimSpace(choice)

	if choice == "" {
		logger.Info("已取消")
		return nil, false
	}

	// 查找订阅
	// 数字选择
	if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(cfg.Subscriptions) {
		return &cfg.Subscriptions[idx-1], true
	}

	// 别名选择
	for _, s := range cfg.Subscriptions {
		if s.Alias == choice {
			return &s, true
		}
	}

	logger.Error("无效的选择")
	return nil, false
}

// findSubscription 查找订阅
func findSubscription(cfg *config.Config, alias string) *config.Subscription {
	for i := range cfg.Subscriptions {
		if cfg.Subscriptions[i].Alias == alias {
			return &cfg.Subscriptions[i]
		}
	}
	return nil
}
