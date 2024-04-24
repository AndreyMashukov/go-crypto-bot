package event_subscriber

type SubscriberInterface interface {
	GetSubscribedEvents() map[string]func(interface{})
}
