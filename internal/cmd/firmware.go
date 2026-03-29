package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"
)

var firmwareCmd = &cobra.Command{
	Use:   "firmware",
	Short: "固件管理",
	Long:  `设备固件版本管理和升级`,
}

// firmware list
var firmwareListDevice string

var firmwareListCmd = &cobra.Command{
	Use:   "list",
	Short: "查看固件版本",
	Long:  `查看设备固件版本信息`,
	Run: func(cmd *cobra.Command, args []string) {
		filters := device.DeviceFilters{}
		if firmwareListDevice != "" {
			filters.Keyword = firmwareListDevice
		}

		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t名称\t区域\t当前固件\t状态")
		fmt.Fprintln(w, "--\t----\t----\t--------\t----")

		for _, d := range devices {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				d.ID, d.Name, d.Region, d.Firmware, d.Status)
		}
		w.Flush()

		fmt.Printf("\n共 %d 台设备\n", len(devices))
	},
}

// firmware check
var firmwareCheckRegion string

var firmwareCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "检查固件更新",
	Long:  `检查设备是否有可用的固件更新`,
	Run: func(cmd *cobra.Command, args []string) {
		filters := device.DeviceFilters{Region: firmwareCheckRegion}
		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		latestVersion := "2.3.1"
		var outdated []device.Device

		for _, d := range devices {
			if d.Firmware < latestVersion {
				outdated = append(outdated, d)
			}
		}

		if len(outdated) == 0 {
			fmt.Println("✅ 所有设备固件已是最新版本")
			return
		}

		fmt.Printf("发现 %d 台设备可升级到 %s:\n\n", len(outdated), latestVersion)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t名称\t区域\t当前固件\t可升级到")
		fmt.Fprintln(w, "--\t----\t----\t--------\t--------")

		for _, d := range outdated {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				d.ID, d.Name, d.Region, d.Firmware, latestVersion)
		}
		w.Flush()
	},
}

// firmware upgrade
var (
	firmwareUpgradeVersion  string
	firmwareUpgradeSchedule string
)

var firmwareUpgradeCmd = &cobra.Command{
	Use:   "upgrade <device-id>",
	Short: "升级固件",
	Long:  `升级指定设备的固件版本`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]

		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		if firmwareUpgradeVersion == "" {
			firmwareUpgradeVersion = "2.3.1"
		}

		fmt.Printf("设备: %s (%s)\n", dev.Name, deviceID)
		fmt.Printf("当前固件: %s\n", dev.Firmware)
		fmt.Printf("目标固件: %s\n", firmwareUpgradeVersion)

		if firmwareUpgradeSchedule != "" {
			fmt.Printf("计划升级时间: %s\n", firmwareUpgradeSchedule)
			fmt.Println("⏰ 已安排定时升级任务")
		} else {
			fmt.Println("🚀 正在升级固件...")
			fmt.Println("✅ 升级指令已发送")
			fmt.Printf("设备 %s 正在升级，预计 5 分钟后完成\n", dev.Name)
		}
	},
}

// firmware rollback
var firmwareRollbackCmd = &cobra.Command{
	Use:   "rollback <device-id>",
	Short: "回滚固件",
	Long:  `将设备固件回滚到上一个版本`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]

		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("设备: %s (%s)\n", dev.Name, deviceID)
		fmt.Printf("当前固件: %s\n", dev.Firmware)
		fmt.Println("↩️  正在回滚到上一版本...")
		fmt.Println("✅ 回滚指令已发送")
		fmt.Printf("设备 %s 正在回滚固件\n", dev.Name)
	},
}

func init() {
	firmwareListCmd.Flags().StringVarP(&firmwareListDevice, "device", "d", "", "按设备筛选")
	firmwareCheckCmd.Flags().StringVarP(&firmwareCheckRegion, "region", "r", "", "按区域筛选")
	firmwareUpgradeCmd.Flags().StringVarP(&firmwareUpgradeVersion, "version", "v", "", "目标版本 (默认最新)")
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradeSchedule, "schedule", "", "计划升级时间 (如 02:00)")

	firmwareCmd.AddCommand(firmwareListCmd)
	firmwareCmd.AddCommand(firmwareCheckCmd)
	firmwareCmd.AddCommand(firmwareUpgradeCmd)
	firmwareCmd.AddCommand(firmwareRollbackCmd)
	rootCmd.AddCommand(firmwareCmd)
}
