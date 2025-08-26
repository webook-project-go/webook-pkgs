package grpcx

import (
	"context"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"time"
)

type GrpcxServer struct {
	*grpc.Server
	Addr     string
	Protocol string
	Name     string
	TTL      int64
	key      string
	client   *clientv3.Client
	kaCancel func()
	em       endpoints.Manager
}

type Config struct {
	TTL      int64
	Name     string
	Addr     string
	Protocol string
}

func NewGrpcxServer(client *clientv3.Client, server *grpc.Server, cfg Config) *GrpcxServer {
	return &GrpcxServer{
		Server:   server,
		client:   client,
		TTL:      cfg.TTL,
		Name:     cfg.Name,
		Protocol: cfg.Protocol,
		Addr:     cfg.Addr,
	}
}

func (g *GrpcxServer) Serve() error {
	l, err := net.Listen(g.Protocol, g.Addr)
	if err != nil {
		return err
	}
	err = g.register()
	if err != nil {
		return err
	}
	return g.Server.Serve(l)
}

func (g *GrpcxServer) register() error {
	em, err := endpoints.NewManager(g.client, "service/"+g.Name)
	if err != nil {
		return err
	}
	g.em = em
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	lease, err := g.client.Grant(ctx, g.TTL)
	cancel()
	if err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	g.key = "service/" + g.Name + "/" + g.Addr
	err = em.AddEndpoint(ctx, g.key, endpoints.Endpoint{
		Addr: g.Addr,
	}, clientv3.WithLease(lease.ID))
	cancel()
	if err != nil {
		return err
	}
	kaCtx, kaCancel := context.WithCancel(context.Background())
	g.kaCancel = kaCancel
	ch, err := g.client.KeepAlive(kaCtx, lease.ID)
	if err != nil {
		return err
	}
	go func() {
		for _ = range ch {
		}
	}()
	return nil
}
func (g *GrpcxServer) Close() error {
	if g.kaCancel != nil {
		g.kaCancel()
	}
	if g.em != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := g.em.DeleteEndpoint(ctx, g.key)
		if err != nil {
			return err
		}
	}
	g.Server.GracefulStop()
	return nil
}
