package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"cost-collect/pkg/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "설정 관리",
	Long:  `cost-collect의 설정을 관리합니다.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "현재 설정 표시",
	Long:  `현재 설정을 표시합니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("설정 로딩 실패: %w", err)
		}

		fmt.Println("=== 현재 설정 ===")
		fmt.Printf("NHN Cloud:\n")
		fmt.Printf("  - Tenant ID: %s\n", cfg.NHNCloud.TenantID)
		fmt.Printf("  - Username: %s\n", cfg.NHNCloud.Username)
		fmt.Printf("  - Password: %s\n", maskPassword(cfg.NHNCloud.Password))
		fmt.Printf("  - Region: %s\n", cfg.NHNCloud.Region)
		fmt.Printf("  - Identity URL: %s\n", cfg.NHNCloud.IdentityURL)
		fmt.Printf("\n모니터링:\n")
		fmt.Printf("  - 간격: %d분\n", cfg.Monitor.IntervalMinutes)
		fmt.Printf("  - 자동 시작: %t\n", cfg.Monitor.AutoStart)
		fmt.Printf("\n저장소:\n")
		fmt.Printf("  - 데이터 디렉토리: %s\n", cfg.Storage.DataDir)
		fmt.Printf("  - 인스턴스 파일: %s\n", cfg.Storage.InstanceFile)
		fmt.Printf("  - 가격 파일: %s\n", cfg.Storage.PriceFile)

		return nil
	},
}


func maskPassword(password string) string {
	if password == "" {
		return "(설정되지 않음)"
	}
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "****" + password[len(password)-2:]
}

func init() {
	configCmd.AddCommand(configShowCmd)
	
	rootCmd.AddCommand(configCmd)
}