package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	tp "taipeion/taipeion"

	"gopkg.in/yaml.v3"
)

// Define a struct for the response
type Response struct {
	Test string `json:"test"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Print Header
	log.Println("Header:", r.Header)

	// Print the request body to the console
	payload, err := tp.DeserializeMessage(body)
	if err != nil {
		log.Println("Error deserializing message:", err)
		// Fallback to printing the raw body
		log.Println("Received payload:", string(body))
	} else {
		log.Println("Received payload:", payload)
	}

	// Create the response object
	response := Response{Test: "Helloworld"}

	// Set the response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	// Encode the response object to JSON and send it
	json.NewEncoder(w).Encode(response)
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

	// Set the handler for the root path
	http.HandleFunc("/", handler)

	// Start the HTTP server.
	log.Println("Starting server on", config.Address+":"+config.Port)
	err = http.ListenAndServe(config.Address+":"+config.Port, nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}

}
