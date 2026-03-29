package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <device-id>",
	Short: "查看设备详情",
	Long:  `查看单台设备的详细信息`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		device, err := Store.GetDevice(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		statusEmoji := map[string]string{
			"online":       "🟢",
			"offline":      "🔴",
			"error":        "⚠️",
			"maintenance":  "🔧",
		}
		emoji := statusEmoji[device.Status]

		fmt.Printf(`
┌─────────────────────────────────────────────────┐
│  %s  %-40s│
├─────────────────────────────────────────────────┤
│  ID:       %-36s│
│  类型:     %-36s│
│  区域:     %-36s│
│  地址:     %-36s│
│  状态:     %s %-34s│
│  固件:     %-36s│
│  安装时间: %-36s│
│  最后心跳: %-36s│
├─────────────────────────────────────────────────┤
│  运行统计                                        │
│  总交易:   %-10d  今日交易: %-10d        │
│  运行时长: %-10d小时  故障次数: %-10d      │
├─────────────────────────────────────────────────┤
│  设备配置                                        │
│  屏幕亮度: %-5d%%    音量: %-5d%%             │
│  交易超时: %-5d秒   自动重启: %-5v %s       │
└─────────────────────────────────────────────────┘
`,
			emoji, device.Name,
			device.ID,
			device.Type,
			device.Region,
			device.Address,
			emoji, device.Status,
			device.Firmware,
			device.InstalledAt.Format("2006-01-02"),
			device.LastHeartbeat.Format("2006-01-02 15:04:05"),
			device.Stats.TotalTransactions, device.Stats.TodayTransactions,
			device.Stats.Uptime, device.Stats.FaultCount,
			device.Config.ScreenBrightness, device.Config.VolumeLevel,
			device.Config.TransactionTimeout, device.Config.AutoRebootEnabled, device.Config.AutoRebootTime,
		)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
