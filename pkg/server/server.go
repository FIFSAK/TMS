package server

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Option func(*Servers) error

type Servers struct {
	grpcServer *grpc.Server
	grpcListen net.Listener
}

func NewServer(opts ...Option) (*Servers, error) {
	s := &Servers{}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	return s, nil
}

func (s *Servers) Run(logger *zap.Logger) error {
	if logger == nil {
		logger = zap.NewNop()
	}

	if s.grpcServer == nil {
		return fmt.Errorf("no servers configured to run")
	}

	if s.grpcServer != nil && s.grpcListen != nil {
		addr := s.grpcListen.Addr().String()
		go func() {
			logger.Info("starting grpc server", zap.String("addr", addr))
			if err := s.grpcServer.Serve(s.grpcListen); err != nil {
				logger.Error("grpc serve failed", zap.String("addr", addr), zap.Error(err))
			}
			logger.Info("grpc server stopped", zap.String("addr", addr))
		}()
	}

	return nil
}

func (s *Servers) Stop(ctx context.Context) error {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	return nil
}

func (s *Servers) GRPCServer() *grpc.Server {
	return s.grpcServer
}

func WithGRPC(addr string, serverOpts ...grpc.ServerOption) Option {
	return func(s *Servers) error {
		if s.grpcServer != nil {
			return fmt.Errorf("grpc server already configured")
		}
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("listen grpc %s: %w", addr, err)
		}
		s.grpcListen = l
		s.grpcServer = grpc.NewServer(serverOpts...)
		return nil
	}
}
