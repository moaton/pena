package sse

import "github.com/r3labs/sse/v2"

func NewServer() *sse.Server {
	server := sse.New()
	return server
}
