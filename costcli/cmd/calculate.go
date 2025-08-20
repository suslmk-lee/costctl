package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"costcli/pkg/config"
	"costcli/pkg/calculator"
	"costcli/pkg/storage"
)

var configPath string
var outputFormat string
var period string

var calculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "비용 계산",
	Long:  `현재 인스턴스 사용량과 할인 정책을 기반으로 비용을 계산합니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("설정 로딩 실패: %w", err)
		}

		stateStorage := storage.NewInstanceStateStorage()
		if err := stateStorage.LoadFromFile(cfg.Storage.InstanceFile); err != nil {
			return fmt.Errorf("인스턴스 상태 로딩 실패: %w", err)
		}

		pricingStorage := storage.NewPricingStorage()
		if err := pricingStorage.LoadFromFile(cfg.Storage.PriceFile); err != nil {
			return fmt.Errorf("가격 정보 로딩 실패: %w", err)
		}

		calc := calculator.NewCostCalculator(pricingStorage)

		var summary *calculator.CostSummary
		switch period {
		case "daily":
			summary, err = calc.CalculateDailyEstimate(stateStorage.GetAllInstances())
		case "monthly":
			summary, err = calc.CalculateMonthlyEstimate(stateStorage.GetAllInstances())
		default:
			now := time.Now()
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			summary, err = calc.CalculateTotalCost(stateStorage.GetAllInstances(), startOfMonth, now)
		}

		if err != nil {
			return fmt.Errorf("비용 계산 실패: %w", err)
		}

		switch outputFormat {
		case "json":
			return outputJSON(summary)
		default:
			return outputTable(summary)
		}
	},
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputTable(summary *calculator.CostSummary) error {
	kst, _ := time.LoadLocation("Asia/Seoul")
	
	fmt.Printf("=== 비용 계산 결과 ===\n")
	fmt.Printf("기간: %s ~ %s\n", summary.Period.StartTime.In(kst).Format("2006-01-02 15:04"), summary.Period.EndTime.In(kst).Format("2006-01-02 15:04"))
	fmt.Printf("총 인스턴스: %d개\n", summary.TotalInstances)
	fmt.Printf("기본 비용: %.2f %s\n", summary.TotalBaseCost, summary.Currency)
	fmt.Printf("총 할인: %.2f %s\n", summary.TotalDiscount, summary.Currency)
	fmt.Printf("최종 비용: %.2f %s\n\n", summary.TotalFinalCost, summary.Currency)

	for _, instance := range summary.InstanceCosts {
		fmt.Printf("인스턴스: %s (%s)\n", instance.InstanceName, instance.InstanceID)
		fmt.Printf("  - Flavor: %s (%.2f %s/시간)\n", instance.FlavorID, instance.BaseHourlyRate, summary.Currency)
		fmt.Printf("  - 실행 시간: %.2f시간\n", instance.TotalRunningHours)
		fmt.Printf("  - 기본 비용: %.2f %s\n", instance.BaseCost, summary.Currency)
		fmt.Printf("  - 할인: %.2f %s\n", instance.TotalDiscount, summary.Currency)
		fmt.Printf("  - 최종 비용: %.2f %s\n", instance.FinalCost, summary.Currency)
		
		if len(instance.AppliedDiscounts) > 0 {
			fmt.Printf("  - 적용된 할인:\n")
			for _, discount := range instance.AppliedDiscounts {
				fmt.Printf("    * %s (%.1f%%): %.2f %s\n", discount.RuleName, discount.DiscountPercent, discount.DiscountAmount, summary.Currency)
			}
		}
		fmt.Println()
	}

	return nil
}

func init() {
	rootCmd.AddCommand(calculateCmd)

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "설정 파일 경로")
	calculateCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "출력 형식 (table, json)")
	calculateCmd.Flags().StringVarP(&period, "period", "p", "current", "계산 기간 (daily, monthly, current)")
}