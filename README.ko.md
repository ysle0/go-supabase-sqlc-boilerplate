# Go, Supabase + SQLC 보일러플레이트

Vertical Slice Architecture와 Supabase를 활용한 프로덕션 레디 Go 마이크로서비스 보일러플레이트

[![English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![한국어](https://img.shields.io/badge/lang-한국어-red.svg)](README.ko.md)
[![Français](https://img.shields.io/badge/lang-Français-yellow.svg)](README.fr.md)
[![Nederlands](https://img.shields.io/badge/lang-Nederlands-orange.svg)](README.nl.md)

## 주요 기능

- **마이크로서비스 아키텍처**: 명확한 관심사 분리를 가진 독립적인 서비스들
- **Vertical Slice Architecture**: 기능별로 완결된 구조, 높은 응집도와 낮은 결합도
- **Supabase 통합**: PostgreSQL 데이터베이스 관리 및 마이그레이션을 Supabase로 간편하게 처리
- **현대적인 스택**: Go 1.25, Chi v5, PostgreSQL (Supabase), Redis
- **실시간 통신**: WebSocket 지원
- **이벤트 드리븐**: Redis Streams 기반 이벤트 처리
- **타입 안전성**: SQLC를 통한 타입 안전 SQL 쿼리
- **우아한 종료**: 적절한 리소스 정리 및 연결 처리

## 프로젝트 구조

```
.
├── servers/                    # Go 마이크로서비스
│   ├── cmd/                    # 서비스 진입점
│   │   ├── api/                # REST API 서비스 (포트 8080)
│   │   ├── ws/                 # WebSocket 서비스 (포트 8081)
│   │   ├── stats/              # 통계 서비스 (포트 8084)
│   │   └── logging/            # 로깅 서비스 (포트 8082)
│   ├── internal/
│   │   ├── feature/            # 비즈니스 기능 (Vertical Slice)
│   │   ├── shared/             # 공유 인프라
│   │   ├── stats/              # 통계 처리
│   │   ├── logging/            # 로깅 서비스
│   │   └── ws_example/         # WebSocket 핸들러
│   └── test/                   # 통합 테스트
├── supabase/                   # Supabase 데이터베이스 관리
│   ├── schemas/                # 데이터베이스 스키마 정의
│   ├── queries/                # SQLC 쿼리 파일
│   ├── migrations/             # 데이터베이스 마이그레이션 (Supabase CLI)
│   └── config.toml             # Supabase 프로젝트 설정
└── script/                     # 코드 생성 및 데이터베이스 관리 스크립트
    ├── gen-sqlc.bash           # SQLC 코드 생성
    ├── gen-proto.bash          # Protocol Buffer 코드 생성
    ├── gen-typing-sb.bash      # TypeScript 타입 생성
    ├── reset-local-sb.bash     # Supabase 로컬 DB 리셋
    └── reset-remote-sb.bash    # Supabase 원격 DB 리셋
```

## 기술 스택

### 코어
- **Go 1.25**: 제네릭 지원
- **Chi v5**: 경량 HTTP 라우터
- **gorilla/websocket**: WebSocket 구현

### 데이터 레이어
- **Supabase**: PostgreSQL 호스팅 및 데이터베이스 관리 플랫폼
- **PostgreSQL**: 메인 데이터베이스 (Supabase에서 호스팅)
- **SQLC**: Go와 TypeScript를 위한 타입 안전 SQL 코드 생성
  - SQL 쿼리로부터 타입 안전 Go 코드 생성
  - Supabase Edge Functions를 위한 TypeScript 코드 생성
  - **주의**: TypeScript 생성은 `:exec`, `:execrows`, `:execresult`, `:batchexec` 어노테이션을 지원하지 않습니다 (대신 `:one` 또는 `:many` 사용)
- **pgx/v5**: 고성능 PostgreSQL 드라이버
- **Supabase CLI**: 로컬 개발 환경 및 마이그레이션 관리

### 캐싱 & 메시징
- **Redis**: 인메모리 데이터 스토어
- **Redis Streams**: 이벤트 스트리밍

## 빠른 시작

### 사전 요구사항

- Go 1.25+
- Supabase CLI ([설치 가이드](https://supabase.com/docs/guides/cli))
- Redis 7+
- Docker (로컬 Supabase 실행용)

### 설치

```bash
# 1. 저장소 클론
git clone https://github.com/your-org/go-monorepo-boilerplate.git
cd go-monorepo-boilerplate

# 2. Supabase 로컬 환경 시작
supabase start
# PostgreSQL 연결 정보가 출력됩니다

# 3. 환경 변수 설정
cd servers
cp .env.example .env
# .env 파일에 Supabase 연결 정보 입력

# 4. 의존성 설치
go mod download

# 5. SQL 쿼리로부터 타입 안전 코드 생성
cd ..
./script/gen-sqlc.bash
# 다음을 생성합니다:
# - 백엔드 서비스를 위한 타입 안전 Go 코드 (servers/internal/sql/)
# - Supabase Edge Functions를 위한 TypeScript 타입 (supabase/functions/_shared/queries/)

# 6. (선택) 데이터베이스 리셋이 필요한 경우
./script/reset-local-sb.bash
```

### 서비스 실행

```bash
cd servers

# API 서비스
go run ./cmd/api

# WebSocket 서비스
go run ./cmd/ws

# 통계 서비스
go run ./cmd/stats

# 로깅 서비스
go run ./cmd/logging
```

## 개발

### 빌드

```bash
cd servers
go build ./...                    # 전체 빌드
go build ./cmd/api                # 특정 서비스 빌드
```

### 테스트

```bash
cd servers
go test ./...                     # 전체 테스트
go test -cover ./...              # 커버리지 포함
go test -v ./internal/feature/... # 특정 패키지 테스트
```

### 코드 생성

```bash
# 저장소 루트에서 실행
./script/gen-sqlc.bash           # SQL로부터 타입 안전 Go 및 TypeScript 코드 생성
                                 # - Go: servers/internal/sql/ (모든 SQLC 어노테이션 완전 지원)
                                 # - TypeScript: supabase/functions/_shared/queries/
                                 #   (제약사항: :exec, :execrows, :execresult, :batchexec 미지원)
./script/gen-proto.bash          # Protocol Buffer 코드 생성
./script/gen-typing-sb.bash      # TypeScript 데이터베이스 스키마 타입 생성
```

**중요**: TypeScript 생성을 위한 SQL 쿼리 작성 시, `:exec` 계열 어노테이션 대신 `:one` 또는 `:many` 어노테이션을 사용하세요. 데이터를 반환하지 않는 쿼리의 경우, `RETURNING` 절과 함께 `:one`을 사용하거나 더미 값을 선택하세요.

### 데이터베이스 관리 (Supabase)

```bash
# Supabase 로컬 환경 관리
supabase start                   # 로컬 Supabase 시작
supabase stop                    # 로컬 Supabase 중지
supabase status                  # Supabase 상태 확인

# 마이그레이션
supabase db reset                # 로컬 DB 리셋 (모든 마이그레이션 재실행)
supabase migration new <name>    # 새 마이그레이션 생성
supabase db push                 # 원격 DB에 마이그레이션 적용

# 스크립트를 통한 DB 리셋
./script/reset-local-sb.bash     # 로컬 Supabase DB 리셋 및 초기 데이터 생성
./script/reset-remote-sb.bash    # 원격 Supabase DB 리셋 (주의!)
```

### Supabase 통합 워크플로우

이 프로젝트는 Supabase를 데이터베이스 관리 플랫폼으로 활용합니다:

1. **로컬 개발**: `supabase start`로 Docker 기반 PostgreSQL 환경 실행
2. **스키마 관리**: `supabase/schemas/`에 테이블 정의, `supabase/migrations/`에 마이그레이션 저장
3. **타입 안전 쿼리**: `supabase/queries/`의 SQL을 SQLC로 Go 코드 생성
4. **배포**: Supabase CLI로 원격 프로젝트에 마이그레이션 적용

**주요 장점**:
- 로컬 개발 환경을 빠르게 구축 (Docker 기반)
- 마이그레이션 버전 관리 자동화
- Supabase Studio로 데이터베이스 시각적 관리
- 프로덕션 배포 간소화
- **타입 안전 코드 생성**: SQL을 한 번 작성하면 `./script/gen-sqlc.bash`를 통해 Go와 TypeScript 타입 안전 코드를 자동 생성

## 아키텍처 패턴

### Vertical Slice Architecture (주요 패턴)

이 프로젝트의 핵심 아키텍처는 **Vertical Slice Architecture**입니다. 각 기능(feature)은 모든 계층(HTTP → 비즈니스 로직 → 데이터 접근)을 포함하는 완결된 수직적 슬라이스입니다.

**특징**:

- 기능별 높은 응집도 (기능에 필요한 모든 코드가 한 곳에 위치)
- 낮은 결합도 (기능 간 의존성 최소화)
- 빠른 개발과 유지보수 (기능 단위로 독립적 작업 가능)

**구조 예시** (`internal/feature/user_profile/`):

```
internal/feature/user_profile/
  ├── router.go              # 라우트 매핑 (MapRoutes 함수)
  ├── get_profile/
  │   ├── endpoint.go        # HTTP 핸들러 (Map 함수)
  │   └── dto.go            # 요청/응답 DTO
  └── update_profile/
      ├── endpoint.go        # HTTP 핸들러 (Map 함수)
      └── dto.go            # 요청/응답 DTO
```

**엔드포인트 패턴**:

각 엔드포인트의 `Map` 함수는 다음을 직접 처리합니다:

1. 컨텍스트에서 로거와 DB 연결 추출
2. `httputil.GetReqBodyWithLog`로 요청 파싱
3. 비즈니스 로직 실행 (쿼리, 검증 등)
4. `httputil.OkWithMsg` 또는 `httputil.ErrWithMsg`로 응답

### 보조 패턴

**컴포넌트 기반 구조** (WebSocket, Stats, Logging 서비스):

- 기술적 관심사별 구조화 (세션, 패킷 처리, 이벤트 소비 등)
- 계층 분리 없이 직접 구현

**이벤트 드리븐 아키텍처**:

- Redis Streams 기반 비동기 처리
- Consumer-Processor 패턴

**리포지토리 패턴** (`internal/repository/`):

- 데이터 접근 추상화를 위한 템플릿
- CRUD 인터페이스 예제

### 주요 공유 컴포넌트

- **Redis Streams Consumer**: 제네릭 기반 이벤트 소비자
- **Database Access**: SQLC 생성 쿼리 또는 직접 pgx 쿼리
- **HTTP Utilities**: 표준화된 요청/응답 처리
- **Graceful Shutdown**: `shared.Closer` 인터페이스 기반

## API 엔드포인트

### API 서비스 (포트 8080)
- `GET /health` - 헬스 체크
- `GET /ready` - 준비 상태 체크
- `GET /api/v1/ping` - Ping
- `POST /api/v1/user-profile/get` - 사용자 프로필 조회
- `POST /api/v1/user-profile/update` - 사용자 프로필 업데이트

### WebSocket 서비스 (포트 8081)
- `GET /health` - 헬스 체크
- `GET /ws` - WebSocket 연결

### 통계 서비스 (포트 8084)
- `GET /health` - 헬스 체크
- `GET /metrics` - 메트릭 조회

## 라이선스

Apache License 2.0 - 자세한 내용은 [LICENSE](LICENSE) 파일 참조

## 기여

Pull Request를 환영합니다!

## 지원

문제가 있으시면 GitHub 이슈를 등록해 주세요.
