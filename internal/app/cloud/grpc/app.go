package grpcapp

import (
	"cloud/internal/config"
	"cloud/internal/grpc/cloud"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	cloudService cloud.Cloud,
	port int,
	cfg config.CloudConfig,
) *App {
	gRPCServer := grpc.NewServer()
	gRPCCloudServer := cloud.New(cloudService, log, cfg)

	cloud.Register(gRPCServer, gRPCCloudServer)

	reflection.Register(gRPCServer)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// MustRun runs gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run runs gRPC server.
func (a *App) Run() error {
	const fn = "grpcapp.Run"
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	a.log.Info("grpc server started", slog.String("addr", lis.Addr().String()))

	if err := a.gRPCServer.Serve(lis); err != nil {
		return fmt.Errorf("%s, %w", fn, err)
	}

	return nil
}

// Stop stops gRPC server.
func (a *App) Stop() {
	const fn = "grpcapp.Stop"

	a.log.With(slog.String("fn", fn)).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
