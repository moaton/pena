package app

import (
	"context"
	"log"
	"net/http"
	"server/internal/db"
	"server/internal/db/bbolt"
	"server/internal/service"
	"server/internal/transport/handlers"
	"server/internal/transport/sse"
)

func Run(ctx context.Context) {

	reportDb, err := bbolt.NewBoltDb("report")
	if err != nil {
		log.Fatalf("bbolt.NewBoltDb err %v", err)
	}

	failsDb, err := bbolt.NewBoltDb("fail")
	if err != nil {
		log.Fatalf("bbolt.NewBoltDb err %v", err)
	}

	reportRepo := db.NewRepository(reportDb)
	failRepo := db.NewRepository(failsDb)

	sse := sse.NewServer()

	service := service.NewService(sse, reportRepo, failRepo)
	go service.StartSender(ctx)

	handlers := handlers.NewHandlers(service)
	router := handlers.NewRouter()

	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatalf("http.ListenAndServe err %v", err)
	}
}
