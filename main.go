package main

import (
	"fmt"
	"net/http"
	"time"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	fmt.Println("Starting web socket server")
	clientManager := NewClientManager()

	go clientManager.StartWebSocketServer()

	http.HandleFunc("/ws", clientManager.WebsocketPage)

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 60 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting web server:", err)
		return
	}
}
