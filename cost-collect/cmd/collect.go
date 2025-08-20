package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"cost-collect/pkg/config"
	"cost-collect/pkg/monitor"
	"github.com/spf13/cobra"
)

var configPath string
var interval int

// getPidFilePath는 PID 파일의 경로를 반환합니다.
func getPidFilePath(cfg *config.Config) string {
	return filepath.Join(cfg.Storage.DataDir, "cost-collect.pid")
}

var collectCmd = &cobra.Command{
	Use:   "start",
	Short: "데이터 수집 시작",
	Long:  `NHN Cloud 인스턴스 상태 데이터 수집을 시작합니다. 백그라운드 실행을 원하시면 'start &' 형태로 실행하세요.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("설정 로딩 실패: %w", err)
		}

		pidFilePath := getPidFilePath(cfg)

		if _, err := os.Stat(pidFilePath); err == nil {
			data, err := os.ReadFile(pidFilePath)
			if err == nil {
				pid, _ := strconv.Atoi(string(data))
				if processExists(pid) {
					return fmt.Errorf("수집기가 이미 실행 중입니다. (PID: %d)", pid)
				}
			}
		}

		if cfg.NHNCloud.TenantID == "" || cfg.NHNCloud.Username == "" || cfg.NHNCloud.Password == "" {
			return fmt.Errorf("NHN Cloud 인증 정보가 설정되지 않았습니다. 설정 파일을 확인해주세요")
		}

		if interval > 0 {
			cfg.Monitor.IntervalMinutes = interval
		}

		// PID 파일 작성
		pid := os.Getpid()
		if err := os.WriteFile(pidFilePath, []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("PID 파일 작성 실패: %w", err)
		}

		m := monitor.NewMonitor(cfg)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := m.Start(ctx); err != nil {
			return fmt.Errorf("데이터 수집 시작 실패: %w", err)
		}

		cleanup := func() {
			fmt.Println("\n데이터 수집을 종료합니다...")
			cancel()
			m.Stop()
			if err := os.Remove(pidFilePath); err != nil {
				log.Printf("PID 파일 삭제 실패: %v", err)
			}
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		fmt.Printf("데이터 수집이 시작되었습니다. (PID: %d)\n", pid)
		fmt.Println("Ctrl+C 또는 'stop' 명령어로 종료할 수 있습니다.")

		<-sigChan // Block until signal
		cleanup()

		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "데이터 수집 중지",
	Long:  `실행 중인 데이터 수집기 프로세스를 중지합니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("설정 로딩 실패: %w", err)
		}

		pidFilePath := getPidFilePath(cfg)
		data, err := os.ReadFile(pidFilePath)
		if err != nil {
			return fmt.Errorf("수집기가 실행 중이지 않거나 PID 파일을 찾을 수 없습니다.")
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			return fmt.Errorf("PID 파일이 손상되었습니다: %w", err)
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("프로세스(PID: %d)를 찾을 수 없습니다: %w", pid, err)
		}

		if err := process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("프로세스(PID: %d)에 종료 신호를 보내는 데 실패했습니다: %w", pid, err)
		}

		fmt.Printf("프로세스(PID: %d)에 종료 신호를 보냈습니다. 5초간 확인합니다...\n", pid)

		for i := 0; i < 5; i++ {
			if !processExists(pid) {
				fmt.Println("수집기가 성공적으로 중지되었습니다.")
				return nil
			}
			time.Sleep(1 * time.Second)
		}

		return fmt.Errorf("프로세스(PID: %d)가 5초 내에 종료되지 않았습니다. 강제 종료가 필요할 수 있습니다.", pid)
	},
}

func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

var onceCmd = &cobra.Command{
	Use:   "once",
	Short: "일회성 데이터 수집",
	Long:  `인스턴스 상태를 한 번만 수집합니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("설정 로딩 실패: %w", err)
		}

		if cfg.NHNCloud.TenantID == "" || cfg.NHNCloud.Username == "" || cfg.NHNCloud.Password == "" {
			return fmt.Errorf("NHN Cloud 인증 정보가 설정되지 않았습니다")
		}

		m := monitor.NewMonitor(cfg)

		fmt.Println("데이터 수집을 시작합니다...")
		start := time.Now()

		if err := m.ForceUpdate(); err != nil {
			return fmt.Errorf("데이터 수집 실패: %w", err)
		}

		elapsed := time.Since(start)
		stats := m.GetStats()

		fmt.Printf("데이터 수집이 완료되었습니다. (소요시간: %v)\n", elapsed)
		fmt.Printf("수집된 인스턴스: %d개\n", stats.TotalInstances)
		fmt.Printf("실행 중: %d개, 정지: %d개\n", stats.RunningInstances, stats.ShutdownInstances)

		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "수집기 상태 확인",
	Long:  `데이터 수집기의 현재 상태를 확인합니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("설정 로딩 실패: %w", err)
		}

		if _, err := os.Stat(cfg.Storage.InstanceFile); os.IsNotExist(err) {
			fmt.Println("수집된 데이터가 없습니다.")
			return nil
		}

		m := monitor.NewMonitor(cfg)

		stats := m.GetStats()

		fmt.Println("=== 데이터 수집기 상태 ===")
		fmt.Printf("마지막 수집: %s\n", stats.LastUpdate.Format("2006-01-02 15:04:05"))
		fmt.Printf("총 인스턴스: %d개\n", stats.TotalInstances)
		fmt.Printf("실행 중: %d개\n", stats.RunningInstances)
		fmt.Printf("정지 상태: %d개\n", stats.ShutdownInstances)
		fmt.Printf("설정된 수집 간격: %d분\n", cfg.Monitor.IntervalMinutes)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(onceCmd)
	rootCmd.AddCommand(statusCmd)

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "설정 파일 경로")

	collectCmd.Flags().IntVarP(&interval, "interval", "i", 0, "수집 간격 (분, 0이면 설정파일 값 사용)")
}