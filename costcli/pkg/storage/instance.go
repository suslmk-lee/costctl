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
	StateHistory      []StateHistoryItem  `json:"state_history"`
	UpdatedAt         *time.Time          `json:"updated_at,omitempty"`
}

type StatusHistoryItem struct {
	Status     string    `json:"status"`
	PowerState int       `json:"power_state"`
	Timestamp  time.Time `json:"timestamp"`
}

type StateHistoryItem struct {
	Status     string    `json:"status"`
	PowerState int       `json:"power_state"`
	Timestamp  time.Time `json:"timestamp"`
	IsRunning  bool      `json:"is_running"`
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
		return fmt.Errorf("파일 읽기 실패: %w", err)
	}

	if err := json.Unmarshal(data, s); err != nil {
		return fmt.Errorf("JSON 파싱 실패: %w", err)
	}

	return nil
}

func (s *InstanceStateStorage) GetAllInstances() map[string]*InstanceState {
	return s.Instances
}

func (i *InstanceState) GetTotalRunningMinutes() int {
	// state_history 우선 사용
	if len(i.StateHistory) > 0 {
		return i.calculateFromStateHistory(true)
	}
	
	// StatusHistory 사용
	if len(i.StatusHistory) > 0 {
		return i.calculateFromStatusHistory(true)
	}
	
	// 기록이 없으면 현재 상태와 updated 시점 활용
	return i.calculateFromCurrentState(true)
}

func (i *InstanceState) calculateFromStateHistory(forRunning bool) int {
	if len(i.StateHistory) == 0 {
		return 0
	}
	
	totalMinutes := 0
	
	// 첫 번째 기록 이전의 시간을 계산 (생성시간부터 첫 모니터링까지)
	firstRecord := &i.StateHistory[0]
	isCurrentlyRunning := i.CurrentStatus == "ACTIVE" && i.CurrentPowerState == 1
	
	// 현재 상태에 따라 생성시간부터 첫 기록까지의 시간을 해당 상태로 간주
	if firstRecord.Timestamp.After(i.CreatedAt) {
		preMonitoringDuration := firstRecord.Timestamp.Sub(i.CreatedAt)
		
		// RUNNING 상태면서 실행시간을 구하거나, SHUTDOWN 상태면서 정지시간을 구할 때
		if (forRunning && isCurrentlyRunning) || (!forRunning && !isCurrentlyRunning) {
			totalMinutes += int(preMonitoringDuration.Minutes())
		}
	}
	
	// 기존 state_history 기반 계산
	for j := 0; j < len(i.StateHistory); j++ {
		current := &i.StateHistory[j]
		isRecordRunning := current.Status == "ACTIVE" && current.PowerState == 1
		
		if (forRunning && isRecordRunning) || (!forRunning && !isRecordRunning) {
			var endTime time.Time
			if j+1 < len(i.StateHistory) {
				endTime = i.StateHistory[j+1].Timestamp
			} else {
				endTime = i.LastUpdated
			}
			
			duration := endTime.Sub(current.Timestamp)
			totalMinutes += int(duration.Minutes())
		}
	}
	
	return totalMinutes
}

func (i *InstanceState) calculateFromStatusHistory(forRunning bool) int {
	totalMinutes := 0
	
	for j := 0; j < len(i.StatusHistory); j++ {
		current := &i.StatusHistory[j]
		isCurrentlyRunning := current.Status == "ACTIVE" && current.PowerState == 1
		
		if (forRunning && isCurrentlyRunning) || (!forRunning && !isCurrentlyRunning) {
			var endTime time.Time
			if j+1 < len(i.StatusHistory) {
				endTime = i.StatusHistory[j+1].Timestamp
			} else {
				endTime = i.LastUpdated
			}
			
			duration := endTime.Sub(current.Timestamp)
			totalMinutes += int(duration.Minutes())
		}
	}
	
	return totalMinutes
}

