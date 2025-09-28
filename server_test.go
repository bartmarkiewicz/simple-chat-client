package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// helper to read from a channel with timeout to avoid deadlocks in tests
func receiveWithTimeout[T any](t *testing.T, ch <-chan T) (T, bool) {
	t.Helper()
	var zero T
	timer := time.NewTimer(time.Millisecond * 100)
	defer timer.Stop()
	select {
	case v := <-ch:
		return v, true
	case <-timer.C:
		return zero, false
	}
}

func getClientManagerAndStartWebSocketServer() *ClientManager {
	clientManager := NewClientManager()
	go clientManager.StartWebSocketServer()
	return clientManager
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		t.Fatalf("Expected: %v Actual %v", expected, actual)
	}
}

func TestStartWebSocketServer_RegisterSendsConnectedToOthers(t *testing.T) {
	clientManager := getClientManagerAndStartWebSocketServer()

	client1 := &Client{Id: "client1", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}
	client2 := &Client{Id: "client2", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}

	clientManager.Register <- client1
	if _, ok := receiveWithTimeout(t, client1.Send); ok {
		t.Fatalf("client1 should not receive a message when it registers alone")
	}

	clientManager.Register <- client2

	raw, ok := receiveWithTimeout(t, client1.Send)
	if !ok {
		t.Fatalf("expected client1 to receive a system 'connected' message")
	}

	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("invalid json in system message: %v", err)
	}
	if msg.Content.Role != "SYSTEM" || msg.Content.Text == "" {
		t.Fatalf("expected SYSTEM role with non-empty text, got: %+v", msg)
	}

	if _, ok := receiveWithTimeout(t, client2.Send); ok {
		t.Fatalf("client2 should not receive a system message about itself")
	}
}

func TestStartWebSocketServer_UnregisterRemovesAndNotifiesOthers(t *testing.T) {
	clientManager := getClientManagerAndStartWebSocketServer()

	client1 := &Client{Id: "client1", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}
	client2 := &Client{Id: "client2", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}

	clientManager.Register <- client1
	_, _ = receiveWithTimeout(t, client1.Send)

	clientManager.Register <- client2
	_, _ = receiveWithTimeout(t, client1.Send)

	clientManager.Unregister <- client2

	raw, ok := receiveWithTimeout(t, client1.Send)
	if !ok {
		t.Fatalf("expected client1 to receive a system 'disconnected' message")
	}

	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("invalid json in system message: %v", err)
	}
	if msg.Content.Role != "SYSTEM" || msg.Content.Text == "" {
		t.Fatalf("expected SYSTEM role with non-empty text, got: %+v", msg)
	}

	for {
		_, ok := <-client2.Send
		if ok {
			t.Fatalf("client2.Send should be closed after unregister")
		} else {
			return
		}
	}
}

func TestStartWebSocketServer_BroadcastToAllActiveClients(t *testing.T) {
	clientManager := getClientManagerAndStartWebSocketServer()

	client1 := &Client{Id: "client1", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}
	client2 := &Client{Id: "client2", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}
	client3 := &Client{Id: "client2", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}

	clientManager.Register <- client1
	clientManager.Register <- client2
	clientManager.Register <- client3

	_, _ = receiveWithTimeout(t, client1.Send)
	_, _ = receiveWithTimeout(t, client1.Send)
	_, _ = receiveWithTimeout(t, client2.Send)

	payload := []byte(`{"hello":"world"}`)
	clientManager.Broadcast <- payload

	got1, ok1 := receiveWithTimeout(t, client1.Send)
	got2, ok2 := receiveWithTimeout(t, client2.Send)
	got3, ok3 := receiveWithTimeout(t, client3.Send)

	if !ok1 || !ok2 || !ok3 {
		t.Fatalf("expected all clients to receive broadcast; client1:%v client2:%v client3:%v", ok1, ok2, ok3)
	}

	assertEqual(t, string(got1), string(payload))
	assertEqual(t, string(got2), string(payload))
	assertEqual(t, string(got3), string(payload))
}

func TestStartWebSocketServer_RemovesClientWhenSendWouldBlock(t *testing.T) {
	clientManager := getClientManagerAndStartWebSocketServer()

	stalled := &Client{Id: "stalled", Socket: &websocket.Conn{}, Send: make(chan []byte)}
	clientManager.Register <- stalled

	deadline := time.Now().Add(2 * time.Second)
	for {
		if clientManager.Clients[stalled] {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("client was not registered in time")
		}
		time.Sleep(50 * time.Millisecond)
	}

	clientManager.Broadcast <- []byte("payload")

	// Poll until the channel is closed or the client is removed.
	waitDeadline := time.Now().Add(2 * time.Second)
	closed := false
	for time.Now().Before(waitDeadline) {
		select {
		case _, ok := <-stalled.Send:
			if !ok {
				closed = true
			}
		default:
		}
		if _, exists := clientManager.Clients[stalled]; !exists {
			closed = true
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if !closed {
		t.Fatalf("expected stalled client's Send channel to be closed and client removed after broadcast")
	}

	if _, exists := clientManager.Clients[stalled]; exists {
		t.Fatalf("expected stalled client to be removed from manager after failed send")
	}
}

func TestSend_SkipsCurrentClient(t *testing.T) {
	clientManager := NewClientManager()

	current := &Client{Id: "sender", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}
	other := &Client{Id: "other", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}

	clientManager.Clients[current] = true
	clientManager.Clients[other] = true

	msg := []byte("hello")
	clientManager.send(msg, current)

	if got, ok := receiveWithTimeout(t, other.Send); !ok || string(got) != string(msg) {
		t.Fatalf("expected other to receive message")
	}

	if _, ok := receiveWithTimeout(t, current.Send); ok {
		t.Fatalf("expected current client to not receive its own message")
	}
}

func TestWebsocketPage_ReturnsNotFoundOnUpgradeFailure(t *testing.T) {
	clientManager := NewClientManager()
	req := httptest.NewRequest("GET", "/web-socket", nil)
	rr := httptest.NewRecorder()

	clientManager.WebsocketPage(rr, req)

	assertEqual(t, http.StatusBadRequest, rr.Code)
}

func TestSend_OnlySelfNoDelivery(t *testing.T) {
	clientManager := NewClientManager()
	self := &Client{Id: "self", Socket: &websocket.Conn{}, Send: make(chan []byte, 1)}
	clientManager.Clients[self] = true

	clientManager.send([]byte("msg"), self)

	if _, ok := receiveWithTimeout(t, self.Send); ok {
		t.Fatalf("sender should not receive its own message")
	}
}
