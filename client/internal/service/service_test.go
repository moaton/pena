package service

import (
	"client/internal/models"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecker(t *testing.T) {
	service := NewService()

	var wg sync.WaitGroup
	wg.Add(1)
	close := make(chan int, 1)
	message := make(chan string, 1)
	service.Checker(&wg, &models.Message{Id: "test", Period: 200}, close, message)
	wg.Wait()
	id := <-message
	assert.Equal(t, "test", id, "they should be equal")

	wg.Add(2)
	message = make(chan string, 2)
	service.Checker(&wg, &models.Message{Id: "test", Period: 200}, close, message)
	service.Checker(&wg, &models.Message{Id: "test2", Period: 300}, close, message)
	wg.Wait()

	ids := make([]string, 0, 2)

	ids = append(ids, <-message)
	ids = append(ids, <-message)

	assert.Equal(t, 2, len(ids), "they should be equal")

	wg.Add(2)
	message = make(chan string, 2)
	service.Checker(&wg, &models.Message{Id: "test", Period: 805}, close, message)
	service.Checker(&wg, &models.Message{Id: "test2", Period: 300}, close, message)
	wg.Wait()

	id = <-message

	assert.Equal(t, "test2", id, "they should be equal")

	wg.Add(1)
	message = make(chan string, 2)
	service.Checker(&wg, &models.Message{Id: "test", Period: 902}, close, message)
	service.Checker(&wg, &models.Message{Id: "test2", Period: 300}, close, message)
	wg.Wait()

	assert.Equal(t, 1, <-close, "they should be equal")
}
