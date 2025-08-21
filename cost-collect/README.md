# cost-collect - CSP 데이터 수집 모듈

NHN Cloud와 같은 CSP의 인스턴스 상태를 주기적으로 수집하고 비용 데이터를 생성하는 독립적인 데이터 수집 모듈입니다.

## 🚀 사용법 (Usage)

### 1. 데이터 수집 시작 (포그라운드)

터미널에 연결된 상태로 데이터 수집을 시작합니다. `Ctrl+C`로 언제든지 중지할 수 있습니다.

```bash
./cost-collect start
```

### 2. 데이터 수집 시작 (백그라운드)

`&` 연산자를 사용하여 백그라운드에서 데이터 수집을 시작합니다. 이 방법으로 실행하면 터미널을 계속 사용할 수 있습니다.

```bash
./cost-collect start &
```

### 3. 수집 중지

백그라운드에서 실행 중인 데이터 수집기를 중지합니다.

```bash
./cost-collect stop
```

### 4. 수집기 상태 확인

수집된 데이터의 최신 상태와 통계를 확인합니다.

```bash
./cost-collect status
```

### 5. 일회성 데이터 수집

데이터를 한 번만 수집하고 즉시 종료합니다.

```bash
./cost-collect once
```

### 6. 설정 확인

현재 애플리케이션의 설정을 표시합니다.

```bash
./cost-collect config show
```

## 🔧 명령어 옵션

### 전역 옵션
- `-c, --config string`: 설정 파일 경로 지정
- `-h, --help`: 도움말 표시

### start 명령어
- `-i, --interval int`: 수집 간격 (분 단위) [기본값: 설정파일 값 사용]

## 🗃️ 수집 데이터 구조

### 인스턴스 상태 데이터 (instances.json)

`state_history`는 각 인스턴스의 상태 변경 이력을 저장하며, 최신 3개의 기록만 유지됩니다.

```json
{
  "instances": {
    "instance-id": {
      "id": "instance-id",
      "name": "instance-name",
      "flavor_id": "flavor-id",
      "current_status": "ACTIVE",
      "current_power_state": 1,
      "created_at": "2025-08-01T10:00:00Z",
      "last_updated": "2025-08-20T13:59:07Z",
      "state_history": [
        {
          "timestamp": "2025-08-20T13:59:07Z",
          "status": "ACTIVE",
          "power_state": 1
        }
      ]
    }
  },
  "last_update": "2025-08-20T13:59:07Z"
}
```

## 🔗 관련 도구

이 모듈에서 수집한 데이터는 [costcli](../costcli/) 도구에서 비용 계산에 사용될 수 있습니다.

## ⚙️ 설정 파일

cost-collect는 costcli와 동일한 설정 파일(`~/.costctl/config.json`)을 사용할 수 있습니다.

**필수 설정 항목:**
- `nhn_cloud.tenant_id`: NHN Cloud 프로젝트 ID
- `nhn_cloud.username`: NHN Cloud 사용자명 (이메일)
- `nhn_cloud.password`: API 비밀번호

**모니터링 설정 항목:**
- `monitor.interval_minutes`: 수집 간격 (분) [기본값: 15]

## 📊 데이터 저장 위치

```
~/.costctl/
├── config.json          # 설정 파일
└── data/
    ├── instances.json    # 인스턴스 상태 데이터
    ├── pricing.json      # 가격 정보 데이터
    └── cost-collect.pid  # 백그라운드 실행 시 생성되는 PID 파일
```
