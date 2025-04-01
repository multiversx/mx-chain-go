package processor

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

// ArgEquivalentProofsInterceptorProcessor is the argument for the interceptor processor used for equivalent proofs
type ArgEquivalentProofsInterceptorProcessor struct {
	EquivalentProofsPool EquivalentProofsPool
	Marshaller           marshal.Marshalizer
	PeerShardMapper      process.PeerShardMapper
	NodesCoordinator     process.NodesCoordinator
}

// equivalentProofsInterceptorProcessor is the processor used when intercepting equivalent proofs
type equivalentProofsInterceptorProcessor struct {
	equivalentProofsPool EquivalentProofsPool
	marshaller           marshal.Marshalizer
	peerShardMapper      process.PeerShardMapper
	nodesCoordinator     process.NodesCoordinator
}

// NewEquivalentProofsInterceptorProcessor creates a new equivalentProofsInterceptorProcessor
func NewEquivalentProofsInterceptorProcessor(args ArgEquivalentProofsInterceptorProcessor) (*equivalentProofsInterceptorProcessor, error) {
	err := checkArgsEquivalentProofs(args)
	if err != nil {
		return nil, err
	}

	return &equivalentProofsInterceptorProcessor{
		equivalentProofsPool: args.EquivalentProofsPool,
		marshaller:           args.Marshaller,
		peerShardMapper:      args.PeerShardMapper,
		nodesCoordinator:     args.NodesCoordinator,
	}, nil
}

func checkArgsEquivalentProofs(args ArgEquivalentProofsInterceptorProcessor) error {
	if check.IfNil(args.EquivalentProofsPool) {
		return process.ErrNilProofsPool
	}
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.PeerShardMapper) {
		return process.ErrNilPeerShardMapper
	}
	if check.IfNil(args.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}

	return nil
}

// Validate checks if the intercepted data was sent by an eligible node
func (epip *equivalentProofsInterceptorProcessor) Validate(data process.InterceptedData, pid core.PeerID) error {
	interceptedProof, ok := data.(interceptedEquivalentProof)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	proof := interceptedProof.GetProof()

	eligibleList, err := epip.nodesCoordinator.GetAllEligibleValidatorsPublicKeysForShard(proof.GetHeaderEpoch(), proof.GetHeaderShardId())
	if err != nil {
		return err
	}

	peerInfo := epip.peerShardMapper.GetPeerInfo(pid)
	for _, eligibleValidator := range eligibleList {
		if string(peerInfo.PkBytes) == eligibleValidator {
			return nil
		}
	}

	return fmt.Errorf("%w, proof sender must be an eligible node", process.ErrInvalidHeaderProof)
}

// Save will save the intercepted equivalent proof inside the proofs tracker
func (epip *equivalentProofsInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	interceptedProof, ok := data.(interceptedEquivalentProof)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	wasAdded := epip.equivalentProofsPool.AddProof(interceptedProof.GetProof())
	if !wasAdded {
		return common.ErrAlreadyExistingEquivalentProof
	}

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming equivalent proofs
func (epip *equivalentProofsInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("equivalentProofsInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (epip *equivalentProofsInterceptorProcessor) IsInterfaceNil() bool {
	return epip == nil
}
