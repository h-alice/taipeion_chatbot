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

## Function Diagrams

<img width="1273" alt="image" src="https://github.com/user-attachments/assets/93a81e98-ee88-4579-a366-0ecfd9cec97a" />

### Concurrent Handling Process
<img width="602" alt="image" src="https://github.com/user-attachments/assets/7924b180-d442-46cd-bb7d-dff771a05766" />

### LLM Service Handler
<img width="719" alt="image" src="https://github.com/user-attachments/assets/535d37b0-7b95-43b6-8b71-96140f49cc3c" />

