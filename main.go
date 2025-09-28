package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Starting web socket server")
	clientManager := NewClientManager()

	go clientManager.StartWebSocketServer()

	http.HandleFunc("/web-socket", clientManager.WebsocketPage)

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 30 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting web server:", err)
		return
	}
}
