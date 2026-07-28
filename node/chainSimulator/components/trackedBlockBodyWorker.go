package components

import (
	"context"
	"sync"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/factory"
)

// trackedBlockBodyWorker keeps the production worker behavior and only reports when a physical
// consensus-group receiver has finished its proposal-body callback. Delivery itself remains
// asynchronous and is still performed by spos.Worker.
type trackedBlockBodyWorker struct {
	factory.ConsensusWorker

	marshalizer    marshal.Marshalizer
	consensusState *spos.ConsensusState
	bodyTracker    blockBodyDeliveryTracker
	headerTracker  blockHeaderDeliveryTracker
	shardID        uint32
	mutTracked     sync.Mutex
	trackedBodies  map[blockBodyDeliveryKey]struct{}
	trackedHeaders map[blockHeaderDeliveryKey]struct{}
}

func newTrackedBlockBodyWorker(
	worker factory.ConsensusWorker,
	marshalizer marshal.Marshalizer,
	consensusState *spos.ConsensusState,
	bodyTracker blockBodyDeliveryTracker,
	headerTracker blockHeaderDeliveryTracker,
	shardID uint32,
) factory.ConsensusWorker {
	if bodyTracker == nil && headerTracker == nil {
		return worker
	}

	trackedWorker := &trackedBlockBodyWorker{
		ConsensusWorker: worker,
		marshalizer:     marshalizer,
		consensusState:  consensusState,
		bodyTracker:     bodyTracker,
		headerTracker:   headerTracker,
		shardID:         shardID,
		trackedBodies:   make(map[blockBodyDeliveryKey]struct{}),
		trackedHeaders:  make(map[blockHeaderDeliveryKey]struct{}),
	}
	if headerTracker != nil {
		headerTracker.registerBlockHeaderDeliveryReceiver(trackedWorker)
	}

	return trackedWorker
}

func (worker *trackedBlockBodyWorker) AddReceivedMessageCall(
	messageType consensus.MessageType,
	receivedMessageCall func(ctx context.Context, cnsDta *consensus.Message) bool,
) {
	if messageType != bls.MtBlockBody || receivedMessageCall == nil {
		worker.ConsensusWorker.AddReceivedMessageCall(messageType, receivedMessageCall)
		return
	}

	worker.ConsensusWorker.AddReceivedMessageCall(messageType, func(ctx context.Context, message *consensus.Message) bool {
		processed := receivedMessageCall(ctx, message)
		if worker.takeTrackedBody(message) {
			worker.bodyTracker.completeBlockBodyDelivery(message)
		}

		return processed
	})
}

func (worker *trackedBlockBodyWorker) ProcessReceivedMessage(
	message p2p.MessageP2P,
	fromConnectedPeer core.PeerID,
	source p2p.MessageHandler,
) ([]byte, error) {
	consensusMessage := worker.decodeConsensusMessage(message)
	shouldTrack := worker.bodyTracker != nil && worker.shouldTrack(consensusMessage)
	if shouldTrack {
		worker.trackBody(consensusMessage)
		worker.bodyTracker.expectBlockBodyDelivery(consensusMessage)
	}
	result, err := worker.ConsensusWorker.ProcessReceivedMessage(message, fromConnectedPeer, source)
	if err != nil && shouldTrack && worker.takeTrackedBody(consensusMessage) {
		// A rejected body will never reach the callback. Fulfil the registered expectation so the
		// simulator does not turn that validation error into an additional wall-clock delay.
		worker.bodyTracker.completeBlockBodyDelivery(consensusMessage)
	}
	return result, err
}

// ReceivedHeader reports completion after the production worker has handled the proposal. Together
// with syncHeadersPool this turns the asynchronous interceptor handoff into a deterministic
// simulator boundary without changing production consensus code.
func (worker *trackedBlockBodyWorker) ReceivedHeader(header data.HeaderHandler, headerHash []byte) {
	worker.ConsensusWorker.ReceivedHeader(header, headerHash)
	if worker.takeTrackedHeader(header) {
		worker.headerTracker.completeBlockHeaderDelivery(header)
	}
}

