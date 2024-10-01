package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

// # LLM User Query Struct
//
// This struct is used to represent a user query that is sent to the LLM server.
type LlmUserQuery struct {
	ChannelId int    `json:"CHANNEL_ID"` // The channel ID.
	UserId    string `json:"USER_ID"`    // The user ID.
	Query     string `json:"USER_QUERY"` // The user query.
}

// # LLM Model Response Struct
//
// This struct is used to represent the response from the LLM server.
type LlmModelResponse struct {
	Reference string `json:"reference_text"` // The reference part of the model response.
	Response  string `json:"response_text"`  // The main model response.
}

// # LLM Connector
//
// This is the main LLM connector struct.
type LlmConnector struct {
	ChannelMap     ChannelIdConfigMap // A map from channel ID to channel configuration.
	LocalDebugMode bool               // Indicates if the LLM connector is in local debug mode.
	waitingCounter int32              // Indicates the current waiting requests.
}

// # New LLM Connector
//
// This function creates a new LLM connector instance.
func NewLlmConnector(channelMap ChannelIdConfigMap, LocalDebugMode bool) *LlmConnector {
	return &LlmConnector{
		ChannelMap:     channelMap,     // Set the channel map.
		LocalDebugMode: LocalDebugMode, // Set the local debug mode.
	}
}

func (c *LlmConnector) LlmCallback(bot *TaipeionBot, event ChatbotWebhookEvent) error {

	// Check if incoming event is text message.
	// We do not support non-text messages for now, we may implement it later.
	if event.Message.Type != "text" {
		log.Println("[PrivateMessageCallback] Received non-text message. Ignoring.")
		return nil
	}

	// Check if channel ID is in the channel map.
	if _, ok := c.ChannelMap[event.Destination]; !ok {
		log.Printf("[LlmCallback] Channel ID (%d) not found in config. Ignoring.\n", event.Destination)
		return nil
	}

	// Information gathering.
	chan_id := event.Destination                               // Channel ID
	userId := event.Source.UserId                              // User ID
	userQuery := event.Message.Text                            // User query
	trigger_word := c.ChannelMap[chan_id].ChannelTriggerPrefix // Trigger word

	log.Printf("[LlmCallback] Received user (%s) query on channel (%d): %s\n", userId, chan_id, userQuery)

	// Check if the user query starts with the trigger word.
	if !strings.HasPrefix(userQuery, trigger_word) {
		log.Printf("[LlmCallback] User query does not start with trigger word (%s). Ignoring.\n", trigger_word)
		return nil
	}

	// Send a friendly message.
	err := bot.SendPrivateMessage(userId, fmt.Sprintf("正在處理您的問題，視當前情況大約需要30秒~數分鐘不等\n感謝您的耐心等待!\n(目前排隊: %d)", c.waitingCounter), chan_id)

	if err != nil {
		return err
	}

	atomic.AddInt32(&c.waitingCounter, 1) // Add waiting counter by 1.

	// Create a new user query.
	userQueryPayload := LlmUserQuery{
		ChannelId: chan_id,
		UserId:    userId,
		Query:     userQuery,
	}

	// Send the user query to the LLM server.
	response, err := c.LlmRequestSender(userQueryPayload)
	if err != nil {
		log.Println("[LlmCallback] Unable to send user query to LLM server:", err)
		return err
	}

	// Create model response.
	concatedResponse := fmt.Sprintf("%s\n\n%s", response.Response, response.Reference)
	log.Printf("[LlmCallback] Model response for user (%s) on channel (%d): %s\n", userId, chan_id, concatedResponse)

	atomic.AddInt32(&c.waitingCounter, -1) // Decrease waiting counter by 1.

	return bot.SendPrivateMessage(userId, concatedResponse, chan_id) // Send final result.
}

// # LLM Request Sender
//
// This function sends a user query to the LLM server and returns the response.
func (c *LlmConnector) LlmRequestSender(prompt LlmUserQuery) (LlmModelResponse, error) {

	// Serialize the user query.
	request_payload, err := json.Marshal(prompt)
	if err != nil {
		log.Println("[LlmConnector] Unable to serialize user query:", err)
		return LlmModelResponse{}, err
	}

	// Create a new HTTP request.
	req, err := http.NewRequest(
		"POST",
		c.ChannelMap[prompt.ChannelId].ChannelLlmEndpoint,
		bytes.NewBuffer(request_payload))

	if err != nil {
		log.Println("[LlmConnector] Unable to create request:", err)
		return LlmModelResponse{}, err
	}

	// Set the request headers.
	req.Header.Set("Content-Type", "application/json")

	// Perform the request.
	client := &http.Client{}    // Create a new HTTP client.
	resp, err := client.Do(req) // Perform the request.
	if err != nil {
		log.Println("[LlmConnector] Unable to perform request:", err)
		return LlmModelResponse{}, err
	}
	defer resp.Body.Close() // Close the response body when done.

	// Create a new model response.
	modelResp := LlmModelResponse{}                     // Create a new model response.
	err = json.NewDecoder(resp.Body).Decode(&modelResp) // Decode the response body.
	if err != nil {
		log.Println("[LlmConnector] Unable to decode response:", err)
		return LlmModelResponse{}, err
	}

	// Return the model response.
	return modelResp, nil
}
