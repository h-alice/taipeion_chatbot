package taipeion

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Type string `json:"type,omitempty"`
	Id   string `json:"id,omitempty"`
	Text string `json:"text,omitempty"`
}

// Received message source.
type MessageSource struct {
	Type   string `json:"type"`
	UserId string `json:"userId"`
}

// The message event from webhook.
type MessageEvent struct {
	Type      string        `json:"type"`
	Timestamp int64         `json:"timestamp"`
	Source    MessageSource `json:"source"`
	Message   Message       `json:"message"`
}

// This is the received payload from the webhook.
type WebhookPayload struct {
	Destination int64         `json:"destination"`
	Events      []MessageEvent `json:"events"`
}

type ChannelMessagePayload struct {
	Ask       string  `json:"ask"`
	Recipient string  `json:"recipient,omitempty"`
	Message   Message `json:"message"`
}

func DeserializeWebhookMessage(data []byte) (*WebhookPayload, error) {
	var payload WebhookPayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (cpd ChannelMessagePayload) Serialize() ([]byte, error) {
	data, err := json.Marshal(cpd)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m Message) String() string {
	return fmt.Sprintf("Type: %s\n, Id: %s, Text: %s", m.Type, m.Id, m.Text)
}
