# Go 마이크로서비스 보일러플레이트

Vertical Slice Architecture와 최신 도구를 적용한 프로덕션 레디 Go 마이크로서비스 보일러플레이트

## 주요 기능

- **마이크로서비스 아키텍처**: 명확한 관심사 분리를 가진 독립적인 서비스들
- **Vertical Slice Architecture**: 기능별로 완결된 구조, 높은 응집도와 낮은 결합도
- **현대적인 스택**: Go 1.25, Chi v5, PostgreSQL, Redis
- **실시간 통신**: WebSocket 지원
- **이벤트 드리븐**: Redis Streams 기반 이벤트 처리
- **타입 안전성**: SQLC를 통한 타입 안전 SQL 쿼리
- **우아한 종료**: 적절한 리소스 정리 및 연결 처리

## 프로젝트 구조

```
.
├── servers/
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
├── supabase/
│   ├── schemas/                # 데이터베이스 스키마
│   ├── queries/                # SQLC 쿼리
│   └── migrations/             # 데이터베이스 마이그레이션
└── script/                     # 코드 생성 스크립트
```

## 기술 스택

### 코어
- **Go 1.25**: 제네릭 지원
- **Chi v5**: 경량 HTTP 라우터
- **gorilla/websocket**: WebSocket 구현

### 데이터 레이어
- **PostgreSQL**: 메인 데이터베이스
- **SQLC**: 타입 안전 SQL 코드 생성
- **pgx/v5**: 고성능 PostgreSQL 드라이버

### 캐싱 & 메시징
- **Redis**: 인메모리 데이터 스토어
- **Redis Streams**: 이벤트 스트리밍

## 빠른 시작

### 사전 요구사항

- Go 1.25+
- PostgreSQL 14+
- Redis 7+

### 설치

```bash
# 1. 저장소 클론
git clone https://github.com/your-org/go-monorepo-boilerplate.git
cd go-monorepo-boilerplate

# 2. 환경 변수 설정
cd servers
cp .env.example .env
# .env 파일 수정

# 3. 의존성 설치
go mod download

# 4. 데이터베이스 설정
psql -U postgres -d your_db < supabase/migrations/20250101000000_initial_schema.sql

# 5. SQLC 코드 생성
cd ..
./script/gen-sqlc.bash
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
./script/gen-sqlc.bash           # SQLC 코드 생성
./script/gen-proto.bash          # Protocol Buffer 코드 생성
./script/gen-typing-sb.bash      # TypeScript 타입 생성
```

### 데이터베이스 관리

```bash
./script/reset-local-sb.bash     # 로컬 데이터베이스 리셋
./script/reset-remote-sb.bash    # 원격 데이터베이스 리셋 (주의!)
```

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
