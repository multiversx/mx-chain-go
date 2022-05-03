package staking

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/stakingcommon"
)

type OwnerStats struct {
	EligibleBlsKeys  map[uint32][][]byte
	WaitingBlsKeys   map[uint32][][]byte
	StakingQueueKeys [][]byte
	TotalStake       *big.Int
}

type InitialNodesConfig struct {
	NumOfShards          uint32
	Owners               map[string]*OwnerStats
	MaxNodesChangeConfig []config.MaxNodesChangeConfig
}

func NewTestMetaProcessorWithCustomNodes(config *InitialNodesConfig) *TestMetaProcessor {
	coreComponents, dataComponents, bootstrapComponents, statusComponents, stateComponents := createComponentHolders(config.NumOfShards)

	_ = dataComponents
	_ = bootstrapComponents
	_ = statusComponents

	queue := createStakingQueueCustomNodes(
		config.Owners,
		coreComponents.InternalMarshalizer(),
		stateComponents.AccountsAdapter(),
	)

	return &TestMetaProcessor{
		NodesConfig: nodesConfig{
			queue: queue,
		},
		AccountsAdapter: stateComponents.AccountsAdapter(),
		Marshaller:      coreComponents.InternalMarshalizer(),
	}
}

func createStakingQueueCustomNodes(
	owners map[string]*OwnerStats,
	marshaller marshal.Marshalizer,
	accountsAdapter state.AccountsAdapter,
) [][]byte {
	queue := make([][]byte, 0)

	for owner, ownerStats := range owners {
		stakingcommon.AddKeysToWaitingList(
			accountsAdapter,
			ownerStats.StakingQueueKeys,
			marshaller,
			[]byte(owner),
			[]byte(owner),
		)

		stakingcommon.AddValidatorData(
			accountsAdapter,
			[]byte(owner),
			ownerStats.StakingQueueKeys,
			ownerStats.TotalStake,
			marshaller,
		)

		queue = append(queue, ownerStats.StakingQueueKeys...)
	}

	return queue
}
