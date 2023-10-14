package app

import (
	"client/internal/service"
	"context"
	"log"
)

func Run(ctx context.Context) {
	service := service.NewService()

	client1, err := service.CreateClient(ctx)
	if err != nil {
		log.Printf("Client hasn't created %v", err)
	}
	go service.Start(ctx, client1)

	client2, err := service.CreateClient(ctx)
	if err != nil {
		log.Printf("Client hasn't created %v", err)
	}
	go service.Start(ctx, client2)

}
