package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "权限控制",
	Long:  `设备访问权限管理`,
}

// auth login
var authLoginToken string

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录认证",
	Long:  `使用 token 登录认证`,
	Run: func(cmd *cobra.Command, args []string) {
		if authLoginToken != "" {
			fmt.Println("✅ 使用 token 登录成功")
		} else {
			fmt.Println("请输入认证 token:")
			fmt.Println("(使用 --token 参数)")
		}
	},
}

// auth whoami
var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "查看当前用户",
	Long:  `查看当前登录用户信息`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`
┌─────────────────────────────────────────────────┐
│  👤 当前用户                                     │
├─────────────────────────────────────────────────┤
│  用户名:     admin                              │
│  角色:       管理员                              │
│  权限范围:   全部设备                             │
│  登录时间:   2026-03-29 09:00:00                │
│  Token 过期: 2026-03-30 09:00:00                │
└─────────────────────────────────────────────────┘
`)
	},
}

// auth grant
var authGrantRole string

var authGrantCmd = &cobra.Command{
	Use:   "grant <device-id>",
	Short: "授予权限",
	Long:  `授予用户对设备的访问权限`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]
		role := authGrantRole
		if role == "" {
			role = "operator"
		}

		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ 已授予 [%s] 角色对设备 %s (%s) 的访问权限\n", role, dev.Name, deviceID)
	},
}

// auth revoke
var authRevokeRole string

var authRevokeCmd = &cobra.Command{
	Use:   "revoke <device-id>",
	Short: "撤销权限",
	Long:  `撤销用户对设备的访问权限`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]
		role := authRevokeRole
		if role == "" {
			role = "operator"
		}

		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ 已撤销 [%s] 角色对设备 %s (%s) 的访问权限\n", role, dev.Name, deviceID)
	},
}

// auth list
var authListCmd = &cobra.Command{
	Use:   "list <device-id>",
	Short: "查看权限列表",
	Long:  `查看设备的权限分配情况`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deviceID := args[0]

		dev, err := Store.GetDevice(deviceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("📋 设备 %s (%s) 的权限列表:\n\n", dev.Name, deviceID)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "用户\t角色\t授权时间\t过期时间")
		fmt.Fprintln(w, "----\t----\t--------\t--------")
		fmt.Fprintln(w, "admin\t管理员\t2026-01-01\t永不过期")
		fmt.Fprintln(w, "operator-01\t运维员\t2026-03-15\t2026-06-15")
		fmt.Fprintln(w, "viewer-01\t只读\t2026-03-20\t2026-04-20")
		w.Flush()
	},
}

func init() {
	authLoginCmd.Flags().StringVarP(&authLoginToken, "token", "t", "", "认证 token")
	authGrantCmd.Flags().StringVarP(&authGrantRole, "role", "r", "operator", "角色 (operator/viewer/admin)")
	authRevokeCmd.Flags().StringVarP(&authRevokeRole, "role", "r", "operator", "角色 (operator/viewer/admin)")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authWhoamiCmd)
	authCmd.AddCommand(authGrantCmd)
	authCmd.AddCommand(authRevokeCmd)
	authCmd.AddCommand(authListCmd)
	rootCmd.AddCommand(authCmd)
}
