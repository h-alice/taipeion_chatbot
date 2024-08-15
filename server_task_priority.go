package main

func ScheduleNormalPriority(callback WebhookEventCallback) eventHandlerEntry {
	return eventHandlerEntry{
		callback:   callback,
		isPriority: false,
	}
}

func ScheduleHighestPriority(callback WebhookEventCallback) eventHandlerEntry {
	return eventHandlerEntry{
		callback:   callback,
		isPriority: true,
	}
}
