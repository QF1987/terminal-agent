// ============================================================
// logs.go - 故障日志命令
// ============================================================
// 实现 "device-ctl logs" 子命令
// 查看设备故障日志，支持按设备、严重程度、类型筛选
// ============================================================

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"
)

var (
	logsDevice   string  // --device：按设备ID筛选
	logsSeverity string  // --severity：按严重程度筛选
	logsType     string  // --type：按日志类型筛选
	logsDays     int     // --days：最近几天
	logsLimit    int     // --limit：返回条数
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "查看故障日志",
	Long:  `查看设备故障日志，支持按设备、严重程度、类型筛选`,
	Run: func(cmd *cobra.Command, args []string) {
		// 构建筛选条件
		filters := device.LogFilters{
			DeviceID: logsDevice,
			Severity: logsSeverity,
			Type:     logsType,
			Days:     logsDays,
			Limit:    logsLimit,
		}

		// 查询日志
		logs, err := Store.GetFaultLogs(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		if len(logs) == 0 {
			fmt.Println("没有找到匹配的日志")
			return
		}

		// 严重程度到 emoji 的映射
		severityEmoji := map[string]string{
			device.SeverityLow:      "ℹ️",
			device.SeverityMedium:   "⚠️",
			device.SeverityHigh:     "🔴",
			device.SeverityCritical: "💀",
		}

		// 格式化输出
		fmt.Printf("\n📋 故障日志（共 %d 条）\n\n", len(logs))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t设备\t时间\t严重程度\t类型\t状态\t描述")
		fmt.Fprintln(w, "--\t----\t----\t--------\t----\t----\t----")

		for _, log := range logs {
			emoji := severityEmoji[log.Severity]
			resolved := "❌ 未解决"
			if log.Resolved {
				resolved = "✅ 已解决"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s\t%s\t%s\n",
				log.ID, log.DeviceID,
				log.Timestamp.Format("01-02 15:04"),
				emoji, log.Severity,
				log.Type, resolved, log.Message)
		}
		w.Flush()
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logsDevice, "device", "D", "", "按设备ID筛选")
	logsCmd.Flags().StringVarP(&logsSeverity, "severity", "S", "", "严重程度 (low/medium/high/critical)")
	logsCmd.Flags().StringVarP(&logsType, "type", "T", "", "日志类型 (hardware/software/network/medicine_stock)")
	logsCmd.Flags().IntVarP(&logsDays, "days", "d", 7, "最近几天")
	logsCmd.Flags().IntVarP(&logsLimit, "limit", "l", 20, "返回条数")
	rootCmd.AddCommand(logsCmd)
}
