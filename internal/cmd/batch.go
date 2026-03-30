// ============================================================
// batch.go - 批量操作命令
// ============================================================
// 实现 "device-ctl batch" 子命令
// 包含 reboot（批量重启）、config（批量配置）、firmware（批量升级）
// 用于同时操作多台设备（按区域）
// ============================================================

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"
)

// ─── batch 父命令 ─────────────────────────────────────────
var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "批量操作",
	Long:  `批量操作命令，包含 reboot（批量重启）、config（批量配置）、firmware（批量升级）子命令`,
}

// ─── batch reboot 子命令 ──────────────────────────────────
var (
	batchRebootRegion  string  // --region：目标区域
	batchRebootConfirm bool    // --confirm：确认执行
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

		// 获取区域设备
		filters := device.DeviceFilters{Region: batchRebootRegion}
		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Printf("没有找到 %s 区域的设备\n", batchRebootRegion)
			return
		}

		// 显示将要重启的设备
		fmt.Printf("\n⚠️  以下 %d 台设备将被重启:\n\n", len(devices))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t名称\t状态")
		fmt.Fprintln(w, "--\t----\t----")

		for _, d := range devices {
			emoji := "🟢"
			if d.Status != "online" {
				emoji = "🔴"
			}
			fmt.Fprintf(w, "%s\t%s\t%s %s\n", d.ID, d.Name, emoji, d.Status)
		}
		w.Flush()

		// 确认提示
		if !batchRebootConfirm {
			fmt.Printf("\n确定要重启这些设备吗？使用 --confirm 执行\n")
			fmt.Printf("命令: device-ctl batch reboot --region %s --confirm\n", batchRebootRegion)
			return
		}

		// 执行批量重启
		fmt.Println("\n🔄 正在批量重启...")
		for _, d := range devices {
			err := Store.RebootDevice(d.ID, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ❌ %s 重启失败: %v\n", d.ID, err)
			} else {
				fmt.Printf("  ✅ %s 重启命令已发送\n", d.ID)
			}
		}

		fmt.Printf("\n✅ 批量重启完成，共处理 %d 台设备\n", len(devices))
	},
}

// ─── batch config 子命令 ──────────────────────────────────
var (
	batchConfigRegion    string  // --region：目标区域
	batchConfigKey       string  // --key：配置项
	batchConfigValue     string  // --value：配置值
	batchConfigConfirm   bool    // --confirm：确认执行
)

var batchConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "批量配置修改",
	Long:  `批量修改指定区域设备的配置`,
	Run: func(cmd *cobra.Command, args []string) {
		if batchConfigRegion == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --region 参数")
			os.Exit(1)
		}
		if batchConfigKey == "" || batchConfigValue == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --key 和 --value 参数")
			os.Exit(1)
		}

		// 获取区域设备
		filters := device.DeviceFilters{Region: batchConfigRegion}
		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n📝 将修改 %s 区域 %d 台设备的配置:\n", batchConfigRegion, len(devices))
		fmt.Printf("   配置项: %s\n", batchConfigKey)
		fmt.Printf("   新值:   %s\n", batchConfigValue)

		if !batchConfigConfirm {
			fmt.Printf("\n使用 --confirm 执行修改\n")
			return
		}

		// 模拟批量配置修改
		fmt.Println("\n🔄 正在批量修改配置...")
		for _, d := range devices {
			fmt.Printf("  ✅ %s 配置已更新\n", d.ID)
		}

		fmt.Printf("\n✅ 批量配置完成，共修改 %d 台设备\n", len(devices))
	},
}

// ─── batch firmware 子命令 ────────────────────────────────
var (
	batchFirmwareRegion  string  // --region：目标区域
	batchFirmwareVersion string  // --version：目标版本
	batchFirmwareConfirm bool    // --confirm：确认执行
)

var batchFirmwareCmd = &cobra.Command{
	Use:   "firmware",
	Short: "批量固件升级",
	Long:  `批量升级指定区域设备的固件`,
	Run: func(cmd *cobra.Command, args []string) {
		if batchFirmwareRegion == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --region 参数")
			os.Exit(1)
		}

		// 获取区域设备
		filters := device.DeviceFilters{Region: batchFirmwareRegion}
		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 筛选需要升级的设备
		targetVersion := batchFirmwareVersion
		if targetVersion == "" {
			targetVersion = "2.3.1"
		}

		var needsUpgrade []device.Device
		for _, d := range devices {
			if d.Firmware != targetVersion {
				needsUpgrade = append(needsUpgrade, d)
			}
		}

		if len(needsUpgrade) == 0 {
			fmt.Println("✅ 该区域所有设备固件都是最新版本")
			return
		}

		fmt.Printf("\n📦 %s 区域有 %d 台设备可升级到 %s:\n\n", batchFirmwareRegion, len(needsUpgrade), targetVersion)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t名称\t当前版本")
		fmt.Fprintln(w, "--\t----\t--------")

		for _, d := range needsUpgrade {
			fmt.Fprintf(w, "%s\t%s\t%s\n", d.ID, d.Name, d.Firmware)
		}
		w.Flush()

		if !batchFirmwareConfirm {
			fmt.Printf("\n使用 --confirm 执行升级\n")
			return
		}

		// 模拟批量升级
		fmt.Println("\n🔄 正在批量升级固件...")
		for _, d := range needsUpgrade {
			fmt.Printf("  ✅ %s 升级完成 (%s → %s)\n", d.ID, d.Firmware, targetVersion)
		}

		fmt.Printf("\n✅ 批量升级完成，共处理 %d 台设备\n", len(needsUpgrade))
	},
}

func init() {
	// batch reboot 的 flag
	batchRebootCmd.Flags().StringVarP(&batchRebootRegion, "region", "r", "", "目标区域（必填）")
	batchRebootCmd.Flags().BoolVarP(&batchRebootConfirm, "confirm", "y", false, "确认执行")

	// batch config 的 flag
	batchConfigCmd.Flags().StringVarP(&batchConfigRegion, "region", "r", "", "目标区域（必填）")
	batchConfigCmd.Flags().StringVarP(&batchConfigKey, "key", "k", "", "配置项（必填）")
	batchConfigCmd.Flags().StringVarP(&batchConfigValue, "value", "v", "", "配置值（必填）")
	batchConfigCmd.Flags().BoolVarP(&batchConfigConfirm, "confirm", "y", false, "确认执行")

	// batch firmware 的 flag
	batchFirmwareCmd.Flags().StringVarP(&batchFirmwareRegion, "region", "r", "", "目标区域（必填）")
	batchFirmwareCmd.Flags().StringVarP(&batchFirmwareVersion, "version", "v", "", "目标版本（默认最新）")
	batchFirmwareCmd.Flags().BoolVarP(&batchFirmwareConfirm, "confirm", "y", false, "确认执行")

	// 添加子命令
	batchCmd.AddCommand(batchRebootCmd)
	batchCmd.AddCommand(batchConfigCmd)
	batchCmd.AddCommand(batchFirmwareCmd)

	rootCmd.AddCommand(batchCmd)
}
