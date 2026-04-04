package service

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
)

type IGRPCService interface {
	GetNetConfig() *NetServiceConfig
	OnRegisterGRPCServer(*grpc.Server)
	OnListenFailed(error) time.Duration
	OnStopping()
	OnStopped()
}

type grpcService struct {
	impl IGRPCService
	srv  *grpc.Server
}

var _ IService = (*grpcService)(nil)

func (srv *grpcService) Listen(ctx context.Context) (IAdapter, error) {
	return srv.impl.GetNetConfig().ListenTCP()
}

func (srv *grpcService) Serve(ctx context.Context, adapter IAdapter) {
	listener, ok := adapter.(net.Listener)
	if !ok {
		return
	}
	srv.srv = grpc.NewServer()
	srv.impl.OnRegisterGRPCServer(srv.srv)
	srv.srv.Serve(listener)
}

func (srv *grpcService) OnListenFailed(err error) time.Duration {
	return srv.impl.OnListenFailed(err)
}

func (srv *grpcService) OnStopping() {
	if srv.srv != nil {
		srv.srv.GracefulStop()
	}
	srv.impl.OnStopping()
}

func (srv *grpcService) OnStopped() {
	srv.impl.OnStopped()
}

func NewGRPCService(impl IGRPCService) *ServiceTask {
	return NewServiceTask(&grpcService{
		impl: impl,
	})
}
