package cmd

import (
	"fmt"
	"sort"
	"time"

	"costcli/pkg/config"
	"costcli/pkg/storage"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "인스턴스 상태 조회",
	Long:  `저장된 인스턴스 상태 정보를 조회합니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("설정 로딩 실패: %w", err)
		}

		stateStorage := storage.NewInstanceStateStorage()
		if err := stateStorage.LoadFromFile(cfg.Storage.InstanceFile); err != nil {
			return fmt.Errorf("인스턴스 상태 로딩 실패: %w", err)
		}

		instances := stateStorage.GetAllInstances()

		switch outputFormat {
		case "json":
			return outputJSON(instances)
		default:
			return outputInstanceStatus(instances, stateStorage.LastUpdate)
		}
	},
}

func outputInstanceStatus(instancesMap map[string]*storage.InstanceState, lastUpdate time.Time) error {
	kst, _ := time.LoadLocation("Asia/Seoul")
	p := message.NewPrinter(language.Korean)

	// Convert map to slice for sorting
	instances := make([]*storage.InstanceState, 0, len(instancesMap))
	for _, instance := range instancesMap {
		instances = append(instances, instance)
	}

	// Sort instances by creation time
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].CreatedAt.Before(instances[j].CreatedAt)
	})

	fmt.Printf("=== 인스턴스 상태 ===\n")
	fmt.Printf("마지막 업데이트: %s\n", lastUpdate.In(kst).Format("2006-01-02 15:04:05"))
	fmt.Printf("총 인스턴스: %d개\n\n", len(instances))

	for _, instance := range instances {
		status := "UNKNOWN"
		if instance.CurrentStatus == "ACTIVE" && instance.CurrentPowerState == 1 {
			status = "RUNNING"
		} else if instance.CurrentStatus == "SHUTOFF" || instance.CurrentPowerState == 4 {
			status = "SHUTDOWN"
		}

		fmt.Printf("인스턴스: %s (%s)\n", instance.Name, instance.ID)
		fmt.Printf("  - 상태: %s\n", status)
		fmt.Printf("  - Flavor: %s\n", instance.FlavorID)
		fmt.Printf("  - 생성: %s\n", instance.CreatedAt.In(kst).Format("2006-01-02 15:04"))
		fmt.Printf("  - 마지막 업데이트: %s\n", instance.LastUpdated.In(kst).Format("2006-01-02 15:04"))
		p.Printf("  - 총 실행 시간: %d분\n", instance.GetTotalRunningMinutes())
		p.Printf("  - 총 정지 시간: %d분\n", instance.GetTotalShutdownMinutes())
		fmt.Printf("  - 현재 상태 지속시간: %s\n", instance.GetCurrentStateDuration().Truncate(time.Second).String())
		fmt.Println()
	}

	return nil
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "출력 형식 (table, json)")
}
