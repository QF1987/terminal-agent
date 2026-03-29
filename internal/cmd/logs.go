package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"
)

var (
	logDeviceID string
	logSeverity string
	logType     string
	logDays     int
	logLimit    int
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "查看故障日志",
	Long:  `查看设备故障日志`,
	Run: func(cmd *cobra.Command, args []string) {
		filters := device.LogFilters{
			DeviceID: logDeviceID,
			Severity: logSeverity,
			Type:     logType,
			Days:     logDays,
			Limit:    logLimit,
		}

		logs, err := Store.GetFaultLogs(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		if len(logs) == 0 {
			fmt.Println("没有找到故障日志")
			return
		}

		typeEmoji := map[string]string{
			device.LogHardware:      "🔧",
			device.LogSoftware:      "💻",
			device.LogNetwork:       "🌐",
			device.LogMedicineStock: "💊",
		}

		severityEmoji := map[string]string{
			device.SeverityLow:      "ℹ️",
			device.SeverityMedium:   "⚠️",
			device.SeverityHigh:     "🔴",
			device.SeverityCritical: "🚨",
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t时间\t设备\t类型\t严重程度\t描述\t状态")
		fmt.Fprintln(w, "--\t----\t----\t----\t--------\t----\t----")

		for _, log := range logs {
			tEmoji := typeEmoji[log.Type]
			sEmoji := severityEmoji[log.Severity]
			status := "✅"
			if !log.Resolved {
				status = "❌"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s %s\t%s\t%s\n",
				log.ID,
				log.Timestamp.Format("01-02 15:04"),
				log.DeviceID,
				tEmoji, log.Type,
				sEmoji, log.Severity,
				log.Message,
				status)
		}
		w.Flush()

		fmt.Printf("\n共 %d 条日志\n", len(logs))
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logDeviceID, "device", "d", "", "按设备ID筛选")
	logsCmd.Flags().StringVar(&logSeverity, "severity", "", "按严重程度筛选 (low/medium/high/critical)")
	logsCmd.Flags().StringVar(&logType, "type", "", "按类型筛选 (hardware/software/network/medicine_stock)")
	logsCmd.Flags().IntVar(&logDays, "days", 7, "查看最近几天的日志")
	logsCmd.Flags().IntVarP(&logLimit, "limit", "n", 20, "返回条数限制")
	rootCmd.AddCommand(logsCmd)
}
