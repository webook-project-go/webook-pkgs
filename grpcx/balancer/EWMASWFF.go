package balancer

import (
	"context"
	"errors"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"sync/atomic"
)

const Name = "EWMASWFFBalancer"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &SWFFBalancerBuilder{}, base.Config{HealthCheck: false})

}

func init() {
	balancer.Register(newBuilder())
}

type SWFFBalancerBuilder struct {
}

func (S *SWFFBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	subcon := make([]*WConn, 0, 16)
	var sumWeight float64
	for conn, subinfo := range info.ReadySCs {
		m, _ := subinfo.Address.Metadata.(map[string]any)
		weight, _ := m["weight"].(float64)
		curcon := &WConn{
			conn:      conn,
			weight:    weight,
			curWeight: weight,
			score:     atomic.Value{},
			alpha:     0.1,
		}
		curcon.score.Store(float64(1))
		subcon = append(subcon, curcon)
		sumWeight += weight

	}
	return &SWFFBalancer{
		conn:      subcon,
		sumWeight: sumWeight,
		mu:        sync.Mutex{},
	}
}

type SWFFBalancer struct {
	conn      []*WConn
	mu        sync.Mutex
	sumWeight float64
}

type WConn struct {
	conn      balancer.SubConn
	weight    float64
	curWeight float64

	score atomic.Value
	alpha float64
}

func (s *SWFFBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var bestWconn *WConn
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, conn := range s.conn {
		score, _ := conn.score.Load().(float64)
		conn.curWeight += conn.weight * score
		if bestWconn == nil {
			bestWconn = conn
		} else {
			if bestWconn.curWeight < conn.curWeight {
				bestWconn = conn
			}
		}
	}
	if bestWconn == nil {
		return balancer.PickResult{}, errors.New("no conn available")
	}
	bestWconn.curWeight -= s.sumWeight
	return balancer.PickResult{
		SubConn: bestWconn.conn,
		Done: func(info balancer.DoneInfo) {
			curscore, _ := bestWconn.score.Load().(float64)
			if info.Err == nil {
				if curscore >= 2 {
					return
				}
				bestWconn.score.Store(
					bestWconn.alpha*curscore*1.1 + (1-bestWconn.alpha)*curscore,
				)
				return
			}
			st, ok := status.FromError(info.Err)
			if !ok {
				if errors.Is(info.Err, context.Canceled) || errors.Is(info.Err, context.DeadlineExceeded) {
					bestWconn.score.Store(
						bestWconn.alpha*curscore*0.9 + (1-bestWconn.alpha)*curscore,
					)
				}
				return
			}
			switch st.Code() {
			case codes.OK:
			case codes.Unavailable:
				bestWconn.score.Store(
					bestWconn.alpha*curscore*0 + (1-bestWconn.alpha)*curscore,
				)
			case codes.ResourceExhausted:
				bestWconn.score.Store(
					bestWconn.alpha*curscore*0.1 + (1-bestWconn.alpha)*curscore,
				)
			case codes.Internal:
				bestWconn.score.Store(
					bestWconn.alpha*curscore*0.5 + (1-bestWconn.alpha)*curscore,
				)
			case codes.InvalidArgument, codes.FailedPrecondition, codes.Unauthenticated, codes.PermissionDenied:
				return
			default:
				if st.Code() > codes.Canceled {
					bestWconn.score.Store(
						bestWconn.alpha*curscore*0.8 + (1-bestWconn.alpha)*curscore,
					)
				}
			}
		},
	}, nil
}
