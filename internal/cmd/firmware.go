// ============================================================
// firmware.go - 固件管理命令
// ============================================================
// 实现 "device-ctl firmware" 子命令
// 包含 check（检查更新）和 upgrade（升级固件）
// ============================================================

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"
)

var firmwareRegion string  // --region：按区域筛选

// ─── firmware 父命令 ──────────────────────────────────────
var firmwareCmd = &cobra.Command{
	Use:   "firmware",
	Short: "固件管理",
	Long:  `固件管理命令，包含 check（检查更新）和 upgrade（升级）子命令`,
}

// ─── firmware check 子命令 ─────────────────────────────────
var firmwareCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "检查固件更新",
	Long:  `检查设备是否有可用的固件更新`,
	Run: func(cmd *cobra.Command, args []string) {
		// 构建筛选条件
		filters := device.DeviceFilters{Region: firmwareRegion}

		// 查询设备
		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 模拟检查固件版本（假设最新是 2.3.1）
		latestVersion := "2.3.1"
		var needsUpgrade []device.Device

		for _, d := range devices {
			// 简单版本比较：如果不是最新版本，标记为需要升级
			if d.Firmware != latestVersion {
				needsUpgrade = append(needsUpgrade, d)
			}
		}

		if len(needsUpgrade) == 0 {
			fmt.Println("✅ 所有设备固件都是最新版本")
			return
		}

		// 输出可升级的设备
		fmt.Printf("\n📦 以下 %d 台设备可升级固件（最新版本: %s）\n\n", len(needsUpgrade), latestVersion)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t名称\t当前版本\t可升级至")
		fmt.Fprintln(w, "--\t----\t--------\t--------")

		for _, d := range needsUpgrade {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				d.ID, d.Name, d.Firmware, latestVersion)
		}
		w.Flush()

		fmt.Printf("\n💡 使用 'device-ctl firmware upgrade <device-id>' 升级单台设备\n")
		fmt.Printf("💡 使用 'device-ctl batch firmware --region %s' 批量升级\n", firmwareRegion)
	},
}

// ─── firmware upgrade 子命令 ───────────────────────────────
var (
	firmwareUpgradeVersion  string  // --version：目标版本
	firmwareUpgradeSchedule string  // --schedule：计划时间
)

var firmwareUpgradeCmd = &cobra.Command{
	Use:   "upgrade <device-id>",
	Short: "升级固件",
	Long:  `升级指定设备的固件版本`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]

		// 检查设备
		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 确定目标版本
		targetVersion := firmwareUpgradeVersion
		if targetVersion == "" {
			targetVersion = "2.3.1"  // 默认最新版本
		}

		// 输出升级信息
		fmt.Printf("🔄 升级设备 %s (%s)\n", dev.Name, deviceID)
		fmt.Printf("   当前版本: %s\n", dev.Firmware)
		fmt.Printf("   目标版本: %s\n", targetVersion)

		if firmwareUpgradeSchedule != "" {
			fmt.Printf("   计划时间: %s\n", firmwareUpgradeSchedule)
			fmt.Println("✅ 升级任务已计划")
		} else {
			fmt.Println("🔄 正在执行升级...")
			fmt.Println("✅ 固件升级完成")
		}
	},
}

func init() {
	// check 的 flag
	firmwareCheckCmd.Flags().StringVarP(&firmwareRegion, "region", "r", "", "按区域筛选")

	// upgrade 的 flag
	firmwareUpgradeCmd.Flags().StringVarP(&firmwareUpgradeVersion, "version", "v", "", "目标版本（默认最新）")
	firmwareUpgradeCmd.Flags().StringVarP(&firmwareUpgradeSchedule, "schedule", "s", "", "计划升级时间（如 02:00）")

	// 添加子命令
	firmwareCmd.AddCommand(firmwareCheckCmd)
	firmwareCmd.AddCommand(firmwareUpgradeCmd)

	rootCmd.AddCommand(firmwareCmd)
}
