package api_test

import (
	"net"
	"testing"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/distributed/api"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const (
	serverNetwork = "tcp"
	serverAddress = ":8081"
)

func TestNewServer(t *testing.T) {
	listener, err := net.Listen(serverNetwork, serverAddress)
	if err != nil {
		t.Fatal(err)
	}

	serverGracefulStop, err := api.NewServer(listener, mock.NewLogger())
	if err != nil {
		t.Fatal(err)
	}

	defer serverGracefulStop()
}
