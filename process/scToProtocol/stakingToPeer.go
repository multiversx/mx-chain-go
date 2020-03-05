package scToProtocol

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("process/scToProtocol")

// ArgStakingToPeer is struct that contain all components that are needed to create a new stakingToPeer object
type ArgStakingToPeer struct {
	AdrConv     state.AddressConverter
	Hasher      hashing.Hasher
	Marshalizer marshal.Marshalizer
	PeerState   state.AccountsAdapter
	BaseState   state.AccountsAdapter

	ArgParser process.ArgumentsParser
	CurrTxs   dataRetriever.TransactionCacher
	ScQuery   external.SCQueryService
}

// stakingToPeer defines the component which will translate changes from staking SC state
// to validator statistics trie
type stakingToPeer struct {
	adrConv     state.AddressConverter
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
	peerState   state.AccountsAdapter
	baseState   state.AccountsAdapter

	argParser process.ArgumentsParser
	currTxs   dataRetriever.TransactionCacher
	scQuery   external.SCQueryService
}

// NewStakingToPeer creates the component which moves from staking sc state to peer state
func NewStakingToPeer(args ArgStakingToPeer) (*stakingToPeer, error) {
	err := checkIfNil(args)
	if err != nil {
		return nil, err
	}

	st := &stakingToPeer{
		adrConv:     args.AdrConv,
		hasher:      args.Hasher,
		marshalizer: args.Marshalizer,
		peerState:   args.PeerState,
		baseState:   args.BaseState,
		argParser:   args.ArgParser,
		currTxs:     args.CurrTxs,
		scQuery:     args.ScQuery,
	}

	return st, nil
}

func checkIfNil(args ArgStakingToPeer) error {
	if args.AdrConv == nil || args.AdrConv.IsInterfaceNil() {
		return process.ErrNilAddressConverter
	}
	if args.Hasher == nil || args.Hasher.IsInterfaceNil() {
		return process.ErrNilHasher
	}
	if args.Marshalizer == nil || args.Marshalizer.IsInterfaceNil() {
		return process.ErrNilMarshalizer
	}
	if args.PeerState == nil || args.PeerState.IsInterfaceNil() {
		return process.ErrNilPeerAccountsAdapter
	}
	if args.BaseState == nil || args.BaseState.IsInterfaceNil() {
		return process.ErrNilAccountsAdapter
	}
	if args.ArgParser == nil || args.ArgParser.IsInterfaceNil() {
		return process.ErrNilArgumentParser
	}
	if args.CurrTxs == nil || args.CurrTxs.IsInterfaceNil() {
		return process.ErrNilTxForCurrentBlockHandler
	}
	if args.ScQuery == nil || args.ScQuery.IsInterfaceNil() {
		return process.ErrNilSCDataGetter
	}

	return nil
}

func (stp *stakingToPeer) getPeerAccount(key []byte) (state.PeerAccountHandler, error) {
	adrSrc, err := stp.adrConv.CreateAddressFromPublicKeyBytes(key)
	if err != nil {
		return nil, err
	}

	account, err := stp.peerState.LoadAccount(adrSrc)
	if err != nil {
		return nil, err
	}

	peerAcc, ok := account.(state.PeerAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return peerAcc, nil
}

// UpdateProtocol applies changes from staking smart contract to peer state and creates the actual peer changes
func (stp *stakingToPeer) UpdateProtocol(body block.Body, _ uint64) error {
	affectedStates, err := stp.getAllModifiedStates(body)
	if err != nil {
		return err
	}

	for _, key := range affectedStates {
		blsPubKey := []byte(key)
		var peerAcc state.PeerAccountHandler
		peerAcc, err = stp.getPeerAccount(blsPubKey)
		if err != nil {
			return err
		}

		log.Trace("get on StakingScAddress called", "blsKey", blsPubKey)

		query := process.SCQuery{
			ScAddress: factory.StakingSCAddress,
			FuncName:  "get",
			Arguments: [][]byte{blsPubKey},
		}
		var vmOutput *vmcommon.VMOutput
		vmOutput, err = stp.scQuery.ExecuteQuery(&query)
		if err != nil {
			return err
		}

		var data []byte
		if len(vmOutput.ReturnData) > 0 {
			data = vmOutput.ReturnData[0]
		}
		// no data under key -> peer can be deleted from trie
		if len(data) == 0 {
			var adrSrc state.AddressContainer
			adrSrc, err = stp.adrConv.CreateAddressFromPublicKeyBytes(blsPubKey)
			if err != nil {
				return err
			}

			err = stp.peerState.RemoveAccount(adrSrc)
			if err != nil {
				return err
			}

			continue
		}

		var stakingData systemSmartContracts.StakedData
		err = stp.marshalizer.Unmarshal(&stakingData, data)
		if err != nil {
			return err
		}

		err = stp.updatePeerState(stakingData, peerAcc, blsPubKey)
		if err != nil {
			return err
		}

		err = stp.peerState.SaveAccount(peerAcc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (stp *stakingToPeer) updatePeerState(
	stakingData systemSmartContracts.StakedData,
	account state.PeerAccountHandler,
	blsPubKey []byte,
) error {

	err := account.SetRewardAddress(stakingData.RewardAddress)
	if err != nil {
		return err
	}

	account.SetBLSPublicKey(blsPubKey)
	account.SetStake(stakingData.StakeValue)

	if stakingData.RegisterNonce != account.GetNonce() {
		account.SetNonce(stakingData.RegisterNonce)
		account.SetNodeInWaitingList(true)
	}
	account.SetUnStakedNonce(stakingData.UnStakedNonce)

	return nil
}

func (stp *stakingToPeer) getAllModifiedStates(body block.Body) ([]string, error) {
	affectedStates := make([]string, 0)

	for _, miniBlock := range body {
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}
		if miniBlock.SenderShardID != core.MetachainShardId {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			tx, err := stp.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			if !bytes.Equal(tx.GetRecvAddress(), factory.StakingSCAddress) {
				continue
			}

			scr, ok := tx.(*smartContractResult.SmartContractResult)
			if !ok {
				return nil, process.ErrWrongTypeAssertion
			}

			storageUpdates, err := stp.argParser.GetStorageUpdates(string(scr.Data))
			if err != nil {
				continue
			}

			for _, storageUpdate := range storageUpdates {
				affectedStates = append(affectedStates, string(storageUpdate.Offset))
			}
		}
	}

	return affectedStates, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stp *stakingToPeer) IsInterfaceNil() bool {
	return stp == nil
}
