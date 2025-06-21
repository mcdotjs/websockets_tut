package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager

	//egress is used to avoid concurrent writes on websocket connectoin
	egress chan []byte
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan []byte),
	}
}

// go routine
func (client *Client) readMessages() {
	//if closed connection in some way ... clean it up
	defer func() {
		client.manager.removeClient(client)
	}()
	for {

		messageType, payload, err := client.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error printing message: $v", err)
			}
			break
		}
		log.Println("messageType", messageType)
		log.Println("payload", string(payload))
		//NOTE: just for test
		for wsclient := range client.manager.clients {
			wsclient.egress <- payload
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
	for {
		select {
		case message, ok := <-client.egress:
			if !ok {

				if err := client.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				return
			}

			if err := client.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("failed to send message", err)
			}
			fmt.Println("message sent")
		}
	}
}
