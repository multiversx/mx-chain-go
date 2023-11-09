package processor

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
)

// ArgHeartbeatInterceptorProcessor is the argument for the interceptor processor used for heartbeat
type ArgHeartbeatInterceptorProcessor struct {
	HeartbeatCacher  storage.Cacher
	ShardCoordinator sharding.Coordinator
	PeerShardMapper  process.PeerShardMapper
}

// heartbeatInterceptorProcessor is the processor used when intercepting heartbeat
type heartbeatInterceptorProcessor struct {
	heartbeatCacher  storage.Cacher
	shardCoordinator sharding.Coordinator
	peerShardMapper  process.PeerShardMapper
}

// NewHeartbeatInterceptorProcessor creates a new heartbeatInterceptorProcessor
func NewHeartbeatInterceptorProcessor(args ArgHeartbeatInterceptorProcessor) (*heartbeatInterceptorProcessor, error) {
	err := checkArgsHeartbeat(args)
	if err != nil {
		return nil, err
	}

	return &heartbeatInterceptorProcessor{
		heartbeatCacher:  args.HeartbeatCacher,
		shardCoordinator: args.ShardCoordinator,
		peerShardMapper:  args.PeerShardMapper,
	}, nil
}

func checkArgsHeartbeat(args ArgHeartbeatInterceptorProcessor) error {
	if check.IfNil(args.HeartbeatCacher) {
		return process.ErrNilHeartbeatCacher
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.PeerShardMapper) {
		return process.ErrNilPeerShardMapper
	}

	return nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (hip *heartbeatInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted heartbeat inside the heartbeat cacher
func (hip *heartbeatInterceptorProcessor) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, _ string) error {
	interceptedHeartbeat, ok := data.(interceptedHeartbeatMessageHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	hip.heartbeatCacher.Put(fromConnectedPeer.Bytes(), interceptedHeartbeat.Message(), interceptedHeartbeat.SizeInBytes())

	return hip.updatePeerInfo(interceptedHeartbeat.Message(), fromConnectedPeer)
}

func (hip *heartbeatInterceptorProcessor) updatePeerInfo(message interface{}, fromConnectedPeer core.PeerID) error {
	heartbeatData, ok := message.(*heartbeat.HeartbeatV2)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	hip.peerShardMapper.PutPeerIdShardId(fromConnectedPeer, hip.shardCoordinator.SelfId())
	hip.peerShardMapper.PutPeerIdSubType(fromConnectedPeer, core.P2PPeerSubType(heartbeatData.GetPeerSubType()))

	log.Trace("Heartbeat message saved")

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming hearbeat
func (hip *heartbeatInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("heartbeatInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (hip *heartbeatInterceptorProcessor) IsInterfaceNil() bool {
	return hip == nil
}
