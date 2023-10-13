package app

import (
	"client/internal/service"
	"context"
)

func Run(ctx context.Context, service service.Service) {
	service.Start(ctx)
}
