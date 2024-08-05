package main

func ScheduleCallbackNormalPriority(callback WebhookEventCallback) eventHandlerEntry {
	return eventHandlerEntry{
		callback:   callback,
		isPriority: false,
	}
}

func ScheduleCallbackHighestPriority(callback WebhookEventCallback) eventHandlerEntry {
	return eventHandlerEntry{
		callback:   callback,
		isPriority: true,
	}
}
