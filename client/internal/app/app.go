package app

import (
	"client/internal/service"
	"context"
)

func Run(ctx context.Context) {
	service := service.NewService()
	go service.CreateClient(ctx)
	go service.CreateClient(ctx)
}
