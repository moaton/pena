package main

import (
	"client/internal/app"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()
	go app.Run(ctx)

	sigs := make(chan os.Signal, 1)
	done := make(chan int, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	log.Println("Running...")

	go func(done chan int) {
		ctx.Done()
		done <- 1
	}(done)

	<-sigs
	log.Println("Graceful shutdown")
	<-done
}
