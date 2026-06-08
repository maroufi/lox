package lock

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidLockName = errors.New("lock name is required")
	ErrInvalidOwner    = errors.New("owner is required")
	ErrInvalidTTL      = errors.New("ttl must be greater than zero")
	ErrAlreadyHeld     = errors.New("lock is already held")
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}

type Lease struct {
	LockName   string
	Owner      string
	TTL        time.Duration
	AcquiredAt time.Time
	ExpiresAt  time.Time
}

type Manager struct {
	mu     sync.Mutex
	clock  Clock
	leases map[string]Lease
}

func NewManager(clock Clock) *Manager {
	if clock == nil {
		clock = RealClock{}
	}

	return &Manager{
		clock:  clock,
		leases: make(map[string]Lease),
	}
}

func (m *Manager) Acquire(lockName, owner string, ttl time.Duration) (Lease, error) {
	lockName = strings.TrimSpace(lockName)
	owner = strings.TrimSpace(owner)

	if lockName == "" {
		return Lease{}, ErrInvalidLockName
	}
	if owner == "" {
		return Lease{}, ErrInvalidOwner
	}
	if ttl <= 0 {
		return Lease{}, ErrInvalidTTL
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.clock.Now()
	if lease, ok := m.leases[lockName]; ok && lease.ExpiresAt.After(now) {
		return Lease{}, ErrAlreadyHeld
	}

	lease := Lease{
		LockName:   lockName,
		Owner:      owner,
		TTL:        ttl,
		AcquiredAt: now,
		ExpiresAt:  now.Add(ttl),
	}
	m.leases[lockName] = lease

	return lease, nil
}
