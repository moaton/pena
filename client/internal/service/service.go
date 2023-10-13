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
)

type Service interface {
	Start(ctx context.Context)
	Handler(ctx context.Context, wg *sync.WaitGroup, ch chan string)
	SendReport(ctx context.Context, ids []string)
}

type service struct {
	client *http.Client
}

func NewService() Service {
	return &service{
		client: &http.Client{},
	}
}

func (s *service) Start(ctx context.Context) {
	batchsize := rand.Intn(10-2) + 2
	// batchsize := 1
	// log.Println("batchsize ", batchsize)
	url := fmt.Sprintf("http://server:3000/task?batchsize=%v", batchsize)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("http.NewRequest err ", err)
	}
	res, err := s.client.Do(req)
	if err != nil {
		log.Println("client.Do err ", err)
	}
	reqCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := make(chan int, 1)

	go func(c chan int, res *http.Response) {
		<-c
		res.Body.Close()
		log.Println("Client fall")
	}(c, res)

	reader := bufio.NewReader(res.Body)
	messages := make(chan string, batchsize)

	var wg sync.WaitGroup
	go s.Handler(reqCtx, &wg, messages)
	wg.Add(batchsize)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading SSE: %v\n", err)
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
			go msg.Checker(&wg, c, messages)
		}
	}
}

func (s *service) Handler(ctx context.Context, wg *sync.WaitGroup, ch chan string) {
	wg.Wait()
	length := len(ch)
	ids := make([]string, 0, length)
	log.Println("Hello after wg")
	select {
	case <-ctx.Done():
		return
	default:
		for i := 0; i < length; i++ {
			ids = append(ids, <-ch)
		}
		go s.SendReport(ctx, ids)
	}
	wg.Add(cap(ch))
	go s.Handler(ctx, wg, ch)
}

func (s *service) SendReport(ctx context.Context, ids []string) {
	log.Println("REPORT ", ids)
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
	resp, err := s.client.Do(req)
	if err != nil {
		log.Println("client.Do err ", err)
		return
	}
	log.Println("RESP ", resp.StatusCode)
}
