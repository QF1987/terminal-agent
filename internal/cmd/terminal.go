// ============================================================
// terminal.go - 终端信息命令
// ============================================================
// 实现 "device-ctl terminal" 子命令
// 包含 info（硬件信息）、network（网络状态）、logs（终端日志）
// ============================================================

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// ─── terminal 父命令 ──────────────────────────────────────
var terminalCmd = &cobra.Command{
	Use:   "terminal",
	Short: "终端信息",
	Long:  `终端信息命令，包含 info（硬件信息）、network（网络状态）、logs（终端日志）子命令`,
}

// ─── terminal info 子命令 ──────────────────────────────────
var terminalInfoCmd = &cobra.Command{
	Use:   "info <device-id>",
	Short: "查看终端硬件信息",
	Long:  `查看设备终端硬件信息`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]

		// 获取设备信息
		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 模拟硬件信息（实际应该从 API 获取）
		fmt.Printf("\n🖥️  终端硬件信息 - %s (%s)\n\n", dev.Name, deviceID)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "项目\t数值")
		fmt.Fprintln(w, "----\t----")
		fmt.Fprintln(w, "CPU\tARM Cortex-A53 1.4GHz")
		fmt.Fprintln(w, "内存\t2GB / 4GB (50%)")
		fmt.Fprintln(w, "存储\t15GB / 32GB (47%)")
		fmt.Fprintln(w, "系统\tLinux 5.4.0")
		fmt.Fprintln(w, "运行时间\t1219 小时")
		fmt.Fprintln(w, "负载\t0.35")
		w.Flush()
	},
}

// ─── terminal network 子命令 ──────────────────────────────
var terminalNetworkCmd = &cobra.Command{
	Use:   "network <device-id>",
	Short: "查看网络状态",
	Long:  `查看设备网络连接状态`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]

		// 获取设备信息
		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 模拟网络信息
		fmt.Printf("\n🌐 网络状态 - %s (%s)\n\n", dev.Name, deviceID)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "项目\t数值")
		fmt.Fprintln(w, "----\t----")
		fmt.Fprintln(w, "连接类型\t4G LTE")
		fmt.Fprintln(w, "IP地址\t10.23.45.67")
		fmt.Fprintln(w, "信号强度\t-65 dBm (良好)")
		fmt.Fprintln(w, "今日流量\t↑ 45.2 MB  ↓ 123.8 MB")
		fmt.Fprintln(w, "VPN状态\t已连接")
		fmt.Fprintln(w, "延迟\t23 ms")
		w.Flush()
	},
}

// ─── terminal logs 子命令 ─────────────────────────────────
var terminalLogsLimit int

var terminalLogsCmd = &cobra.Command{
	Use:   "logs <device-id>",
	Short: "查看终端日志",
	Long:  `查看设备终端日志`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]

		// 获取设备信息
		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 模拟终端日志
		fmt.Printf("\n📝 终端日志 - %s (%s)\n\n", dev.Name, deviceID)

		logs := []string{
			"[2026-03-29 20:30:15] INFO  心跳包发送成功",
			"[2026-03-29 20:29:45] INFO  交易 #12345 完成",
			"[2026-03-29 20:28:30] WARN  4G信号弱 (-78dBm)",
			"[2026-03-29 20:25:00] INFO  系统自动检查更新",
			"[2026-03-29 20:20:15] INFO  交易 #12344 完成",
			"[2026-03-29 20:15:00] INFO  温度传感器正常 (25.3°C)",
		}

		limit := terminalLogsLimit
		if limit > len(logs) {
			limit = len(logs)
		}

		for i := 0; i < limit; i++ {
			fmt.Println(logs[i])
		}
	},
}

func init() {
	// logs 的 flag
	terminalLogsCmd.Flags().IntVarP(&terminalLogsLimit, "limit", "l", 20, "返回条数")

	// 添加子命令
	terminalCmd.AddCommand(terminalInfoCmd)
	terminalCmd.AddCommand(terminalNetworkCmd)
	terminalCmd.AddCommand(terminalLogsCmd)

	rootCmd.AddCommand(terminalCmd)
}
