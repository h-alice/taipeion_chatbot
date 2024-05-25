package taipeion

import (
	"fmt"
)

type Message struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Text string `json:"text"`
}

func (m *Message) String() string {
	return fmt.Sprintf("Message{Type: %s, Id: %s, Text: %s}", m.Type, m.Id, m.Text)
}
