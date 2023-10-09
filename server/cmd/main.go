package main

import (
	"log"
	"server/internal/models"
)

func main() {
	ch := make(chan int, 1)
	msg := models.NewMessage()
	log.Println("Server Hi ", *msg)
	<-ch
}
