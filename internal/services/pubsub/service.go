package pubsub

import (
	"VK_task/pkg/subpub"
	"context"
	"log/slog"
	"sync"

	subpub1 "github.com/SinevArtem/protos_subpub/gen/go/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type serverAPI struct {
	subpub1.UnimplementedPubSubServer
	SUBPUB subpub.SubPub
	mu     sync.Mutex
	log    *slog.Logger
}

func NewService(subpub subpub.SubPub, log *slog.Logger) *serverAPI {
	return &serverAPI{
		SUBPUB: subpub,
		log:    log,
	}
}

func (s *serverAPI) Subscribe(req *subpub1.SubscribeRequest, stream subpub1.PubSub_SubscribeServer) error {
	key := req.GetKey()
	if key == "" {
		return status.Error(codes.InvalidArgument, "key is required")
	}

	ctx := stream.Context()
	eventChan := make(chan string, 100)
	defer close(eventChan)

	sub, err := s.SUBPUB.Subscribe(key, func(msg interface{}) {
		select {
		case eventChan <- msg.(string):
		default:
			s.log.Info("event channel full ")
		}
	})

	if err != nil {
		return status.Error(codes.Internal, "subscribe failed")
	}

	defer func() {
		s.mu.Lock()
		sub.Unsubscribe()
		s.mu.Unlock()
	}()

	for {
		select {
		case data := <-eventChan:
			if err := stream.Send(&subpub1.Event{Data: data}); err != nil {
				s.log.Info("stream send error")
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

func (s *serverAPI) Publish(ctx context.Context, req *subpub1.PublishRequest) (*emptypb.Empty, error) {
	key := req.GetKey()
	data := req.GetData()

	if key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is required")
	}

	if err := s.SUBPUB.Publish(key, data); err != nil {
		return nil, status.Error(codes.Internal, "publish failed")
	}

	s.log.Info("published to %s: %s", key, data)

	return &emptypb.Empty{}, nil
}
