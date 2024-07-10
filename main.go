package main

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func DelayedLoggingEventCallback(bot *TaipeionBot, event ChatbotWebhookEvent) error {
	// Check if incoming event is text message.
	if event.Message.Type != "text" {
		log.Println("[DelayedLogger] Received non-text message. Ignoring.")
		return nil
	}

	log.Printf("[DelayedLogger] >>>>>>>>>> Received event: %#v\n", event)

	log.Println("[DelayedLogger] >>>>>>>>>> Entering DelayedLoggingEventCallback, sleeping for 5 seconds.")

	time.Sleep(5 * time.Second)

	log.Println("[DelayedLogger] <<<<<<<<<< Leaving DelayedLoggingEventCallback.")
	return nil
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

	//llm := NewLlmConnector(config.LlmEndpoint)

	// Register callbacks.
	//bot.RegisterWebhookEventCallback(llm.LlmCallback)

	bot.RegisterWebhookEventCallback(SimpleWebhookEventCallback)
	// Start the chatbot
	_ = bot.Start()

}
