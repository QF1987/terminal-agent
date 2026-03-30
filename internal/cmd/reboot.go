// ============================================================
// reboot.go - 重启设备命令
// ============================================================
// 实现 "device-ctl reboot <device-id>" 子命令
// 远程重启指定设备
// ============================================================

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rebootForce bool  // --force：强制重启

var rebootCmd = &cobra.Command{
	Use:   "reboot <device-id>",
	Short: "重启设备",
	Long:  `远程重启指定设备`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]

		// 检查设备是否存在
		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 确认提示（除非用了 --force）
		if !rebootForce {
			fmt.Printf("⚠️  确定要重启设备 %s (%s) 吗？[y/N]: ", dev.Name, deviceID)
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Println("已取消")
				return
			}
		}

		// 执行重启
		fmt.Printf("🔄 正在重启设备 %s...\n", deviceID)
		err = Store.RebootDevice(deviceID, rebootForce)
		if err != nil {
			fmt.Fprintf(os.Stderr, "重启失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ 设备 %s 重启命令已发送\n", deviceID)
	},
}

func init() {
	rebootCmd.Flags().BoolVarP(&rebootForce, "force", "f", false, "强制重启，跳过确认")
	rootCmd.AddCommand(rebootCmd)
}
