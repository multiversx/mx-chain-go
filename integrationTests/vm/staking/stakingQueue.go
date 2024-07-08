package staking

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/stakingcommon"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
)

func createStakingQueue(
	numOfNodesInStakingQueue uint32,
	totalNumOfNodes uint32,
	marshaller marshal.Marshalizer,
	accountsAdapter state.AccountsAdapter,
) [][]byte {
	ownerWaitingNodes := make([][]byte, 0)
	if numOfNodesInStakingQueue == 0 {
		return ownerWaitingNodes
	}

	owner := generateAddress(totalNumOfNodes)
	totalNumOfNodes += 1
	for i := totalNumOfNodes; i < totalNumOfNodes+numOfNodesInStakingQueue; i++ {
		ownerWaitingNodes = append(ownerWaitingNodes, generateAddress(i))
	}

	stakingcommon.AddKeysToWaitingList(
		accountsAdapter,
		ownerWaitingNodes,
		marshaller,
		owner,
		owner,
	)

	stakingcommon.RegisterValidatorKeys(
		accountsAdapter,
		owner,
		owner,
		ownerWaitingNodes,
		big.NewInt(int64(2*nodePrice*numOfNodesInStakingQueue)),
		marshaller,
	)

	return ownerWaitingNodes
}

func createStakingQueueCustomNodes(
	owners map[string]*OwnerStats,
	marshaller marshal.Marshalizer,
	accountsAdapter state.AccountsAdapter,
) [][]byte {
	queue := make([][]byte, 0)

	for owner, ownerStats := range owners {
		stakingcommon.RegisterValidatorKeys(
			accountsAdapter,
			[]byte(owner),
			[]byte(owner),
			ownerStats.StakingQueueKeys,
			ownerStats.TotalStake,
			marshaller,
		)

		stakingcommon.AddKeysToWaitingList(
			accountsAdapter,
			ownerStats.StakingQueueKeys,
			marshaller,
			[]byte(owner),
			[]byte(owner),
		)

		queue = append(queue, ownerStats.StakingQueueKeys...)
	}

	return queue
}

func (tmp *TestMetaProcessor) getWaitingListKeys() [][]byte {
	stakingSCAcc := stakingcommon.LoadUserAccount(tmp.AccountsAdapter, vm.StakingSCAddress)

	waitingList := &systemSmartContracts.WaitingList{
		FirstKey:      make([]byte, 0),
		LastKey:       make([]byte, 0),
		Length:        0,
		LastJailedKey: make([]byte, 0),
	}
	marshaledData, _, _ := stakingSCAcc.RetrieveValue([]byte("waitingList"))
	if len(marshaledData) == 0 {
		return nil
	}

	err := tmp.Marshaller.Unmarshal(waitingList, marshaledData)
	if err != nil {
		return nil
	}

	index := uint32(1)
	nextKey := make([]byte, len(waitingList.FirstKey))
	copy(nextKey, waitingList.FirstKey)

	allPubKeys := make([][]byte, 0)
	for len(nextKey) != 0 && index <= waitingList.Length {
		allPubKeys = append(allPubKeys, nextKey[2:]) // remove "w_" prefix

		element, errGet := stakingcommon.GetWaitingListElement(stakingSCAcc, tmp.Marshaller, nextKey)
		if errGet != nil {
			return nil
		}

		nextKey = make([]byte, len(element.NextKey))
		if len(element.NextKey) == 0 {
			break
		}
		index++
		copy(nextKey, element.NextKey)
	}
	return allPubKeys
}
