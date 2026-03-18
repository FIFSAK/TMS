package handler

import (
	"github.com/FIFSAK/TMS/internal/config"
	"github.com/FIFSAK/TMS/internal/service"
)

type Dependencies struct {
	Configs  *config.Configs
	Services *service.Services
}

type Configuration func(h *Handlers) error

type Handlers struct {
	dependencies Dependencies
	//TODO add grpc
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

//TODO add with grpc
