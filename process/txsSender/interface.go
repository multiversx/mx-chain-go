package txsSender

import (
	"io"
)

// NetworkMessenger defines the basic functionality of a network messenger
// to broadcast buffer data on a channel, for a given topic
type NetworkMessenger interface {
	io.Closer

	BroadcastOnChannel(channel string, topic string, buff []byte)
	IsInterfaceNil() bool
}
