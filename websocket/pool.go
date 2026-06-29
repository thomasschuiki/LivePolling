package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"nhooyr.io/websocket"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
	Store      *Store
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
		Store:      NewStore(),
	}
}

func (c *Client) Send(ctx context.Context, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.Conn.Write(ctx, websocket.MessageText, data)
}

func (pool *Pool) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-pool.Register:
			client.ID = fmt.Sprintf("id-%d", len(pool.Clients)+1)
			client.Store = pool.Store
			fmt.Println("registering client:", client.ID)
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool:", len(pool.Clients))
		case client := <-pool.Unregister:
			fmt.Println("unregistering client:", client.ID)
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool:", len(pool.Clients))
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for client := range pool.Clients {
				if err := client.Send(context.Background(), message); err != nil {
					fmt.Println("error broadcasting:", err)
				}
			}
		}
	}
}

func (c *Client) WriteJSON(msg Message) error {
	return c.Send(context.Background(), msg)
}
