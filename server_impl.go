package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	tp "taipeion/core"

	api_platform "github.com/h-alice/tcg-api-platform-client"

	"golang.org/x/sync/semaphore"
)

// # Enqueue an incoming webhook event.
func (tpb *TaipeionBot) enqueueWebhookIncomingEvent(event ChatbotWebhookEvent) {
	tpb.eventQueue <- event
}

// # Broadcast message sender
//
// This method uses the message API to send a broadcast message to all users who have subscribed to the channel.
//
// Parameters:
// - message: The message to be sent.
// - target_channel: The channel's ID to send the message to.
func (tpb *TaipeionBot) SendBroadcastMessage(message string, target_channel int) error {
	// Craft the message
	ch_payload := tp.ChannelMessagePayload{
		Ask: "broadcastMessage",
		Message: tp.Message{
			Type: "text",
			Text: message,
		},
	}

	// Send the message
	return tpb.DoEndpointPostRequest(tpb.Endpoint, ch_payload, target_channel)
}

// # Private message sender
//
// This method uses the message API to send a private message to a user.
//
// Parameters:
// - userId: The user's ID to send the message to.
// - message: The message to be sent.
// - target_channel: The channel's ID to send the message to.
func (tpb *TaipeionBot) SendPrivateMessage(userId string, message string, target_channel int) error {
	// Craft the message
	ch_payload := tp.ChannelMessagePayload{
		Ask:       "sendMessage",
		Recipient: userId,
		Message: tp.Message{
			Type: "text",
			Text: message,
		},
	}

	// Send the message
	return tpb.DoEndpointPostRequest(tpb.Endpoint, ch_payload, target_channel)

}

