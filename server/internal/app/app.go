package app

import (
	"context"
	"log"
	"net/http"
	"server/internal/handlers"
	"server/internal/service"
	"server/pkg/bolt"
	"server/pkg/sse"
)

func Run(ctx context.Context) {
	reportDb, err := bolt.NewBolt("reports")
	if err != nil {
		log.Fatalf("bolt.NewBolt() err %v", err)
	}
	failesDb, err := bolt.NewBolt("failes")
	if err != nil {
		log.Fatalf("bolt.NewBolt() err %v", err)
	}
	server := sse.NewServer()
	service := service.NewService(server, reportDb, failesDb)
	go service.StartSender(context.Background())
	router := handlers.NewRouter(service)

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatalf("http.ListenAndServe err %v", err)
	}
}
