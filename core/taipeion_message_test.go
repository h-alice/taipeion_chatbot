package taipeion_core

import (
	"testing"
)

func TestMessageSerializeWithOmittedField(t *testing.T) {

	excepted := `{"ask":"broadcastMessage","message":{"type":"text","text":"Hello, World!"}}`
	// Create a new message
	message := Message{
		Type: "text",
		Text: "Hello, World!",
	}
	payload := ChannelMessagePayload{
		Ask:     "broadcastMessage",
		Message: message,
	}

	// Serialize the message
	data, err := payload.Serialize()
	if err != nil {
		t.Errorf("Error serializing message: %v", err)
	}
	t.Log("Serialized message:", string(data))

	if string(data) != excepted {
		t.Errorf("Expected %s, got %s", excepted, string(data))
	}

}
