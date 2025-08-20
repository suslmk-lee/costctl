package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cost-collect",
	Short: "CSP 비용 데이터 수집 도구",
	Long:  `NHN Cloud와 같은 CSP의 인스턴스 상태를 수집하고 비용 데이터를 생성하는 도구입니다.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cost-collect - CSP 비용 데이터 수집 도구")
		fmt.Println("사용법: cost-collect [command]")
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 설정 파일 및 플래그 초기화
}