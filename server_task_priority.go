package main

func ScheduleCallbackNormalPriority(callback WebhookEventCallback) eventHandlerEntry {
	return eventHandlerEntry{
		Callback:   callback,
		IsPriority: false,
	}
}

func ScheduleCallbackHighestPriority(callback WebhookEventCallback) eventHandlerEntry {
	return eventHandlerEntry{
		Callback:   callback,
		IsPriority: true,
	}
}
