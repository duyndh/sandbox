
package grpc

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"

	"github.com/ngray/sandbox/pkg/api/v1"
)

func RunServer(ctx context.Context, v1API v1.ToDoServiceServer, port string) error {
	listen, error := net.Listen("tcp", ":"+port)
	if error != nil {
		return error
	}

	// Register GRPC server
	server := grpc.NewServer()
	v1.RegisterTodoServiceServer(server, v1API)

	// Singal to shutdown
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, os.Interrupt)
	go func() {
		for range cancel {
			log.Println("Shutting down GRPC server")
			server.GracefulStop()
			<- ctx.Done()
		}
	}()

	log.Println("Starting GRPC server")
	return server.Serve(listen)
}