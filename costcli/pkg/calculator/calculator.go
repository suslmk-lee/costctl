package calculator

import (
	"fmt"
	"time"

	"costcli/pkg/storage"
)

type CostCalculator struct {
	pricingStorage *storage.PricingStorage
}

type CostSummary struct {
	Period           TimePeriod      `json:"period"`
	TotalInstances   int             `json:"total_instances"`
	TotalBaseCost    float64         `json:"total_base_cost"`
	TotalDiscount    float64         `json:"total_discount"`
	TotalFinalCost   float64         `json:"total_final_cost"`
	Currency         string          `json:"currency"`
	InstanceCosts    []InstanceCost  `json:"instance_costs"`
}

type TimePeriod struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type InstanceCost struct {
	InstanceID          string            `json:"instance_id"`
	InstanceName        string            `json:"instance_name"`
	FlavorID            string            `json:"flavor_id"`
	BaseHourlyRate      float64           `json:"base_hourly_rate"`
	TotalRunningHours   float64           `json:"total_running_hours"`
	BaseCost            float64           `json:"base_cost"`
	TotalDiscount       float64           `json:"total_discount"`
	FinalCost           float64           `json:"final_cost"`
	AppliedDiscounts    []DiscountDetail  `json:"applied_discounts"`
}

type DiscountDetail struct {
	RuleName        string  `json:"rule_name"`
	DiscountPercent float64 `json:"discount_percent"`
	DiscountAmount  float64 `json:"discount_amount"`
}

func NewCostCalculator(pricingStorage *storage.PricingStorage) *CostCalculator {
	return &CostCalculator{
		pricingStorage: pricingStorage,
	}
}

func (c *CostCalculator) CalculateTotalCost(instances map[string]*storage.InstanceState, startTime, endTime time.Time) (*CostSummary, error) {
	summary := &CostSummary{
		Period: TimePeriod{
			StartTime: startTime,
			EndTime:   endTime,
		},
		TotalInstances: len(instances),
		Currency:       "KRW",
		InstanceCosts:  make([]InstanceCost, 0, len(instances)),
	}

	for _, instance := range instances {
		cost, err := c.calculateInstanceCost(instance, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("인스턴스 %s 비용 계산 실패: %w", instance.ID, err)
		}

		summary.InstanceCosts = append(summary.InstanceCosts, *cost)
		summary.TotalBaseCost += cost.BaseCost
		summary.TotalDiscount += cost.TotalDiscount
		summary.TotalFinalCost += cost.FinalCost
	}

	return summary, nil
}

func (c *CostCalculator) CalculateDailyEstimate(instances map[string]*storage.InstanceState) (*CostSummary, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	return c.CalculateTotalCost(instances, startOfDay, endOfDay)
}

func (c *CostCalculator) CalculateMonthlyEstimate(instances map[string]*storage.InstanceState) (*CostSummary, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	
	return c.CalculateTotalCost(instances, startOfMonth, endOfMonth)
}

func (c *CostCalculator) calculateInstanceCost(instance *storage.InstanceState, startTime, endTime time.Time) (*InstanceCost, error) {
	flavorPrice, exists := c.pricingStorage.GetFlavorPrice(instance.FlavorID)
	if !exists {
		return nil, fmt.Errorf("flavor %s에 대한 가격 정보를 찾을 수 없습니다", instance.FlavorID)
	}

	runningHours := c.calculateRunningHours(instance, startTime, endTime)
	baseCost := runningHours * flavorPrice.HourlyPrice

	cost := &InstanceCost{
		InstanceID:        instance.ID,
		InstanceName:      instance.Name,
		FlavorID:          instance.FlavorID,
		BaseHourlyRate:    flavorPrice.HourlyPrice,
		TotalRunningHours: runningHours,
		BaseCost:          baseCost,
		AppliedDiscounts:  []DiscountDetail{},
	}

	c.applyDiscounts(cost, instance)

	cost.FinalCost = cost.BaseCost - cost.TotalDiscount

	return cost, nil
}

func (c *CostCalculator) calculateRunningHours(instance *storage.InstanceState, startTime, endTime time.Time) float64 {
	totalHours := 0.0

	for i := 0; i < len(instance.StatusHistory); i++ {
		current := &instance.StatusHistory[i]

		if current.Status == "ACTIVE" && current.PowerState == 1 {
			periodStart := current.Timestamp
			if periodStart.Before(startTime) {
				periodStart = startTime
			}

			var periodEnd time.Time
			if i+1 < len(instance.StatusHistory) {
				periodEnd = instance.StatusHistory[i+1].Timestamp
			} else {
				periodEnd = instance.LastUpdated
			}

			if periodEnd.After(endTime) {
				periodEnd = endTime
			}

			if periodEnd.After(periodStart) {
				duration := periodEnd.Sub(periodStart)
				totalHours += duration.Hours()
			}
		}
	}

	return totalHours
}

func (c *CostCalculator) applyDiscounts(cost *InstanceCost, instance *storage.InstanceState) {
	if c.isEligibleForShutdownDiscount(instance) {
		discountPercent := 90.0
		discountAmount := cost.BaseCost * (discountPercent / 100.0)
		
		cost.AppliedDiscounts = append(cost.AppliedDiscounts, DiscountDetail{
			RuleName:        "NHN 인스턴스 셧다운 90일 할인",
			DiscountPercent: discountPercent,
			DiscountAmount:  discountAmount,
		})
		
		cost.TotalDiscount += discountAmount
	}
}

func (c *CostCalculator) isEligibleForShutdownDiscount(instance *storage.InstanceState) bool {
	return instance.CurrentStatus == "SHUTOFF" || instance.CurrentPowerState == 4
}