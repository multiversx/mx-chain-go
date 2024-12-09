package nodesCoordinator

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type sovereignHashValidatorShuffler struct {
	*randHashShuffler
}

func newSovereignHashValidatorShuffler(baseShuffler NodesShuffler) (*sovereignHashValidatorShuffler, error) {
	if check.IfNil(baseShuffler) {
		return nil, ErrNilShuffler
	}

	randShuffler, castOK := baseShuffler.(*randHashShuffler)
	if !castOK {
		return nil, ErrWrongTypeAssertion
	}

	return &sovereignHashValidatorShuffler{
		randShuffler,
	}, nil
}

// UpdateNodeLists will update node lists w.r.t to metachain original shuffling, only adjusted to work for one shard
func (ss *sovereignHashValidatorShuffler) UpdateNodeLists(args ArgsUpdateNodes) (*ResUpdateNodes, error) {
	ss.updateShufflerConfig(args.Epoch)
	eligibleAfterReshard := copyValidatorMap(args.Eligible)
	waitingAfterReshard := copyValidatorMap(args.Waiting)

	args.AdditionalLeaving = removeDupplicates(args.UnStakeLeaving, args.AdditionalLeaving)

	ss.mutShufflerParams.RLock()
	nodesPerShard := ss.nodesShard
	nodesMeta := uint32(0)
	ss.mutShufflerParams.RUnlock()

	args.NbShards = 1

	return shuffleSovereignNodes(shuffleNodesArg{
		eligible:                           eligibleAfterReshard,
		waiting:                            waitingAfterReshard,
		unstakeLeaving:                     args.UnStakeLeaving,
		additionalLeaving:                  args.AdditionalLeaving,
		newNodes:                           args.NewNodes,
		auction:                            args.Auction,
		randomness:                         args.Rand,
		nodesMeta:                          nodesMeta,
		nodesPerShard:                      nodesPerShard,
		nbShards:                           args.NbShards,
		distributor:                        ss.validatorDistributor,
		maxNodesToSwapPerShard:             ss.activeNodesConfig.NodesToShufflePerShard,
		flagBalanceWaitingLists:            true,
		flagStakingV4Step2:                 true,
		flagStakingV4Step3:                 true,
		maxNumNodes:                        ss.activeNodesConfig.MaxNumNodes,
		flagCleanupAuctionOnLowWaitingList: true,
	})
}

func shuffleSovereignNodes(arg shuffleNodesArg) (*ResUpdateNodes, error) {
	waitingCopy := copyValidatorMap(arg.waiting)
	if waitingCopy[core.SovereignChainShardId] == nil {
		waitingCopy[core.SovereignChainShardId] = make([]Validator, 0)
	}

	numToRemove, err := computeSovereignNumToRemove(arg)
	if err != nil {
		return nil, err
	}

	return baseShuffleNodes(arg, waitingCopy, numToRemove)
}

func computeSovereignNumToRemove(arg shuffleNodesArg) (map[uint32]int, error) {
	numToRemove := make(map[uint32]int)
	if arg.nbShards == 0 {
		return numToRemove, nil
	}

	shardId := core.SovereignChainShardId
	maxToRemove, err := computeNumToRemovePerShard(
		len(arg.eligible[shardId]),
		len(arg.waiting[shardId]),
		int(arg.nodesPerShard))
	if err != nil {
		return nil, fmt.Errorf("%w shard=%v", err, shardId)
	}

	numToRemove[shardId] = maxToRemove
	return numToRemove, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ss *sovereignHashValidatorShuffler) IsInterfaceNil() bool {
	return ss == nil
}
