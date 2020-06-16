package mock

import (
	"reflect"

	"github.com/libp2p/go-libp2p-core/event"
)

// EventBusStub -
type EventBusStub struct {
	SubscribeCalled        func(eventType interface{}, opts ...event.SubscriptionOpt) (event.Subscription, error)
	EmitterCalled          func(eventType interface{}, opts ...event.EmitterOpt) (event.Emitter, error)
	GetAllEventTypesCalled func() []reflect.Type
}

// Subscribe -
func (ebs *EventBusStub) Subscribe(eventType interface{}, opts ...event.SubscriptionOpt) (event.Subscription, error) {
	if ebs.SubscribeCalled != nil {
		return ebs.SubscribeCalled(eventType, opts...)
	}

	return &EventSubscriptionStub{}, nil
}

// Emitter -
func (ebs *EventBusStub) Emitter(eventType interface{}, opts ...event.EmitterOpt) (event.Emitter, error) {
	if ebs.EmitterCalled != nil {
		return ebs.EmitterCalled(eventType, opts...)
	}

	return nil, nil
}

// GetAllEventTypes -
func (ebs *EventBusStub) GetAllEventTypes() []reflect.Type {
	if ebs.GetAllEventTypesCalled != nil {
		return ebs.GetAllEventTypesCalled()
	}

	return make([]reflect.Type, 0)
}
