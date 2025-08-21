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
		// 새로운 인스턴스 - updated 시간 기반 히스토리 생성
		s.createInitialHistory(newInstance)
		s.limitHistorySize(newInstance)
		s.Instances[newInstance.ID] = newInstance
		return
	}

	// 기존 인스턴스 업데이트
	s.updateExistingInstance(existingInstance, newInstance)
	s.limitHistorySize(existingInstance)
}

// createInitialHistory는 새로운 인스턴스의 초기 히스토리를 생성합니다.
// updated 시간이 created 시간과 다르면 이전 상태를 추론합니다.
func (s *InstanceStateStorage) createInitialHistory(instance *InstanceState) {
	history := []StatusHistoryItem{}
	
	// updated 시간이 created 시간과 다르면 상태 변경이 있었음을 의미
	if !instance.LastUpdated.Equal(instance.CreatedAt) && 
	   instance.LastUpdated.After(instance.CreatedAt.Add(1*time.Minute)) {
		
		// 이전 상태 추론: 현재 상태의 반대 상태로 가정
		var previousStatus string
		var previousPowerState int
		
		if instance.CurrentStatus == "ACTIVE" {
			previousStatus = "SHUTOFF"
			previousPowerState = 4
		} else {
			previousStatus = "ACTIVE" 
			previousPowerState = 1
		}
		
		// 이전 상태를 created 시간에 추가
		previousHistoryItem := StatusHistoryItem{
			Status:     previousStatus,
			PowerState: previousPowerState,
			Timestamp:  instance.CreatedAt,
		}
		history = append(history, previousHistoryItem)
	}
	
	// 현재 상태 추가
	currentHistoryItem := StatusHistoryItem{
		Status:     instance.CurrentStatus,
		PowerState: instance.CurrentPowerState,
		Timestamp:  instance.LastUpdated,
	}
	history = append(history, currentHistoryItem)
	
	instance.StatusHistory = history
}

// updateExistingInstance는 기존 인스턴스를 업데이트합니다.
func (s *InstanceStateStorage) updateExistingInstance(existingInstance, newInstance *InstanceState) {
	// LastUpdated 시간 비교로 실제 상태 변경 감지
	isRealUpdate := newInstance.LastUpdated.After(existingInstance.LastUpdated)
	
	// 기존 인스턴스 정보 업데이트
	existingInstance.CurrentStatus = newInstance.CurrentStatus
	existingInstance.CurrentPowerState = newInstance.CurrentPowerState
	existingInstance.LastUpdated = newInstance.LastUpdated

	// updated 시간이 실제로 변경된 경우에만 히스토리 추가
	// (API의 updated 시간이 변경되었다는 것은 실제 상태 변경이 있었음을 의미)
	if isRealUpdate {
		newHistoryItem := StatusHistoryItem{
			Status:     newInstance.CurrentStatus,
			PowerState: newInstance.CurrentPowerState,
			Timestamp:  newInstance.LastUpdated,
		}
		existingInstance.StatusHistory = append(existingInstance.StatusHistory, newHistoryItem)
	}
}

// limitHistorySize는 인스턴스의 히스토리를 최신 3개로 제한합니다.
func (s *InstanceStateStorage) limitHistorySize(instance *InstanceState) {
	const maxHistorySize = 3
	if len(instance.StatusHistory) > maxHistorySize {
		instance.StatusHistory = instance.StatusHistory[len(instance.StatusHistory)-maxHistorySize:]
	}
}

func (s *InstanceStateStorage) GetAllInstances() map[string]*InstanceState {
	return s.Instances
}