package processor

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

type shardProvider interface {
	ShardID() uint32
}

// ArgShardValidatorInfoInterceptorProcessor is the argument for the interceptor processor used for shard validator info
type ArgShardValidatorInfoInterceptorProcessor struct {
	Marshaller      marshal.Marshalizer
	PeerShardMapper process.PeerShardMapper
}

type shardValidatorInfoInterceptorProcessor struct {
	marshaller      marshal.Marshalizer
	peerShardMapper process.PeerShardMapper
}

// NewShardValidatorInfoInterceptorProcessor creates an instance of shardValidatorInfoInterceptorProcessor
func NewShardValidatorInfoInterceptorProcessor(args ArgShardValidatorInfoInterceptorProcessor) (*shardValidatorInfoInterceptorProcessor, error) {
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.PeerShardMapper) {
		return nil, process.ErrNilPeerShardMapper
	}

	return &shardValidatorInfoInterceptorProcessor{
		marshaller:      args.Marshaller,
		peerShardMapper: args.PeerShardMapper,
	}, nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (processor *shardValidatorInfoInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted shard validator info into peer shard mapper
func (processor *shardValidatorInfoInterceptorProcessor) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, _ string) error {
	shardValidatorInfo, ok := data.(shardProvider)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	processor.peerShardMapper.PutPeerIdShardId(fromConnectedPeer, shardValidatorInfo.ShardID())

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming shard validator info
func (processor *shardValidatorInfoInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("shardValidatorInfoInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (processor *shardValidatorInfoInterceptorProcessor) IsInterfaceNil() bool {
	return processor == nil
}
