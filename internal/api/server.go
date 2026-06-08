package api

import (
	"context"
	"errors"

	"github.com/maroufi/lox/gen/go/proto/lox/v1"
	"github.com/maroufi/lox/internal/lock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	loxv1.UnimplementedLockServiceServer

	manager *lock.Manager
}

func NewServer(manager *lock.Manager) *Server {
	return &Server{manager: manager}
}

func (s *Server) Acquire(ctx context.Context, req *loxv1.AcquireRequest) (*loxv1.AcquireResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	ttl := req.GetTtl().AsDuration()
	lease, err := s.manager.Acquire(req.GetLockName(), req.GetOwner(), ttl)
	if err != nil {
		return nil, acquireError(err)
	}

	return &loxv1.AcquireResponse{
		Status:  loxv1.AcquireStatus_ACQUIRE_STATUS_ACQUIRED,
		Message: "lock acquired",
		Lease: &loxv1.Lease{
			LockName:   lease.LockName,
			Owner:      lease.Owner,
			Ttl:        durationpb.New(lease.TTL),
			AcquiredAt: timestamppb.New(lease.AcquiredAt),
			ExpiresAt:  timestamppb.New(lease.ExpiresAt),
		},
	}, nil
}

func acquireError(err error) error {
	switch {
	case errors.Is(err, lock.ErrInvalidLockName),
		errors.Is(err, lock.ErrInvalidOwner),
		errors.Is(err, lock.ErrInvalidTTL):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, lock.ErrAlreadyHeld):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "acquire failed")
	}
}
