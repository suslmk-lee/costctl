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
	FlavorName          string            `json:"flavor_name"`
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
		FlavorName:        c.pricingStorage.GetFlavorName(instance.FlavorID),
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
	
	// StateHistory 우선 사용
	if len(instance.StateHistory) > 0 {
		totalHours += c.calculateHoursFromStateHistory(instance, startTime, endTime, true)
	} else if len(instance.StatusHistory) > 0 {
		totalHours += c.calculateHoursFromStatusHistory(instance, startTime, endTime, true)
	} else {
		// 기록이 없으면 현재 상태 기반 계산
		totalHours += c.calculateHoursFromCurrentState(instance, startTime, endTime, true)
	}

	return totalHours
}

func (c *CostCalculator) calculateHoursFromStateHistory(instance *storage.InstanceState, startTime, endTime time.Time, forRunning bool) float64 {
	totalHours := 0.0
	
	// 생성시간이 계산 기간 내에 있고 현재 상태가 RUNNING이면 생성~첫 기록까지도 RUNNING으로 간주
	if len(instance.StateHistory) > 0 {
		firstRecord := &instance.StateHistory[0]
		isCurrentlyRunning := instance.CurrentStatus == "ACTIVE" && instance.CurrentPowerState == 1
		
		// 생성시간부터 첫 기록까지의 시간 계산
		if firstRecord.Timestamp.After(instance.CreatedAt) && firstRecord.Timestamp.After(startTime) {
			preStart := instance.CreatedAt
			if preStart.Before(startTime) {
				preStart = startTime
			}
			
			preEnd := firstRecord.Timestamp
			if preEnd.After(endTime) {
				preEnd = endTime
			}
			
			if preEnd.After(preStart) && ((forRunning && isCurrentlyRunning) || (!forRunning && !isCurrentlyRunning)) {
				duration := preEnd.Sub(preStart)
				totalHours += duration.Hours()
			}
		}
	}
	
	// StateHistory 기록 계산
	for i := 0; i < len(instance.StateHistory); i++ {
		current := &instance.StateHistory[i]
		isRecordRunning := current.Status == "ACTIVE" && current.PowerState == 1

		if (forRunning && isRecordRunning) || (!forRunning && !isRecordRunning) {
			periodStart := current.Timestamp
			if periodStart.Before(startTime) {
				periodStart = startTime
			}

			var periodEnd time.Time
			if i+1 < len(instance.StateHistory) {
				periodEnd = instance.StateHistory[i+1].Timestamp
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

func (c *CostCalculator) calculateHoursFromStatusHistory(instance *storage.InstanceState, startTime, endTime time.Time, forRunning bool) float64 {
	totalHours := 0.0

	for i := 0; i < len(instance.StatusHistory); i++ {
		current := &instance.StatusHistory[i]
		isRecordRunning := current.Status == "ACTIVE" && current.PowerState == 1

		if (forRunning && isRecordRunning) || (!forRunning && !isRecordRunning) {
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

func (c *CostCalculator) calculateHoursFromCurrentState(instance *storage.InstanceState, startTime, endTime time.Time, forRunning bool) float64 {
	isCurrentlyRunning := instance.CurrentStatus == "ACTIVE" && instance.CurrentPowerState == 1
	
	if (forRunning && !isCurrentlyRunning) || (!forRunning && isCurrentlyRunning) {
		return 0
	}
	
	// 인스턴스가 계산 기간과 겹치는 시간 계산
	instanceStart := instance.CreatedAt
	if instanceStart.Before(startTime) {
		instanceStart = startTime
	}
	
	instanceEnd := instance.LastUpdated
	if instanceEnd.After(endTime) {
		instanceEnd = endTime
	}
	
	if instanceEnd.After(instanceStart) {
		duration := instanceEnd.Sub(instanceStart)
		return duration.Hours()
	}
	
	return 0
}

func (c *CostCalculator) applyDiscounts(cost *InstanceCost, instance *storage.InstanceState) {
	// Apply legacy hardcoded discount if new format is not available
	if !c.pricingStorage.IsNewFormat() {
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
		return
	}
	
	// Apply discounts from new pricing schema
	defaultCSP := c.pricingStorage.DefaultCSP
	if defaultCSP == "" {
		defaultCSP = "nhn"
	}
	
	// Apply global discount rules first
	for _, rule := range c.pricingStorage.GlobalDiscountRules {
		if rule.Enabled && c.evaluateDiscountRule(&rule, cost, instance) {
			discountAmount := cost.BaseCost * (rule.DiscountPercent / 100.0)
			cost.AppliedDiscounts = append(cost.AppliedDiscounts, DiscountDetail{
				RuleName:        rule.Name,
				DiscountPercent: rule.DiscountPercent,
				DiscountAmount:  discountAmount,
			})
			cost.TotalDiscount += discountAmount
		}
	}
	
	// Apply CSP-specific discount rules
	if csp, exists := c.pricingStorage.CSPs[defaultCSP]; exists {
		for _, rule := range csp.DiscountRules {
			if rule.Enabled && c.evaluateDiscountRule(&rule, cost, instance) {
				discountAmount := cost.BaseCost * (rule.DiscountPercent / 100.0)
				cost.AppliedDiscounts = append(cost.AppliedDiscounts, DiscountDetail{
					RuleName:        rule.Name,
					DiscountPercent: rule.DiscountPercent,
					DiscountAmount:  discountAmount,
				})
				cost.TotalDiscount += discountAmount
			}
		}
	}
}

func (c *CostCalculator) evaluateDiscountRule(rule *storage.DiscountRule, cost *InstanceCost, instance *storage.InstanceState) bool {
	for _, condition := range rule.Conditions {
		if !c.evaluateCondition(&condition, cost, instance) {
			return false
		}
	}
	return true
}

func (c *CostCalculator) evaluateCondition(condition *storage.DiscountCondition, cost *InstanceCost, instance *storage.InstanceState) bool {
	now := time.Now()
	
	switch condition.Type {
	case "instance_age_days":
		ageInDays := int(now.Sub(instance.CreatedAt).Hours() / 24)
		return c.compareValues(float64(ageInDays), condition.Operator, condition.Value)
		
	case "running_hours":
		return c.compareValues(cost.TotalRunningHours, condition.Operator, condition.Value)
		
	case "running_minutes":
		runningMinutes := cost.TotalRunningHours * 60
		return c.compareValues(runningMinutes, condition.Operator, condition.Value)
		
	case "monthly_hours":
		// For monthly hours, we need to calculate total hours in current month
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		monthlyHours := c.calculateRunningHours(instance, startOfMonth, endOfMonth)
		return c.compareValues(monthlyHours, condition.Operator, condition.Value)
		
	case "instance_status":
		expectedStatus := fmt.Sprintf("%v", condition.Value)
		currentStatus := "RUNNING"
		if instance.CurrentStatus == "SHUTOFF" || instance.CurrentPowerState == 4 {
			currentStatus = "SHUTDOWN"
		}
		return currentStatus == expectedStatus
		
	case "shutdown_age_days":
		// NHN Cloud specific: 90일 이내 생성된 인스턴스가 SHUTDOWN 상태일 때만 할인
		ageInDays := int(now.Sub(instance.CreatedAt).Hours() / 24)
		return c.compareValues(float64(ageInDays), condition.Operator, condition.Value)
		
	default:
		return false
	}
}

func (c *CostCalculator) compareValues(actual float64, operator string, expectedValue any) bool {
	var expected float64
	
	// Type assertion for expected value
	switch v := expectedValue.(type) {
	case float64:
		expected = v
	case int:
		expected = float64(v)
	case string:
		// Try to parse string as number
		if parsed, err := fmt.Sscanf(v, "%f", &expected); err != nil || parsed != 1 {
			return false
		}
	default:
		return false
	}
	
	switch operator {
	case "<=":
		return actual <= expected
	case ">=":
		return actual >= expected
	case "<":
		return actual < expected
	case ">":
		return actual > expected
	case "==":
		return actual == expected
	default:
		return false
	}
}

// Legacy discount check for backward compatibility
func (c *CostCalculator) isEligibleForShutdownDiscount(instance *storage.InstanceState) bool {
	// 1. SHUTDOWN 상태 확인
	isShutdown := instance.CurrentStatus == "SHUTOFF" || instance.CurrentPowerState == 4
	if !isShutdown {
		return false
	}
	
	// 2. NHN Cloud 정작: 생성된지 90일 이내인지 확인 (이전 코드가 잘못되었음)
	now := time.Now()
	ageInDays := now.Sub(instance.CreatedAt).Hours() / 24
	isWithin90Days := ageInDays <= 90
	
	return isWithin90Days
}