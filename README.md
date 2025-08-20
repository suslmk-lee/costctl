# costctl - CSP 비용 관리 도구

NHN Cloud와 같은 CSP의 서비스 사용에 따른 비용을 산출하고 관리하는 도구 모음입니다.

## 프로젝트 구조

이 프로젝트는 두 개의 독립적인 모듈로 구성되어 있습니다:

### 📊 costcli - 비용 계산 CLI 도구
비용 계산, 분석, 리포팅을 담당하는 명령줄 도구입니다.

### 📡 cost-collect - 데이터 수집 모듈
NHN Cloud 인스턴스 상태를 주기적으로 수집하는 독립적인 모듈입니다.

## 주요 기능

- **인스턴스 모니터링**: NHN Cloud 인스턴스 상태를 주기적으로 모니터링
- **상태 추적**: 인스턴스의 running/shutdown 시간 추적
- **할인 정책 적용**: NHN Cloud의 90일 할인 정책 등 자동 적용
- **비용 계산**: 실제 사용량과 할인 정책을 기반으로 정확한 비용 산출
- **JSON 데이터 저장**: 모든 데이터를 JSON 형태로 저장하여 관리

## 빠른 시작

### 1. 프로젝트 빌드

```bash
git clone <repository>
cd costctl01

# 비용 계산 CLI 도구 빌드
cd costcli && go build -o costcli main.go && cd ..

# 데이터 수집 모듈 빌드
cd cost-collect && go build -o cost-collect main.go && cd ..
```

### 2. 설정 초기화

```bash
# 설정 파일 생성
./costcli/costcli config init
```

## 초기 설정

### 1. 설정 파일 생성

```bash
./costctl config init
```

### 3. NHN Cloud 인증 정보 설정

```bash
# NHN Cloud 프로젝트 ID 설정
./costcli/costcli config set nhn.tenant_id "your-tenant-id"

# 사용자명 (이메일) 설정
./costcli/costcli config set nhn.username "your-email@example.com"

# API 비밀번호 설정
./costcli/costcli config set nhn.password "your-api-password"
```

### 4. 현재 설정 확인

```bash
./costcli/costcli config show
```

## 사용법

### 📡 데이터 수집 (cost-collect)

```bash
# 일회성 데이터 수집
./cost-collect/cost-collect once

# 지속적인 데이터 수집 시작 (Ctrl+C로 중지)
./cost-collect/cost-collect start

# 데이터 수집기 상태 확인
./cost-collect/cost-collect status

# 사용자 정의 간격으로 수집 시작 (예: 10분 간격)
./cost-collect/cost-collect start --interval 10
```

### 📊 비용 계산 및 분석 (costcli)

```bash
# 인스턴스 상태 조회
./costcli/costcli status

# JSON 형태로 조회
./costcli/costcli status -o json

# 현재 월 비용 계산
./costcli/costcli calculate

# 일별 비용 계산
./costcli/costcli calculate -p daily

# 월별 비용 계산
./costcli/costcli calculate -p monthly

# JSON 형태로 출력
./costcli/costcli calculate -o json
```

## 워크플로우

1. **설정**: `costcli config init`으로 초기 설정
2. **데이터 수집**: `cost-collect start`로 지속적인 모니터링 시작
3. **비용 분석**: `costcli calculate`로 비용 계산 및 분석

## NHN Cloud 인증 정보 획득

### 1. 테넌트 ID 획득
- NHN Cloud 콘솔 > Compute > Instance > 관리 페이지에서 확인

### 2. API 비밀번호 설정
- NHN Cloud 계정 비밀번호와 별도로 API 전용 비밀번호 설정 필요

### 3. 사용자명
- NHN Cloud 계정의 이메일 주소 사용

## 할인 정책

### NHN Cloud 90일 할인 정책
- 인스턴스 생성 후 90일 이내에 shutdown 시 90% 할인 적용
- 자동으로 인스턴스 생성일과 shutdown 기간을 체크하여 할인 적용

### 장기 사용 할인
- 30일 이상 연속 사용 시 10% 할인 (예시)

## 데이터 저장 위치

기본적으로 다음 위치에 데이터가 저장됩니다:

```
~/.costctl/
├── config.json          # 설정 파일
└── data/
    ├── instances.json    # 인스턴스 상태 데이터
    └── pricing.json      # 가격 정보 및 할인 정책
```

## 명령어 옵션

### costcli 명령어 옵션

#### 전역 옵션
- `-c, --config`: 설정 파일 경로 지정

#### calculate 명령어
- `-o, --output`: 출력 형식 (table, json)
- `-p, --period`: 계산 기간 (daily, monthly, current)

#### status 명령어
- `-o, --output`: 출력 형식 (table, json)

### cost-collect 명령어 옵션

#### 전역 옵션
- `-c, --config`: 설정 파일 경로 지정

#### start 명령어
- `-i, --interval`: 수집 간격 (분 단위)
- `-d, --daemon`: 데몬 모드로 실행

## 예시 출력

### 비용 계산 결과
```
=== 비용 계산 결과 ===
기간: 2024-01-01 00:00 ~ 2024-01-31 23:59
총 인스턴스: 2개
기본 비용: 72000.00 KRW
총 할인: 64800.00 KRW
최종 비용: 7200.00 KRW

인스턴스: web-server (12345678-1234-1234-1234-123456789012)
  - Flavor: c2.m4 (150.00 KRW/시간)
  - 실행 시간: 240.00시간
  - 기본 비용: 36000.00 KRW
  - 할인: 32400.00 KRW
  - 최종 비용: 3600.00 KRW
  - 적용된 할인:
    * NHN 인스턴스 셧다운 90일 할인 (90.0%): 32400.00 KRW
```

### 인스턴스 상태
```
=== 인스턴스 상태 ===
마지막 업데이트: 2024-01-15 14:30:00
총 인스턴스: 2개

인스턴스: web-server (12345678-1234-1234-1234-123456789012)
  - 상태: SHUTDOWN
  - Flavor: c2.m4
  - 생성: 2024-01-01 10:00
  - 마지막 업데이트: 2024-01-15 14:30
  - 총 실행 시간: 14400분
  - 총 정지 시간: 5760분
  - 현재 상태 지속시간: 2h30m0s
```

## 트러블슈팅

### 인증 오류
- NHN Cloud 인증 정보가 올바른지 확인
- API 비밀번호가 계정 비밀번호와 다른지 확인
- 테넌트 ID가 정확한지 확인

### 네트워크 오류
- 인터넷 연결 상태 확인
- NHN Cloud API 서비스 상태 확인

### 설정 파일 오류
- `costctl config show`로 현재 설정 확인
- `costctl config init`으로 설정 파일 재생성

## 라이선스

MIT License