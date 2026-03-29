package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rebootForce bool

var rebootCmd = &cobra.Command{
	Use:   "reboot <device-id>",
	Short: "重启设备",
	Long:  `远程重启指定设备`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]

		device, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("正在重启设备: %s (%s)\n", device.Name, deviceID)

		if rebootForce {
			fmt.Println("⚡ 强制重启模式")
		}

		err = Store.RebootDevice(deviceID, rebootForce)
		if err != nil {
			fmt.Fprintf(os.Stderr, "重启失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ 重启指令已发送")
		fmt.Printf("设备 %s 正在重启，预计 30 秒后恢复\n", device.Name)
	},
}

func init() {
	rebootCmd.Flags().BoolVarP(&rebootForce, "force", "f", false, "强制重启")
	rootCmd.AddCommand(rebootCmd)
}
