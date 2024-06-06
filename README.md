# TaipeiON Messenger Chatbot

## Design A Callback Function
The callback function defines the action which will be triggered when the chatbot receives a webhook event. 
 
The signature of the callback function should be `func(bot *TaipeionBot, event ChatbotWebhookEvent) error`. 

The `TaipeionBot` struct contains the chatbot's configuration and the `ChatbotWebhookEvent` struct contains the incoming event data.

While designing the callback function, you should filter the incoming event type first. For example, if you only want to handle text messages, you can check the `event.Message.Type` field.

```go
func SimpleWebhookEventCallback(bot *TaipeionBot, event ChatbotWebhookEvent) error {
	// Check if incoming event is text message.
	if event.Message.Type != "text" {
		log.Println("[SimpleStdCallback] Received non-text message. Ignoring.")
		return nil
	}

	log.Printf("[SimpleStdCallback] Received event: %#v\n", event)
	return nil
}
```


