package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type InstanceState struct {
	ID                string              `json:"id"`
	Name              string              `json:"name"`
	FlavorID          string              `json:"flavor_id"`
	CurrentStatus     string              `json:"current_status"`
	CurrentPowerState int                 `json:"current_power_state"`
	CreatedAt         time.Time           `json:"created_at"`
	LastUpdated       time.Time           `json:"last_updated"`
	StatusHistory     []StatusHistoryItem `json:"status_history"`
}

type StatusHistoryItem struct {
	Status     string    `json:"status"`
	PowerState int       `json:"power_state"`
	Timestamp  time.Time `json:"timestamp"`
}

type InstanceStateStorage struct {
	LastUpdate time.Time                  `json:"last_update"`
	Instances  map[string]*InstanceState  `json:"instances"`
}

func NewInstanceStateStorage() *InstanceStateStorage {
	return &InstanceStateStorage{
		Instances: make(map[string]*InstanceState),
	}
}

func (s *InstanceStateStorage) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 파일이 없으면 에러 없이 넘어감
		}
		return fmt.Errorf("파일 읽기 실패: %w", err)
	}
	if len(data) == 0 {
		return nil // 파일이 비어있으면 무시
	}
	if err := json.Unmarshal(data, s); err != nil {
		return fmt.Errorf("JSON 파싱 실패: %w", err)
	}

	return nil
}

func (s *InstanceStateStorage) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 마샬링 실패: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("파일 쓰기 실패: %w", err)
	}

	return nil
}

func (s *InstanceStateStorage) UpdateInstance(newInstance *InstanceState) {
	existingInstance, ok := s.Instances[newInstance.ID]
	if !ok {
		// 새로운 인스턴스
		newHistoryItem := StatusHistoryItem{
			Status:     newInstance.CurrentStatus,
			PowerState: newInstance.CurrentPowerState,
			Timestamp:  newInstance.LastUpdated,
		}
		newInstance.StatusHistory = []StatusHistoryItem{newHistoryItem}
		s.Instances[newInstance.ID] = newInstance
		return
	}

	// 기존 인스턴스 업데이트
	existingInstance.CurrentStatus = newInstance.CurrentStatus
	existingInstance.CurrentPowerState = newInstance.CurrentPowerState
	existingInstance.LastUpdated = newInstance.LastUpdated

	// 히스토리 추가 로직
	shouldAddHistory := true
	if len(existingInstance.StatusHistory) > 0 {
		lastHistory := existingInstance.StatusHistory[len(existingInstance.StatusHistory)-1]
		if lastHistory.Status == newInstance.CurrentStatus && lastHistory.PowerState == newInstance.CurrentPowerState {
			// 상태 변경이 없으면 히스토리를 추가하지 않음
			shouldAddHistory = false
		}
	}

	if shouldAddHistory {
		newHistoryItem := StatusHistoryItem{
			Status:     newInstance.CurrentStatus,
			PowerState: newInstance.CurrentPowerState,
			Timestamp:  newInstance.LastUpdated,
		}
		existingInstance.StatusHistory = append(existingInstance.StatusHistory, newHistoryItem)
	}

	// 히스토리를 최신 3개로 제한
	if len(existingInstance.StatusHistory) > 3 {
		existingInstance.StatusHistory = existingInstance.StatusHistory[len(existingInstance.StatusHistory)-3:]
	}
}

func (s *InstanceStateStorage) GetAllInstances() map[string]*InstanceState {
	return s.Instances
}