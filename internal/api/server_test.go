package api

import (
	"context"
	"testing"
	"time"

	"github.com/maroufi/lox/gen/go/proto/lox/v1"
	"github.com/maroufi/lox/internal/lock"
	"google.golang.org/protobuf/types/known/durationpb"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestAcquireFreeLock(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	server := NewServer(lock.NewManager(fixedClock{now: now}))

	resp, err := server.Acquire(context.Background(), &loxv1.AcquireRequest{
		LockName: "orders:123",
		Owner:    "worker-a",
		Ttl:      durationpb.New(30 * time.Second),
	})
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	if resp.GetStatus() != loxv1.AcquireStatus_ACQUIRE_STATUS_ACQUIRED {
		t.Fatalf("Acquire() status = %v, want %v", resp.GetStatus(), loxv1.AcquireStatus_ACQUIRE_STATUS_ACQUIRED)
	}

	lease := resp.GetLease()
	if lease == nil {
		t.Fatal("Acquire() lease is nil")
	}
	if lease.GetLockName() != "orders:123" {
		t.Fatalf("Lease.LockName = %q, want %q", lease.GetLockName(), "orders:123")
	}
	if lease.GetOwner() != "worker-a" {
		t.Fatalf("Lease.Owner = %q, want %q", lease.GetOwner(), "worker-a")
	}
	if got := lease.GetTtl().AsDuration(); got != 30*time.Second {
		t.Fatalf("Lease.TTL = %v, want %v", got, 30*time.Second)
	}
	if got := lease.GetAcquiredAt().AsTime(); !got.Equal(now) {
		t.Fatalf("Lease.AcquiredAt = %v, want %v", got, now)
	}
	if want := now.Add(30 * time.Second); !lease.GetExpiresAt().AsTime().Equal(want) {
		t.Fatalf("Lease.ExpiresAt = %v, want %v", lease.GetExpiresAt().AsTime(), want)
	}
}
