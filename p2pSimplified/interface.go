package components

import pubsub "github.com/libp2p/go-libp2p-pubsub"

// MessageProcessor defines what a p2p message processor can do
type MessageProcessor interface {
	ProcessMessage(msg *pubsub.Message) error
}
