package grpcapi

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cargov1 "github.com/roma-glushko/cargo/gen/cargo/v1"
	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/handling"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/voyage"
)

type HandlingServer struct {
	cargov1.UnimplementedHandlingServiceServer
	service *handling.Service
}

func (s *HandlingServer) RegisterHandlingEvent(ctx context.Context, req *cargov1.RegisterHandlingEventRequest) (*cargov1.RegisterHandlingEventResponse, error) {
	if req.CompletionTime == nil {
		return nil, status.Error(codes.InvalidArgument, "completion_time is required")
	}

	eventType, ok := cargo.ParseEventType(req.Type)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unknown event type: %s", req.Type)
	}

	var errs []string
	for _, tid := range req.TrackingIds {
		err := s.service.RegisterHandlingEvent(
			ctx,
			req.CompletionTime.AsTime(),
			cargo.TrackingID(tid),
			voyage.Number(req.VoyageNumber),
			location.UNLocode(req.UnLocode),
			eventType,
		)

		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", tid, err.Error()))
		}
	}

	if len(errs) > 0 {
		return nil, status.Error(codes.InvalidArgument, strings.Join(errs, "; "))
	}

	return &cargov1.RegisterHandlingEventResponse{}, nil
}
