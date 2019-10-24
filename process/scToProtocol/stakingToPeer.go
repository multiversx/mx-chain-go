package scToProtocol

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
)

// stakingToPeer defines the component which will translate changes from staking SC state
// to validator statistics trie
type stakingToPeer struct {
	peerState state.AccountsAdapter
	baseState state.AccountsAdapter

	argParser process.ArgumentsParser
}

func NewStakingToPeer() (*stakingToPeer, error) {
	return nil, nil
}

// UpdateProtocol applies changes from staking smart contract to peer state and creates the actual peer changes
func (stp *stakingToPeer) UpdateProtocol(body block.Body) error {
	affectedStates := make(map[string]struct{})

	for _, miniBlock := range body {
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			tx, err := process.GetTxForCurrentBlock(txHash)
			if err != nil {
				return err
			}

			if !bytes.Equal(tx.GetRecvAddress(), factory.StakingSCAddress) {
				continue
			}

			scr, ok := tx.(*smartContractResult.SmartContractResult)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			storageUpdates, err := stp.argParser.GetStorageUpdates(scr.Data)
			if err != nil {
				return err
			}

			for _, storageUpdate := range storageUpdates {
				affectedStates[string(storageUpdate.Offset)] = struct{}{}
			}
		}
	}

	return nil
}

func (stp *stakingToPeer) PeerChanges() {

}
