package integrationTests

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// CountInterceptor represents an interceptors that counts received messages on topics
type CountInterceptor struct {
	mutMessagesCount sync.RWMutex
	messagesCount    map[string]int
}

// NewCountInterceptor creates a new CountInterceptor instance
func NewCountInterceptor() *CountInterceptor {
	return &CountInterceptor{
		messagesCount: make(map[string]int),
	}
}

// ProcessReceivedMessage is called each time a new message is received
func (ci *CountInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	ci.mutMessagesCount.Lock()
	ci.messagesCount[message.Topic()]++
	ci.mutMessagesCount.Unlock()

	return nil
}

// MessageCount returns the number of messages received on the provided topic
func (ci *CountInterceptor) MessageCount(topic string) int {
	ci.mutMessagesCount.RLock()
	defer ci.mutMessagesCount.RUnlock()

	return ci.messagesCount[topic]
}

// Reset resets the counters values
func (ci *CountInterceptor) Reset() {
	ci.mutMessagesCount.Lock()
	ci.messagesCount = make(map[string]int)
	ci.mutMessagesCount.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ci *CountInterceptor) IsInterfaceNil() bool {
	return ci == nil
}
