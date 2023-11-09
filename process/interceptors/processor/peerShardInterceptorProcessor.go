package processor

import (
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type shardProvider interface {
	ShardID() string
}

// ArgPeerShardInterceptorProcessor is the argument for the interceptor processor used for peer shard message
type ArgPeerShardInterceptorProcessor struct {
	PeerShardMapper process.PeerShardMapper
}

type peerShardInterceptorProcessor struct {
	peerShardMapper process.PeerShardMapper
}

// NewPeerShardInterceptorProcessor creates an instance of peerShardInterceptorProcessor
func NewPeerShardInterceptorProcessor(args ArgPeerShardInterceptorProcessor) (*peerShardInterceptorProcessor, error) {
	if check.IfNil(args.PeerShardMapper) {
		return nil, process.ErrNilPeerShardMapper
	}

	return &peerShardInterceptorProcessor{
		peerShardMapper: args.PeerShardMapper,
	}, nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (processor *peerShardInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted validator info into peer shard mapper
func (processor *peerShardInterceptorProcessor) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, _ string) error {
	shardPeerShard, ok := data.(shardProvider)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	shardID, err := strconv.Atoi(shardPeerShard.ShardID())
	if err != nil {
		return err
	}

	processor.peerShardMapper.PutPeerIdShardId(fromConnectedPeer, uint32(shardID))

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming shard validator info, currently not implemented
func (processor *peerShardInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("peerShardInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (processor *peerShardInterceptorProcessor) IsInterfaceNil() bool {
	return processor == nil
}
