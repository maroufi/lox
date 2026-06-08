# MVP Semantics

This document records the currently implemented Phase 1 behavior.

## Implemented

### Acquire a free lock

A client can call `Acquire(lock_name, owner, ttl)` for a lock that is not currently held. The service records an in-memory lease and returns:

- `ACQUIRE_STATUS_ACQUIRED`
- lock name
- owner
- requested TTL
- acquired timestamp
- expiration timestamp

The lease is stored only in process memory. It is lost when the service exits.

## Validation

`Acquire` rejects missing lock names, missing owners, and TTL values less than or equal to zero.

## Pending Phase 1 Behavior

The remaining Phase 1 use cases are still pending, including held-lock rejection tests, acquire-after-expiration behavior, release, renew, status, and concurrent acquire coverage.

## Phase 1 Limitations

Persistence, wait queues, blocking acquire, fencing token guarantees, high availability, and multi-instance coordination are intentionally out of scope for Phase 1.
