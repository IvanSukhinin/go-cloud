package cloud

import (
	"cloud/internal/app/cloud/cloud"
	"cloud/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const envDev = "dev"

type App struct {
	cfg   *config.Config
	log   *slog.Logger
	cloud *cloud.App
}

func New() *App {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	return &App{
		cfg:   cfg,
		log:   log,
		cloud: cloud.New(log, cfg),
	}
}

func (a *App) Run() {
	a.log.Info("start", slog.Any("config", a.cfg))

	go a.cloud.GRPCServer.MustRun()

	// gracefull shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sign := <-stop

	a.cloud.GRPCServer.Stop()
	a.log.Info("app stopped by signal " + sign.String())
}

func setupLogger(env string) *slog.Logger {
	level := slog.LevelInfo
	if env == envDev {
		level = slog.LevelDebug
	}
	return slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
	)
}
