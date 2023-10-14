package service

import (
	"bufio"
	"client/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Service interface {
	CreateClient(ctx context.Context) (*models.Client, error)
	Start(ctx context.Context, client *models.Client)
	Checker(wg *sync.WaitGroup, msg *models.Message, close chan int, messages chan string)
	Handler(ctx context.Context, client *models.Client, wg *sync.WaitGroup, ch chan string)
	// CreateClient(ctx context.Context)
}

type service struct {
	// Client HTTPClient
}

func NewService() Service {
	return &service{}
}

func (s *service) CreateClient(ctx context.Context) (*models.Client, error) {
	batchsize := rand.Intn(10-2) + 2
	log.Println("batchsize ", batchsize)

	url := fmt.Sprintf("http://server:3000/task?batchsize=%v", batchsize)
	taskReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("http.NewRequest err ", err)
		return nil, err
	}

	reportReq, err := http.NewRequest("POST", "http://server:3000/report", nil)
	if err != nil {
		log.Println("http.NewRequest err ", err)
		return nil, err
	}
	reportReq.Header.Set("Content-Type", "application/json")

	return &models.Client{
		Batchsize: batchsize,
		Task: models.Http{
			Do:      http.DefaultClient.Do,
			Request: taskReq,
		},
		Report: models.Http{
			Do:      http.DefaultClient.Do,
			Request: reportReq,
		},
	}, nil
}

func (s *service) Start(ctx context.Context, client *models.Client) {
	var err error

	reqCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := make(chan int, 1)

	client.Task.Response, err = client.Task.Do(client.Task.Request)
	if err != nil {
		log.Println("client.Do err ", err)
	}
	go client.Close(c)

	client.GetUUID()

	reader := bufio.NewReader(client.Task.Response.Body)

	messages := make(chan string, client.Batchsize)

	var wg sync.WaitGroup
	go s.Handler(reqCtx, client, &wg, messages)
	wg.Add(client.Batchsize)

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

func (s *service) Handler(ctx context.Context, client *models.Client, wg *sync.WaitGroup, ch chan string) {
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
		go s.SendReport(ctx, client, ids)
	}
	wg.Add(cap(ch))
	go s.Handler(ctx, client, wg, ch)
}

func (s *service) SendReport(ctx context.Context, client *models.Client, ids []string) {
	var request struct {
		Ids []string `json:"ids"`
	}
	request.Ids = ids
	idsStr, err := json.Marshal(request)
	if err != nil {
		log.Println("json.Marshal err ", err)
		return
	}

	client.Report.Request.AddCookie(&http.Cookie{
		Name:  "uuid",
		Value: client.ID,
	})
	bodyReader := strings.NewReader(string(idsStr))
	client.Report.Request.Body = io.NopCloser(bodyReader)

	_, err = client.Report.Do(client.Report.Request)
	if err != nil {
		log.Println("client.Do err ", err)
		return
	}
}
