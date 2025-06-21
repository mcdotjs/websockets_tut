package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	//take http request and upgrade it to websocket connection
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")
	//upgrade regular connection to websocket
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("webosocket upgrade error", err)
		return
	}

	conn.Close()
}
