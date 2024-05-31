package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	tp "taipeion/core"

	"golang.org/x/sync/semaphore"
)

type WebhookEventCallback func(*TaipeionBot, ChatbotWebhookEvent) error

// Define a struct for the response
type response struct {
	Status string `json:"status"`
}

type ServerConfig struct {
	Endpoint           string `yaml:"taipeion-endpoint"`
	ChannelSecret      string `yaml:"channel-secret"`
	ChannelAccessToken string `yaml:"channel-access-token"`
	Address            string `yaml:"address"`
	Port               int16  `yaml:"port"`
}

type ChatbotWebhookEvent struct {
	Destination int64
	tp.MessageEvent
}

type TaipeionBot struct {
	Endpoint           string
	ChannelSecret      string
	ChannelAccessToken string
	ServerAddress      string
	ServerPort         int16
	eventQueue         chan ChatbotWebhookEvent
	eventHandlers      []WebhookEventCallback
	eventSemaphore     *semaphore.Weighted
}

// # Simple webhook event callback.
//
// This callback puts the incoming message to the log (stdout).
func SimpleWebhookEventCallback(bot *TaipeionBot, event ChatbotWebhookEvent) error {
	// Check if incoming event is text message.
	if event.Message.Type != "text" {
		log.Println("[SimpleStdCallback] Received non-text message. Ignoring.")
		return nil
	}

	log.Printf("[SimpleStdCallback] Received event: %#v\n", event)
	return nil
}

// # Private message callback.
//
// This callback is used to send a reply to a cer
func PrivateMessageCallback(bot *TaipeionBot, event ChatbotWebhookEvent) error {
	// Check if incoming event is text message.
	if event.Message.Type != "text" {
		log.Println("[PrivateMessageCallback] Received non-text message. Ignoring.")
		return nil
	}

	reply_message := event.Message.Text
	receiver := event.Source.UserId

	// Process the incoming message.
	reply_message = fmt.Sprintf("Bot: %s", reply_message)

	return bot.SendPrivateMessage(receiver, reply_message)

}

// # Enqueue an incoming webhook event.
func (tpb *TaipeionBot) enqueueWebhookIncomingEvent(event ChatbotWebhookEvent) {
	tpb.eventQueue <- event
}

// # Send a broadcast message to all users.
//
// This method uses the message API to send a broadcast message to all users who have subscribed to the channel.
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

// # Send a private message to a user.
//
// This method uses the message API to send a private message to a user.
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
	log.Println(string(data))

	req, err := http.NewRequest("POST", tpb.Endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", tpb.ChannelAccessToken)

	log.Println(req.Header)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Verbose logging
	log.Println("Response Status:", resp.Status)
	log.Println("Response Headers:", resp.Header)
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
			response := response{Status: "no content"}
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
		log.Println("[EventListener] Received header:", r.Header)

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
			internal_event := ChatbotWebhookEvent{
				Destination:  payload.Destination,
				MessageEvent: event,
			}

			// Enqueue the event
			tpb.enqueueWebhookIncomingEvent(internal_event)
		}

		// Create the response object
		response := response{Status: "success"}

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

// # The main event processor loop.
//
// This function is the main loop for processing incoming events.
// It waits for incoming events and processes them using the registered event handlers.
func (tpb *TaipeionBot) EventProcessorLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done(): // Check if the context is cancelled.
			log.Println("Context cancelled. Exiting event processor loop.")
			return nil

		case event := <-tpb.eventQueue: // Wait for incoming events.
			log.Printf("[EventProcessor] Processing event: %#v\n", event)
			for _, handler := range tpb.eventHandlers { // Iterate over the event handlers.
				log.Printf("[EventProcessor] Processing event with handler: %#v\n", handler)
				go tpb.eventProcessorInternalCallbackWrapper(ctx, handler, event) // Call the handler in a goroutine.
			}
		}
	}
}

// # The wrapper for the event processor callback.
//
// Since we've simplified the callback to a single function, we can use this wrapper to handle the semaphore.
// So there's no need to deal with the semaphore or context in the callback function.
//
// The function is for internal use only.
func (tpb *TaipeionBot) eventProcessorInternalCallbackWrapper(ctx context.Context, event_handler WebhookEventCallback, event ChatbotWebhookEvent) error {
	tpb.eventSemaphore.Acquire(ctx, 1) // Acquire the semaphore, wait until available.
	err := event_handler(tpb, event)   // Call the event handler.
	tpb.eventSemaphore.Release(1)      // Release the semaphore if callback is done.
	return err
}

// # Register a webhook event callback.
func (tpb *TaipeionBot) RegisterWebhookEventCallback(callback WebhookEventCallback) {
	tpb.eventHandlers = append(tpb.eventHandlers, callback)
}

func (tpb *TaipeionBot) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tpb.eventSemaphore = semaphore.NewWeighted(1) // TODO: Adjust the semaphore weight.

	go tpb.EventProcessorLoop(ctx) // Start the event processor loop

	tpb.eventQueue = make(chan ChatbotWebhookEvent, 100)
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
