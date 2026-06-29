package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/thomasschuiki/LivePolling/server/websocket"
)

func clientHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("browsing: ", r.URL.Path)
	p := "." + r.URL.Path
	if p == "./" {
		p = "static/index.html"
	}
	http.ServeFile(w, r, p)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("browsing: ", r.URL.Path)
	p := "." + r.URL.Path
	if p == "./admin/" {
		p = "static/admin.html"
	}
	http.ServeFile(w, r, p)
}

func wsHandler(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Accept(w, r)
	if err != nil {
		log.Println(err)
		return
	}

	client := &websocket.Client{Conn: ws, Pool: pool}
	pool.Register <- client
	client.Read(context.Background())
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := websocket.NewPool()
	go pool.Start(ctx)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", clientHandler)
	http.HandleFunc("/admin/", adminHandler)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(pool, w, r)
	})

	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Starting Server on Port 8080")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	select {}
}
