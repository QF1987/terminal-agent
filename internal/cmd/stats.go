// ============================================================
// stats.go - 设备统计命令
// ============================================================
// 实现 "device-ctl stats <device-id>" 子命令
// 显示设备的运行统计数据（交易量、运行时长、故障次数）
// ============================================================

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// --days：统计天数参数
var statsDays int

var statsCmd = &cobra.Command{
	Use:   "stats <device-id>",
	Short: "查看设备统计",
	Long:  `查看设备运行统计数据`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 获取设备信息（为了显示名称）
		dev, err := Store.GetDevice(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 获取统计数据
		stats, err := Store.GetDeviceStats(args[0], statsDays)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 格式化输出
		fmt.Printf("\n📊 设备统计 - %s (%s)\n", dev.Name, dev.ID)
		fmt.Printf("统计周期: 最近 %d 天\n\n", statsDays)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "指标\t数值")
		fmt.Fprintln(w, "----\t----")
		fmt.Fprintf(w, "运行时长\t%d 小时\n", stats.Uptime)
		fmt.Fprintf(w, "故障次数\t%d 次\n", stats.FaultCount)
		w.Flush()
	},
}

func init() {
	statsCmd.Flags().IntVarP(&statsDays, "days", "d", 7, "统计天数")
	rootCmd.AddCommand(statsCmd)
}
