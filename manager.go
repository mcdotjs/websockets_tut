package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	//take http request and upgrade it to websocket connection
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
)

type Manager struct {
	clients ClientList
	sync.RWMutex

	otps RetentionMap

	handlers map[string]EventHandler
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		otps:     NewRetentionMap(ctx, 5*time.Second),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	// checking if have event type on event
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("There is no such a handler")
	}
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	//NOTE: before we connect we need check otp from quuery string
	otp := r.URL.Query().Get("otp")
	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !m.otps.VerifyOTP(otp) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
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
	go client.writeMessages()

}

func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	type userLoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req userLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "user" && req.Password == "123" {
		// on fe we expect otp in login reponse
		type response struct {
			OTP string `json:"otp"`
		}
		otp := m.otps.NewOTP()

		resp := response{
			OTP: otp.Key,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("error mashaling otp when login", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
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

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	fmt.Println("send message be", event)
	return nil
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	switch origin {
	case "http://localhost:8080":
		return true
	default:
		return false

	}
}
