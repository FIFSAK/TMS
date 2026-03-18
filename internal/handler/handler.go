package handler

import (
	"github.com/FIFSAK/TMS/internal/config"
	grpcHandler "github.com/FIFSAK/TMS/internal/handler/grpc"
	"github.com/FIFSAK/TMS/internal/service"
	pb "github.com/FIFSAK/TMS/pkg/pb/shipment/v1"
	"google.golang.org/grpc"
)

type Dependencies struct {
	Configs  *config.Configs
	Services *service.Services
}

type Configuration func(h *Handlers) error

type Handlers struct {
	dependencies Dependencies
	Shipment     *grpcHandler.ShipmentHandler
}

func New(d Dependencies, configs ...Configuration) (h *Handlers, err error) {
	h = &Handlers{
		dependencies: d,
	}

	for _, cfg := range configs {
		if err = cfg(h); err != nil {
			return
		}
	}

	return
}

func WithShipmentHandler() Configuration {
	return func(h *Handlers) error {
		h.Shipment = grpcHandler.NewShipmentHandler(h.dependencies.Services.Shipment)
		return nil
	}
}

func (h *Handlers) RegisterGRPC(server *grpc.Server) {
	if h.Shipment != nil {
		pb.RegisterShipmentServiceServer(server, h.Shipment)
	}
}
