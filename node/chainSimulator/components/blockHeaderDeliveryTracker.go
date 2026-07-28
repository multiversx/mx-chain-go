package components

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

// blockHeaderDeliveryReceiver is implemented by the simulator's consensus worker wrapper. Before
// a proposal header is broadcast, the shared registry asks every physical worker whether that
// worker is an active receiver for the proposal.
type blockHeaderDeliveryReceiver interface {
	expectBlockHeaderDelivery(header data.HeaderHandler, leaderPublicKey []byte) (bool, uint32)
}

// blockHeaderDeliveryTracker provides a simulator-only barrier around the production header
// interceptor. The interceptor starts its processing in a goroutine, so BroadcastHeader can return
// before a follower's consensus worker has observed the proposal.
type blockHeaderDeliveryTracker interface {
	registerBlockHeaderDeliveryReceiver(receiver blockHeaderDeliveryReceiver)
	prepareBlockHeaderDelivery(header data.HeaderHandler, leaderPublicKey []byte)
	completeBlockHeaderDelivery(header data.HeaderHandler)
	waitBlockHeaderDelivery(header data.HeaderHandler, timeout time.Duration) bool
}

type blockHeaderDeliveryKey struct {
	shardID uint32
	round   uint64
	nonce   uint64
}

type blockHeaderDeliveryState struct {
	expected  uint32
	required  uint32
	completed uint32
	done      chan struct{}
}

type blockHeaderDeliveryRegistry struct {
	mutOperation sync.Mutex
	receivers    []blockHeaderDeliveryReceiver
	deliveries   map[blockHeaderDeliveryKey]*blockHeaderDeliveryState
}

func newBlockHeaderDeliveryRegistry() *blockHeaderDeliveryRegistry {
	return &blockHeaderDeliveryRegistry{
		deliveries: make(map[blockHeaderDeliveryKey]*blockHeaderDeliveryState),
	}
}

func (registry *blockHeaderDeliveryRegistry) registerBlockHeaderDeliveryReceiver(
	receiver blockHeaderDeliveryReceiver,
) {
	if receiver == nil {
		return
	}

	registry.mutOperation.Lock()
	registry.receivers = append(registry.receivers, receiver)
	registry.mutOperation.Unlock()
}

func (registry *blockHeaderDeliveryRegistry) prepareBlockHeaderDelivery(
	header data.HeaderHandler,
	leaderPublicKey []byte,
) {
	if check.IfNil(header) {
		return
	}

	key := makeBlockHeaderDeliveryKey(header)
	registry.mutOperation.Lock()
	receivers := append([]blockHeaderDeliveryReceiver{}, registry.receivers...)
	state := &blockHeaderDeliveryState{
		done: make(chan struct{}, 1),
	}
	registry.deliveries[key] = state
	registry.mutOperation.Unlock()

	var expected uint32
	var required uint32
	for _, receiver := range receivers {
		shouldExpect, receiverRequired := receiver.expectBlockHeaderDelivery(header, leaderPublicKey)
		if shouldExpect {
			expected++
			if receiverRequired > required {
				required = receiverRequired
			}
		}
	}
	if required > expected {
		required = expected
	}

	registry.mutOperation.Lock()
	currentState := registry.deliveries[key]
	if currentState == state {
		state.expected = expected
		state.required = required
		if state.completed >= state.required {
			signalBlockHeaderDelivery(state)
		}
	}
	registry.mutOperation.Unlock()
}

func (registry *blockHeaderDeliveryRegistry) completeBlockHeaderDelivery(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	registry.mutOperation.Lock()
	state := registry.deliveries[makeBlockHeaderDeliveryKey(header)]
	if state != nil {
		state.completed++
		if state.completed >= state.required {
			signalBlockHeaderDelivery(state)
		}
	}
	registry.mutOperation.Unlock()
}

func (registry *blockHeaderDeliveryRegistry) waitBlockHeaderDelivery(
	header data.HeaderHandler,
	timeout time.Duration,
) bool {
	if check.IfNil(header) {
		return true
	}

	key := makeBlockHeaderDeliveryKey(header)
	registry.mutOperation.Lock()
	state := registry.deliveries[key]
	if state == nil {
		registry.mutOperation.Unlock()
		return true
	}
	if state.completed >= state.required {
		signalBlockHeaderDelivery(state)
	}
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
				"block header delivery tracker timed out",
				"shard", key.shardID,
				"round", key.round,
				"nonce", key.nonce,
				"expected", currentState.expected,
				"required", currentState.required,
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

func signalBlockHeaderDelivery(state *blockHeaderDeliveryState) {
	select {
	case state.done <- struct{}{}:
	default:
	}
}

func makeBlockHeaderDeliveryKey(header data.HeaderHandler) blockHeaderDeliveryKey {
	return blockHeaderDeliveryKey{
		shardID: header.GetShardID(),
		round:   header.GetRound(),
		nonce:   header.GetNonce(),
	}
}
