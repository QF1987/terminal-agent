// ============================================================
// list.go - 设备列表命令
// ============================================================
// 实现 "device-ctl list" 子命令
// 支持按区域、状态、类型、关键字筛选设备
// ============================================================

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"  // tabwriter：格式化表格输出（自动对齐列）

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"  // cobra：CLI 框架
)

// ─── 命令参数 ─────────────────────────────────────────────
// var 块：定义命令的 flag 参数
// 每个 flag 都是一个全局变量，cobra 会自动绑定
var (
	listRegion  string  // --region / -r
	listStatus  string  // --status / -s
	listType    string  // --type / -t
	listKeyword string  // --keyword / -k
)

// ─── 命令定义 ─────────────────────────────────────────────
// cobra.Command：定义一个子命令
var listCmd = &cobra.Command{
	Use:   "list",                    // 命令名称
	Short: "列出设备",                // 简短描述
	Long:  `列出自助购药机设备列表，支持按区域、状态、类型筛选`,  // 详细描述

	// Run：命令执行函数（核心逻辑）
	// cmd：命令对象（可以获取 flag 值）
	// args：位置参数（device-ctl list arg1 arg2 ...）
	Run: func(cmd *cobra.Command, args []string) {
		// 1. 构建筛选条件
		filters := device.DeviceFilters{
			Region:  listRegion,
			Status:  listStatus,
			Type:    listType,
			Keyword: listKeyword,
		}

		// 2. 调用 Store 查询设备
		// Store 是全局变量，在 root.go 中定义
		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)  // 打印错误到 stderr
			os.Exit(1)  // 退出码 1 表示错误
		}

		// 3. 如果没有结果，提示用户
		if len(devices) == 0 {
			fmt.Println("没有找到匹配的设备")
			return
		}

		// 4. 格式化输出表格
		// tabwriter.NewWriter：创建表格写入器
		// 参数：输出目标, 最小单元格宽度, 制表符宽度, 填充字符, 对齐方式
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// 打印表头
		fmt.Fprintln(w, "ID\t名称\t区域\t状态\t固件\t最后心跳")
		fmt.Fprintln(w, "---\t----\t----\t----\t----\t--------")

		// 状态到 emoji 的映射
		statusEmoji := map[string]string{  // map：字典
			device.StatusOnline:      "🟢",
			device.StatusOffline:     "🔴",
			device.StatusError:       "⚠️",
			device.StatusMaintenance: "🔧",
		}

		// 遍历设备，打印每行
		for _, d := range devices {
			emoji := statusEmoji[d.Status]
			if emoji == "" {
				emoji = "❓"
			}
			// %s：字符串占位符
			// %-36s：左对齐，宽度 36
			// d.LastHeartbeat.Format("15:04:05")：格式化时间
			fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s\t%s\n",
				d.ID, d.Name, d.Region, emoji, d.Status,
				d.Firmware, d.LastHeartbeat.Format("15:04:05"))
		}

		// Flush：刷新缓冲区，输出所有内容
		w.Flush()

		// 打印统计
		fmt.Printf("\n共 %d 台设备\n", len(devices))
	},
}

// ─── 初始化 ───────────────────────────────────────────────
// init()：Go 的特殊函数，在包加载时自动执行
// 这里用来注册命令和绑定 flag
func init() {
	// 绑定 flag 到变量
	// StringVarP：字符串 flag，P 表示支持短选项（-r）
	listCmd.Flags().StringVarP(&listRegion, "region", "r", "", "按区域筛选 (华东/华南/华北/西南/华中)")
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "按状态筛选 (online/offline/error/maintenance)")
	listCmd.Flags().StringVarP(&listType, "type", "t", "", "按类型筛选")
	listCmd.Flags().StringVarP(&listKeyword, "keyword", "k", "", "按关键字搜索 (名称/地址)")

	// 把 listCmd 添加到根命令
	// 之后就可以用 "device-ctl list" 调用
	rootCmd.AddCommand(listCmd)
}
