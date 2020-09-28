package api_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/distributed/api"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const contextTimeout = 10 * time.Second

func TestNewClient(t *testing.T) {
	listener, err := net.Listen(serverNetwork, serverAddress)
	if err != nil {
		t.Fatal(err)
	}

	serverGracefulStop, err := api.NewServer(listener, mock.NewLogger())
	if err != nil {
		t.Fatal(err)
	}

	defer serverGracefulStop()

	client, err := api.NewClient(mock.NewLogger(), serverAddress)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := client.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelCtx()

	getLogLevelStream, err := client.GetLogLevel(ctx, &api.GetLogLevelParameters{})
	if err != nil {
		t.Fatal(err)
	}

	for {
		select {
		case <-getLogLevelStream.Context().Done():
			t.Fatal(err)
		default:
			logLevel, err := getLogLevelStream.Recv()
			if err != nil {
				t.Fatal(err)
			}

			if logLevel.Level != client.Log.Level() {
				client.Log.UpdateLevel(logLevel.Level)
			}
		}
	}
}