func (worker *trackedBlockBodyWorker) expectBlockHeaderDelivery(
	header data.HeaderHandler,
	leaderPublicKey []byte,
) (bool, uint32) {
	if check.IfNil(header) || header.GetShardID() != worker.shardID {
		return false, 0
	}
	if !worker.consensusState.IsSubroundFinished(bls.SrStartRound) {
		return false, 0
	}
	// In v1 the proposal carries body and header in the consensus message, so consensusState
	// already holds the header by the time EndRound broadcasts the finalized header. The headers
	// pool deduplicates that second delivery and will not invoke ReceivedHeader again; registering
	// an expectation here would therefore force the simulator to wait for a callback that cannot
	// arrive. In v2 the separately broadcast proposal header is not set yet, so it is still tracked.
	if worker.consensusState.IsHeaderAlreadyReceived() {
		return false, 0
	}
	if worker.consensusState.SelfPubKey() == string(leaderPublicKey) ||
		worker.consensusState.IsKeyManagedBySelf(leaderPublicKey) {
		return false, 0
	}
	if !worker.consensusState.IsNodeInConsensusGroup(worker.consensusState.SelfPubKey()) &&
		!worker.consensusState.IsMultiKeyInConsensusGroup() {
		return false, 0
	}

	worker.mutTracked.Lock()
	worker.trackedHeaders[makeBlockHeaderDeliveryKey(header)] = struct{}{}
	worker.mutTracked.Unlock()

	// The leader processes its own proposal before broadcasting it. The delivery barrier therefore
	// needs only PBFT-threshold-minus-one follower completions, not every physical follower. Waiting
	// for all followers serializes consensus on the slowest validator and is stricter than SPoS.
	threshold := worker.consensusState.Threshold(bls.SrSignature)
	if threshold <= 1 {
		return true, 0
	}

	return true, uint32(threshold - 1)
}

func (worker *trackedBlockBodyWorker) decodeConsensusMessage(message p2p.MessageP2P) *consensus.Message {
	if message == nil || message.Data() == nil {
		return nil
	}

	consensusMessage := &consensus.Message{}
	err := worker.marshalizer.Unmarshal(consensusMessage, message.Data())
	if err != nil {
		return nil
	}

	return consensusMessage
}

func (worker *trackedBlockBodyWorker) shouldTrack(message *consensus.Message) bool {
	if message == nil || consensus.MessageType(message.MsgType) != bls.MtBlockBody {
		return false
	}
	if !worker.consensusState.IsSubroundFinished(bls.SrStartRound) {
		return false
	}
	if worker.consensusState.SelfPubKey() == string(message.PubKey) ||
		worker.consensusState.IsKeyManagedBySelf(message.PubKey) {
		return false
	}

	return worker.consensusState.IsNodeInConsensusGroup(worker.consensusState.SelfPubKey()) ||
		worker.consensusState.IsMultiKeyInConsensusGroup()
}

func (worker *trackedBlockBodyWorker) trackBody(message *consensus.Message) {
	worker.mutTracked.Lock()
	worker.trackedBodies[makeBlockBodyDeliveryKey(message)] = struct{}{}
	worker.mutTracked.Unlock()
}

func (worker *trackedBlockBodyWorker) takeTrackedBody(message *consensus.Message) bool {
	if message == nil {
		return false
	}

	key := makeBlockBodyDeliveryKey(message)
	worker.mutTracked.Lock()
	_, tracked := worker.trackedBodies[key]
	delete(worker.trackedBodies, key)
	worker.mutTracked.Unlock()

	return tracked
}

func (worker *trackedBlockBodyWorker) takeTrackedHeader(header data.HeaderHandler) bool {
	if check.IfNil(header) {
		return false
	}

	key := makeBlockHeaderDeliveryKey(header)
	worker.mutTracked.Lock()
	_, tracked := worker.trackedHeaders[key]
	delete(worker.trackedHeaders, key)
	worker.mutTracked.Unlock()

	return tracked
}
