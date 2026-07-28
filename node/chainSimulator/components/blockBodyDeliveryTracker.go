package components

import (
	"crypto/sha256"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-go/consensus"
)

// blockBodyDeliveryTracker provides a simulator-only barrier between the separately broadcast v2
// proposal body and header. The production worker executes the body callback asynchronously; the
// simulator must not advance delivery to the header until the consensus-group receivers have run
// that callback.
type blockBodyDeliveryTracker interface {
	prepareBlockBodyDelivery(message *consensus.Message)
	expectBlockBodyDelivery(message *consensus.Message)
	completeBlockBodyDelivery(message *consensus.Message)
	waitBlockBodyDelivery(message *consensus.Message, timeout time.Duration) bool
}

type blockBodyDeliveryKey struct {
	roundIndex int64
	publicKey  string
	bodyHash   [sha256.Size]byte
}

type blockBodyDeliveryState struct {
	expected  uint32
	completed uint32
	sealed    bool
	done      chan struct{}
}

type blockBodyDeliveryRegistry struct {
	mutOperation sync.Mutex
	deliveries   map[blockBodyDeliveryKey]*blockBodyDeliveryState
}

func newBlockBodyDeliveryRegistry() *blockBodyDeliveryRegistry {
	return &blockBodyDeliveryRegistry{
		deliveries: make(map[blockBodyDeliveryKey]*blockBodyDeliveryState),
	}
}

func (registry *blockBodyDeliveryRegistry) prepareBlockBodyDelivery(message *consensus.Message) {
	if message == nil {
		return
	}

	key := makeBlockBodyDeliveryKey(message)
	registry.mutOperation.Lock()
	registry.deliveries[key] = &blockBodyDeliveryState{
		done: make(chan struct{}, 1),
	}
	registry.mutOperation.Unlock()
}

func (registry *blockBodyDeliveryRegistry) expectBlockBodyDelivery(message *consensus.Message) {
	if message == nil {
		return
	}

	registry.mutOperation.Lock()
	state := registry.deliveries[makeBlockBodyDeliveryKey(message)]
	if state != nil {
		state.expected++
	}
	registry.mutOperation.Unlock()
}

func (registry *blockBodyDeliveryRegistry) completeBlockBodyDelivery(message *consensus.Message) {
	if message == nil {
		return
	}

	registry.mutOperation.Lock()
	state := registry.deliveries[makeBlockBodyDeliveryKey(message)]
	if state != nil {
		state.completed++
		registry.signalIfComplete(state)
	}
	registry.mutOperation.Unlock()
}

func (registry *blockBodyDeliveryRegistry) waitBlockBodyDelivery(message *consensus.Message, timeout time.Duration) bool {
	if message == nil {
		return true
	}

	key := makeBlockBodyDeliveryKey(message)
	registry.mutOperation.Lock()
	state := registry.deliveries[key]
	if state == nil {
		registry.mutOperation.Unlock()
		return true
	}

	state.sealed = true
	registry.signalIfComplete(state)
	done := state.done
	registry.mutOperation.Unlock()

	completed := false
	select {
	case <-done:
		completed = true
	case <-time.After(timeout):
		registry.mutOperation.Lock()
		currentState := registry.deliveries[key]
		if currentState != nil {
			log.Debug(
				"block body delivery tracker timed out",
				"round", key.roundIndex,
				"expected", currentState.expected,
				"completed", currentState.completed,
			)
		}
		registry.mutOperation.Unlock()
	}

	registry.mutOperation.Lock()
	delete(registry.deliveries, key)
	registry.mutOperation.Unlock()

	return completed
}

func (registry *blockBodyDeliveryRegistry) signalIfComplete(state *blockBodyDeliveryState) {
	if !state.sealed || state.completed < state.expected {
		return
	}

	select {
	case state.done <- struct{}{}:
	default:
	}
}

func makeBlockBodyDeliveryKey(message *consensus.Message) blockBodyDeliveryKey {
	return blockBodyDeliveryKey{
		roundIndex: message.RoundIndex,
		publicKey:  string(message.PubKey),
		bodyHash:   sha256.Sum256(message.Body),
	}
}
