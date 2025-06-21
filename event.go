package main

import "encoding/json"

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NOTE: like in FE... use type field to route it to proper event handler
// then I will use it in manager
// generic json payload... is up to function which receive Event to handle payload
type EventHandler func(event Event, c *Client) error

// when EventSendMessage is triggered we await json with message and from
var (
	EventSendMessage = "send_message"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}
