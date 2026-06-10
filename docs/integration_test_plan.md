# Integration Test Plan: Acquire Already Held Lock

## File to Create
`internal/api/server_integration_test.go`

## Build Tag
`//go:build integration`

## Test Function
`TestAcquireAlreadyHeldLock`

## Scenario
Two gRPC clients contend for the same lock. Client A acquires it successfully; Client B receives a `FailedPrecondition` error.

## Steps

1. **Setup gRPC server**
   - Create `lock.NewManager(lock.RealClock{})`
   - Create `bufconn.Listener` (in-memory, no real TCP port)
   - Start `grpc.NewServer()` and register `loxv1.RegisterLockServiceServer`
   - Serve on the bufconn listener in a goroutine

2. **Create client connections**
   - Two `grpc.Dial` calls with `grpc.WithTransportCredentials(insecure.NewCredentials())`
   - Custom `grpc.WithContextDialer` dials the bufconn listener
   - Create two `loxv1.NewLockServiceClient` instances

3. **Client A — Acquire lock**
   - `Acquire("orders:123", "worker-a", 30s)`
   - Assert: no error
   - Assert: `status == ACQUIRE_STATUS_ACQUIRED`
   - Assert: lease fields match request (lock_name, owner, ttl)

4. **Client B — Acquire same lock**
   - `Acquire("orders:123", "worker-b", 30s)`
   - Assert: error is not nil
   - Assert: gRPC status code == `codes.FailedPrecondition`

5. **Cleanup**
   - Close both client connections
   - Stop the gRPC server
   - Close the bufconn listener

## Run Command
```bash
go test -tags=integration -v -run TestAcquireAlreadyHeldLock ./internal/api/
```

## Dependencies
- `google.golang.org/grpc/test/bufconn` — available via existing `google.golang.org/grpc` dependency
- No new `go.mod` changes required
