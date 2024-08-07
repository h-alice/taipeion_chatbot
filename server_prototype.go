package main

import (
	tp "taipeion/core"

	api_platform "github.com/h-alice/tcg-api-platform-client"

	"golang.org/x/sync/semaphore"
)

type WebhookEventCallback func(*TaipeionBot, ChatbotWebhookEvent) error

type eventHandlerEntry struct {
	Callback   WebhookEventCallback
	IsPriority bool
}

// Define a struct for the response
type response struct {
	Status string `json:"status"`
}

// Definiton of the channel struct.
type Channel struct {
	ChannelSecret      string `yaml:"channel-secret"`
	ChannelAccessToken string `yaml:"channel-access-token"`
	ChannelLlmEndpoint string `yaml:"llm-endpoint"`
}

type ChannelIdConfigMap map[int]Channel

type ServerConfig struct {
	Endpoint               string             `yaml:"taipeion-endpoint"`
	Channels               ChannelIdConfigMap `yaml:"channels"`
	Address                string             `yaml:"address"`
	Port                   int16              `yaml:"port"`
	ApiPlatformEndpoint    string             `yaml:"api-platform-endpoint"`
	ApiPlatformClientId    string             `yaml:"api-platform-client-id"`
	ApiPlatformClientToken string             `yaml:"api-platform-client-token"`
	MaxConcurrentEvent     int                `yaml:"max-concurrent-event-handlers"`
}

type ChatbotWebhookEvent struct {
	Destination int
	tp.MessageEvent
}

type TaipeionBot struct {
	Endpoint       string
	Channels       map[int]Channel
	ServerAddress  string
	ServerPort     int16
	eventQueue     chan ChatbotWebhookEvent
	eventHandlers  []eventHandlerEntry
	eventSemaphore *semaphore.Weighted
	maxConcurrent  int
	api_client     *api_platform.ApiPlatformClient
}
