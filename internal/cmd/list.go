package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/QF1987/terminal-agent-go/internal/device"
	"github.com/spf13/cobra"
)

var (
	listRegion  string
	listStatus  string
	listType    string
	listKeyword string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出设备",
	Long:  `列出自助购药机设备列表，支持按区域、状态、类型筛选`,
	Run: func(cmd *cobra.Command, args []string) {
		filters := device.DeviceFilters{
			Region:  listRegion,
			Status:  listStatus,
			Type:    listType,
			Keyword: listKeyword,
		}

		devices, err := Store.ListDevices(filters)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Println("没有找到匹配的设备")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t名称\t区域\t状态\t固件\t最后心跳")
		fmt.Fprintln(w, "---\t----\t----\t----\t----\t--------")

		statusEmoji := map[string]string{
			device.StatusOnline:      "🟢",
			device.StatusOffline:     "🔴",
			device.StatusError:       "⚠️",
			device.StatusMaintenance: "🔧",
		}

		for _, d := range devices {
			emoji := statusEmoji[d.Status]
			if emoji == "" {
				emoji = "❓"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s %s\t%s\t%s\n",
				d.ID, d.Name, d.Region, emoji, d.Status,
				d.Firmware, d.LastHeartbeat.Format("15:04:05"))
		}
		w.Flush()

		fmt.Printf("\n共 %d 台设备\n", len(devices))
	},
}

func init() {
	listCmd.Flags().StringVarP(&listRegion, "region", "r", "", "按区域筛选 (华东/华南/华北/西南/华中)")
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "按状态筛选 (online/offline/error/maintenance)")
	listCmd.Flags().StringVarP(&listType, "type", "t", "", "按类型筛选")
	listCmd.Flags().StringVarP(&listKeyword, "keyword", "k", "", "按关键字搜索 (名称/地址)")
	rootCmd.AddCommand(listCmd)
}
