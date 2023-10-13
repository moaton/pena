package sse

import (
	"context"
	"log"
	"server/internal/models"
	"sync"
)

type Server interface {
	AddConnection(ctx context.Context, addr string, conn *models.Conn, f chan models.Msg)
	GetConnections() map[string]*models.Conn
	Sent(addr string, msg models.Msg)
	Close(addr string)
	ClearQueue(ids []string)

	// Serve(w http.ResponseWriter, r *http.Request)
	// StartSender(cl chan int)
	// close(addr string)
	// GetFails() []string
}

type server struct {
	Connections map[string]*models.Conn
	SentTasks   map[string]string
	cl          chan int
	full        int
	// failesDb    *bbolt.DB
	mu sync.RWMutex
}

func NewServer() Server {
	return &server{
		Connections: make(map[string]*models.Conn),
		SentTasks:   make(map[string]string, 100),
		cl:          make(chan int, 1),
		// failesDb:    failesDb,
	}
}

// Добавляем коннект
func (s *server) AddConnection(ctx context.Context, addr string, conn *models.Conn, f chan models.Msg) {
	s.mu.Lock()
	s.Connections[addr] = conn
	log.Println("s.Connections ", s.Connections)

	go conn.Resender(ctx, f)
	s.mu.Unlock()
}

// Получение коннектов
func (s *server) GetConnections() map[string]*models.Conn {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.full == len(s.Connections) {
		return map[string]*models.Conn{}
	}
	return s.Connections
}

// Закрытие коннекта
func (s *server) Close(addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Println("CLOSE")
	delete(s.Connections, addr)
	if len(s.Connections) == 0 {
		s.cl <- 1
	}
}

func (s *server) Sent(addr string, msg models.Msg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Connections[addr].Sended[msg.Id] = msg
	s.SentTasks[msg.Id] = addr
	if s.Connections[addr].Batchsize == len(s.Connections[addr].Sended) {
		s.full++
	}
	log.Println("s.Connections ", s.Connections)
}

func (s *server) ClearQueue(ids []string) {
	log.Println("CLEAR")
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range ids {
		if _, ok := s.SentTasks[id]; ok {
			addr := s.SentTasks[id]
			delete(s.Connections[addr].Sended, id)
		}
	}
	s.full--
	// if s.full > 0 {
	// }
	log.Println("SEND chan ", len(s.Connections[addr].Sended))
	s.Connections[addr].Rs <- 1
	log.Println("After chan")
	// if _, ok := s.Connections[addr]; ok {
	// 	for _, id := range ids {
	// 		delete(s.Connections[addr].sended, id)
	// 	}
	// 	if len(s.Connections[addr].sended) > 0 {
	// 		go s.Resend(addr)
	// 	}
	// 	s.full--
	// }
	log.Println("FULL: ", s.full)
}

// func (s *server) Serve(w http.ResponseWriter, r *http.Request) {
// 	batchsizeStr := r.FormValue("batchsize")
// 	batchsize, err := strconv.Atoi(batchsizeStr)
// 	if err != nil {
// 		util.ResponseError(w, http.StatusBadRequest, err.Error())
// 		return
// 	}
// 	if batchsize == 0 {
// 		util.ResponseError(w, http.StatusBadRequest, "batchsize equal 0")
// 		return
// 	}

// 	flusher, ok := w.(http.Flusher)
// 	if !ok {
// 		util.ResponseError(w, http.StatusBadRequest, "streaming unsupported")
// 	}
// 	ctx := r.Context()
// 	s.mu.Lock()
// 	if len(s.Connections) == 0 {
// 		go s.StartSender(s.cl)
// 	}
// 	addr := strings.TrimSuffix(r.RemoteAddr, ":")
// 	log.Println("SSE addr ", addr)
// 	s.Connections[r.RemoteAddr] = &models.Conn{
// 		Writer:    w,
// 		Flusher:   flusher,
// 		ReqCtx:    ctx,
// 		Batchsize: batchsize,
// 		Sended:    make(map[string]models.Msg, batchsize),
// 	}
// 	log.Println("s.Connections ", s.Connections)
// 	s.mu.Unlock()

// 	defer s.Close(r.RemoteAddr)

// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Connection", "keep-alive")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")

// 	<-ctx.Done()
// }

// func (s *server) StartSender(cl chan int) {
// 	log.Println("StartSender started")
// 	for {
// 		select {
// 		case <-cl:
// 			log.Println("StartSender finished")
// 			return
// 		default:
// 			s.mu.Lock()
// 			log.Println("after lock ", s.full)
// 			if s.full < len(s.Connections) {
// 				log.Println("Send")
// 				for client, conn := range s.Connections {
// 					id := xid.New().String()
// 					period := uint64(rand.Intn(999) + 1)
// 					msg := models.Msg{
// 						Id:     id,
// 						Period: period,
// 					}
// 					log.Println("Batchsize ", conn.Batchsize, conn.Sended)
// 					if conn.Batchsize > len(conn.Sended) {
// 						msgBytes, err := json.Marshal(msg)
// 						if err != nil {
// 							log.Println("Send json.Marshal err ", err)
// 						}
// 						message := []byte("data:" + string(msgBytes) + "\n\n")
// 						_, err = conn.Writer.Write(message)
// 						if err != nil {
// 							log.Println("err ", err)
// 							s.Close(client)
// 							continue
// 						}
// 						conn.Sended[msg.Id] = msg
// 						conn.Flusher.Flush()
// 					} else {
// 						s.full++
// 						log.Println("INCREMENT ", s.full)
// 					}
// 				}
// 			}
// 			s.mu.Unlock()
// 			log.Println("After unlock")
// 			<-time.After(1 * time.Second)
// 		}
// 	}
// }

// func (s *server) Resend(addr string) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	for _, msg := range s.Connections[addr].sended {
// 		if msg.Retry == 3 {
// 			// go s.SaveFailed(msg)
// 			continue
// 		}
// 		log.Println("RESEND ", msg.Id)
// 		msgBytes, err := json.Marshal(msg)
// 		if err != nil {
// 			log.Println("Send json.Marshal err ", err)
// 		}
// 		message := []byte("data:" + string(msgBytes) + "\n\n")
// 		_, err = s.Connections[addr].writer.Write(message)
// 		if err != nil {
// 			log.Println("err ", err)
// 			s.close(addr)
// 			continue
// 		}
// 		s.Connections[addr].flusher.Flush()
// 	}
// }

// func (s *server) SaveFailed(msg models.Msg) {
// 	s.failesDb.Update(func(tx *bbolt.Tx) error {
// 		b := tx.Bucket([]byte("fail"))
// if b == nil {
// 	b, _ = tx.CreateBucket([]byte("fail"))
// }
// 		b.Put([]byte(msg.Id), []byte("fail"))
// 		return nil
// 	})
// }

// func (s *server) GetFails() []string {
// 	ids := make([]string, 0, 100)
// 	s.failesDb.View(func(tx *bbolt.Tx) error {
// 		b := tx.Bucket([]byte("fail"))
// 		if b == nil {
// 			return errors.New("bucket doesn't exist")
// 		}
// 		err := b.ForEach(func(k, v []byte) error {
// 			ids = append(ids, string(k))
// 			return nil
// 		})
// 		return err
// 	})
// 	return ids
// }