// # Perform a POST request to the TaipeiON endpoint
func (tpb *TaipeionBot) DoEndpointPostRequest(endpoint string, channelPayload tp.ChannelMessagePayload, target_channel int) error {

	_, err := tpb.api_client.RequestAccessToken()
	if err != nil {
		log.Printf("[ReqSender] Error: Unable to request access token: %s\n", err)
		return err
	}

	_, err = tpb.api_client.RequestSignBlock()
	if err != nil {
		log.Printf("[ReqSender] Error: Unable to request sign block: %s\n", err)
		return err
	}

	log.Println(channelPayload)

	headers := map[string]string{
		"Content-Type": "application/json",
		"backAuth":     tpb.Channels[target_channel].ChannelAccessToken,
	}

	// Perform the request
	resp, err := tpb.api_client.SendRequest(tpb.Endpoint, "POST", headers, channelPayload, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Verbose logging
	body, _ := io.ReadAll(resp.Body)
	log.Printf("[ReqSender] Response (%d): %s\n", resp.StatusCode, string(body))

	return nil
}

// # Income Request Handler Factory
//
// This function creates a handler for incoming requests.
func (tpb *TaipeionBot) incomeRequestHandlerFactory() func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		// Check header for content length.
		if r.Header.Get("Content-Length") == "" || r.Header.Get("Content-Length") == "0" {
			// Ignore and send OK.
			response := response{Status: "no content"}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "[EvHandler] Error: Unable to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Deserialize the message
		payload, err := tp.DeserializeWebhookMessage(body)

		if err != nil {
			log.Println("[EvHandler] Error: Unable to deserializing message:", err)
			// DEBUG: Fallback to printing the raw body
			log.Println("[EvHandler] Received payload:", string(body))
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
	}
}

func (tpb *TaipeionBot) webhookEventListener() error {

	http.HandleFunc("/", tpb.incomeRequestHandlerFactory()) // Set the default handler.

	// Start the server.
	full_server_address := fmt.Sprintf("%s:%d", tpb.ServerAddress, tpb.ServerPort)
	log.Println("[EvListener] Starting server at ", full_server_address)

	return http.ListenAndServe(full_server_address, nil) // Serve until error.
}

// # The Main Event Processor Loop
//
// This function is the main loop for processing incoming events.
// It waits for incoming events and processes them using the registered event handlers.
func (tpb *TaipeionBot) EventProcessorLoop(ctx context.Context) error {
	log.Println("[EvLoop] Starting event processor loop.")
	for {
		select {
		case <-ctx.Done(): // Check if the context is cancelled.
			log.Println("[EvLoop] Context cancelled. Exiting event processor loop.")
			return nil

		case event := <-tpb.eventQueue: // Wait for incoming events.
			log.Printf("[EvProcessor] Processing event: %#v\n", event)
			for _, event_handler := range tpb.eventHandlers { // Iterate over the event handlers.
				log.Printf("[EvProcessor] Processing event with handler: %#v\n", event_handler.Callback)
				go tpb.eventProcessorInternalCallbackWrapper(ctx, event_handler, event) // Call the handler in a goroutine.
			}
		}
	}
}

// # Event Processor Callback Wrapper
//
// Since we've simplified the callback to a single function, we can use this wrapper to handle the semaphore.
// So there's no need to deal with the semaphore or context in the callback function.
//
// The function is for internal use only.
func (tpb *TaipeionBot) eventProcessorInternalCallbackWrapper(ctx context.Context, event_handler_entry eventHandlerEntry, event ChatbotWebhookEvent) error {
	ctx = context.WithoutCancel(ctx)
	if event_handler_entry.IsPriority {
		return event_handler_entry.Callback(tpb, event) // Directly call the event handler.
	} else {
		tpb.eventSemaphore.Acquire(ctx, 1)              // Acquire the semaphore, wait until available.
		err := event_handler_entry.Callback(tpb, event) // Call the event handler.

		tpb.eventSemaphore.Release(1) // Release the semaphore if callback is done.
		return err
	}
}

// # Webhook Event Registration
//
// Register a webhook event callback.
// All registered callbacks will be called when an event is received.
func (tpb *TaipeionBot) RegisterWebhookEventCallback(ev_handler_entry eventHandlerEntry) {
	tpb.eventHandlers = append(tpb.eventHandlers, ev_handler_entry)
}

// # Main loop
//
// The main loop of the chatbot.
func (tpb *TaipeionBot) mainLoop(ctx context.Context) error {

	// Create a child context.
	ctx_child, cancel := context.WithCancel(ctx)
	defer cancel() // All child coroutines will be cancelled upon main loop termination.

	// Create a channel for errors.
	subroutine_err := make(chan error)

	// Create a semaphore for concurrency control.
	tpb.eventSemaphore = semaphore.NewWeighted(int64(tpb.maxConcurrent))

	// Start all child coroutines.
	log.Println("[Daemon] Starting all child coroutines.")

	// Start the event processor loop.
	go func(ctx context.Context, err_chan chan error) {
		select {
		case <-ctx.Done(): // Check if the context is cancelled.
			log.Println("[EvListener] Received cancel signal.")
			return
		default:
			err := tpb.EventProcessorLoop(ctx_child) // Start the event processor loop.
			if err != nil {                          // The subroutine has returned an error.
				err_chan <- err
			}
		}

	}(ctx_child, subroutine_err)

	// Start the webhook event listener.
	go func(ctx context.Context, err_chan chan error) {

		select {
		case <-ctx.Done(): // Check if the context is cancelled.
			log.Println("[EvListener] Received cancel signal.")
			return

		default:
			err := tpb.webhookEventListener()
			if err != nil { // The subroutine has returned an error.
				err_chan <- err
			}
		}
	}(ctx_child, subroutine_err)

	// Wait for signals.
	select {

	case <-ctx.Done(): // Check if the context is cancelled.
		log.Println("[Daemon] Received cancelling signal, terminating all subroutines.")
		return nil

	default:
		// Wait for the first error to occur.
		err := <-subroutine_err
		if err != nil {
			return err
		}
	}
	return nil
}

func (tpb *TaipeionBot) Start() error {

	// Create the event queue.
	// NOTE: Consider design a queue flushing machanism.
	tpb.eventQueue = make(chan ChatbotWebhookEvent, 100)

	for {
		// Create a new context.
		ctx, cancel := context.WithCancel(context.Background())

		// Start the main loop.
		err := tpb.mainLoop(ctx)
		if err != nil { // The main loop has returned an error.
			log.Println("[Daemon] Main loop returned an error:", err)
			cancel()
		}

		continue // Restart the main loop.
	}
}

// # New Chatbot Instance
//
// Create a new chatbot instance.
func NewChatbotInstance(
	endpoint string,
	channels map[int]Channel,
	serverAddress string,
	serverPort int16,
	apiPlatformEndpoint string,
	apiPlatformClientId string,
	apiPlatformClientToken string,
	maxConcurrentEvent int) *TaipeionBot {

	return &TaipeionBot{
		Endpoint:      endpoint,
		Channels:      channels,
		ServerAddress: serverAddress,
		ServerPort:    serverPort,
		maxConcurrent: maxConcurrentEvent,
		api_client:    api_platform.NewApiPlatformClient(apiPlatformEndpoint, apiPlatformClientId, apiPlatformClientToken),
	}
}

// # New Chatbot Instance from Configuration
//
// Create a new chatbot instance from a configuration.
func NewChatbotFromConfig(config ServerConfig) *TaipeionBot {
	return NewChatbotInstance(
		config.Endpoint,
		config.Channels,
		config.Address,
		config.Port,
		config.ApiPlatformEndpoint,
		config.ApiPlatformClientId,
		config.ApiPlatformClientToken,
		config.MaxConcurrentEvent)
}