func (i *InstanceState) calculateFromCurrentState(forRunning bool) int {
	isCurrentlyRunning := i.CurrentStatus == "ACTIVE" && i.CurrentPowerState == 1
	
	if (forRunning && !isCurrentlyRunning) || (!forRunning && isCurrentlyRunning) {
		return 0
	}
	
	// UpdatedAt이 있으면 상태 변경 시점부터 계산
	if i.UpdatedAt != nil && (forRunning == isCurrentlyRunning) {
		// RUNNING 상태면 UpdatedAt부터 현재까지가 실행시간
		// SHUTDOWN 상태면 UpdatedAt부터 현재까지가 정지시간
		duration := i.LastUpdated.Sub(*i.UpdatedAt)
		return int(duration.Minutes())
	}
	
	// UpdatedAt이 없으면 생성시간부터 계산 (기존 로직)
	if (forRunning && isCurrentlyRunning) || (!forRunning && !isCurrentlyRunning) {
		duration := i.LastUpdated.Sub(i.CreatedAt)
		return int(duration.Minutes())
	}
	
	return 0
}

func (i *InstanceState) GetTotalShutdownMinutes() int {
	// state_history 우선 사용
	if len(i.StateHistory) > 0 {
		return i.calculateFromStateHistory(false)
	}
	
	// StatusHistory 사용
	if len(i.StatusHistory) > 0 {
		return i.calculateFromStatusHistory(false)
	}
	
	// 기록이 없으면 현재 상태와 updated 시점 활용
	return i.calculateFromCurrentState(false)
}

func (i *InstanceState) GetCurrentStateDuration() time.Duration {
	currentRunning := i.CurrentStatus == "ACTIVE" && i.CurrentPowerState == 1
	
	// state_history에서 마지막 상태 변경 시점 찾기
	if len(i.StateHistory) > 0 {
		// 역순으로 찾아서 상태가 바뀐 마지막 시점 확인
		for j := len(i.StateHistory) - 1; j >= 0; j-- {
			record := &i.StateHistory[j]
			recordRunning := record.Status == "ACTIVE" && record.PowerState == 1
			
			// 현재 상태와 다른 기록을 찾으면, 그 다음 기록이 현재 상태 시작점
			if recordRunning != currentRunning {
				if j+1 < len(i.StateHistory) {
					stateChangePoint := i.StateHistory[j+1].Timestamp
					return i.LastUpdated.Sub(stateChangePoint)
				}
				break
			}
		}
		
		// 모든 기록이 현재 상태와 동일하면, 첫 번째 기록부터 현재까지
		// 단, 생성시간이 더 이르면 생성시간부터 계산
		firstRecord := i.StateHistory[0]
		if firstRecord.Timestamp.After(i.CreatedAt) {
			return i.LastUpdated.Sub(i.CreatedAt)
		}
		return i.LastUpdated.Sub(firstRecord.Timestamp)
	}
	
	// StatusHistory 사용 (기존 로직)
	if len(i.StatusHistory) > 0 {
		for j := len(i.StatusHistory) - 1; j >= 0; j-- {
			record := &i.StatusHistory[j]
			recordRunning := record.Status == "ACTIVE" && record.PowerState == 1
			
			if recordRunning != currentRunning {
				if j+1 < len(i.StatusHistory) {
					stateChangePoint := i.StatusHistory[j+1].Timestamp
					return i.LastUpdated.Sub(stateChangePoint)
				}
				break
			}
		}
		
		firstRecord := i.StatusHistory[0]
		if firstRecord.Timestamp.After(i.CreatedAt) {
			return i.LastUpdated.Sub(i.CreatedAt)
		}
		return i.LastUpdated.Sub(firstRecord.Timestamp)
	}
	
	// UpdatedAt이 있으면 상태 변경 시점부터 계산
	if i.UpdatedAt != nil {
		return i.LastUpdated.Sub(*i.UpdatedAt)
	}
	
	// 기록이 없으면 생성시간부터 현재까지
	return i.LastUpdated.Sub(i.CreatedAt)
}