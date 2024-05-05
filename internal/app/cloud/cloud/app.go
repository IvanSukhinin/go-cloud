package cloud

import (
	grpcapp "cloud/internal/app/cloud/grpc"
	"cloud/internal/config"
	"cloud/internal/services/cloud"
	"cloud/internal/storage/drive"
	"log/slog"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
) *App {
	// data layer
	storage, err := drive.New(cfg.Storage)
	if err != nil {
		panic(err)
	}

	// service layer
	cloudService := cloud.New(log, storage)

	// transport layer
	grpcApp := grpcapp.New(log, cloudService, cfg.GRPC.Port, cfg.Cloud)

	return &App{
		GRPCServer: grpcApp,
	}
}
