package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func loadConfig() ServerConfig {
	var configPath string
	var configFile []byte
	var err error

	// Check if a config file is provided as a command-line argument
	if len(os.Args) > 1 {
		configPath = os.Args[1]
		configFile, err = os.ReadFile(configPath)
		if err == nil {
			fmt.Printf("Using config file: %s\n", configPath)
		}
	}

	// If no config file was provided or couldn't be read, try the default "config.yaml"
	if configFile == nil {
		configPath = "config.yaml"
		configFile, err = os.ReadFile(configPath)
		if err == nil {
			fmt.Printf("Using default config file: %s\n", configPath)
		}
	}

	// If still no config file, suggest proper usage and exit
	if configFile == nil {
		fmt.Println("Error: No config file found.")
		fmt.Println("Usage: ./program [path_to_config_file]")
		fmt.Println("If no config file is specified, the program will try to use 'config.yaml' in the current directory.")
		os.Exit(1)
	}

	// Parse the configuration
	var config ServerConfig
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal("Error parsing config file:", err)
	}

	return config
}

func main() {
	// Load the configuration
	config := loadConfig()

	// Create a new chatbot instance
	bot := NewChatbotFromConfig(config)

	llm := NewLlmConnector(config.Channels)

	// Register callbacks.
	bot.RegisterWebhookEventCallback(
		ScheduleCallbackNormalPriority(llm.LlmCallback),
	)

	bot.RegisterWebhookEventCallback(
		ScheduleCallbackHighestPriority(SimpleWebhookEventCallback),
	)
	// Start the chatbot
	_ = bot.Start()
}
