package client

import (
	"cloud/internal/app/client/client"
	"log/slog"
	"os"
)

type App struct {
	log       *slog.Logger
	clientApp *client.App
}

func New() *App {
	log := setupLogger()
	clientApp, err := client.New(log)
	if err != nil {
		panic(err)
	}
	return &App{
		log:       log,
		clientApp: clientApp,
	}
}

func (a *App) Run() {
	a.clientApp.Run()
}

func setupLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
}
