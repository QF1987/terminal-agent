package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"
)

var (
	monitorDevice   string
	monitorSeverity string
	monitorLimit    int
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "设备监控",
	Long:  `查看设备监控信息和告警`,
}

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "查看告警",
	Long:  `查看设备告警信息`,
	Run: func(cmd *cobra.Command, args []string) {
		filters := device.LogFilters{
			DeviceID: monitorDevice,
			Severity: monitorSeverity,
			Days:     7,
			Limit:    monitorLimit,
		}

		logs, err := Store.GetFaultLogs(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		if len(logs) == 0 {
			fmt.Println("✅ 没有告警信息")
			return
		}

		severityEmoji := map[string]string{
			device.SeverityLow:      "ℹ️",
			device.SeverityMedium:   "⚠️",
			device.SeverityHigh:     "🔴",
			device.SeverityCritical: "🚨",
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "时间\t设备\t严重程度\t类型\t描述\t状态")
		fmt.Fprintln(w, "----\t----\t--------\t----\t----\t----")

		for _, log := range logs {
			emoji := severityEmoji[log.Severity]
			status := "✅ 已解决"
			if !log.Resolved {
				status = "❌ 未解决"
			}
			fmt.Fprintf(w, "%s\t%s\t%s %s\t%s\t%s\t%s\n",
				log.Timestamp.Format("01-02 15:04"),
				log.DeviceID,
				emoji, log.Severity,
				log.Type,
				log.Message,
				status)
		}
		w.Flush()

		fmt.Printf("\n共 %d 条告警\n", len(logs))
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看设备状态概览",
	Long:  `查看所有设备的状态概览`,
	Run: func(cmd *cobra.Command, args []string) {
		devices, err := Store.ListDevices(device.DeviceFilters{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		counts := make(map[string]int)
		for _, d := range devices {
			counts[d.Status]++
		}

		fmt.Printf(`
┌─────────────────────────────────────────────────┐
│  📡 设备状态概览                                 │
├─────────────────────────────────────────────────┤
│                                                 │
│  🟢 在线:     %-10d                        │
│  🔴 离线:     %-10d                        │
│  ⚠️  故障:     %-10d                        │
│  🔧 维护中:   %-10d                        │
│  ─────────────────────                         │
│  📦 总计:     %-10d                        │
│                                                 │
└─────────────────────────────────────────────────┘
`,
			counts[device.StatusOnline],
			counts[device.StatusOffline],
			counts[device.StatusError],
			counts[device.StatusMaintenance],
			len(devices),
		)
	},
}

func init() {
	monitorCmd.AddCommand(alertsCmd)
	monitorCmd.AddCommand(statusCmd)
	monitorCmd.PersistentFlags().StringVarP(&monitorDevice, "device", "d", "", "按设备筛选")
	monitorCmd.PersistentFlags().StringVar(&monitorSeverity, "severity", "", "按严重程度筛选 (low/medium/high/critical)")
	monitorCmd.PersistentFlags().IntVarP(&monitorLimit, "limit", "n", 20, "返回条数限制")
	rootCmd.AddCommand(monitorCmd)
}
