package server

import (
	"VK_task/internal/config"
	"VK_task/internal/services/pubsub"
	"VK_task/pkg/subpub"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	subpub1 "github.com/SinevArtem/protos_subpub/gen/go/pubsub"
	"google.golang.org/grpc"
)

func MustRun(cfg config.Config, log *slog.Logger) {
	if err := Run(cfg, log); err != nil {
		panic(err)
	}
}

func Run(cfg config.Config, log *slog.Logger) error {
	subpub := subpub.NewSubPub()
	defer subpub.Close(context.Background())

	grpcServer := grpc.NewServer()

	pubsubService := pubsub.NewService(subpub, log)

	subpub1.RegisterPubSubServer(grpcServer, pubsubService)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))
		if err := grpcServer.Serve(l); err != nil {
			log.Info("server failed ")
			return
		}
	}()

	<-stop

	log.Info("shutting down server")
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcServer.GracefulStop()
	log.Info("server stopped gracefully")

	return nil
}
