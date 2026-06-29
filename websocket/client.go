package websocket

import (
	"context"
	"encoding/json"
	"log"

	"nhooyr.io/websocket"
)

type Client struct {
	ID             string
	Conn           *websocket.Conn
	Pool           *Pool
	Store          *Store
	IsAdmin        bool
	Authenticated  bool
}

type Message struct {
	Type          string   `json:"type"`
	Data          json.RawMessage `json:"data,omitempty"`
	Text          string   `json:"text,omitempty"`
	Question      string   `json:"question,omitempty"`
	Choices       []string `json:"choices,omitempty"`
	Possibilities []string `json:"possibilities,omitempty"`
	Votes         []int    `json:"votes,omitempty"`
	ID            string   `json:"id,omitempty"`
}

func (c *Client) Read(ctx context.Context) {
	defer func() {
		c.Pool.Unregister <- c
		c.Store.RemoveAdmin(c)
		c.Conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		_, p, err := c.Conn.Read(ctx)
		if err != nil {
			log.Printf("error reading msg: %v", err)
			return
		}
		HandleMessage(c, c.Pool, c.Store, p)
	}
}
