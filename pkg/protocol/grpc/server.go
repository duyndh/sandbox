
package grpc

import (
	"context"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	"github.com/ngray1747/sandbox/pkg/api/v1"
	"github.com/ngray1747/sandbox/pkg/logger"
	"github.com/ngray1747/sandbox/pkg/protocol/grpc/middleware"
)

func RunServer(ctx context.Context, v1API v1.TodoServiceServer, port string) error {
	listen, error := net.Listen("tcp", ":"+port)
	if error != nil {
		return error
	}

	// gRPC server statup options
	opts := []grpc.ServerOption{}

	// add middleware
	opts = middleware.AddLogging(logger.Log, opts)

	// Register GRPC server
	server := grpc.NewServer(opts...)
	v1.RegisterTodoServiceServer(server, v1API)

	// Singal to shutdown
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, os.Interrupt)
	go func() {
		for range cancel {
			logger.Log.Warn("Shutting down GRPC server")
			server.GracefulStop()
			<- ctx.Done()
		}
	}()

	logger.Log.Info("Starting GRPC server")
	return server.Serve(listen)
}