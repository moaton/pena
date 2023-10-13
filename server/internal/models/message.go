package models

import (
	"math/rand"
	"time"

	"github.com/rs/xid"
)

type Msg struct {
	Id     string `json:"id"`
	Period uint64 `json:"period"`
	Retry  int    `json:"retry,omitempty"`
}

func NewMessage() *Msg {
	r := rand.NewSource(time.Now().UnixNano())
	return &Msg{
		Id:     xid.New().String(),
		Period: uint64(r.Int63() + 1000 - 1),
	}
}
