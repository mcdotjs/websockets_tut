package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
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
	}
}

//NOTE: write message
// connection (gorilla) can write one connection at a time
// thats problem when doing concurent stuff ..gorilla documentation ...
// write msg to unbuffered channel, which is than used to read from
