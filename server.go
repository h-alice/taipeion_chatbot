package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	tp "taipeion/taipeion"

	"gopkg.in/yaml.v3"
)

// Define a struct for the response
type Response struct {
	Status string `json:"test"`
}

type ServerConfig struct {
	Endpoint           string `yaml:"taipeion-endpoint"`
	ChannelSecret      string `yaml:"channel-secret"`
	ChannelAccessToken string `yaml:"channel-access-token"`
	Address            string `yaml:"address"`
	Port               int16  `yaml:"port"`
}

type InternalWebhookEvent struct {
	Destination string
	tp.MessageEvent
}

type TaipeionBot struct {
	Endpoint           string
	ChannelSecret      string
	ChannelAccessToken string
	ServerAddress      string
	ServerPort         int16
	eventQueue         chan InternalWebhookEvent
}

func (tpb *TaipeionBot) EnqueueWebhookIncomingEvent(event InternalWebhookEvent) {
	tpb.eventQueue <- event
}

func (tpb *TaipeionBot) SendBroadcastMessage(message string) error {
	// Craft the message
	ch_payload := tp.ChannelMessagePayload{
		Ask: "broadcastMessage",
		Message: tp.Message{
			Type: "text",
			Text: message,
		},
	}

	// Serialize the message
	data, err := ch_payload.Serialize()
	if err != nil {
		return err
	}

	// Send the message
	return tpb.DoEndpointPostRequest(tpb.Endpoint, data)
}

func (tpb *TaipeionBot) SendPrivateMessage(userId string, message string) error {
	// Craft the message
	ch_payload := tp.ChannelMessagePayload{
		Ask:       "sendMessage",
		Recipient: userId,
		Message: tp.Message{
			Type: "text",
			Text: message,
		},
	}

	// Serialize the message
	data, err := ch_payload.Serialize()
	if err != nil {
		return err
	}

	// Send the message
	return tpb.DoEndpointPostRequest(tpb.Endpoint, data)

}

func (tpb *TaipeionBot) DoEndpointPostRequest(endpoint string, data []byte) error {
	req, err := http.NewRequest("POST", tpb.Endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", tpb.ChannelAccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Verbose logging
	log.Println("Response Status:", resp.Status)
	log.Panicln("Response Headers:", resp.Header)
	body, _ := io.ReadAll(resp.Body)
	log.Println("Response Body:", string(body))

	return nil
}

func (tpb *TaipeionBot) webhookEventListener() error {

	// We embed the handler function inside the function to access the bot instance.
	webhookIncomeRequestHandler := func(w http.ResponseWriter, r *http.Request) {

		// Check header for content length.
		if r.Header.Get("Content-Length") == "" || r.Header.Get("Content-Length") == "0" {
			// Ignore and send OK.
			response := Response{Status: "no content"}
			w.WriteHeader(http.StatusAccepted)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Print Header
		log.Println("Header:", r.Header)

		// Deserialize the message
		payload, err := tp.DeserializeWebhookMessage(body)

		if err != nil {
			log.Println("Error deserializing message:", err)
			// Fallback to printing the raw body
			log.Println("Received payload:", string(body))
			http.Error(w, "Malformed payload.", http.StatusBadRequest)
			return
		}

		// Iterate over the events.
		for _, event := range payload.Events {
			// Create an internal event
			internal_event := InternalWebhookEvent{
				Destination:  payload.Destination,
				MessageEvent: event,
			}

			// Enqueue the event
			tpb.EnqueueWebhookIncomingEvent(internal_event)
		}

		// Create the response object
		response := Response{Status: "success"}

		// Set the response status to OK
		w.WriteHeader(http.StatusOK)

		// Set the response header to indicate JSON content
		w.Header().Set("Content-Type", "application/json")

		// Encode the response object to JSON and send it
		json.NewEncoder(w).Encode(response)
	}

	http.HandleFunc("/", webhookIncomeRequestHandler) // Set the default handler.

	// Start the server.
	full_server_address := fmt.Sprintf("%s:%d", tpb.ServerAddress, tpb.ServerPort)
	log.Println("[EventListener] Starting server at ", full_server_address)

	return http.ListenAndServe(full_server_address, nil) // Serve until error.
}

func (tpb *TaipeionBot) EventProcessorLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled. Exiting event processor loop.")
			return nil

		case event := <-tpb.eventQueue:
			// Process the event
			// For current implementation, we just echo the message back.
			// Print the message
			log.Printf("Received event: %#v\n", event)
		}

	}
}

func (tpb *TaipeionBot) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go tpb.EventProcessorLoop(ctx) // Start the event processor loop

	tpb.eventQueue = make(chan InternalWebhookEvent, 100)
	return tpb.webhookEventListener()
}

func NewChatbotInstance(endpoint string, channelSecret string, channelAccessToken string, serverAddress string, serverPort int16) *TaipeionBot {
	return &TaipeionBot{
		Endpoint:           endpoint,
		ChannelSecret:      channelSecret,
		ChannelAccessToken: channelAccessToken,
		ServerAddress:      serverAddress,
		ServerPort:         serverPort,
	}
}

func NewChatbotFromConfig(config ServerConfig) *TaipeionBot {
	return NewChatbotInstance(config.Endpoint, config.ChannelSecret, config.ChannelAccessToken, config.Address, config.Port)
}

func main() {
	// Read the server configuration from file.
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	// Parse the configuration
	var config ServerConfig
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal("Error parsing config file:", err)
	}

	// Create a new chatbot instance
	bot := NewChatbotFromConfig(config)

	// Start the chatbot
	_ = bot.Start()

}
