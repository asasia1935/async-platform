# Async Processing Platform - Failure Scenarios

## 1. 개요

본 문서는 Async Processing Platform에서 발생할 수 있는 주요 실패 시나리오와 현재 시스템의 동작 방식을 설명합니다.

특히 다음을 중심으로 정리합니다:

- Handler 실패
- Retry 실패 및 DLQ 처리
- Redis 장애 상황
- Graceful Shutdown 중 처리
- 현재 MVP 한계

---

## 2. Handler 실패

### 상황

Worker가 메시지를 처리하는 과정에서 Handler가 에러를 반환하는 경우입니다.

### 동작

1. Worker가 메시지를 dequeue 합니다.
2. Dispatcher를 통해 Handler를 실행합니다.
3. Handler가 에러를 반환합니다.
4. Worker는 해당 메시지의 Retry 값을 증가시킵니다.
5. 메시지를 다시 Queue에 enqueue 합니다.

### 결과

- 일시적인 장애는 자동으로 복구됩니다.
- 메시지는 다시 처리 대상이 됩니다.

---

## 3. Retry 반복 실패

### 상황

동일 메시지가 여러 번 처리 실패하는 경우입니다.

### 동작

1. 메시지 처리 실패 시 Retry 값 증가
2. 메시지를 다시 Queue에 enqueue
3. 동일 흐름 반복

### 문제

- 무한 retry가 발생할 수 있음
- 시스템 리소스를 지속적으로 사용

### 해결

본 시스템은 `maxRetry` 값을 통해 retry 횟수를 제한합니다.

---

## 4. DLQ (Dead Letter Queue)

### 상황

Retry 횟수가 `maxRetry`를 초과한 경우입니다.

### 동작

1. Worker가 Retry 증가
2. Retry > maxRetry 조건 확인
3. 메시지를 일반 Queue에 넣지 않음
4. DLQ(`default:dlq`)로 enqueue

### 결과

- 반복 실패 메시지는 격리됩니다.
- 정상 메시지 처리 흐름에 영향을 주지 않습니다.

---

## 5. Redis 장애

### 상황

Redis 서버가 다운되거나 연결이 불가능한 경우입니다.

### 동작

1. Worker가 dequeue 시도
2. Redis 연결 실패 → error 반환
3. Worker는 에러 로그 출력
4. 일정 시간 sleep 후 재시도

### 결과

- 메시지 처리는 일시적으로 중단됩니다.
- Redis 복구 시 자동으로 정상 처리 재개

### 한계

- 장애 동안 메시지 처리 불가
- retry backoff 없음
- circuit breaker 없음

---

## 6. Dequeue Timeout

### 상황

Queue에 메시지가 없는 경우입니다.

### 동작

1. Worker가 BRPOP 수행
2. 일정 시간(timeout) 동안 메시지 없음
3. timeout 에러 반환
4. Worker는 해당 에러를 무시하고 루프 지속

### 결과

- Worker는 idle 상태 유지
- CPU 낭비 없이 대기 가능
- shutdown 신호를 체크할 수 있음

---

## 7. Graceful Shutdown 중 처리

### 상황

Worker가 메시지를 처리 중일 때 종료 신호가 들어오는 경우입니다.

### 동작

1. Main이 종료 신호 수신
2. `cancel()` 호출
3. Worker가 `ctx.Done()` 감지
4. 현재 처리 중인 작업은 계속 수행
5. 작업 완료 후 Worker 종료

### 결과

- 작업 중단 없이 안전한 종료
- 데이터 손실 최소화

---

## 8. 잘못된 메시지 구조

### 상황

Queue에서 가져온 데이터가 예상과 다른 형식인 경우입니다.

### 동작

1. 메시지 파싱 시도 (json.Unmarshal)
2. 실패 시 에러 반환
3. Worker는 해당 메시지를 실패 처리

### 결과

- 잘못된 메시지는 정상 처리되지 않음
- 필요 시 DLQ로 이동 가능

---

## 9. 현재 MVP 한계

본 시스템은 MVP 수준의 구현으로, 다음과 같은 한계를 가지고 있습니다.

- retry backoff 전략 미적용
- idempotency 보장 없음
- poison message 별도 처리 없음
- Redis 장애 대응 전략 제한적
- observability (metrics, tracing) 미구현
- HTTP/DB 작업 연동 미구현

---

## 10. 향후 개선 방향

다음과 같은 방향으로 확장이 가능합니다:

- exponential backoff 기반 retry
- circuit breaker 도입
- idempotency key 적용
- 메시지 상태 추적 (processing / failed)
- metrics 및 모니터링 추가
- Kafka 기반 확장

---

## 11. 요약

본 시스템은 다음과 같은 실패 대응 전략을 포함합니다:

- Handler 실패 시 자동 retry
- maxRetry 초과 시 DLQ 격리
- Redis 장애 시 재시도 기반 복구
- Graceful Shutdown으로 안전한 종료

이를 통해 단순 비동기 처리에서 나아가,
기본적인 신뢰성(Reliability)을 갖춘 처리 구조를 제공합니다.

---