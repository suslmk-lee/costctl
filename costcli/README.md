# costcli - CSP 비용 계산 CLI 도구

NHN Cloud와 같은 CSP의 서비스 사용에 따른 비용을 계산하고 분석하는 명령줄 도구입니다.

## 📊 주요 기능

- **비용 계산**: 실제 사용량과 할인 정책을 기반으로 정확한 비용 산출
- **인스턴스 상태 조회**: 저장된 인스턴스 상태 정보 확인
- **할인 정책 적용**: NHN Cloud의 90일 할인 정책 등 자동 적용
- **다양한 출력 형식**: 테이블 및 JSON 형식 지원
- **기간별 계산**: 일별, 월별, 사용자 정의 기간 비용 계산

## 🚀 빠른 시작

### 1. 빌드

```bash
cd costcli
go build -o costcli main.go
```

### 2. 설정 초기화

```bash
# 설정 파일 생성
./costcli config init
```

### 3. NHN Cloud 인증 정보 설정

```bash
# 프로젝트 ID 설정
./costcli config set nhn.tenant_id "your-tenant-id"

# 사용자명 설정
./costcli config set nhn.username "your-email@example.com"

# API 비밀번호 설정
./costcli config set nhn.password "your-api-password"
```

### 4. 설정 확인

```bash
./costcli config show
```

## 📋 사용법

### 비용 계산

```bash
# 현재 월 비용 계산
./costcli calculate

# 일별 비용 계산
./costcli calculate --period daily

# 월별 비용 계산
./costcli calculate --period monthly

# JSON 형태로 출력
./costcli calculate --output json
```

### 인스턴스 상태 조회

```bash
# 테이블 형태로 조회
./costcli status

# JSON 형태로 조회
./costcli status --output json
```

### 설정 관리

```bash
# 현재 설정 표시
./costcli config show

# 설정 값 변경
./costcli config set [key] [value]

# 설정 파일 초기화
./costcli config init
```

## 🔧 명령어 옵션

### 전역 옵션
- `-c, --config string`: 설정 파일 경로 지정
- `-h, --help`: 도움말 표시

### calculate 명령어
- `-o, --output string`: 출력 형식 (table, json) [기본값: table]
- `-p, --period string`: 계산 기간 (daily, monthly, current) [기본값: current]

### status 명령어
- `-o, --output string`: 출력 형식 (table, json) [기본값: table]

## 📄 예시 출력

### 비용 계산 결과 (테이블 형식)

```
=== 비용 계산 결과 ===
기간: 2025-08-01 00:00 ~ 2025-08-20 13:01
총 인스턴스: 25개
기본 비용: 76800.00 KRW
총 할인: 69120.00 KRW
최종 비용: 7680.00 KRW

인스턴스: web-server (12345678-1234-1234-1234-123456789012)
  - Flavor: c2.m4 (150.00 KRW/시간)
  - 실행 시간: 240.00시간
  - 기본 비용: 36000.00 KRW
  - 할인: 32400.00 KRW
  - 최종 비용: 3600.00 KRW
  - 적용된 할인:
    * NHN 인스턴스 셧다운 90일 할인 (90.0%): 32400.00 KRW
```

### 인스턴스 상태 (테이블 형식)

```
=== 인스턴스 상태 ===
마지막 업데이트: 2025-08-20 13:01:00
총 인스턴스: 25개

인스턴스: web-server (12345678-1234-1234-1234-123456789012)
  - 상태: SHUTDOWN
  - Flavor: c2.m4
  - 생성: 2025-08-01 10:00
  - 마지막 업데이트: 2025-08-20 13:01
  - 총 실행 시간: 14400분
  - 총 정지 시간: 5760분
  - 현재 상태 지속시간: 2h30m0s
```

## 🔗 데이터 소스

costcli는 [cost-collect](../cost-collect/) 모듈에서 수집한 데이터를 사용합니다.

**데이터 수집이 필요한 경우:**
```bash
# 상위 디렉토리의 cost-collect 사용
cd ../cost-collect
./cost-collect once
```

## 🗂️ 설정 파일

설정 파일은 기본적으로 `~/.costctl/config.json`에 저장됩니다.

```json
{
  "nhn_cloud": {
    "tenant_id": "your-tenant-id",
    "username": "your-email@example.com",
    "password": "your-api-password",
    "region": "KR1",
    "identity_url": "https://api-identity-infrastructure.nhncloudservice.com"
  },
  "storage": {
    "data_dir": "/Users/username/.costctl/data",
    "instance_file": "/Users/username/.costctl/data/instances.json",
    "price_file": "/Users/username/.costctl/data/pricing.json"
  }
}
```

## 🚨 트러블슈팅

### 인스턴스 상태 로딩 실패
- cost-collect로 데이터 수집이 선행되어야 합니다
- 설정 파일의 데이터 경로가 올바른지 확인하세요

### 가격 정보 로딩 실패
- pricing.json 파일이 존재하지 않거나 손상되었을 수 있습니다
- cost-collect를 한 번 실행하여 기본 가격 정보를 생성하세요

### 설정 로딩 실패
- `~/.costctl/config.json` 파일의 JSON 형식이 올바른지 확인하세요
- `costcli config init`으로 설정 파일을 재생성하세요

## 📝 라이선스

MIT License