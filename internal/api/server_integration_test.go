//go:build integration

package api

import (
	"context"
	"net"
	"testing"
	"time"

	loxv1 "github.com/maroufi/lox/gen/go/proto/lox/v1"
	"github.com/maroufi/lox/internal/lock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestAcquireAlreadyHeldLock(t *testing.T) {
	lis := bufconn.Listen(256 * 1024)

	grpcServer := grpc.NewServer()
	loxv1.RegisterLockServiceServer(grpcServer, NewServer(lock.NewManager(nil)))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("server serve error: %v", err)
		}
	}()
	defer grpcServer.Stop()
	defer lis.Close()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	connA, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("client A dial error: %v", err)
	}
	defer connA.Close()

	connB, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("client B dial error: %v", err)
	}
	defer connB.Close()

	clientA := loxv1.NewLockServiceClient(connA)
	clientB := loxv1.NewLockServiceClient(connB)

	ctx := context.Background()

	respA, err := clientA.Acquire(ctx, &loxv1.AcquireRequest{
		LockName: "orders:123",
		Owner:    "worker-a",
		Ttl:      durationpb.New(30 * time.Second),
	})
	if err != nil {
		t.Fatalf("client A Acquire() error = %v", err)
	}

	if respA.GetStatus() != loxv1.AcquireStatus_ACQUIRE_STATUS_ACQUIRED {
		t.Fatalf("client A status = %v, want ACQUIRE_STATUS_ACQUIRED", respA.GetStatus())
	}

	lease := respA.GetLease()
	if lease == nil {
		t.Fatal("client A lease is nil")
	}
	if lease.GetLockName() != "orders:123" {
		t.Fatalf("client A lease.LockName = %q, want %q", lease.GetLockName(), "orders:123")
	}
	if lease.GetOwner() != "worker-a" {
		t.Fatalf("client A lease.Owner = %q, want %q", lease.GetOwner(), "worker-a")
	}
	if got := lease.GetTtl().AsDuration(); got != 30*time.Second {
		t.Fatalf("client A lease.TTL = %v, want %v", got, 30*time.Second)
	}

	_, err = clientB.Acquire(ctx, &loxv1.AcquireRequest{
		LockName: "orders:123",
		Owner:    "worker-b",
		Ttl:      durationpb.New(30 * time.Second),
	})
	if err == nil {
		t.Fatal("client B Acquire() expected error, got nil")
	}

	if code := status.Code(err); code.String() != "FailedPrecondition" {
		t.Fatalf("client B error code = %v, want FailedPrecondition", code)
	}
}

func TestAcquireDifferentLocks(t *testing.T) {
	lis := bufconn.Listen(256 * 1024)

	grpcServer := grpc.NewServer()
	loxv1.RegisterLockServiceServer(grpcServer, NewServer(lock.NewManager(nil)))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("server serve error: %v", err)
		}
	}()
	defer grpcServer.Stop()
	defer lis.Close()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	connA, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("client A dial error: %v", err)
	}
	defer connA.Close()

	connB, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("client B dial error: %v", err)
	}
	defer connB.Close()

	clientA := loxv1.NewLockServiceClient(connA)
	clientB := loxv1.NewLockServiceClient(connB)

	ctx := context.Background()

	respA, err := clientA.Acquire(ctx, &loxv1.AcquireRequest{
		LockName: "orders:123",
		Owner:    "worker-a",
		Ttl:      durationpb.New(30 * time.Second),
	})
	if err != nil {
		t.Fatalf("client A Acquire() error = %v", err)
	}

	if respA.GetStatus() != loxv1.AcquireStatus_ACQUIRE_STATUS_ACQUIRED {
		t.Fatalf("client A status = %v, want ACQUIRE_STATUS_ACQUIRED", respA.GetStatus())
	}

	leaseA := respA.GetLease()
	if leaseA == nil {
		t.Fatal("client A lease is nil")
	}
	if leaseA.GetLockName() != "orders:123" {
		t.Fatalf("client A lease.LockName = %q, want %q", leaseA.GetLockName(), "orders:123")
	}
	if leaseA.GetOwner() != "worker-a" {
		t.Fatalf("client A lease.Owner = %q, want %q", leaseA.GetOwner(), "worker-a")
	}

	respB, err := clientB.Acquire(ctx, &loxv1.AcquireRequest{
		LockName: "orders:456",
		Owner:    "worker-b",
		Ttl:      durationpb.New(30 * time.Second),
	})
	if err != nil {
		t.Fatalf("client B Acquire() error = %v", err)
	}

	if respB.GetStatus() != loxv1.AcquireStatus_ACQUIRE_STATUS_ACQUIRED {
		t.Fatalf("client B status = %v, want ACQUIRE_STATUS_ACQUIRED", respB.GetStatus())
	}

	leaseB := respB.GetLease()
	if leaseB == nil {
		t.Fatal("client B lease is nil")
	}
	if leaseB.GetLockName() != "orders:456" {
		t.Fatalf("client B lease.LockName = %q, want %q", leaseB.GetLockName(), "orders:456")
	}
	if leaseB.GetOwner() != "worker-b" {
		t.Fatalf("client B lease.Owner = %q, want %q", leaseB.GetOwner(), "worker-b")
	}
}
