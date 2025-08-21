package monitor

import (
	"context"
	"log"
	"time"

	"cost-collect/pkg/config"
	"cost-collect/pkg/nhncloud"
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
	nhnClient       *nhncloud.Client
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

	nhnClient := nhncloud.NewClient(&cfg.NHNCloud)

	return &Monitor{
		config:          cfg,
		instanceStorage: storage,
		nhnClient:       nhnClient,
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

	// 1. Fetch data from NHN Cloud API
	instances, err := m.nhnClient.GetInstances()
	if err != nil {
		log.Printf("오류: NHN Cloud API에서 데이터를 가져오지 못했습니다: %v", err)
		return
	}

	log.Printf("NHN Cloud API에서 %d개의 인스턴스를 가져왔습니다.", len(instances))

	// 2. Update instance storage
	for _, instance := range instances {
		m.instanceStorage.UpdateInstance(instance)
	}

	// 3. Save to file
	if err := m.instanceStorage.SaveToFile(m.config.Storage.InstanceFile); err != nil {
		log.Printf("오류: 인스턴스 데이터를 파일에 저장하지 못했습니다: %v", err)
	}

	// 4. Recalculate summary stats
	m.recalculateStats()

	log.Printf("업데이트 완료. 총 %d개 인스턴스 (실행 중: %d개, 정지: %d개)", 
		m.stats.TotalInstances, m.stats.RunningInstances, m.stats.ShutdownInstances)
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

