package api

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
	grpc "google.golang.org/grpc"
)

const interval = 5 * time.Second

type Server struct {
	Log logging.Interface
	UnimplementedAPIServer
}

func NewServer(listener net.Listener, log logging.Interface) (gracefulStop func(), err error) {
	server := &Server{Log: log}
	grpcServer := grpc.NewServer()

	RegisterAPIServer(grpcServer, server)

	go func() {
		err := grpcServer.Serve(listener)
		if err != nil {
			server.Log.WithError(err).Errorf("gRPC server error")
		}
	}()

	return grpcServer.GracefulStop, nil
}

func (server *Server) GetLogLevel(_ *GetLogLevelParameters, stream API_GetLogLevelServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		default:
			time.Sleep(interval)

			err := stream.Send(&LogLevel{Level: server.Log.Level()})
			if err != nil {
				return fmt.Errorf("error sending GetLogLevel response: %w", err)
			}
		}
	}
}

func (server *Server) SetLogLevel(ctx context.Context, logLevel *LogLevel) (*SetLogLevelResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if logLevel.Level != server.Log.Level() {
			server.Log.UpdateLevel(logLevel.Level)
		}

		return &SetLogLevelResponse{}, nil
	}
}
