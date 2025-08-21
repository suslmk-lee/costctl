package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

// Legacy pricing structure for backward compatibility
type FlavorPrice struct {
	FlavorID    string  `json:"flavor_id"`
	HourlyPrice float64 `json:"hourly_price"`
	Currency    string  `json:"currency"`
}

type LegacyPricingStorage struct {
	Flavors map[string]*FlavorPrice `json:"flavors"`
}

// New pricing structure
type InstanceType struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	DisplayName        string             `json:"display_name"`
	VCPU              int                `json:"vcpu"`
	MemoryMB          int                `json:"memory_mb"`
	DiskGB            int                `json:"disk_gb"`
	NetworkPerformance string             `json:"network_performance"`
	Pricing           map[string]Pricing `json:"pricing"`
	Availability      []string           `json:"availability"`
}

type Pricing struct {
	Hourly  float64 `json:"hourly"`
	Monthly float64 `json:"monthly"`
	Yearly  float64 `json:"yearly"`
}

type Region struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type DiscountCondition struct {
	Type     string `json:"type"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

type DiscountRule struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Type            string              `json:"type"`
	Conditions      []DiscountCondition `json:"conditions"`
	DiscountPercent float64             `json:"discount_percent"`
	EffectiveFrom   string              `json:"effective_from"`
	EffectiveTo     *string             `json:"effective_to"`
	Enabled         bool                `json:"enabled"`
}

type CSPProvider struct {
	Name            string                   `json:"name"`
	DisplayName     string                   `json:"display_name"`
	APIURL          string                   `json:"api_url"`
	DefaultCurrency string                   `json:"default_currency"`
	Regions         []Region                 `json:"regions"`
	InstanceTypes   map[string]InstanceType  `json:"instance_types"`
	DiscountRules   []DiscountRule           `json:"discount_rules"`
}

type CurrencyConversion struct {
	Rates       map[string]float64 `json:"rates"`
	LastUpdated string             `json:"last_updated"`
}

type GlobalSettings struct {
	CacheTTLSeconds     int                 `json:"cache_ttl_seconds"`
	DefaultDuration     string              `json:"default_duration"`
	SupportedFormats    []string            `json:"supported_formats"`
	CurrencyConversion  CurrencyConversion  `json:"currency_conversion"`
}

type NewPricingSchema struct {
	Version             string                  `json:"version"`
	LastUpdated         string                  `json:"last_updated"`
	DefaultCSP          string                  `json:"default_csp"`
	CSPs                map[string]CSPProvider `json:"csps"`
	GlobalSettings      GlobalSettings          `json:"global_settings"`
	GlobalDiscountRules []DiscountRule          `json:"global_discount_rules"`
}

// Main pricing storage that supports both legacy and new formats
type PricingStorage struct {
	// Legacy format
	Flavors map[string]*FlavorPrice `json:"flavors,omitempty"`
	
	// New format (embedded for direct access)
	*NewPricingSchema
}

func NewPricingStorage() *PricingStorage {
	return &PricingStorage{
		Flavors: make(map[string]*FlavorPrice),
	}
}

func (p *PricingStorage) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("파일 읽기 실패: %w", err)
	}

	// Try to determine format by checking for new schema structure
	var structureCheck struct {
		Version string             `json:"version"`
		CSPs    map[string]any     `json:"csps"`
		Flavors map[string]any     `json:"flavors"`
	}
	
	if err := json.Unmarshal(data, &structureCheck); err == nil {
		// Check if it's new format (has CSPs field and version >= 2.0.0)
		if structureCheck.CSPs != nil && len(structureCheck.CSPs) > 0 {
			// New format with CSPs
			var newSchema NewPricingSchema
			if err := json.Unmarshal(data, &newSchema); err != nil {
				return fmt.Errorf("새 형식 JSON 파싱 실패: %w", err)
			}
			p.NewPricingSchema = &newSchema
		} else if structureCheck.Flavors != nil {
			// Legacy format with flavors
			var legacy LegacyPricingStorage
			if err := json.Unmarshal(data, &legacy); err != nil {
				return fmt.Errorf("기존 형식 JSON 파싱 실패: %w", err)
			}
			p.Flavors = legacy.Flavors
		} else {
			return fmt.Errorf("알 수 없는 JSON 형식")
		}
	} else {
		return fmt.Errorf("JSON 구조 분석 실패: %w", err)
	}

	return nil
}

// GetFlavorPrice supports both legacy and new formats
func (p *PricingStorage) GetFlavorPrice(flavorID string) (*FlavorPrice, bool) {
	// Check legacy format first
	if len(p.Flavors) > 0 {
		price, exists := p.Flavors[flavorID]
		return price, exists
	}
	
	// Check new format
	if p.NewPricingSchema != nil {
		defaultCSP := p.DefaultCSP
		if defaultCSP == "" {
			defaultCSP = "nhn" // fallback
		}
		
		if csp, exists := p.CSPs[defaultCSP]; exists {
			if instanceType, exists := csp.InstanceTypes[flavorID]; exists {
				currency := csp.DefaultCurrency
				if pricing, exists := instanceType.Pricing[currency]; exists {
					return &FlavorPrice{
						FlavorID:    flavorID,
						HourlyPrice: pricing.Hourly,
						Currency:    currency,
					}, true
				}
			}
		}
	}
	
	return nil, false
}

// GetInstanceTypeInfo returns detailed instance information from new format
func (p *PricingStorage) GetInstanceTypeInfo(cspName, instanceTypeID string) (*InstanceType, bool) {
	if p.NewPricingSchema == nil {
		return nil, false
	}
	
	if csp, exists := p.CSPs[cspName]; exists {
		if instanceType, exists := csp.InstanceTypes[instanceTypeID]; exists {
			return &instanceType, true
		}
	}
	
	return nil, false
}

// IsNewFormat returns true if using new pricing schema
func (p *PricingStorage) IsNewFormat() bool {
	return p.NewPricingSchema != nil
}

// GetFlavorName returns human-readable flavor name
func (p *PricingStorage) GetFlavorName(flavorID string) string {
	// Check new format first for detailed names
	if p.NewPricingSchema != nil {
		defaultCSP := p.DefaultCSP
		if defaultCSP == "" {
			defaultCSP = "nhn"
		}
		
		if csp, exists := p.CSPs[defaultCSP]; exists {
			if instanceType, exists := csp.InstanceTypes[flavorID]; exists {
				return instanceType.Name
			}
		}
	}
	
	// Fallback to flavor ID for legacy format
	return flavorID
}