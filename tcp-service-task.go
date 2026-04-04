package service

import (
	"context"
	"log"
	"net"
	"runtime/debug"
	"time"
)

type ITCPService interface {
	GetAddress() string
	OnNewConnection(context.Context, net.Conn) bool // DO NOT block, start a TCPConnectionTask or handle the connection in a new goroutine if needed, return false to close the connection immediately
	OnListenFailed(error) time.Duration
	OnStopping()
	OnStopped()
}

type tcpService struct {
	impl ITCPService
}

var _ IService = (*tcpService)(nil)

func (srv *tcpService) Listen(context.Context) (IAdapter, error) {
	return net.Listen("tcp", srv.impl.GetAddress())
}

func (srv *tcpService) Serve(ctx context.Context, adapter IAdapter) {
	listen, ok := adapter.(net.Listener)
	if !ok {
		return
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			// handle error
			return
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("【OnNewConnection Panic】Recover: %v\nStack:\n%s", r, debug.Stack())
					conn.Close()
				}
			}()
			if !srv.impl.OnNewConnection(ctx, conn) {
				conn.Close()
			}
		}()
	}
}

func (srv *tcpService) OnListenFailed(err error) time.Duration {
	return srv.impl.OnListenFailed(err)
}

func (srv *tcpService) OnStopping() {
	srv.impl.OnStopping()
}

func (srv *tcpService) OnStopped() {
	srv.impl.OnStopped()
}

func NewTCPServiceTask(delegate ITCPService) *ServiceTask {
	return NewServiceTask(&tcpService{impl: delegate})
}
