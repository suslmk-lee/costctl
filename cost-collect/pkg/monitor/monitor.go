package monitor

import (
	"context"
	"log"
	"time"

	"cost-collect/pkg/config"
	"cost-collect/pkg/storage"
)

// Stats holds monitoring statistics.
type Stats struct {
	LastUpdate        time.Time
	TotalInstances    int
	RunningInstances  int
	ShutdownInstances int
}

// Monitor manages the collection of instance data.
type Monitor struct {
	config          *config.Config
	instanceStorage *storage.InstanceStateStorage
	stats           Stats
	ticker          *time.Ticker
	done            chan bool
}

// NewMonitor creates a new Monitor.
func NewMonitor(cfg *config.Config) *Monitor {
	storage := storage.NewInstanceStateStorage()
	if err := storage.LoadFromFile(cfg.Storage.InstanceFile); err != nil {
		log.Printf("경고: 기존 인스턴스 데이터 로딩 실패 (%s): %v", cfg.Storage.InstanceFile, err)
	}

	return &Monitor{
		config:          cfg,
		instanceStorage: storage,
		done:            make(chan bool),
	}
}

// Start begins the monitoring loop.
func (m *Monitor) Start(ctx context.Context) error {
	log.Println("모니터링을 시작합니다...")
	// Perform an initial update
	m.update()

	// Start the ticker for periodic updates
	m.ticker = time.NewTicker(time.Duration(m.config.Monitor.IntervalMinutes) * time.Minute)

	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.update()
			case <-m.done:
				m.ticker.Stop()
				return
			}
		}
	}()

	return nil
}

// Stop halts the monitoring loop.
func (m *Monitor) Stop() {
	m.done <- true
}

// ForceUpdate performs a single, immediate update.
func (m *Monitor) ForceUpdate() error {
	m.update()
	return nil
}

// GetStats returns the current statistics.
func (m *Monitor) GetStats() Stats {
	// Ensure stats are up-to-date
	m.recalculateStats()
	return m.stats
}

// update contains the core logic to fetch data from the cloud provider.
func (m *Monitor) update() {
	log.Println("데이터를 수집하고 업데이트합니다...")

	// 1. Fetch data from cloud provider (Simulated)
	// In a real implementation, this would be a call to NHN Cloud API
	simulatedInstances := m.simulateFetchFromCloud()

	// 2. Update instance storage
	for _, instance := range simulatedInstances {
		m.instanceStorage.UpdateInstance(instance)
	}

	// 3. Save to file
	if err := m.instanceStorage.SaveToFile(m.config.Storage.InstanceFile); err != nil {
		log.Printf("오류: 인스턴스 데이터를 파일에 저장하지 못했습니다: %v", err)
	}

	// 4. Recalculate summary stats
	m.recalculateStats()

	log.Printf("업데이트 완료. 총 %d개 인스턴스.", m.stats.TotalInstances)
}

func (m *Monitor) recalculateStats() {
	instances := m.instanceStorage.GetAllInstances()
	m.stats.TotalInstances = len(instances)
	m.stats.RunningInstances = 0
	m.stats.ShutdownInstances = 0
	for _, instance := range instances {
		if instance.CurrentStatus == "ACTIVE" {
			m.stats.RunningInstances++
		} else {
			m.stats.ShutdownInstances++
		}
	}
	m.stats.LastUpdate = time.Now()
}

// simulateFetchFromCloud simulates fetching instance data from the cloud.
func (m *Monitor) simulateFetchFromCloud() []*storage.InstanceState {
	// This is a placeholder. It should be replaced with actual API calls.
	// We'll create a dummy instance for demonstration.
	return []*storage.InstanceState{
		{
			ID:                "fe37386d-f2aa-451c-8590-83dc45601a3c",
			Name:              "karmada-host1",
			FlavorID:          "edc79d63-98c3-4b77-a2d4-482d70e6b554",
			CurrentStatus:     "ACTIVE",
			CurrentPowerState: 1,
			CreatedAt:         time.Now().Add(-48 * time.Hour),
			LastUpdated:       time.Now(),
		},
		{
			ID:                "2c45b048-189a-4d50-adc4-826542472476",
			Name:              "ImageHubGw",
			FlavorID:          "a4b6a0f7-aeff-4d78-a8d5-7de9f007012d",
			CurrentStatus:     "SHUTOFF",
			CurrentPowerState: 4,
			CreatedAt:         time.Now().Add(-100 * time.Hour),
			LastUpdated:       time.Now(),
		},
	}
}