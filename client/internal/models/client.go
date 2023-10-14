package models

import (
	"errors"
	"log"
	"net/http"
)

type Http struct {
	Do       func(req *http.Request) (*http.Response, error)
	Request  *http.Request
	Response *http.Response
}

type Client struct {
	ID        string
	Batchsize int
	Task      Http
	Report    Http
}

func (c *Client) GetUUID() error {
	if len(c.Task.Response.Cookies()) == 0 {
		return errors.New("cookie empty")
	}
	uuid := c.Task.Response.Cookies()[0].Value
	c.ID = uuid
	return nil
}

func (c *Client) Close(closer chan int) {
	<-closer

	c.Task.Response.Body.Close()
	log.Println("Client fall ", c.ID)
}
