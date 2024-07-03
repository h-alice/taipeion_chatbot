package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// # LLM User Query Struct
//
// This struct is used to represent a user query that is sent to the LLM server.
type LlmUserQuery struct {
	ChannelId int    `json:"CHANNEL_ID"`
	UserId    string `json:"USER_ID"`
	Query     string `json:"USER_QUERY"`
}

// # LLM Model Response Struct
//
// This struct is used to represent the response from the LLM server.
type LlmModelResponse struct {
	Reference string `json:"reference_text"`
	Response  string `json:"response_text"`
}

// # LLM Connector
//
// This is the main LLM connector struct.
type LlmConnector struct {
	LlmEndpoint string // The URL of the LLM server.
}

// # New LLM Connector
//
// This function creates a new LLM connector instance.
func NewLlmConnector(llmEndpoint string) *LlmConnector {
	return &LlmConnector{
		LlmEndpoint: llmEndpoint,
	}
}

func (c *LlmConnector) LlmCallback(bot *TaipeionBot, event ChatbotWebhookEvent) error {

	// Check if incoming event is text message.
	if event.Message.Type != "text" {
		log.Println("[PrivateMessageCallback] Received non-text message. Ignoring.")
		return nil
	}

	// Information gathering.
	chan_id := event.Destination    // Channel ID
	userId := event.Source.UserId   // User ID
	userQuery := event.Message.Text // User query

	log.Printf("[LlmCallback] Received user (%s) query on channel (%d): %s\n", userId, chan_id, userQuery)

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

	concatedResponse := fmt.Sprintf("%s\n\n%s", response.Response, response.Reference)
	log.Printf("[LlmCallback] Model response for user (%s) on channel (%d): %s\n", userId, chan_id, concatedResponse)

	return bot.SendPrivateMessage(userId, concatedResponse, chan_id)
}

func (c *LlmConnector) LlmRequestSender(prompt LlmUserQuery) (LlmModelResponse, error) {

	request_payload, err := json.Marshal(prompt)
	if err != nil {
		log.Println("[LlmConnector] Unable to serialize user query:", err)
		return LlmModelResponse{}, err
	}

	req, err := http.NewRequest("POST", c.LlmEndpoint, bytes.NewBuffer(request_payload))
	if err != nil {
		log.Println("[LlmConnector] Unable to create request:", err)
		return LlmModelResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	log.Println(req.Header)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[LlmConnector] Unable to perform request:", err)
		return LlmModelResponse{}, err
	}
	defer resp.Body.Close()

	modelResp := LlmModelResponse{}

	err = json.NewDecoder(resp.Body).Decode(&modelResp)
	if err != nil {
		log.Println("[LlmConnector] Unable to decode response:", err)
		return LlmModelResponse{}, err
	}

	return modelResp, nil
}
