package client

import (
	"cloud/internal/app/client/params"
	"cloud/internal/clients/cloud/cloudgrpc"
	"fmt"
	"log/slog"
)

const (
	uploadMethod   = "upload"
	downloadMethod = "download"
	listMethod     = "list"
)

type App struct {
	params *params.Params
	api    *cloudgrpc.Client
}

func New(log *slog.Logger) (*App, error) {
	p := params.New()

	api, err := cloudgrpc.New(p.Addr, log)
	if err != nil {
		return nil, err
	}

	return &App{
		params: p,
		api:    api,
	}, nil
}

func (c *App) Run() error {
	var err = fmt.Errorf("unknown method")
	switch c.params.Method {
	case uploadMethod:
		err = c.api.Upload(c.params.Src)
	case downloadMethod:
		err = c.api.Download(c.params.Dest, c.params.Filename)
	case listMethod:
		err = c.api.List()
	}
	return err
}
