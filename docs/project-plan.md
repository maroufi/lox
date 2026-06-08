# Lox Project Plan

## Overview

**Lox** is a single-instance, durable, blocking lock service for distributed or multi-process systems.

The project should be built in small phases. Each phase delivers a useful slice of behavior while keeping the implementation aligned with the long-term goals:

- Correct mutual exclusion
- Blocking lock acquisition with FIFO ordering
- TTL-based expiration
- Durable state across restarts
- Fencing tokens for correctness
- Clear semantics that can evolve toward a future distributed or HA design

Lox has three main layers:

1. API Layer (gRPC)
2. Lock Manager (in-memory state and concurrency control)
3. Persistence Layer (SQLite)

---

## Phase 1 - MVP

Build the smallest usable version of Lox with an in-memory lock table and non-blocking operations.

### Tasks

- Define the initial gRPC API for `Acquire`, `Release`, `Renew`, and `Status`.
- Implement an in-memory lock table keyed by lock name.
- Track the current owner, lease TTL, expiration time, and basic lock metadata.
- Implement non-blocking `Acquire`:
  - Succeeds immediately when the lock is free.
  - Fails immediately when the lock is already held.
- Implement `Release`:
  - Only the current owner can release the lock.
  - Releasing a missing, expired, or differently owned lock returns a clear error.
- Implement `Renew`:
  - Only the current owner can extend the lease.
  - Renewal updates the expiration time.
- Implement `Status`:
  - Returns whether a lock is free, held, or expired.
  - Includes owner and expiration metadata when appropriate.
- Add unit tests for basic acquire, release, renew, status, ownership checks, and TTL validation.
- Document the MVP semantics and known limitations.

### Acceptance Criteria

- A client can acquire, release, renew, and inspect locks through the API.
- Mutual exclusion is enforced for non-blocking acquisition.
- No state is persisted across process restarts.
- Blocking behavior, wait queues, and durability are explicitly out of scope for this phase.

---

## Phase 2 - Blocking

Add wait queues and blocking acquisition semantics on top of the in-memory lock manager.

### Tasks

- Add a FIFO wait queue per lock.
- Extend `Acquire` with `wait_timeout_ms`.
- Define acquire modes:
  - Non-blocking when `wait_timeout_ms` is zero or omitted.
  - Blocking when `wait_timeout_ms` is greater than zero.
- When a lock is held, enqueue blocking acquire requests in arrival order.
- Promote the next waiter when the current owner releases the lock.
- Promote the next waiter when the current owner expires.
- Remove waiters when their `wait_timeout_ms` elapses.
- Handle client cancellation or disconnect while waiting.
- Ensure queue operations are protected by the lock manager concurrency model.
- Add tests for FIFO ordering, timeout behavior, cancellation, release promotion, and expiration promotion.
- Document blocking acquire semantics and timeout behavior.

### Acceptance Criteria

- Blocking acquire requests are served in FIFO order.
- Waiters either acquire the lock, time out, or are removed on cancellation.
- Releasing or expiring a lock promotes the next eligible waiter.
- Non-blocking acquire behavior from Phase 1 still works.

---

## Phase 3 - Durability

Add persistent storage so lock state survives process restarts.

### Tasks

- Design the SQLite schema for active locks and metadata.
- Add a persistence interface between the lock manager and SQLite.
- Persist lock acquisition, release, renewal, expiration, and owner changes with write-through behavior.
- Recover active lock state from SQLite on startup.
- Decide and document restart behavior for blocked waiters.
- Persist monotonic counters needed for future fencing token guarantees.
- Add integration tests using a real SQLite database.
- Add restart recovery tests for held locks, expired locks, renewed locks, and released locks.
- Add failure-path tests for persistence errors during lock state changes.

### Acceptance Criteria

- Active lock state survives a clean process restart.
- Expired locks are not recovered as valid held locks.
- Persistence writes happen as part of lock state transitions.
- Recovery preserves the data needed for monotonically increasing fencing tokens.

---

## Phase 4 - Hardening

Make the service safer to operate and complete the correctness mechanisms required by the goals.

### Tasks

- Fully wire fencing tokens into successful acquire responses.
- Ensure fencing tokens are monotonically increasing per lock.
- Persist fencing token counters and restore them correctly on startup.
- Add tests proving fencing tokens are never reused after release, expiration, renewal, or restart.
- Add background expiry cleanup for held locks whose TTL has elapsed.
- Ensure expiry cleanup promotes queued waiters fairly.
- Add structured logging for acquire, release, renew, expiration, queueing, recovery, and persistence errors.
- Add health checks for process status and persistence availability.
- Add metrics-ready boundaries for lock counts, wait queue sizes, acquire latency, timeouts, expirations, and persistence failures.
- Expand concurrency tests to cover racing acquire, release, renew, timeout, and expiration paths.
- Document operational behavior and troubleshooting guidance.

### Acceptance Criteria

- Fencing tokens are part of the public acquire contract.
- Expired locks are cleaned up without requiring a client request.
- Operators can inspect basic health and understand service behavior through logs.
- Core correctness properties are covered by unit, integration, concurrency, and restart tests.

---

## Phase 5 - Prepare for HA

Refine the design so future distributed or highly available work can be added without rewriting the service.

### Tasks

- Separate API handling, lock manager logic, clock/timer behavior, and persistence behind clear interfaces.
- Keep single-instance assumptions explicit in code and documentation.
- Document which guarantees currently depend on a single process.
- Document future HA requirements for leader election, replicated state, consensus, failover, and client retry behavior.
- Review persistence boundaries to identify what would need to change for replicated storage.
- Review fencing token generation to identify requirements for distributed monotonicity.
- Provide architecture notes for future distributed deployment.
- Keep high availability, multi-instance coordination, horizontal scalability, and consensus algorithms out of the implementation for now.

### Acceptance Criteria

- The codebase has clear boundaries between API, lock management, timing, and persistence.
- Current single-instance semantics are documented.
- Future HA work has a written design starting point.
- No HA, consensus, or multi-instance behavior is implemented in this phase.

---

## Current Non-Goals

The following are intentionally out of scope for the current roadmap:

- High availability
- Multi-instance coordination
- Horizontal scalability
- Consensus algorithms such as Raft or Paxos
