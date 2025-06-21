package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	//take http request and upgrade it to websocket connection
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
	clients ClientList
	sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients: make(ClientList),
	}
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")
	//upgrade regular connection to websocket
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("webosocket upgrade error", err)
		return
	}

	client := NewClient(conn, m)
	m.addClient(client)

	//client processes
	go client.readMessages()

}

func (m *Manager) addClient(c *Client) {
	m.Lock()
	defer m.Unlock()
	m.clients[c] = true
	// 	if _, ok := m.clients[c]; !ok {
	// 		m.clients[c] = true
	// 	}
}

func (m *Manager) removeClient(c *Client) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[c]; ok {
		c.connection.Close()
		delete(m.clients, c)
	}
}
