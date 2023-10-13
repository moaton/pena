package service

import (
	"math/rand"
	"server/internal/models"

	"github.com/rs/xid"
)

func (s *service) GenerateMsg() models.Msg {
	xid := xid.New().String()
	period := uint64(rand.Intn(999) + 1)
	msg := models.Msg{
		ID:     xid,
		Period: period,
	}
	return msg
}
