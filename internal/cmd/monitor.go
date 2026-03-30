// ============================================================
// monitor.go - 监控命令
// ============================================================
// 实现 "device-ctl monitor" 子命令
// 包含两个子命令：status（概览）和 alerts（告警）
// ============================================================

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"
)

// ─── monitor 父命令 ───────────────────────────────────────
// monitor 本身不执行，只是 status 和 alerts 的"父命令"
// 类似 Linux 命令的嵌套：git remote add / git remote remove
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "设备监控",
	Long:  `设备监控命令，包含 status（状态概览）和 alerts（告警）子命令`,
}

// ─── monitor status 子命令 ────────────────────────────────
var monitorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "设备状态概览",
	Long:  `显示所有设备的状态统计`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取所有设备（不筛选）
		devices, _ := Store.ListDevices(device.DeviceFilters{})

		// 按状态统计
		// map[string]int：类似 Record<string, number>
		counts := make(map[string]int)
		for _, d := range devices {
			counts[d.Status]++
		}

		// 输出统计
		fmt.Println("\n📊 设备状态概览")
		fmt.Println("──────────────────")
		fmt.Printf("🟢 在线:   %d 台\n", counts["online"])
		fmt.Printf("🔴 离线:   %d 台\n", counts["offline"])
		fmt.Printf("⚠️  故障:   %d 台\n", counts["error"])
		fmt.Printf("🔧 维护中: %d 台\n", counts["maintenance"])
		fmt.Printf("──────────────────\n")
		fmt.Printf("📦 总计:   %d 台\n", len(devices))
	},
}

// ─── monitor alerts 子命令 ────────────────────────────────
var (
	monitorAlertsDevice   string
	monitorAlertsSeverity string
	monitorAlertsLimit    int
)

var monitorAlertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "查看告警",
	Long:  `查看设备告警信息`,
	Run: func(cmd *cobra.Command, args []string) {
		// 筛选条件：severity 高/严重 + 未解决
		filters := device.LogFilters{
			DeviceID: monitorAlertsDevice,
			Severity: monitorAlertsSeverity,
			Days:     7,
			Limit:    monitorAlertsLimit,
		}

		logs, err := Store.GetFaultLogs(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 只显示 high 和 critical 级别的告警
		var alerts []device.FaultLog
		for _, log := range logs {
			if log.Severity == device.SeverityHigh || log.Severity == device.SeverityCritical {
				alerts = append(alerts, log)
			}
		}

		if len(alerts) == 0 {
			fmt.Println("✅ 没有高严重程度的告警")
			return
		}

		fmt.Printf("\n🚨 告警信息（共 %d 条）\n\n", len(alerts))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "设备\t时间\t严重程度\t描述")
		fmt.Fprintln(w, "----\t----\t--------\t----")

		for _, alert := range alerts {
			emoji := "🔴"
			if alert.Severity == device.SeverityCritical {
				emoji = "💀"
			}
			fmt.Fprintf(w, "%s\t%s\t%s %s\t%s\n",
				alert.DeviceID,
				alert.Timestamp.Format("01-02 15:04"),
				emoji, alert.Severity,
				alert.Message)
		}
		w.Flush()
	},
}

func init() {
	// 添加子命令到父命令
	monitorCmd.AddCommand(monitorStatusCmd)
	monitorCmd.AddCommand(monitorAlertsCmd)

	// alerts 的 flag
	monitorAlertsCmd.Flags().StringVarP(&monitorAlertsDevice, "device", "D", "", "按设备ID筛选")
	monitorAlertsCmd.Flags().StringVarP(&monitorAlertsSeverity, "severity", "S", "", "严重程度")
	monitorAlertsCmd.Flags().IntVarP(&monitorAlertsLimit, "limit", "l", 20, "返回条数")

	// 把 monitor 添加到根命令
	rootCmd.AddCommand(monitorCmd)
}
