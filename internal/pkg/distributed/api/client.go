package api

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

type Client struct {
	APIClient
	Log        logging.Interface
	connection *grpc.ClientConn
}

func NewClient(log logging.Interface, serverAddress string) (*Client, error) {
	connection, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("error dialing gRPC server: %w", err)
	}

	return &Client{
		APIClient:  NewAPIClient(connection),
		Log:        log,
		connection: connection,
	}, nil
}

func (client *Client) GetLogLevel(ctx context.Context, in *GetLogLevelParameters, opts ...grpc.CallOption) (API_GetLogLevelClient, error) {
	stream, err := client.APIClient.GetLogLevel(ctx, in, opts)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil, ctx.Err()
		default:
			return &aPIGetLogLevelClient{}, nil
		}
	}
}

func (client *Client) SetLogLevel(ctx context.Context, in *LogLevel, opts ...grpc.CallOption) (*SetLogLevelResponse, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return &SetLogLevelResponse{}, nil
		}
	}
}

func (client *Client) Close() error {
	return client.connection.Close()
}
