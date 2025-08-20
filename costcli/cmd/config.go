package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"costcli/pkg/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "설정 관리",
	Long:  `costcli의 설정을 관리합니다.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "설정 파일 초기화",
	Long:  `기본 설정 파일을 생성합니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := &config.Config{
			NHNCloud: config.NHNCloudConfig{
				TenantID:    "",
				Username:    "",
				Password:    "",
				Region:      "KR1",
				IdentityURL: "https://api-identity-infrastructure.nhncloudservice.com",
			},
			Monitor: config.MonitorConfig{
				IntervalMinutes: 15,
				AutoStart:       false,
			},
			Storage: config.StorageConfig{},
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("홈 디렉토리 조회 실패: %w", err)
		}

		configDir := filepath.Join(homeDir, ".costctl")
		dataDir := filepath.Join(configDir, "data")
		
		cfg.Storage.DataDir = dataDir
		cfg.Storage.InstanceFile = filepath.Join(dataDir, "instances.json")
		cfg.Storage.PriceFile = filepath.Join(dataDir, "pricing.json")

		configPath := filepath.Join(configDir, "config.json")

		if err := cfg.Save(configPath); err != nil {
			return fmt.Errorf("설정 파일 저장 실패: %w", err)
		}

		fmt.Printf("설정 파일이 생성되었습니다: %s\n", configPath)
		fmt.Println("NHN Cloud 인증 정보를 설정 파일에 추가해주세요:")
		fmt.Println("  - tenant_id: NHN Cloud 프로젝트 ID")
		fmt.Println("  - username: NHN Cloud 사용자명 (이메일)")
		fmt.Println("  - password: API 비밀번호")
		fmt.Println("")
		fmt.Println("설정 방법:")
		fmt.Println("  costcli config set nhn.tenant_id \"your-tenant-id\"")
		fmt.Println("  costcli config set nhn.username \"your-email@example.com\"")
		fmt.Println("  costcli config set nhn.password \"your-api-password\"")

		return nil
	},
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
		fmt.Printf("\n저장소:\n")
		fmt.Printf("  - 데이터 디렉토리: %s\n", cfg.Storage.DataDir)
		fmt.Printf("  - 인스턴스 파일: %s\n", cfg.Storage.InstanceFile)
		fmt.Printf("  - 가격 파일: %s\n", cfg.Storage.PriceFile)

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "설정 값 변경",
	Long:  `설정 값을 변경합니다.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("설정 로딩 실패: %w", err)
		}

		switch key {
		case "nhn.tenant_id":
			cfg.NHNCloud.TenantID = value
		case "nhn.username":
			cfg.NHNCloud.Username = value
		case "nhn.password":
			cfg.NHNCloud.Password = value
		case "nhn.region":
			cfg.NHNCloud.Region = value
		default:
			return fmt.Errorf("알 수 없는 설정 키: %s", key)
		}

		configFilePath := configPath
		if configFilePath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("홈 디렉토리 조회 실패: %w", err)
			}
			configFilePath = filepath.Join(homeDir, ".costctl", "config.json")
		}

		if err := cfg.Save(configFilePath); err != nil {
			return fmt.Errorf("설정 저장 실패: %w", err)
		}

		fmt.Printf("설정이 변경되었습니다: %s = %s\n", key, value)
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
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	
	rootCmd.AddCommand(configCmd)
}