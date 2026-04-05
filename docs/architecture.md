# Async Processing Platform - Architecture

## 1. 개요 (Overview)

본 프로젝트는 **비동기 작업 처리를 위한 경량 플랫폼**을 목표로 설계되었습니다.

동기 처리의 한계를 해결하기 위해,  
Producer → Queue → Worker 구조를 기반으로 **비동기 메시지 처리 시스템**을 구현하였습니다.

특히 다음을 핵심 목표로 설정하였습니다:

- 작업 요청과 실행의 분리
- 병렬 처리 기반 처리량 향상
- 실패 대응 (Retry / DLQ)
- 안전한 종료 (Graceful Shutdown)

---

## 2. 전체 구조 (High-Level Architecture)

```
Producer → Redis Queue → Worker Pool → Handler
                              ↓
                              DLQ
```

### 구성 요소

| 구성 요소 | 설명 |
|----------|------|
| Producer | 메시지를 생성하고 Queue에 적재 |
| Queue (Redis) | 메시지를 저장하는 중간 버퍼 |
| Worker | 메시지를 가져와 처리 |
| Worker Pool | 여러 Worker로 병렬 처리 |
| Handler | 실제 비즈니스 로직 수행 |
| DLQ | 실패 메시지 저장 |

---

## 3. 주요 컴포넌트 설명

### 3.1 Producer

Producer는 메시지를 생성하여 Queue에 전달합니다.

- 메시지는 `Type`, `Payload`, `Retry` 필드를 포함합니다.
- 현재는 테스트용 Producer를 사용하며, 향후 HTTP 기반 요청으로 확장 가능합니다.

---

### 3.2 Queue (Redis List 기반)

Queue는 Redis List를 사용하여 구현되었습니다.

- `LPUSH`를 통해 메시지를 삽입합니다.
- `BRPOP`을 통해 메시지를 소비합니다.
- Blocking 방식으로 Worker가 대기하며 효율적으로 메시지를 처리합니다.

또한 다음과 같은 큐를 사용합니다:

- `default`: 일반 메시지 처리 큐
- `default:dlq`: 실패 메시지 저장 큐

---

### 3.3 Worker & Worker Pool

Worker는 Queue에서 메시지를 가져와 처리하는 역할을 합니다.

- Worker는 무한 루프 기반으로 동작합니다.
- 여러 Worker를 생성하여 Worker Pool을 구성합니다.
- 각 Worker는 독립적으로 메시지를 처리하여 병렬성을 확보합니다.

---

### 3.4 Dispatcher & Handler

Worker는 메시지의 `Type`에 따라 Handler로 분기합니다.

```go
switch msg.Type {
case "test":
    handleTest(msg)
}
```

- Dispatcher는 분기 역할만 수행합니다.
- Handler는 실제 작업을 수행합니다.
- Worker는 비즈니스 로직을 직접 알지 않도록 분리하였습니다.

### 3.5 Retry & DLQ

작업 실패 시 다음과 같은 전략을 사용합니다:

- 실패 시 Retry 값을 증가시켜 재큐잉
- maxRetry 초과 시 DLQ로 이동

```
Fail → Retry++ → Requeue → Fail → DLQ
```

이를 통해:

- 일시적인 실패는 자동 복구
- 반복 실패는 격리 처리

### 3.6 Graceful Shutdown

Worker는 Context 기반으로 종료 신호를 전달받습니다.

- context.Context를 통해 종료 신호 전파
- WaitGroup을 통해 모든 Worker 종료 대기

```
Signal → cancel() → ctx.Done() → worker 종료 → WaitGroup 완료
```

이를 통해:

- 처리 중인 작업을 안전하게 마무리
- 중간 데이터 손실 방지

## 4. 메시지 구조 (Message Structure)

```go
type Message struct {
    Type    string
    Payload string
    Retry   int
}
```

- ``Type``: 작업 종류
- ``Payload``: 작업 데이터
- ``Retry``: 재시도 횟수

## 5. 에러 처리 전략 (Error Handling)

본 시스템은 sentinel error 기반 에러 처리를 사용합니다.

- 문자열 비교 대신 errors.Is() 사용
- 도메인별 에러 분리
    - queue: ErrDequeueTimeout
    - worker: ErrTaskFailed, ErrUnknownMessageType

이를 통해:

에러 비교 안정성 확보
의미 기반 에러 처리 가능
6. 로깅 전략 (Logging)

로그는 구조화된 형태로 통일하였습니다.

level=INFO worker=1 action=dequeue type=test payload="..." retry=1

특징:

key=value 형태
action 중심 이벤트 로그
retry / DLQ / shutdown 흐름 추적 가능
7. 설계 특징 요약
Redis List 기반 단순한 Queue 구현
Worker Pool을 통한 병렬 처리
Retry / DLQ 기반 실패 대응
Context + WaitGroup 기반 안전한 종료
구조화된 로그를 통한 관찰성 확보