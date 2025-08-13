package grpcx

import (
	"google.golang.org/grpc"
	"net"
)

type GrpcxServer struct {
	*grpc.Server
	Addr   string
	Protocol string
}

func (g *GrpcxServer) Serve() error {
	l, err := net.Listen(g.Addr, g.Protocol)
	if err != nil {
		return err
	}
	return g.Server.Serve(l)
}
