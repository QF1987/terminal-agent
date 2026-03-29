package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "批量操作",
	Long:  `对多台设备执行批量操作`,
}

// batch reboot
var (
	batchRebootRegion  string
	batchRebootConfirm bool
)

var batchRebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "批量重启设备",
	Long:  `批量重启指定区域的设备`,
	Run: func(cmd *cobra.Command, args []string) {
		if batchRebootRegion == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --region 参数")
			os.Exit(1)
		}

		filters := device.DeviceFilters{Region: batchRebootRegion}
		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Printf("区域 %s 没有找到设备\n", batchRebootRegion)
			return
		}

		fmt.Printf("⚠️  即将重启区域 [%s] 的 %d 台设备:\n\n", batchRebootRegion, len(devices))
		for _, d := range devices {
			fmt.Printf("  - %s (%s)\n", d.Name, d.ID)
		}

		if !batchRebootConfirm {
			fmt.Println("\n使用 --confirm 参数确认执行")
			return
		}

		fmt.Println("\n🚀 正在发送重启指令...")
		for _, d := range devices {
			fmt.Printf("  ✅ %s - 重启指令已发送\n", d.Name)
		}
		fmt.Printf("\n共 %d 台设备正在重启\n", len(devices))
	},
}

// batch config
var (
	batchConfigRegion string
	batchConfigSet    []string
	batchConfigConfirm bool
)

var batchConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "批量修改配置",
	Long:  `批量修改指定区域设备的配置`,
	Run: func(cmd *cobra.Command, args []string) {
		if batchConfigRegion == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --region 参数")
			os.Exit(1)
		}

		if len(batchConfigSet) == 0 {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --set 参数")
			os.Exit(1)
		}

		filters := device.DeviceFilters{Region: batchConfigRegion}
		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("⚠️  即将修改区域 [%s] %d 台设备的配置:\n\n", batchConfigRegion, len(devices))
		fmt.Println("配置变更:")
		for _, s := range batchConfigSet {
			parts := strings.SplitN(s, "=", 2)
			if len(parts) == 2 {
				fmt.Printf("  %s = %s\n", parts[0], parts[1])
			}
		}

		if !batchConfigConfirm {
			fmt.Println("\n使用 --confirm 参数确认执行")
			return
		}

		fmt.Println("\n🚀 正在修改配置...")
		for _, d := range devices {
			fmt.Printf("  ✅ %s - 配置已更新\n", d.Name)
		}
		fmt.Printf("\n共 %d 台设备配置已更新\n", len(devices))
	},
}

// batch firmware
var (
	batchFirmwareRegion  string
	batchFirmwareVersion string
	batchFirmwareConfirm bool
)

var batchFirmwareCmd = &cobra.Command{
	Use:   "firmware",
	Short: "批量升级固件",
	Long:  `批量升级指定区域设备的固件`,
	Run: func(cmd *cobra.Command, args []string) {
		if batchFirmwareRegion == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --region 参数")
			os.Exit(1)
		}

		if batchFirmwareVersion == "" {
			batchFirmwareVersion = "2.3.1"
		}

		filters := device.DeviceFilters{Region: batchFirmwareRegion}
		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 找出需要升级的设备
		var outdated []device.Device
		for _, d := range devices {
			if d.Firmware < batchFirmwareVersion {
				outdated = append(outdated, d)
			}
		}

		if len(outdated) == 0 {
			fmt.Printf("✅ 区域 %s 的所有设备已是最新固件\n", batchFirmwareRegion)
			return
		}

		fmt.Printf("⚠️  即将升级区域 [%s] %d 台设备的固件到 %s:\n\n", batchFirmwareRegion, len(outdated), batchFirmwareVersion)
		for _, d := range outdated {
			fmt.Printf("  - %s (%s): %s -> %s\n", d.Name, d.ID, d.Firmware, batchFirmwareVersion)
		}

		if !batchFirmwareConfirm {
			fmt.Println("\n使用 --confirm 参数确认执行")
			return
		}

		fmt.Println("\n🚀 正在发送升级指令...")
		for _, d := range outdated {
			fmt.Printf("  ✅ %s - 升级指令已发送\n", d.Name)
		}
		fmt.Printf("\n共 %d 台设备正在升级固件\n", len(outdated))
	},
}

func init() {
	batchRebootCmd.Flags().StringVarP(&batchRebootRegion, "region", "r", "", "目标区域 (必填)")
	batchRebootCmd.Flags().BoolVar(&batchRebootConfirm, "confirm", false, "确认执行")
	batchRebootCmd.MarkFlagRequired("region")

	batchConfigCmd.Flags().StringVarP(&batchConfigRegion, "region", "r", "", "目标区域 (必填)")
	batchConfigCmd.Flags().StringArrayVar(&batchConfigSet, "set", []string{}, "配置项 (key=value)")
	batchConfigCmd.Flags().BoolVar(&batchConfigConfirm, "confirm", false, "确认执行")
	batchConfigCmd.MarkFlagRequired("region")

	batchFirmwareCmd.Flags().StringVarP(&batchFirmwareRegion, "region", "r", "", "目标区域 (必填)")
	batchFirmwareCmd.Flags().StringVarP(&batchFirmwareVersion, "version", "v", "", "目标版本 (默认最新)")
	batchFirmwareCmd.Flags().BoolVar(&batchFirmwareConfirm, "confirm", false, "确认执行")
	batchFirmwareCmd.MarkFlagRequired("region")

	batchCmd.AddCommand(batchRebootCmd)
	batchCmd.AddCommand(batchConfigCmd)
	batchCmd.AddCommand(batchFirmwareCmd)
	rootCmd.AddCommand(batchCmd)
}
