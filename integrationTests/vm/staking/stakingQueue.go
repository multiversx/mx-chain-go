package staking

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/stakingcommon"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
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

	stakingcommon.AddValidatorData(
		accountsAdapter,
		owner,
		ownerWaitingNodes,
		big.NewInt(int64(2*nodePrice*numOfNodesInStakingQueue)),
		marshaller,
	)

	return ownerWaitingNodes
}

func (tmp *TestMetaProcessor) getWaitingListKeys() [][]byte {
	stakingSCAcc := stakingcommon.LoadUserAccount(tmp.AccountsAdapter, vm.StakingSCAddress)

	waitingList := &systemSmartContracts.WaitingList{
		FirstKey:      make([]byte, 0),
		LastKey:       make([]byte, 0),
		Length:        0,
		LastJailedKey: make([]byte, 0),
	}
	marshaledData, _ := stakingSCAcc.DataTrieTracker().RetrieveValue([]byte("waitingList"))
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
		allPubKeys = append(allPubKeys, nextKey)

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
