package main

import (
	"bytes"
	"encoding/json"
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

// NOTE: We use stdout printing only to handling error.
func LlmRequestSender(prompt LlmUserQuery, llmEndpoint string) string {

	request_payload, err := json.Marshal(prompt)
	if err != nil {
		log.Println("[LlmConnector] Unable to serialize user query:", err)
	}

	req, err := http.NewRequest("POST", llmEndpoint, bytes.NewBuffer(request_payload))
	if err != nil {
		log.Println("[LlmConnector] Unable to create request:", err)
	}

	req.Header.Set("Content-Type", "application/json")

	log.Println(req.Header)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[LlmConnector] Unable to perform request:", err)
	}
	defer resp.Body.Close()

	return ""
}
