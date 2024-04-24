package service

import "github.com/AndreyMashukov/go-crypto-bot/server/src/event_subscriber"

type EventDispatcher struct {
	Subscribers []event_subscriber.SubscriberInterface
	Enabled     bool
}

func (d *EventDispatcher) Dispatch(event interface{}, eventName string) {
	if !d.Enabled {
		return
	}

	for _, subscriber := range d.Subscribers {
		eventMap := subscriber.GetSubscribedEvents()
		callback, ok := eventMap[eventName]
		if ok {
			callback(event)
		}
	}
}
