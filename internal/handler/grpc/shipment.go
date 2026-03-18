package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	domain "github.com/FIFSAK/TMS/internal/domain/shipment"
	shipmentSvc "github.com/FIFSAK/TMS/internal/service/shipment"
	pb "github.com/FIFSAK/TMS/pkg/pb/shipment/v1"
	"github.com/FIFSAK/TMS/pkg/store"
)

type ShipmentHandler struct {
	pb.UnimplementedShipmentServiceServer
	service shipmentSvc.Service
}

func NewShipmentHandler(service shipmentSvc.Service) *ShipmentHandler {
	return &ShipmentHandler{service: service}
}

func (h *ShipmentHandler) CreateShipment(ctx context.Context, req *pb.CreateShipmentRequest) (*pb.CreateShipmentResponse, error) {
	s, err := h.service.CreateShipment(ctx, shipmentSvc.CreateRequest{
		ReferenceNumber: req.GetReferenceNumber(),
		Origin:          req.GetOrigin(),
		Destination:     req.GetDestination(),
		DriverName:      req.GetDriverName(),
		UnitNumber:      req.GetUnitNumber(),
		ShipmentAmount:  req.GetShipmentAmount(),
		DriverRevenue:   req.GetDriverRevenue(),
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "create shipment: %v", err)
	}

	return &pb.CreateShipmentResponse{Shipment: shipmentToProto(s)}, nil
}

func (h *ShipmentHandler) GetShipment(ctx context.Context, req *pb.GetShipmentRequest) (*pb.GetShipmentResponse, error) {
	s, err := h.service.GetShipment(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			return nil, status.Errorf(codes.NotFound, "shipment not found")
		}
		return nil, status.Errorf(codes.Internal, "get shipment: %v", err)
	}

	return &pb.GetShipmentResponse{Shipment: shipmentToProto(s)}, nil
}

func (h *ShipmentHandler) ListShipments(ctx context.Context, _ *pb.ListShipmentsRequest) (*pb.ListShipmentsResponse, error) {
	shipments, err := h.service.ListShipments(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list shipments: %v", err)
	}

	pbShipments := make([]*pb.Shipment, len(shipments))
	for i, s := range shipments {
		pbShipments[i] = shipmentToProto(s)
	}

	return &pb.ListShipmentsResponse{Shipments: pbShipments}, nil
}

func (h *ShipmentHandler) AddShipmentEvent(ctx context.Context, req *pb.AddShipmentEventRequest) (*pb.AddShipmentEventResponse, error) {
	domainStatus, err := protoStatusToDomain(req.GetStatus())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid status: %v", err)
	}

	event, err := h.service.AddEvent(ctx, req.GetShipmentId(), domainStatus, req.GetComment())
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			return nil, status.Errorf(codes.NotFound, "shipment not found")
		}
		return nil, status.Errorf(codes.InvalidArgument, "add event: %v", err)
	}

	return &pb.AddShipmentEventResponse{Event: eventToProto(event)}, nil
}

func (h *ShipmentHandler) GetShipmentEvents(ctx context.Context, req *pb.GetShipmentEventsRequest) (*pb.GetShipmentEventsResponse, error) {
	events, err := h.service.GetEvents(ctx, req.GetShipmentId())
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			return nil, status.Errorf(codes.NotFound, "shipment not found")
		}
		return nil, status.Errorf(codes.Internal, "get events: %v", err)
	}

	pbEvents := make([]*pb.ShipmentEvent, len(events))
	for i, e := range events {
		pbEvents[i] = eventToProto(e)
	}

	return &pb.GetShipmentEventsResponse{Events: pbEvents}, nil
}

func shipmentToProto(s domain.Shipment) *pb.Shipment {
	return &pb.Shipment{
		Id:              s.ID,
		ReferenceNumber: s.ReferenceNumber,
		Origin:          s.Origin,
		Destination:     s.Destination,
		Status:          domainStatusToProto(s.Status),
		DriverName:      s.DriverName,
		UnitNumber:      s.UnitNumber,
		ShipmentAmount:  s.ShipmentAmount,
		DriverRevenue:   s.DriverRevenue,
		CreatedAt:       timestamppb.New(s.CreatedAt),
		UpdatedAt:       timestamppb.New(s.UpdatedAt),
	}
}

func eventToProto(e domain.Event) *pb.ShipmentEvent {
	return &pb.ShipmentEvent{
		Id:         e.ID,
		ShipmentId: e.ShipmentID,
		Status:     domainStatusToProto(e.Status),
		Comment:    e.Comment,
		CreatedAt:  timestamppb.New(e.CreatedAt),
	}
}

func domainStatusToProto(s domain.Status) pb.ShipmentStatus {
	switch s {
	case domain.StatusPending:
		return pb.ShipmentStatus_SHIPMENT_STATUS_PENDING
	case domain.StatusPickedUp:
		return pb.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP
	case domain.StatusInTransit:
		return pb.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT
	case domain.StatusDelivered:
		return pb.ShipmentStatus_SHIPMENT_STATUS_DELIVERED
	case domain.StatusCancelled:
		return pb.ShipmentStatus_SHIPMENT_STATUS_CANCELLED
	default:
		return pb.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED
	}
}

func protoStatusToDomain(s pb.ShipmentStatus) (domain.Status, error) {
	switch s {
	case pb.ShipmentStatus_SHIPMENT_STATUS_PENDING:
		return domain.StatusPending, nil
	case pb.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP:
		return domain.StatusPickedUp, nil
	case pb.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT:
		return domain.StatusInTransit, nil
	case pb.ShipmentStatus_SHIPMENT_STATUS_DELIVERED:
		return domain.StatusDelivered, nil
	case pb.ShipmentStatus_SHIPMENT_STATUS_CANCELLED:
		return domain.StatusCancelled, nil
	default:
		return "", errors.New("unknown shipment status")
	}
}
