package consumer

import (
	"context"
)

// Consumer는 이벤트 스트림을 소비하는 컴포넌트의 기본 인터페이스입니다.
// 이 인터페이스를 구현하면 다양한 이벤트 소스(Redis Streams, Kafka, RabbitMQ 등)로부터
// 이벤트를 소비하는 공통 패턴을 제공할 수 있습니다.
type Consumer interface {
	// Start는 이벤트 소비를 시작합니다.
	// context가 취소되면 소비를 중단해야 합니다.
	Start(ctx context.Context) error

	// Shutdown은 소비자를 우아하게 종료합니다.
	// 처리 중인 이벤트를 완료하고 리소스를 정리합니다.
	Shutdown(ctx context.Context) error
}

// Processor는 개별 이벤트를 처리하는 인터페이스입니다.
// 제네릭 타입 T를 사용하여 다양한 이벤트 타입을 처리할 수 있습니다.
type Processor[T any] interface {
	// ProcessEvent는 단일 이벤트를 처리합니다.
	// 처리 중 오류가 발생하면 error를 반환합니다.
	ProcessEvent(ctx context.Context, event T) error
}

// MetricsProvider는 메트릭 정보를 제공하는 인터페이스입니다.
// 제네릭 타입 T를 사용하여 다양한 메트릭 타입을 지원합니다.
type MetricsProvider[T any] interface {
	// GetMetrics는 현재 메트릭 요약 정보를 반환합니다.
	// 반환된 메트릭은 thread-safe한 복사본이어야 합니다.
	GetMetrics() T
}

// ProcessorWithMetrics는 Processor와 MetricsProvider를 결합한 인터페이스입니다.
// 이벤트 처리와 메트릭 조회를 모두 지원하는 컴포넌트에 사용합니다.
type ProcessorWithMetrics[T any, M any] interface {
	Processor[T]
	MetricsProvider[M]
}
