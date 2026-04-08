// ============================================================
// info.go - 设备详情命令
// ============================================================
// 实现 "device-ctl info <device-id>" 子命令
// 显示单台设备的详细信息（框线表格输出）
// ============================================================

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <device-id>",  // <device-id> 是位置参数
	Short: "查看设备详情",
	Long:  `查看单台设备的详细信息`,
	Args:  cobra.ExactArgs(1), // 要求恰好 1 个位置参数
	Run: func(cmd *cobra.Command, args []string) {
		// args[0] 就是设备 ID（如 DEV-001）
		device, err := Store.GetDevice(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 状态 emoji
		statusEmoji := map[string]string{
			"online":       "🟢",
			"offline":      "🔴",
			"error":        "⚠️",
			"maintenance":  "🔧",
		}
		emoji := statusEmoji[device.Status]

		// 框线表格输出（类似 ASCII art）
		// ┌─┐│├─┤└─┘：Unicode 框线字符
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
│  运行时长: %-10d小时  故障次数: %-10d      │
├─────────────────────────────────────────────────┤
│  设备配置                                        │
│  屏幕亮度: %-5d%%    音量: %-5d%%             │
│  自动重启: %-5v %s                          │
└─────────────────────────────────────────────────┘
`,
			emoji, device.Name,
			device.ID,
			device.Type,
			device.Region,
			device.Address,
			emoji, device.Status,
			device.Firmware,
			device.InstalledAt.Format("2006-01-02"),  // Go 的时间格式化（2006-01-02 是参考时间）
			device.LastHeartbeat.Format("2006-01-02 15:04:05"),
			device.Stats.Uptime, device.Stats.FaultCount,
			device.Config.ScreenBrightness, device.Config.VolumeLevel,
			device.Config.AutoRebootEnabled, device.Config.AutoRebootTime,
		)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
