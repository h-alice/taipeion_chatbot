package taipeion

import (
	"encoding/json"
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

func DeserializeMessage(data []byte) (*Payload, error) {
	var payload Payload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (m Message) String() string {
	return fmt.Sprintf("Type: %s\n, Id: %s, Text: %s", m.Type, m.Id, m.Text)
}
