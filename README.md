# Async Platform (Go)

Platform-oriented asynchronous processing infrastructure

---

## 🎯 Objective

단순한 작업 큐가 아니라,  
서비스에서 공통적으로 사용할 수 있는 **비동기 처리 플랫폼**을 설계/구현하는 것이 목표입니다.

이 프로젝트는 다음을 중점적으로 다룹니다:

- 요청 처리와 비동기 작업의 분리
- 안정적인 작업 처리 (retry / DLQ)
- worker 기반 처리 모델
- 향후 Event-Driven Architecture로 확장 가능한 구조 설계

---

## 🏗 Architecture (Concept)

Client / Service  
  ↓  
Producer (enqueue)  
  ↓  
Queue (Redis List)  
  ↓  
Worker (consumer)  
  ↓  
Task Execution  

---

## 📦 MVP v1 Scope

- Redis List 기반 Queue
- Producer / Consumer 분리 구조
- Worker Pool
- Retry 처리
- Dead Letter Queue (DLQ)
- Message Envelope 표준화
- Graceful Shutdown
- 기본 Logging

---

## 🚫 Out of Scope (v1)

- Kafka / Event Streaming
- Redis Streams
- Scheduler / Cron
- Workflow Engine
- Exactly-once 처리
- 복잡한 Event Sourcing

---

## 💡 Design Goals

- 단순 Queue 구현이 아닌 **플랫폼 관점 설계**
- 메시지 구조를 표준화하여 확장성 확보
- Consumer 로직과 Queue 인프라 분리
- 향후 Kafka / EDA로 전환 가능한 구조 유지

---

## 🔄 Future Evolution

- Redis List → Redis Streams
- Redis → Kafka
- Job Queue → Event-Driven Architecture
- 단일 Consumer → Consumer Group

---

## 🧠 Key Concepts

- At-least-once delivery
- Idempotent processing
- Retry / Backoff 전략
- DLQ 처리 전략
- Worker lifecycle 관리

---

## 🚀 Status

- [ ] Project initialized
- [ ] Redis queue 구현
- [ ] Worker pool 구현
- [ ] Retry / DLQ 구현