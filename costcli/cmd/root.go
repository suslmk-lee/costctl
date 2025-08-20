package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "costcli",
	Short: "CSP 비용 계산 CLI 도구",
	Long:  `NHN Cloud와 같은 CSP의 서비스 사용에 따른 비용을 계산하고 분석하는 CLI 도구입니다.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("costcli - CSP 비용 계산 CLI 도구")
		fmt.Println("사용법: costcli [command]")
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 설정 파일 및 플래그 초기화
}