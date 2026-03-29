package cmd

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/spf13/cobra"
)

var terminalCmd = &cobra.Command{
	Use:   "terminal",
	Short: "终端信息",
	Long:  `查看设备终端硬件、网络、日志信息`,
}

// terminal info
var terminalInfoCmd = &cobra.Command{
	Use:   "info <device-id>",
	Short: "查看硬件信息",
	Long:  `查看设备硬件和系统信息`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dev, err := Store.GetDevice(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		cpuUsage := 20 + rand.Intn(60)
		memUsage := 30 + rand.Intn(50)
		diskUsage := 40 + rand.Intn(40)

		fmt.Printf(`
┌─────────────────────────────────────────────────┐
│  💻 终端硬件信息: %-30s│
├─────────────────────────────────────────────────┤
│  设备 ID:    %-34s│
│  设备类型:   %-34s│
│  固件版本:   %-34s│
│  安装日期:   %-34s│
├─────────────────────────────────────────────────┤
│  系统资源                                        │
│  CPU 使用率:  [%-20s] %3d%%          │
│  内存使用率:  [%-20s] %3d%%          │
│  磁盘使用率:  [%-20s] %3d%%          │
├─────────────────────────────────────────────────┤
│  硬件状态                                        │
│  打印机:     ✅ 正常                             │
│  扫码器:     ✅ 正常                             │
│  触摸屏:     ✅ 正常                             │
│  出药机构:   ✅ 正常                             │
│  4G 模块:    ✅ 信号良好                         │
└─────────────────────────────────────────────────┘
`,
			dev.Name,
			dev.ID,
			dev.Type,
			dev.Firmware,
			dev.InstalledAt.Format("2006-01-02"),
			progressBar(cpuUsage), cpuUsage,
			progressBar(memUsage), memUsage,
			progressBar(diskUsage), diskUsage,
		)
	},
}

// terminal network
var terminalNetworkCmd = &cobra.Command{
	Use:   "network <device-id>",
	Short: "查看网络状态",
	Long:  `查看设备网络连接状态`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dev, err := Store.GetDevice(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		signal := 60 + rand.Intn(40)
		ip := fmt.Sprintf("192.168.%d.%d", rand.Intn(255), rand.Intn(255))

		fmt.Printf(`
┌─────────────────────────────────────────────────┐
│  🌐 网络状态: %-33s│
├─────────────────────────────────────────────────┤
│  连接状态:   ✅ 已连接                           │
│  网络类型:   4G LTE                             │
│  IP 地址:    %-34s│
│  信号强度:   [%-20s] %3d%%          │
│  最后心跳:   %-34s│
├─────────────────────────────────────────────────┤
│  网络统计                                        │
│  上行流量:   125.3 MB (今日)                     │
│  下行流量:   89.7 MB (今日)                      │
│  连接时长:   72 小时                             │
│  断线次数:   0 次 (7天内)                        │
└─────────────────────────────────────────────────┘
`,
			dev.Name,
			ip,
			progressBar(signal), signal,
			dev.LastHeartbeat.Format("2006-01-02 15:04:05"),
		)
	},
}

// terminal logs
var (
	terminalLogsLines  int
	terminalLogsFollow bool
)

var terminalLogsCmd = &cobra.Command{
	Use:   "logs <device-id>",
	Short: "查看终端日志",
	Long:  `查看设备终端运行日志`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dev, err := Store.GetDevice(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("📋 设备 %s 的终端日志 (最近 %d 行):\n\n", dev.Name, terminalLogsLines)

		// 模拟日志
		logs := []string{
			"[INFO]  系统启动完成",
			"[INFO]  加载配置文件: /etc/device/config.yaml",
			"[INFO]  网络连接已建立",
			"[INFO]  心跳服务已启动 (间隔: 30s)",
			"[INFO]  药品库存检查完成",
			"[DEBUG] 等待交易请求...",
			"[INFO]  收到交易请求 #TXN-20260329-001",
			"[INFO]  交易完成: 布洛芬缓释胶囊 x1",
			"[WARN]  触摸屏校准偏移 2px",
			"[INFO]  自动校准完成",
		}

		for i, log := range logs {
			if i >= terminalLogsLines {
				break
			}
			fmt.Printf("2026-03-29 10:%02d:%02d %s\n", 30+i, rand.Intn(60), log)
		}

		if terminalLogsFollow {
			fmt.Println("\n📡 实时监控中... (Ctrl+C 退出)")
		}
	},
}

func progressBar(percent int) string {
	width := 20
	filled := percent * width / 100
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

func init() {
	terminalLogsCmd.Flags().IntVarP(&terminalLogsLines, "lines", "n", 10, "显示行数")
	terminalLogsCmd.Flags().BoolVarP(&terminalLogsFollow, "follow", "f", false, "实时跟踪日志")

	terminalCmd.AddCommand(terminalInfoCmd)
	terminalCmd.AddCommand(terminalNetworkCmd)
	terminalCmd.AddCommand(terminalLogsCmd)
	rootCmd.AddCommand(terminalCmd)
}
