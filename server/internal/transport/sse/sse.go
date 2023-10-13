package sse

import "server/internal/models"

type Server interface {
	Serve(client *models.Client, c chan struct{})
}

type sse struct {
}

func NewServer() Server {
	return &sse{}
}

func (s *sse) Serve(client *models.Client, c chan struct{}) {
	client.Writer.Header().Set("Content-Type", "text/event-stream")
	client.Writer.Header().Set("Cache-Control", "no-cache")
	client.Writer.Header().Set("Connection", "keep-alive")
	client.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	<-client.ReqCtx.Done()
	c <- struct{}{}
}
