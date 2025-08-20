# cost-collect - CSP 데이터 수집 모듈

NHN Cloud와 같은 CSP의 인스턴스 상태를 주기적으로 수집하고 비용 데이터를 생성하는 독립적인 데이터 수집 모듈입니다.

## 📡 주요 기능

- **실시간 데이터 수집**: NHN Cloud API를 통한 인스턴스 상태 실시간 수집
- **상태 추적**: 인스턴스의 running/shutdown 시간 정밀 추적
- **주기적 모니터링**: 사용자 정의 간격으로 자동 데이터 수집
- **데이터 저장**: JSON 형태로 구조화된 데이터 저장
- **할인 정책 관리**: NHN Cloud 할인 정책 정보 관리

## 🚀 빠른 시작

### 1. 빌드

```bash
cd cost-collect
go build -o cost-collect main.go
```

### 2. 설정 확인

```bash
# 현재 설정 확인 (costcli에서 설정한 정보 사용)
./cost-collect config show
```

### 3. 일회성 데이터 수집

```bash
# 인스턴스 상태를 한 번 수집
./cost-collect once
```

### 4. 지속적인 데이터 수집

```bash
# 기본 간격(15분)으로 데이터 수집 시작
./cost-collect start

# 사용자 정의 간격(10분)으로 수집
./cost-collect start --interval 10

# 데몬 모드로 백그라운드 실행
./cost-collect start --daemon
```

## 📋 사용법

### 데이터 수집

```bash
# 일회성 수집
./cost-collect once

# 지속적인 수집 시작 (Ctrl+C로 중지)
./cost-collect start

# 사용자 정의 간격으로 수집 (분 단위)
./cost-collect start --interval 30

# 데몬 모드로 실행
./cost-collect start --daemon
```

### 수집기 상태 확인

```bash
# 수집기 상태 및 통계 확인
./cost-collect status
```

### 설정 관리

```bash
# 현재 설정 표시
./cost-collect config show
```

## 🔧 명령어 옵션

### 전역 옵션
- `-c, --config string`: 설정 파일 경로 지정
- `-h, --help`: 도움말 표시

### start 명령어
- `-i, --interval int`: 수집 간격 (분 단위) [기본값: 설정파일 값 사용]
- `-d, --daemon`: 데몬 모드로 실행

## 📄 예시 출력

### 일회성 수집 결과

```
데이터 수집을 시작합니다...
2025/08/20 13:59:06 인증 성공. 토큰 만료 시간: 2025-08-20 16:59:06 +0000 UTC
2025/08/20 13:59:07 인스턴스 25개를 발견했습니다.
데이터 수집이 완료되었습니다. (소요시간: 1.2s)
수집된 인스턴스: 25개
실행 중: 12개, 정지: 13개
```

### 지속적인 수집 시작

```
데이터 수집이 시작되었습니다. (간격: 15분)
Ctrl+C로 종료할 수 있습니다.

2025/08/20 13:59:07 인스턴스 25개를 발견했습니다.
2025/08/20 14:14:07 인스턴스 25개를 발견했습니다.
2025/08/20 14:29:07 인스턴스 25개를 발견했습니다.
...
```

### 수집기 상태

```
=== 데이터 수집기 상태 ===
마지막 수집: 2025-08-20 13:59:07
총 인스턴스: 25개
실행 중: 12개
정지 상태: 13개
설정된 수집 간격: 15분
```

## 🗃️ 수집 데이터 구조

### 인스턴스 상태 데이터 (instances.json)

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
          "power_state": 1,
          "is_running": true
        }
      ],
      "running_periods": [
        {
          "start_time": "2025-08-01T10:00:00Z",
          "end_time": "2025-08-01T15:00:00Z",
          "duration_minutes": 300
        }
      ],
      "shutdown_periods": [
        {
          "start_time": "2025-08-01T15:00:00Z",
          "end_time": null,
          "duration_minutes": 480
        }
      ]
    }
  },
  "last_update": "2025-08-20T13:59:07Z",
  "version": "1.0.0"
}
```

### 가격 정보 데이터 (pricing.json)

```json
{
  "version": "1.0.0",
  "currency": "KRW",
  "flavor_pricing": {
    "flavor-id": {
      "flavor_id": "flavor-id",
      "name": "c2.m4",
      "vcpu": 2,
      "ram_mb": 4096,
      "disk_gb": 20,
      "hourly_rate": 150.0,
      "monthly_rate": 108000.0
    }
  },
  "discount_rules": [
    {
      "id": "nhn_shutdown_90day",
      "name": "NHN 인스턴스 셧다운 90일 할인",
      "description": "인스턴스 생성 후 90일 이내에 셧다운 시 90% 할인",
      "type": "shutdown",
      "discount_percent": 90.0,
      "conditions": [
        {
          "type": "instance_age_days",
          "operator": "<=",
          "value": 90
        }
      ],
      "enabled": true
    }
  ]
}
```

## 🔗 관련 도구

이 모듈에서 수집한 데이터는 [costcli](../costcli/) 도구에서 비용 계산에 사용됩니다.

**비용 계산이 필요한 경우:**
```bash
# 상위 디렉토리의 costcli 사용
cd ../costcli
./costcli calculate
```

## ⚙️ 설정 파일

cost-collect는 costcli와 동일한 설정 파일(`~/.costctl/config.json`)을 사용합니다.

**필수 설정 항목:**
- `nhn_cloud.tenant_id`: NHN Cloud 프로젝트 ID
- `nhn_cloud.username`: NHN Cloud 사용자명 (이메일)
- `nhn_cloud.password`: API 비밀번호

**모니터링 설정 항목:**
- `monitor.interval_minutes`: 수집 간격 (분) [기본값: 15]

## 📊 데이터 저장 위치

```
~/.costctl/
├── config.json          # 설정 파일 (costcli와 공유)
└── data/
    ├── instances.json    # 인스턴스 상태 데이터 (이 모듈에서 생성)
    └── pricing.json      # 가격 정보 및 할인 정책 (이 모듈에서 관리)
```

## 🔄 워크플로우

1. **설정**: costcli에서 `config init` 및 인증 정보 설정
2. **수집**: cost-collect로 데이터 수집
3. **분석**: costcli로 비용 계산 및 분석

## 🚨 트러블슈팅

### 인증 실패
- NHN Cloud 인증 정보가 올바른지 확인하세요
- API 비밀번호가 계정 비밀번호와 다른지 확인하세요
- 테넌트 ID가 정확한지 확인하세요

### 네트워크 오류
- 인터넷 연결 상태를 확인하세요
- NHN Cloud API 서비스 상태를 확인하세요
- 방화벽 설정을 확인하세요

### 설정 파일 오류
- `cost-collect config show`로 현재 설정을 확인하세요
- costcli에서 `config init`으로 설정 파일을 재생성하세요

### 데이터 저장 실패
- 데이터 디렉토리에 쓰기 권한이 있는지 확인하세요
- 디스크 공간이 충분한지 확인하세요

## 📈 성능 고려사항

- **수집 간격**: 너무 짧은 간격은 API 호출 제한에 걸릴 수 있습니다
- **인스턴스 수**: 대량의 인스턴스가 있는 경우 수집 시간이 길어질 수 있습니다
- **네트워크**: 안정적인 네트워크 연결이 필요합니다

## 📝 라이선스

MIT License