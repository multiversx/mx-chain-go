package processor

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
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
	// eligibleNodesMap holds a map with key shard id and value
	//		a map with key epoch and value
	//		a map with key eligible node pk and value struct{}
	// this is needed because:
	// 	- a shard node will intercept proof for self shard and meta
	// 	- a meta node will intercept proofs from all shards
	eligibleNodesMap    map[uint32]map[uint32]map[string]struct{}
	mutEligibleNodesMap sync.RWMutex
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
		eligibleNodesMap:     map[uint32]map[uint32]map[string]struct{}{},
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
	epochForEligibleList := getEpochForEligibleList(proof)

	epip.mutEligibleNodesMap.Lock()
	defer epip.mutEligibleNodesMap.Unlock()

	var err error
	eligibleNodesForShard, hasShardCached := epip.eligibleNodesMap[proof.GetHeaderShardId()]
	if !hasShardCached {
		// if no eligible list is cached for the current shard, get it from nodesCoordinator
		eligibleNodesForShardInEpoch, err := epip.getEligibleNodesForShardInEpoch(epochForEligibleList, proof.GetHeaderShardId())
		if err != nil {
			return err
		}

		eligibleNodesForShard = map[uint32]map[string]struct{}{
			epochForEligibleList: eligibleNodesForShardInEpoch,
		}
		epip.eligibleNodesMap[proof.GetHeaderShardId()] = eligibleNodesForShard
	}

	eligibleNodesForShardInEpoch, hasEpochCachedForShard := eligibleNodesForShard[epochForEligibleList]
	if !hasEpochCachedForShard {
		// get the eligible list for the new epoch from nodesCoordinator
		eligibleNodesForShardInEpoch, err = epip.getEligibleNodesForShardInEpoch(epochForEligibleList, proof.GetHeaderShardId())
		if err != nil {
			return err
		}

		// if no eligible list was cached for this epoch, it means that it is a new one
		// it is safe to clean the cached eligible list for other epochs for this shard
		epip.eligibleNodesMap[proof.GetHeaderShardId()] = map[uint32]map[string]struct{}{
			epochForEligibleList: eligibleNodesForShardInEpoch,
		}
	}

	// check the eligible list from the proof's epoch and shard
	peerInfo := epip.peerShardMapper.GetPeerInfo(pid)
	_, isNodeEligible := eligibleNodesForShardInEpoch[string(peerInfo.PkBytes)]
	if !isNodeEligible {
		return fmt.Errorf("%w, proof sender must be an eligible node", process.ErrInvalidHeaderProof)
	}

	return nil
}

func (epip *equivalentProofsInterceptorProcessor) getEligibleNodesForShardInEpoch(epoch uint32, shardID uint32) (map[string]struct{}, error) {
	eligibleList, err := epip.nodesCoordinator.GetAllEligibleValidatorsPublicKeysForShard(epoch, shardID)
	if err != nil {
		return nil, err
	}

	eligibleNodesForShardInEpoch := make(map[string]struct{}, len(eligibleList))
	for _, eligibleNode := range eligibleList {
		eligibleNodesForShardInEpoch[eligibleNode] = struct{}{}
	}

	return eligibleNodesForShardInEpoch, nil
}

func getEpochForEligibleList(proof data.HeaderProofHandler) uint32 {
	epochForEligibleList := proof.GetHeaderEpoch()
	if proof.GetIsStartOfEpoch() && epochForEligibleList > 0 {
		return epochForEligibleList - 1
	}

	return epochForEligibleList
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
