package processor

import (
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
)

type shardProvider interface {
	ShardID() string
}

// ArgDirectConnectionInfoInterceptorProcessor is the argument for the interceptor processor used for direct connection info
type ArgDirectConnectionInfoInterceptorProcessor struct {
	PeerShardMapper process.PeerShardMapper
}

type directConnectionInfoInterceptorProcessor struct {
	peerShardMapper process.PeerShardMapper
}

// NewDirectConnectionInfoInterceptorProcessor creates an instance of directConnectionInfoInterceptorProcessor
func NewDirectConnectionInfoInterceptorProcessor(args ArgDirectConnectionInfoInterceptorProcessor) (*directConnectionInfoInterceptorProcessor, error) {
	if check.IfNil(args.PeerShardMapper) {
		return nil, process.ErrNilPeerShardMapper
	}

	return &directConnectionInfoInterceptorProcessor{
		peerShardMapper: args.PeerShardMapper,
	}, nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (processor *directConnectionInfoInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted validator info into peer shard mapper
func (processor *directConnectionInfoInterceptorProcessor) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, _ string) error {
	shardDirectConnectionInfo, ok := data.(shardProvider)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	shardID, err := strconv.Atoi(shardDirectConnectionInfo.ShardID())
	if err != nil {
		return err
	}

	processor.peerShardMapper.PutPeerIdShardId(fromConnectedPeer, uint32(shardID))

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming shard validator info, currently not implemented
func (processor *directConnectionInfoInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("directConnectionInfoInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (processor *directConnectionInfoInterceptorProcessor) IsInterfaceNil() bool {
	return processor == nil
}
