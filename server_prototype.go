package main

import (
	tp "taipeion/core"

	api_platform "github.com/h-alice/tcg-api-platform-client"

	"golang.org/x/sync/semaphore"
)

// The prototype of the webhook event callback.
type WebhookEventCallback func(*TaipeionBot, ChatbotWebhookEvent) error

type eventHandlerEntry struct {
	Callback   WebhookEventCallback // The callback function.
	IsPriority bool                 // Indicates if the callback is a priority callback (bypass the concuurancy limit).
}

// Define a struct for the response
type response struct {
	Status string `json:"status"` // Status of an empty response.
}

// Definiton of the channel struct.
type Channel struct {
	ChannelSecret      string `yaml:"channel-secret"`       // The secret of the channel, get it from TaipeiON admin panel.
	ChannelAccessToken string `yaml:"channel-access-token"` // The access token of the channel.
	ChannelLlmEndpoint string `yaml:"llm-endpoint"`         // The endpoint of the LLM server for this channel.
}

type ChannelIdConfigMap map[int]Channel // A map from channel ID to channel configuration.

type ServerConfig struct {
	Endpoint               string             `yaml:"taipeion-endpoint"`             // The endpoint of the Taipeion server.
	Channels               ChannelIdConfigMap `yaml:"channels"`                      // The configuration of the channels.
	Address                string             `yaml:"address"`                       // Local IP to listen on.
	Port                   int16              `yaml:"port"`                          // Local port to listen on.
	ApiPlatformEndpoint    string             `yaml:"api-platform-endpoint"`         // The endpoint of the API platform.
	ApiPlatformClientId    string             `yaml:"api-platform-client-id"`        // The client ID of the API platform.
	ApiPlatformClientToken string             `yaml:"api-platform-client-token"`     // The client token of the API platform.
	MaxConcurrentEvent     int                `yaml:"max-concurrent-event-handlers"` // Maximum number of concurrent event handlers.
}

type ChatbotWebhookEvent struct {
	Destination     int // ID of incoming channel. Since the Destination field is not in the event object, we need to add it.
	tp.MessageEvent     // The message event.
}

type TaipeionBot struct {
	Endpoint       string                          // The endpoint of the Taipeion server.
	Channels       map[int]Channel                 // A map from channel ID to channel configuration.
	ServerAddress  string                          // The address to listen on.
	ServerPort     int16                           // The port to listen on.
	eventQueue     chan ChatbotWebhookEvent        // Event queue, every incoming event will be put into this queue.
	eventHandlers  []eventHandlerEntry             // Event handlers.
	eventSemaphore *semaphore.Weighted             // Semaphore for event handlers.
	maxConcurrent  int                             // Maximum number of concurrent event handlers.
	api_client     *api_platform.ApiPlatformClient // Insrance of the API platform client.
}
