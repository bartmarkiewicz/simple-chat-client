package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

type ClientManager struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

type Client struct {
	Id     string
	Socket *websocket.Conn
	Send   chan []byte
}

type MessageContent struct {
	Text string `json:"text"`
	Role string `json:"role"`
}

type Message struct {
	Sender  string         `json:"sender"`
	Content MessageContent `json:"content"`
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (clientManager *ClientManager) WebsocketPage(res http.ResponseWriter, req *http.Request) {
	connection, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if err != nil {
		http.NotFound(res, req)
		return
	}

	client := &Client{Id: uuid.NewV4().String(), Socket: connection, Send: make(chan []byte)}

	clientManager.Register <- client

	go client.read(clientManager)
	go client.write()
}

func (clientManager *ClientManager) StartWebSocketServer() {
	for {
		select {
		case connection := <-clientManager.Register:
			clientManager.Clients[connection] = true
			jsonMessage, _ := json.Marshal(&Message{Content: MessageContent{
				Text: "A new client has connected.",
				Role: "SYSTEM",
			}})
			clientManager.send(jsonMessage, connection)
		case connection := <-clientManager.Unregister:
			if _, ok := clientManager.Clients[connection]; ok {
				close(connection.Send)
				delete(clientManager.Clients, connection)
				jsonMessage, _ := json.Marshal(&Message{Content: MessageContent{Text: "A client has disconnected.", Role: "SYSTEM"}})
				clientManager.send(jsonMessage, connection)
			}
		case message := <-clientManager.Broadcast:
			for client := range clientManager.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(clientManager.Clients, client)
				}
			}
		}
	}
}

func (clientManager *ClientManager) send(message []byte, currentClient *Client) {
	for client := range clientManager.Clients {

		fmt.Printf("\nSending message to client from %v to %v\n", currentClient.Id, client.Id)
		if client.Id != currentClient.Id {
			client.Send <- message
		} else {
			fmt.Printf("Not sending because same client")
		}
	}
}

func (client *Client) read(manager *ClientManager) {
	defer func() {
		manager.Unregister <- client
		err := client.Socket.Close()
		if err != nil {
			fmt.Printf("Error encountered when closing socket %v", err)
			return
		}
	}()

	for {
		_, message, err := client.Socket.ReadMessage()
		if err != nil {
			manager.Unregister <- client
			err2 := client.Socket.Close()
			if err2 != nil {
				fmt.Printf("Error encountered when failing to close socket after failing to reading message %v %v", err, err2)
				return
			}
		}

		jsonMessage, _ := json.Marshal(&Message{Sender: client.Id, Content: MessageContent{Text: string(message), Role: "USER"}})

		manager.Broadcast <- jsonMessage
	}
}

func (client *Client) write() {
	defer func() {
		err := client.Socket.Close()
		if err != nil {
			fmt.Printf("Error encountered when closing socket %v", err)
			return
		}
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				err := client.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					fmt.Printf("Error writing close message to web socket%v", err)
					return
				}
			}

			err := client.Socket.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Printf("Error writing to web socket %v to client %v", err, client.Id)
				return
			}
		}
	}
}
