package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var statsDays int

var statsCmd = &cobra.Command{
	Use:   "stats <device-id>",
	Short: "查看设备统计",
	Long:  `查看设备运行统计数据`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		device, err := Store.GetDevice(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf(`
┌─────────────────────────────────────────────────┐
│  📊 设备统计: %-35s│
├─────────────────────────────────────────────────┤
│                                                 │
│  总交易数:     %-10d                        │
│  今日交易:     %-10d                        │
│  运行时长:     %-10d 小时                    │
│  故障次数:     %-10d                        │
│                                                 │
│  日均交易:     %-10.1f                        │
│  故障率:       %-10.2f%%                       │
│                                                 │
└─────────────────────────────────────────────────┘
`,
			device.Name,
			device.Stats.TotalTransactions,
			device.Stats.TodayTransactions,
			device.Stats.Uptime,
			device.Stats.FaultCount,
			float64(device.Stats.TotalTransactions)/float64(max(device.Stats.Uptime/24, 1)),
			float64(device.Stats.FaultCount)/float64(max(device.Stats.Uptime/24, 1))*100,
		)
	},
}

func init() {
	statsCmd.Flags().IntVarP(&statsDays, "days", "d", 7, "统计天数")
	rootCmd.AddCommand(statsCmd)
}
