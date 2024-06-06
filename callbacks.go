package main

import (
	"fmt"
	"log"
)

/*
# Guide: How to design a chatbot callback

- Filter the incoming event to make sure it is what you want to process:

```go
// Check if incoming event is text message.
if event.Message.Type != "text" {
	log.Println("Received non-text message. Ignoring.")
	return nil
}
```

- Interacting with the channel/user:

Private message:
```go
err := bot.SendPrivateMessage(receiver, reply_message)
```

Broadcast message:
```go
err := bot.SendBroadcastMessage(reply_message)
```

- Register the callback (in the main function):

```go
bot.RegisterWebhookEventCallback(YourCallbackFunction)
```

Followings are some examples of chatbot callbacks.
*/

// # Simple webhook event callback.
//
// This callback puts the incoming message to the log (stdout).
func SimpleWebhookEventCallback(bot *TaipeionBot, event ChatbotWebhookEvent) error {
	// Check if incoming event is text message.
	if event.Message.Type != "text" {
		log.Println("[SimpleStdCallback] Received non-text message. Ignoring.")
		return nil
	}

	log.Printf("[SimpleStdCallback] Received event: %#v\n", event)
	return nil
}

// # Private message callback.
//
// This callback is used to send a reply to a cer
func PrivateMessageCallback(bot *TaipeionBot, event ChatbotWebhookEvent) error {
	// Check if incoming event is text message.
	if event.Message.Type != "text" {
		log.Println("[PrivateMessageCallback] Received non-text message. Ignoring.")
		return nil
	}

	chan_id := event.Destination
	reply_message := event.Message.Text
	receiver := event.Source.UserId

	// Process the incoming message.
	reply_message = fmt.Sprintf("ChannelId: %d\nEcho: %s", chan_id, reply_message)

	return bot.SendPrivateMessage(receiver, reply_message, chan_id)

}
