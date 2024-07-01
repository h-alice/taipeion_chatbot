package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type LlmUserQuery struct {
	ChannelId int    `json:"channel-id"`
	UserId    string `json:"user-id"`
	Query     string `json:"query"`
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
