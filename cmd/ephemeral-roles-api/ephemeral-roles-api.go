package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env"
	"google.golang.org/grpc"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/distributed/api"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

type environmentVariables struct {
	Port                 string `env:"PORT" envDefault:"9000"`
	LogLevel             string `env:"LOG_LEVEL" envDefault:"info"`
	LogTimezoneLocation  string `env:"LOG_TIMEZONE_LOCATION" envDefault:"UTC"`
	DiscordrusWebHookURL string `env:"DISCORDRUS_WEBHOOK_URL"`
}

func stopChan() chan os.Signal {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGTERM)

	return stop
}

func createListener(port string) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, fmt.Errorf("error listening on TCP port %s: %w", port, err)
	}

	return listener, nil
}

func grpcServer() *grpc.Server {
	grpcServer := grpc.NewServer()

	api.RegisterAPIServer(grpcServer, &api.Server{})

	return grpcServer
}

func runServer(log logging.Interface, listener net.Listener, grpcServer *grpc.Server) {
	stop := stopChan()

	go func() {
		err := grpcServer.Serve(listener)
		if err != nil {
			log.WithError(err).Error("Error serving gRPC connection")
			stop <- syscall.SIGTERM
		}
	}()

	<-stop // Block until signal to stop

	grpcServer.GracefulStop()
}

func main() {
	envVars := &environmentVariables{}

	err := env.Parse(envVars)
	if err != nil {
		fmt.Printf("Error parsing environment variables: %s\n", err)
		os.Exit(1)
	}

	log := logging.New(
		logging.OptionalLogLevel(envVars.LogLevel),
		logging.OptionalTimezoneLocation(envVars.LogTimezoneLocation),
		logging.OptionalDiscordrus(envVars.DiscordrusWebHookURL),
	)

	listener, err := createListener(envVars.Port)
	if err != nil {
		log.WithError(err).Fatal("Error creating listener")
	}

	runServer(log, listener, grpcServer())
}
