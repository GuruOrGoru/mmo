package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/guruorgoru/go-mmo/server/internal/clients"
	"github.com/guruorgoru/go-mmo/server/internal/server"
)

func main() {
	config := server.LoadConfig()
	hub := server.NewHub(config)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewWebSocketClient, w, r)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	go hub.Run()

	addr := fmt.Sprintf(":%v", config.Port)

	log.Println("Server starting at port:", config.Port)

	log.Fatal(http.ListenAndServe(addr, nil))
}
