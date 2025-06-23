package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager

	//egress is used to avoid concurrent writes on websocket connectoin
	egress chan Event
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

// go routine
func (client *Client) readMessages() {
	//if closed connection in some way ... clean it up
	defer func() {
		client.manager.removeClient(client)
	}()

	//start timer
	if err := client.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println("never...", err)
	}

	client.connection.SetPongHandler(client.pongHandler)

	for {

		_, payload, err := client.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error printing message: $v", err)
			}
			break
		}
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			//NOTE: maybe dont have to break here
			break
		}
		if err := client.manager.routeEvent(request, client); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
}

// NOTE: write message
// connection (gorilla) can write one connection at a time
// thats problem when doing concurent stuff ..gorilla documentation ...
// write msg to unbuffered channel, which is than used to read from
// messages than will be taken one by one from channel
func (client *Client) writeMessages() {
	defer func() {
		client.manager.removeClient(client)
	}()

	ticker := time.NewTicker(pingInterval)
	for {
		select {
		case message, ok := <-client.egress:
			if !ok {

				if err := client.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				return
			}
			data, err := json.Marshal(message)
			if err != nil {
				fmt.Errorf("Writing message, marshal error: %v", err)
				return
			}

			if err := client.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("failed to send data", err)
			}
			fmt.Println("message sent")
		case <-ticker.C:
			log.Println("Ping")

			if err := client.connection.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Println("ping error", err)
				return
			}
		}
	}
}

func (client *Client) pongHandler(pongMsg string) error {
	log.Println("pong handler")
	//need to reset timer here
	return client.connection.SetReadDeadline(time.Now().Add(pongWait))
}
