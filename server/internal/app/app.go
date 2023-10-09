package app

import (
	"context"
	"server/pkg/sse"
)

func Run(ctx context.Context) {
	server := sse.NewServer()
	defer server.Close()

}
