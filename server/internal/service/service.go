package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"server/internal/models"
	"server/pkg/sse"
	"server/pkg/util"
	"strconv"
	"time"

	"go.etcd.io/bbolt"
)

type Service interface {
	Serve(w http.ResponseWriter, r *http.Request)
	StartSender(ctx context.Context)
	AddReport(ids []string)
	GetReports() []string
	GetFails() []string
	FailsSave(ctx context.Context, f chan models.Msg)
	// GetFails() []string
}

type service struct {
	server   sse.Server
	reportDb *bbolt.DB
	failesDb *bbolt.DB
}

func NewService(server sse.Server, reportDb, failesDb *bbolt.DB) Service {
	return &service{
		server:   server,
		reportDb: reportDb,
		failesDb: failesDb,
	}
}

func (s *service) Serve(w http.ResponseWriter, r *http.Request) {
	// s.server.Serve(w, r)
	batchsizeStr := r.FormValue("batchsize")
	batchsize, err := strconv.Atoi(batchsizeStr)
	if err != nil {
		util.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}
	if batchsize == 0 {
		util.ResponseError(w, http.StatusBadRequest, "batchsize equal 0")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		util.ResponseError(w, http.StatusBadRequest, "streaming unsupported")
	}
	ctx := r.Context()
	f := make(chan models.Msg)
	go s.FailsSave(ctx, f)
	s.server.AddConnection(ctx, r.RemoteAddr, &models.Conn{
		Writer:    w,
		Flusher:   flusher,
		ReqCtx:    ctx,
		Batchsize: batchsize,
		Rs:        make(chan int),
		Sended:    make(map[string]models.Msg, batchsize),
	}, f)
	defer s.server.Close(r.RemoteAddr)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	<-ctx.Done()
}

func (s *service) StartSender(ctx context.Context) {
	log.Println("StartSender started")
	for {
		select {
		case <-ctx.Done():
			log.Println("StartSender finished")
			return
		default:
			conns := s.server.GetConnections()
			if conns == nil {
				continue
			}
			for addr, conn := range conns {
				log.Println("len(conn.Sended) ", len(conn.Sended))
				if conn.Batchsize > len(conn.Sended) {

					msg := s.MessageGenerator()

					msgBytes, err := json.Marshal(msg)
					if err != nil {
						log.Println("Send json.Marshal err ", err)
					}
					message := []byte("data:" + string(msgBytes) + "\n\n")

					err = conn.Send(message)
					if err != nil {
						continue
					}
					s.server.Sent(addr, msg)
				}
			}

			<-time.After(1 * time.Second)
		}
	}
}

func (s *service) AddReport(ids []string) {
	s.reportDb.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b == nil {
			b, _ = tx.CreateBucket([]byte("report"))
		}
		for _, id := range ids {
			b.Put([]byte(id), []byte("success"))
		}
		return nil
	})
	s.server.ClearQueue(ids)
}

func (s *service) GetReports() []string {
	ids := make([]string, 0, 100)
	s.reportDb.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b == nil {
			return errors.New("bucket doesn't exist")
		}
		err := b.ForEach(func(k, v []byte) error {
			ids = append(ids, string(k))
			return nil
		})
		return err
	})
	return ids
}

func (s *service) GetFails() []string {
	ids := make([]string, 0, 100)
	s.failesDb.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("fail"))
		if b == nil {
			return errors.New("bucket doesn't exist")
		}
		err := b.ForEach(func(k, v []byte) error {
			ids = append(ids, string(k))
			return nil
		})
		return err
	})
	return ids
}

func (s *service) FailsSave(ctx context.Context, f chan models.Msg) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-f:
			log.Println("FAIL ", msg)
			s.failesDb.Update(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte("fail"))
				if b == nil {
					b, _ = tx.CreateBucket([]byte("fail"))
				}
				err := b.Put([]byte(msg.Id), []byte("fail"))
				log.Println("Put ", err)
				return err
			})
		}
	}
}

// func (s *service) GetFails() []string {
// 	return s.server.GetFails()
// }
