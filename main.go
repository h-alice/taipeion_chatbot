package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

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

	llm := NewLlmConnector(config.Channels)

	// Register callbacks.
	bot.RegisterWebhookEventCallback(ScheduleNormalPriority(llm.LlmCallback))

	bot.RegisterWebhookEventCallback(ScheduleHighestPriority(SimpleWebhookEventCallback))
	// Start the chatbot
	_ = bot.Start()

}
