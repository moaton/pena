package service

import (
	"context"
	"net/http"
	"server/internal/db"
	"server/internal/models"
	"server/internal/transport/sse"
	"sync"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

type mockWriter struct{}

func NewMockWriter() http.ResponseWriter {
	return &mockWriter{}
}

func (m *mockWriter) Header() http.Header        { return http.Header{} }
func (m *mockWriter) Write([]byte) (int, error)  { return 0, nil }
func (m *mockWriter) WriteHeader(statusCode int) {}

type mockFlusher struct{}

func NewMockFlusher() http.Flusher {
	return &mockFlusher{}
}

func (m *mockFlusher) Flush() {}

type mockServer struct{}

func NewMockServer() sse.Server {
	return &mockServer{}
}

func (m *mockServer) Serve(client *models.Client, c chan struct{}) {}

type mockDb struct{}

func NewMockDb() db.Repository {
	return &mockDb{}
}

func (m *mockDb) Get() []string {
	return []string{}
}

func (m *mockDb) Save(id string, period uint64) {}

type mockService struct {
	clients     map[string]*models.Client
	tasks       map[string]map[string]*models.Msg
	resendTasks map[string]int
	reportDb    db.Repository
	failDb      db.Repository
	sse         sse.Server
	mu          sync.RWMutex
}

func NewMockService(sse sse.Server, reportRepo, failRepo db.Repository) Service {
	return &mockService{
		sse:         sse,
		clients:     make(map[string]*models.Client, 10),
		tasks:       make(map[string]map[string]*models.Msg, 10),
		resendTasks: make(map[string]int),
		reportDb:    reportRepo,
		failDb:      failRepo,
	}
}

func (m *mockService) GetTasks(client *models.Client)         {}
func (m *mockService) StartSender(ctx context.Context)        {}
func (m *mockService) Report(id string, ids []string)         {}
func (m *mockService) GetReports() []string                   { return []string{} }
func (m *mockService) GetFailTasks() []string                 { return []string{} }
func (s *mockService) saveFailTask(msg *models.Msg)           {}
func (s *mockService) saveReport(msg *models.Msg)             {}
func (s *mockService) GenerateMsg(id xidGenerator) models.Msg { return models.Msg{} }

func TestNewService(t *testing.T) {
	server := NewMockServer()
	mockDb := NewMockDb()

	expected := &service{
		clients:     make(map[string]*models.Client),
		tasks:       make(map[string]map[string]*models.Msg),
		resendTasks: make(map[string]int),
		reportDb:    mockDb,
		failDb:      mockDb,
		sse:         server,
	}

	service := NewService(server, mockDb, mockDb)
	assert.Equal(t, expected, service, "they should be equal")
}

func TestGetTasks(t *testing.T) {
	server := NewMockServer()
	mockDb := NewMockDb()

	service := &service{
		clients:     make(map[string]*models.Client),
		tasks:       make(map[string]map[string]*models.Msg),
		resendTasks: make(map[string]int),
		reportDb:    mockDb,
		failDb:      mockDb,
		sse:         server,
	}

	client := &models.Client{ID: "1"}
	service.GetTasks(client)
	client = &models.Client{ID: "2"}
	service.GetTasks(client)
	assert.Equal(t, 0, len(service.clients), "they should be equal")
}

func TestReport(t *testing.T) {
	server := NewMockServer()
	mockDb := NewMockDb()

	service := &service{
		clients:     make(map[string]*models.Client),
		tasks:       make(map[string]map[string]*models.Msg),
		resendTasks: make(map[string]int),
		reportDb:    mockDb,
		failDb:      mockDb,
		sse:         server,
	}
	writer := NewMockWriter()
	flusher := NewMockFlusher()
	service.clients["test"] = &models.Client{
		Writer:  writer,
		Flusher: flusher,
	}

	service.tasks["test"] = make(map[string]*models.Msg)
	service.tasks["test"]["test"] = &models.Msg{ID: "test", Period: 100}

	service.Report("test", []string{"test"})

	assert.Equal(t, 0, len(service.tasks["test"]), "they should be equal")

	service.tasks["test"] = make(map[string]*models.Msg)
	service.tasks["test"]["test"] = &models.Msg{ID: "test", Period: 100}
	service.tasks["test"]["test2"] = &models.Msg{ID: "test2", Period: 100}

	service.Report("test", []string{"test"})

	assert.Equal(t, 1, len(service.tasks["test"]), "they should be equal")
}

func TestGenerateMsg(t *testing.T) {
	server := NewMockServer()
	mockDb := NewMockDb()

	service := &service{
		clients:     make(map[string]*models.Client),
		tasks:       make(map[string]map[string]*models.Msg),
		resendTasks: make(map[string]int),
		reportDb:    mockDb,
		failDb:      mockDb,
		sse:         server,
	}
	msg1 := service.GenerateMsg(xid.New)

	assert.NotEqual(t, nil, msg1, "they shouldn't be equal")

	msg2 := service.GenerateMsg(xid.New)

	assert.NotEqual(t, msg2.ID, msg1.ID, "they shouldn't be equal")

	id := xid.New()
	msg3 := service.GenerateMsg(func() xid.ID { return id })

	assert.Equal(t, id.String(), msg3.ID, "they should be equal")
}
