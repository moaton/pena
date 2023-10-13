package service

import (
	"context"
	"encoding/json"
	"log"
	"server/internal/db"
	"server/internal/models"
	"server/internal/transport/sse"
	"sync"
	"time"

	"github.com/rs/xid"
)

type Service interface {
	GetTasks(client *models.Client)
	StartSender(ctx context.Context)
	Report(id string, ids []string)
	GenerateMsg(id xidGenerator) models.Msg
	GetReports() []string
	GetFailTasks() []string
	saveFailTask(msg *models.Msg)
	saveReport(msg *models.Msg)
}

type service struct {
	clients     map[string]*models.Client
	tasks       map[string]map[string]*models.Msg
	resendTasks map[string]int
	reportDb    db.Repository
	failDb      db.Repository
	sse         sse.Server
	mu          sync.RWMutex
}

func NewService(sse sse.Server, reportRepo, failRepo db.Repository) Service {
	return &service{
		sse:         sse,
		clients:     make(map[string]*models.Client, 10),
		tasks:       make(map[string]map[string]*models.Msg, 10),
		resendTasks: make(map[string]int),
		reportDb:    reportRepo,
		failDb:      failRepo,
	}
}

func (s *service) StartSender(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.mu.Lock()
			for _, client := range s.clients {
				if client.Batchsize > len(s.tasks[client.ID]) {
					msg := s.GenerateMsg(xid.New)
					msgBytes, err := json.Marshal(msg)
					if err != nil {
						log.Println("Send json.Marshal err ", err)
					}
					message := []byte("data:" + string(msgBytes) + "\n\n")
					_, err = client.Writer.Write(message)
					if err != nil {
						log.Println("err ", err)
					}
					if _, ok := s.tasks[client.ID]; !ok {
						s.tasks[client.ID] = make(map[string]*models.Msg, client.Batchsize)
					}
					s.tasks[client.ID][msg.ID] = &msg

					client.Flusher.Flush()
				}
			}
			s.mu.Unlock()
			<-time.After(1 * time.Second)
		}
	}
}

func (s *service) GetTasks(client *models.Client) {
	log.Println("client.ID added ", client.ID)
	s.mu.Lock()
	s.clients[client.ID] = client
	s.mu.Unlock()
	c := make(chan struct{}, 1)

	s.sse.Serve(client, c)
	log.Println("client.ID delete ", client.ID)
	s.mu.Lock()
	delete(s.clients, client.ID)
	delete(s.tasks, client.ID)
	s.mu.Unlock()
}

func (s *service) Report(id string, ids []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; ok {
		for _, taskId := range ids {
			s.saveReport(s.tasks[id][taskId])
			delete(s.tasks[id], taskId)
		}
		if len(s.tasks[id]) > 0 {
			for _, msg := range s.tasks[id] {
				if _, ok := s.resendTasks[msg.ID]; ok && s.resendTasks[msg.ID] == 3 {
					s.saveFailTask(msg)
					delete(s.tasks[id], msg.ID)
					continue
				}
				msgBytes, err := json.Marshal(msg)
				if err != nil {
					log.Println("Send json.Marshal err ", err)
				}
				message := []byte("data:" + string(msgBytes) + "\n\n")

				_, err = s.clients[id].Writer.Write(message)
				if err != nil {
					log.Println("err ", err)
				}
				s.resendTasks[msg.ID]++
				s.clients[id].Flusher.Flush()
			}
		}
	}
}

func (s *service) saveFailTask(msg *models.Msg) {
	s.failDb.Save(msg.ID, msg.Period)
}

func (s *service) saveReport(msg *models.Msg) {
	s.reportDb.Save(msg.ID, msg.Period)
}

func (s *service) GetFailTasks() []string {
	return s.failDb.Get()
}

func (s *service) GetReports() []string {
	return s.reportDb.Get()
}
