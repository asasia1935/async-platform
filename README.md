# Async Processing Platform

## 1. Problem

동기 방식의 작업 처리는 다음과 같은 한계를 가지고 있습니다.

- 요청 처리 시간이 길어짐 (Latency 증가)
- 특정 작업 실패 시 전체 흐름에 영향
- 트래픽 증가 시 확장 어려움
- 외부 API / DB 작업 시 블로킹 발생

이러한 문제를 해결하기 위해  
**작업 요청과 실행을 분리하는 비동기 처리 구조**가 필요하다고 판단하였습니다.

---

## 2. Design

본 프로젝트는 다음과 같은 구조로 설계되었습니다.

```
Producer → Redis Queue → Worker Pool → Handler
                          ↓
                          DLQ
```

### 핵심 구성 요소

- **Producer**
  - 메시지를 생성하여 Queue에 적재

- **Queue (Redis List)**
  - `LPUSH / BRPOP` 기반 비동기 큐

- **Worker Pool**
  - 여러 Worker를 통해 병렬 처리

- **Dispatcher**
  - 메시지 타입 기반 Handler 분기

- **Handler**
  - 실제 작업 수행

- **DLQ (Dead Letter Queue)**
  - 반복 실패 메시지 격리

---

## 3. Key Features

### 3.1 Worker Pool (병렬 처리)

- 여러 Worker 고루틴을 실행하여 작업을 분산 처리
- 동시에 여러 메시지를 처리하여 처리량 향상

---

### 3.2 Retry Mechanism

- Handler 실패 시 메시지를 재큐잉
- `Retry` 값을 기반으로 재시도 횟수 관리

---

### 3.3 DLQ (Dead Letter Queue)

- `maxRetry` 초과 시 메시지를 DLQ로 이동
- 반복 실패 메시지를 격리하여 시스템 안정성 확보

---

### 3.4 Graceful Shutdown

- `context.Context` 기반 종료 신호 전파
- `WaitGroup`을 통해 Worker 종료 대기

```
Signal → cancel() → ctx.Done() → worker 종료 → WaitGroup 완료
```

---

### 3.5 Structured Logging

- key=value 형태 로그 적용

```
level=INFO worker=1 action=dequeue type=test payload="..." retry=1
```

- 이벤트 기반 로그로 흐름 추적 가능

---

## 4. Message Flow

### 정상 흐름

```
Producer → Queue → Worker → Handler → Done
```

### 실패 흐름

```
Handler 실패 → Retry 증가 → Queue 재적재
```

### DLQ 흐름

```
Retry 초과 → DLQ 이동
```

---

## 5. Failure Handling

본 시스템은 다음과 같은 실패 대응 전략을 포함합니다.

- Handler 실패 시 자동 Retry
- maxRetry 초과 시 DLQ 격리
- Redis 장애 시 재시도 기반 복구
- Graceful Shutdown으로 안전한 종료

자세한 내용은 [`docs/failure-scenarios.md`](docs/failure-scenarios.md) 참고

---

## 6. Trade-offs

### Redis List 선택

- 장점
  - 단순하고 빠른 구현
  - MVP에 적합

- 단점
  - 메시지 재처리/ack 구조 없음
  - Kafka 대비 기능 제한

---

### Retry 전략

- 장점
  - 단순한 구조
  - 구현 용이

- 단점
  - backoff (바로 재시도 하지 않고 대기 후 재시도) 없음
  - 동일 메시지 반복 처리 가능

---

### Worker 기반 Retry

- 장점
  - 구조 단순
  - 빠른 구현

- 단점
  - retry 정책이 worker에 종속됨

---

## 7. Limitations (MVP)

현재 구현은 MVP 수준으로 다음과 같은 한계가 있습니다.

- retry backoff 미적용
- idempotency 보장 없음
- poison message 별도 처리 없음 (현재는 임의로 DLQ에만 저장)
- Redis 장애 대응 전략 제한적
- observability (metrics, tracing) 미구현
- HTTP/DB 작업 연동 미구현

---

## 8. Future Improvements

- exponential backoff retry
- circuit breaker (문제가 생긴 외부 시스템에 요청하지 않고 차단하는 것)
- idempotency key
- Kafka 기반 확장
- metrics / monitoring
- distributed tracing

---

## 9. How to Run

### 1. Redis 실행

```bash
docker run -d -p 6379:6379 --name redis redis
```

### 2. Producer 실행

```
go run ./cmd/producer
```

### 3. Worker 실행

```
go run ./cmd/worker
```

### 10. Project Structure

```
cmd/
  producer/
  worker/

internal/
  queue/
  worker/
  message/

docs/
  architecture.md
  message-flows.md
  failure-scenarios.md
```

### 11. Summary

본 프로젝트는 다음을 목표로 설계되었습니다.

- 비동기 작업 처리 구조 이해
- Worker Pool 기반 병렬 처리
- Retry / DLQ 기반 실패 대응
- Graceful Shutdown 기반 안정성 확보

단순한 Queue 구현을 넘어
**확장 가능한 Async Processing Platform**의 기초 구조를 만드는 것을 목표로 하였습니다.

---
