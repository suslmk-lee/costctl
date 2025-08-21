package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	NHNCloud NHNCloudConfig `json:"nhn_cloud"`
	Monitor  MonitorConfig  `json:"monitor"`
	Storage  StorageConfig  `json:"storage"`
}

type NHNCloudConfig struct {
	TenantID    string `json:"tenant_id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Region      string `json:"region"`
	IdentityURL string `json:"identity_url"`
	ComputeURL  string `json:"compute_url"`
}

type MonitorConfig struct {
	IntervalMinutes int  `json:"interval_minutes"`
	AutoStart       bool `json:"auto_start"`
}

type StorageConfig struct {
	DataDir      string `json:"data_dir"`
	InstanceFile string `json:"instance_file"`
	PriceFile    string `json:"price_file"`
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("홈 디렉토리 조회 실패: %w", err)
		}
		configPath = filepath.Join(homeDir, ".costctl", "config.json")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("설정 파일 파싱 실패: %w", err)
	}

	return &config, nil
}

func (c *Config) Save(configPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("설정 직렬화 실패: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %w", err)
	}

	return nil
}