package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type LlmUserQuery struct {
	ChannelId int    `json:"CHANNEL_ID"`
	UserId    string `json:"USER_ID"`
	Query     string `json:"USER_QUERY"`
}

type LlmModelResponse struct {
	Reference string `json:"reference_text"`
	Response  string `json:"response_text"`
}

type LlmConnector struct {
	LlmEndpoint string // The URL of the LLM server.
}

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

	chan_id := event.Destination
	userId := event.Source.UserId
	userQuery := event.Message.Text

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

	finalResponse := fmt.Sprintf("%s\n%s", response.Response, response.Reference)

	return bot.SendPrivateMessage(userId, finalResponse, chan_id)
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
