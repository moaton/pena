package service

import (
	"bufio"
	"bytes"
	"client/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Service interface {
	CreateClient(ctx context.Context)
}

type service struct {
}

func NewService() Service {
	return &service{}
}

func (s *service) CreateClient(ctx context.Context) {
	reqCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	batchsize := rand.Intn(10-2) + 2
	log.Println("batchsize ", batchsize)
	url := fmt.Sprintf("http://server:3000/task?batchsize=%v", batchsize)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("http.NewRequest err ", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("client.Do err ", err)
	}
	uuid := res.Cookies()[0].Value

	reader := bufio.NewReader(res.Body)
	c := make(chan int, 1)
	go func(c chan int, res *http.Response, uuid string) {
		<-c
		res.Body.Close()
		log.Println("Client fall ", uuid)
	}(c, res, uuid)

	messages := make(chan string, batchsize)

	var wg sync.WaitGroup
	go s.Handler(reqCtx, uuid, &wg, messages)
	wg.Add(batchsize)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)

		if line != "" {
			var msg models.Message
			data := strings.TrimPrefix(line, "data:")
			err := json.Unmarshal([]byte(data), &msg)
			if err != nil {
				log.Println("data json.Unmarshal err ", err)
			}
			go s.Checker(&wg, &msg, c, messages)
		}
	}
}

func (s *service) Checker(wg *sync.WaitGroup, msg *models.Message, close chan int, messages chan string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Gorutina caught the panic")
			wg.Done()
		}
	}()

	log.Println("msg ", msg)
	if msg.Period > 800 && msg.Period < 900 {
		log.Panicln("Panic")
	} else if msg.Period > 900 {
		close <- 1
		return
	}
	<-time.After(time.Duration(msg.Period) * time.Millisecond)
	messages <- msg.Id
	wg.Done()
}

func (s *service) Handler(ctx context.Context, uuid string, wg *sync.WaitGroup, ch chan string) {
	wg.Wait()
	length := len(ch)
	ids := make([]string, 0, length)
	select {
	case <-ctx.Done():
		return
	default:
		for i := 0; i < length; i++ {
			ids = append(ids, <-ch)
		}
		go s.SendReport(ctx, uuid, ids)
	}
	wg.Add(cap(ch))
	go s.Handler(ctx, uuid, wg, ch)
}

func (s *service) SendReport(ctx context.Context, uuid string, ids []string) {
	var request struct {
		Ids []string `json:"ids"`
	}
	request.Ids = ids
	idsStr, err := json.Marshal(request)
	if err != nil {
		log.Println("json.Marshal err ", err)
		return
	}
	req, err := http.NewRequest("POST", "http://server:3000/report", bytes.NewBuffer(idsStr))
	if err != nil {
		log.Println("http.NewRequest err ", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "uuid",
		Value: uuid,
	})
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Println("client.Do err ", err)
		return
	}
}
