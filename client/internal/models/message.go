package models

import (
	"log"
	"sync"
	"time"
)

type Msg interface {
	Checker()
}

type Message struct {
	Id     string `json:"id"`
	Period uint64 `json:"period"`
}

type MessageStatus struct {
	Id     string `json:"id"`
	Status bool   `json:"status"`
}

func (m *Message) Checker(wg *sync.WaitGroup, c chan int, messages chan string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Gorutina caught the panic")
			wg.Done()
		}
	}()
	// m.Period = 982
	if m.Period > 900 {
		m.Period = 805
	}
	log.Println("Period ", m.Period)
	if m.Period > 800 && m.Period < 900 {
		log.Panicln("Panic")
	} else if m.Period > 900 {
		c <- 1
		return
	}
	<-time.After(time.Duration(m.Period) * time.Millisecond)
	messages <- m.Id
	wg.Done()
	log.Println("CHECKER WG.done")
}
