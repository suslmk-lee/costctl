package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

type FlavorPrice struct {
	FlavorID    string  `json:"flavor_id"`
	HourlyPrice float64 `json:"hourly_price"`
	Currency    string  `json:"currency"`
}

type PricingStorage struct {
	Flavors map[string]*FlavorPrice `json:"flavors"`
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

	if err := json.Unmarshal(data, p); err != nil {
		return fmt.Errorf("JSON 파싱 실패: %w", err)
	}

	return nil
}

func (p *PricingStorage) GetFlavorPrice(flavorID string) (*FlavorPrice, bool) {
	price, exists := p.Flavors[flavorID]
	return price, exists
}