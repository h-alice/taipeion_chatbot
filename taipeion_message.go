package taipeion

import (
	"fmt"
)

type Message struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Text string `json:"text"`
}
type MessageSource struct {
	Type   string `json:"type"`
	UserId string `json:"userId"`
}

type MessageEvent struct {
	Type      string        `json:"type"`
	Timestamp int64         `json:"timestamp"`
	Source    MessageSource `json:"source"`
	Message   Message       `json:"message"`
}

type Payload struct {
	Destination string         `json:"destination"`
	Events      []MessageEvent `json:"events"`
}

func (m *Message) String() string {
	return fmt.Sprintf("Message{Type: %s, Id: %s, Text: %s}", m.Type, m.Id, m.Text)
}
