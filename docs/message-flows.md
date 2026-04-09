# Async Processing Platform - Message Flows

## 1. 개요

본 문서는 Async Processing Platform의 주요 메시지 처리 흐름을 설명합니다.

특히 다음 흐름을 중심으로 정리합니다:

- 정상 처리 흐름
- 실패 후 Retry 흐름
- 최대 재시도 초과 후 DLQ 이동 흐름
- Graceful Shutdown 흐름

---

## 2. 정상 처리 흐름

정상적인 메시지 처리 흐름은 다음과 같습니다.

1. Producer가 메시지를 생성합니다.
2. Producer가 Queue에 메시지를 적재합니다.
3. Worker가 Queue에서 메시지를 가져옵니다.
4. Worker는 Dispatcher를 통해 적절한 Handler로 분기합니다.
5. Handler가 작업을 성공적으로 수행합니다.
6. 메시지 처리가 종료됩니다.

### Sequence

```text
Producer -> Queue : Enqueue(message)
Worker -> Queue : Dequeue()
Queue -> Worker : message
Worker -> Dispatcher : dispatch(message)
Dispatcher -> Handler : handle(message)
Handler -> Worker : success
Worker : done
```

---

## 3. 실패 후 Retry 흐름

작업 처리 중 Handler에서 실패가 발생하면, Worker는 해당 메시지를 다시 Queue에 적재하여 재시도합니다.

1. Worker가 메시지를 가져옵니다.
2. Dispatcher가 Handler를 실행합니다.
3. Handler가 실패(error)를 반환합니다.
4. Worker가 Retry 횟수를 증가시킵니다.
5. Worker가 메시지를 다시 Queue에 적재합니다.
6. 이후 Worker가 다시 해당 메시지를 처리합니다.

### Sequence

```
Worker -> Queue : Dequeue()
Queue -> Worker : message
Worker -> Dispatcher : dispatch(message)
Dispatcher -> Handler : handle(message)
Handler -> Worker : error
Worker : Retry++
Worker -> Queue : Re-enqueue(message)
```

### 예시

- 초기 상태: Retry=0
- 첫 번째 실패 후: Retry=1
- 두 번째 실패 후: Retry=2
- 세 번째 실패 후: Retry=3

---

## 4. 최대 재시도 초과 후 DLQ 이동 흐름

메시지가 반복적으로 실패하여 최대 재시도 횟수를 초과하면, Worker는 해당 메시지를 DLQ(Dead Letter Queue)로 이동시킵니다.

1. Worker가 메시지를 처리합니다.
2. Handler가 실패를 반환합니다.
3. Worker가 Retry 횟수를 증가시킵니다.
4. Retry 횟수가 maxRetry를 초과하면 더 이상 일반 Queue에 넣지 않습니다.
5. Worker가 메시지를 DLQ에 적재합니다.
6. 일반 처리 흐름에서는 제외됩니다.

### Sequence

```
Worker -> Queue : Dequeue()
Queue -> Worker : message
Worker -> Dispatcher : dispatch(message)
Dispatcher -> Handler : handle(message)
Handler -> Worker : error
Worker : Retry++
Worker : check maxRetry
Worker -> DLQ : Enqueue(message)
```

### 정책

본 프로젝트에서는 maxRetry = 3 으로 설정하였습니다.

즉:

- 최초 처리 1회
- 재시도 최대 3회
- 이후 실패 시 DLQ 이동

---

## 5. Graceful Shutdown 흐름

Worker는 종료 신호를 받더라도 즉시 강제 종료되지 않고, 현재 처리 상태를 기준으로 안전하게 종료됩니다.

### 종료 흐름

1. OS 종료 신호(SIGINT, SIGTERM)가 발생합니다.
2. Main Goroutine이 signal을 수신합니다.
3. Main이 cancel()을 호출합니다.
4. Worker는 ctx.Done()을 감지합니다.
5. Worker는 현재 루프를 종료하고 빠져나옵니다.
6. WaitGroup이 모든 Worker 종료를 기다립니다.
7. 모든 Worker 종료 후 Main이 종료됩니다.

### Sequence

```
OS Signal -> Main : interrupt
Main -> Context : cancel()
Context -> Worker : ctx.Done()
Worker -> Worker : stop loop
Worker -> WaitGroup : Done()
Main -> WaitGroup : Wait()
WaitGroup -> Main : all workers stopped
```

---

## 6. 처리 중 종료 시 동작

Graceful Shutdown 중 Worker가 이미 작업을 처리 중인 경우, 현재 구현에서는 처리 중인 작업을 먼저 마친 뒤 종료됩니다.

즉:

- 새 작업은 더 이상 받지 않음
- 현재 처리 중인 Handler는 끝까지 수행
- 작업 완료 후 Worker 종료

### 예시 흐름

```
Worker -> Handler : handle(message)
OS Signal -> Main : interrupt
Main -> Context : cancel()
Handler : continue processing
Handler -> Worker : done
Worker : shutdown
```

이 방식은 갑작스러운 중단으로 인한 작업 손실을 줄이기 위한 선택입니다.

---

## 7. Redis 장애 시 흐름

Redis가 일시적으로 다운되면 Worker는 메시지를 가져오지 못하고 dequeue 에러를 반환받습니다.

현재 MVP에서는 다음과 같이 동작합니다.

1. Worker가 Queue에서 메시지를 가져오려고 시도합니다.
2. Redis 연결 실패 시 dequeue 에러가 발생합니다.
3. Worker는 에러 로그를 남깁니다.
4. Worker는 짧은 대기 후 다시 dequeue를 시도합니다.
5. Redis가 복구되면 Worker는 정상적으로 처리 흐름을 재개합니다.

### Sequence

```
Worker -> Redis Queue : Dequeue()
Redis Queue -> Worker : connection error
Worker : log error
Worker : sleep briefly
Worker -> Redis Queue : retry dequeue
```

### 현재 한계

- Redis 장애 중에는 메시지 처리가 중단됩니다.
- 장애가 길어지면 반복 에러 로그가 발생할 수 있습니다.
- retry backoff, circuit breaker, metrics 등은 현재 MVP 범위에서 제외하였습니다.

---

## 8. 요약

본 프로젝트의 메시지 흐름은 다음 4가지 축으로 구성됩니다.

- 정상 처리
- 실패 후 Retry
- 최대 재시도 초과 시 DLQ 이동
- Graceful Shutdown

이를 통해 단순한 비동기 처리뿐 아니라, 실패 대응과 안전한 종료까지 포함한 기본적인 Async Processing Platform의 뼈대를 구성하였습니다.

---