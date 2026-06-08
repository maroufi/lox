package lock

import (
	"errors"
	"testing"
	"time"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestAcquireFreeLock(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	manager := NewManager(fixedClock{now: now})

	lease, err := manager.Acquire("orders:123", "worker-a", 30*time.Second)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	if lease.LockName != "orders:123" {
		t.Fatalf("Lease.LockName = %q, want %q", lease.LockName, "orders:123")
	}
	if lease.Owner != "worker-a" {
		t.Fatalf("Lease.Owner = %q, want %q", lease.Owner, "worker-a")
	}
	if lease.TTL != 30*time.Second {
		t.Fatalf("Lease.TTL = %v, want %v", lease.TTL, 30*time.Second)
	}
	if !lease.AcquiredAt.Equal(now) {
		t.Fatalf("Lease.AcquiredAt = %v, want %v", lease.AcquiredAt, now)
	}
	if want := now.Add(30 * time.Second); !lease.ExpiresAt.Equal(want) {
		t.Fatalf("Lease.ExpiresAt = %v, want %v", lease.ExpiresAt, want)
	}
}

func TestAcquireValidatesInput(t *testing.T) {
	manager := NewManager(fixedClock{now: time.Now().UTC()})

	tests := []struct {
		name     string
		lockName string
		owner    string
		ttl      time.Duration
		wantErr  error
	}{
		{
			name:     "missing lock name",
			lockName: "",
			owner:    "worker-a",
			ttl:      time.Second,
			wantErr:  ErrInvalidLockName,
		},
		{
			name:     "missing owner",
			lockName: "orders:123",
			owner:    "",
			ttl:      time.Second,
			wantErr:  ErrInvalidOwner,
		},
		{
			name:     "zero ttl",
			lockName: "orders:123",
			owner:    "worker-a",
			ttl:      0,
			wantErr:  ErrInvalidTTL,
		},
		{
			name:     "negative ttl",
			lockName: "orders:123",
			owner:    "worker-a",
			ttl:      -time.Second,
			wantErr:  ErrInvalidTTL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.Acquire(tt.lockName, tt.owner, tt.ttl)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Acquire() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
