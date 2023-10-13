package service

import (
	"math/rand"
	"server/internal/models"

	"github.com/rs/xid"
)

type (
	xidGenerator func() xid.ID
)

func (s *service) GenerateMsg(id xidGenerator) models.Msg {
	period := uint64(rand.Intn(999) + 1)
	msg := models.Msg{
		ID:     id().String(),
		Period: period,
	}
	return msg
}
