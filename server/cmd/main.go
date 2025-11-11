package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/guruorgoru/go-mmo/server/internal/server"
	"github.com/guruorgoru/go-mmo/server/internal/clients"
)

func main() {
	port := "8414"
	hub := server.NewHub()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewWebSocketClient, w, r)
	})

	go hub.Run()

	addr := fmt.Sprintf(":%v", port)

	log.Println("Server starting at port:", port)

	log.Fatal(http.ListenAndServe(addr, nil))
}
