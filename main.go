package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// loadConfig loads the server configuration from a file.
func loadConfig(configPath string) ServerConfig {
	// Try to read the specified config file
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("[Init] Could not read specified config file: %s\n", configPath)
		fmt.Println("[Init] Trying default config file: config.yaml")

		// Try to read the default config file
		configFile, err = os.ReadFile("config.yaml")
		if err != nil {
			fmt.Println("[Init] Error: No config file found.")
			fmt.Println("Usage: ./program --config [path_to_config_file]")
			fmt.Println("If no config file is specified, the program will try to use 'config.yaml' in the current directory.")
			os.Exit(1)
		}
	}

	fmt.Printf("[Init] Using config file: %s\n", configPath)

	// Parse the configuration
	var config ServerConfig
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("[Init] Error parsing config file: %v", err)
	}

	return config
}

func main() {
	// Define command-line flags
	configPath := flag.String("config", "config.yaml", "Path to the config file")
	llmDebug := flag.Bool("llm-debug", false, "Enable debug mode for LLM")

	// Parse command-line flags
	flag.Parse()

	// Print debug information if --llm-debug flag is set
	if *llmDebug {
		fmt.Println("[Init] LLM Local Debug mode is enabled. Requests won't be sent to the LLM endpoint.")
	}

	// Load the configuration
	config := loadConfig(*configPath)

	// Create a new chatbot instance
	bot := NewChatbotFromConfig(config)

	llm := NewLlmConnector(config.Channels, *llmDebug)

	// Register callbacks.
	bot.RegisterWebhookEventCallback(
		ScheduleCallbackNormalPriority(llm.LlmCallback),
	)

	bot.RegisterWebhookEventCallback(
		ScheduleCallbackHighestPriority(SimpleWebhookEventCallback),
	)
	// Start the chatbot
	if err := bot.Start(); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}
}
