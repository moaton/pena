package models

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Conn struct {
	Writer    http.ResponseWriter
	Flusher   http.Flusher
	ReqCtx    context.Context
	Sended    map[string]Msg
	Rs        chan int
	Batchsize int
}

func (c *Conn) Send(msg []byte) error {
	_, err := c.Writer.Write(msg)
	if err != nil {
		log.Println("err ", err)
		return errors.New("Writer.Write err")
	}
	c.Flusher.Flush()
	return nil
}

func (c *Conn) Resender(ctx context.Context, f chan Msg) {
	resends := make(map[string]int, 100)
	for {
		select {
		case <-c.Rs:
			for _, msg := range c.Sended {
				log.Println("RESENDER ", msg.Id)
				resends[msg.Id]++
				if resends[msg.Id] == 3 {
					delete(resends, msg.Id)
					delete(c.Sended, msg.Id)
					f <- msg
				}
				msgBytes, err := json.Marshal(msg)
				if err != nil {
					log.Println("Send json.Marshal err ", err)
				}
				message := []byte("data:" + string(msgBytes) + "\n\n")
				_, err = c.Writer.Write(message)
				if err != nil {
					log.Println("err ", err)
					continue
				}
				c.Flusher.Flush()
			}
		}
	}
}
