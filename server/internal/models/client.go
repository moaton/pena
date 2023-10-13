package models

import (
	"context"
	"net/http"
)

type Client struct {
	ID        string
	Batchsize int

	Writer  http.ResponseWriter
	Flusher http.Flusher
	ReqCtx  context.Context
}
