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
		go func(s event_subscriber.SubscriberInterface, e interface{}, n string) {
			eventMap := s.GetSubscribedEvents()
			callback, ok := eventMap[n]
			if ok {
				callback(e)
			}
		}(subscriber, event, eventName)
	}
}
