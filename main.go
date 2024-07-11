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

	//llm := NewLlmConnector(config.LlmEndpoint)

	// Register callbacks.
	//bot.RegisterWebhookEventCallback(llm.LlmCallback)

	bot.RegisterWebhookEventCallback(SimpleWebhookEventCallback)
	// Start the chatbot
	_ = bot.Start()

}
