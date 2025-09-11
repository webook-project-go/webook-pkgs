package trace

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type TracerBuilder struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

func NewTracerBuilder(tracer trace.Tracer, propagator propagation.TextMapPropagator) *TracerBuilder {
	return &TracerBuilder{tracer: tracer, propagator: propagator}
}
func (t *TracerBuilder) BuildServer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = t.extract(ctx)
		ctx, span := t.tracer.Start(ctx, info.FullMethod)
		defer span.End()
		attrs := []attribute.KeyValue{
			semconv.RPCSystemKey.String("GRPC"),
			attribute.Key("rpc.grpc.kind").String("unary"),
			attribute.Key("rpc.component").String("server"),
		}
		span.SetAttributes(attrs...)
		resp, err = handler(ctx, req)
		if err != nil {
			span.SetStatus(codes.Error, "failed")
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "ok")
		}
		return

	}
}

func (t *TracerBuilder) BuildClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = t.inject(ctx)
		ctx, span := t.tracer.Start(ctx, method)
		defer span.End()
		attrs := []attribute.KeyValue{
			attribute.Key("rpc.grpc.kind").String("unary"),
			attribute.Key("rpc.component").String("client"),
		}
		span.SetAttributes(attrs...)
		err := invoker(ctx, method, req, reply, cc)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed")
		} else {
			span.SetStatus(codes.Ok, "ok")
		}
		return err
	}
}

func (t *TracerBuilder) inject(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	t.propagator.Inject(ctx, GrpcMD(md))
	return metadata.NewOutgoingContext(ctx, md)
}

type GrpcMD metadata.MD

func (g GrpcMD) Get(key string) string {
	val := metadata.MD(g).Get(key)
	if len(val) > 0 {
		return val[0]
	}
	return ""
}

func (g GrpcMD) Set(key string, value string) {
	metadata.MD(g).Set(key, value)
}

func (g GrpcMD) Keys() []string {
	keys := make([]string, 0, len(g))
	for key := range metadata.MD(g) {
		keys = append(keys, key)
	}
	return keys
}

func (t *TracerBuilder) extract(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	return t.propagator.Extract(ctx, GrpcMD(md))
}
