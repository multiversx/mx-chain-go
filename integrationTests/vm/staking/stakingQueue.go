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

	// We need to save one key and then add keys to waiting list because there is a bug in those functions
	// TODO: FIX bug in testscommon.AddKeysToWaitingList to also init staking queue if there are no keys in list
	stakingcommon.SaveOneKeyToWaitingList(
		accountsAdapter,
		ownerWaitingNodes[0],
		marshaller,
		owner,
		owner,
	)
	if numOfNodesInStakingQueue > 1 {
		stakingcommon.AddKeysToWaitingList(
			accountsAdapter,
			ownerWaitingNodes[1:],
			marshaller,
			owner,
			owner,
		)
	}
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

		element, errGet := tmp.getWaitingListElement(stakingSCAcc, nextKey)
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

func (tmp *TestMetaProcessor) getWaitingListElement(stakingSCAcc state.UserAccountHandler, key []byte) (*systemSmartContracts.ElementInList, error) {
	marshaledData, _ := stakingSCAcc.DataTrieTracker().RetrieveValue(key)
	if len(marshaledData) == 0 {
		return nil, vm.ErrElementNotFound
	}

	element := &systemSmartContracts.ElementInList{}
	err := tmp.Marshaller.Unmarshal(element, marshaledData)
	if err != nil {
		return nil, err
	}

	return element, nil
}
