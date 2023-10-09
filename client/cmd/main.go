package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Client Hello")
	<-time.After(5 * time.Hour)
}
