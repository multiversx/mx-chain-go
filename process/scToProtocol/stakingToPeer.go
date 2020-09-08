package scToProtocol

import (
	"bytes"
	"math"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.SmartContractToProtocolHandler = (*stakingToPeer)(nil)

var log = logger.GetOrCreate("process/scToProtocol")

// ArgStakingToPeer is struct that contain all components that are needed to create a new stakingToPeer object
type ArgStakingToPeer struct {
	PubkeyConv  core.PubkeyConverter
	Hasher      hashing.Hasher
	Marshalizer marshal.Marshalizer
	PeerState   state.AccountsAdapter
	BaseState   state.AccountsAdapter

	ArgParser   process.ArgumentsParser
	CurrTxs     dataRetriever.TransactionCacher
	ScQuery     external.SCQueryService
	RatingsData process.RatingsInfoHandler
}

// stakingToPeer defines the component which will translate changes from staking SC state
// to validator statistics trie
type stakingToPeer struct {
	pubkeyConv  core.PubkeyConverter
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
	peerState   state.AccountsAdapter
	baseState   state.AccountsAdapter

	argParser    process.ArgumentsParser
	currTxs      dataRetriever.TransactionCacher
	scQuery      external.SCQueryService
	startRating  uint32
	unJailRating uint32
	jailRating   uint32
}

// NewStakingToPeer creates the component which moves from staking sc state to peer state
func NewStakingToPeer(args ArgStakingToPeer) (*stakingToPeer, error) {
	err := checkIfNil(args)
	if err != nil {
		return nil, err
	}

	st := &stakingToPeer{
		pubkeyConv:   args.PubkeyConv,
		hasher:       args.Hasher,
		marshalizer:  args.Marshalizer,
		peerState:    args.PeerState,
		baseState:    args.BaseState,
		argParser:    args.ArgParser,
		currTxs:      args.CurrTxs,
		scQuery:      args.ScQuery,
		startRating:  args.RatingsData.StartRating(),
		unJailRating: args.RatingsData.StartRating(),
		jailRating:   args.RatingsData.MinRating(),
	}

	return st, nil
}

func checkIfNil(args ArgStakingToPeer) error {
	if check.IfNil(args.PubkeyConv) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.PeerState) {
		return process.ErrNilPeerAccountsAdapter
	}
	if check.IfNil(args.BaseState) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.ArgParser) {
		return process.ErrNilArgumentParser
	}
	if check.IfNil(args.CurrTxs) {
		return process.ErrNilTxForCurrentBlockHandler
	}
	if check.IfNil(args.ScQuery) {
		return process.ErrNilSCDataGetter
	}
	if check.IfNil(args.RatingsData) {
		return process.ErrNilRatingsInfoHandler
	}

	return nil
}

func (stp *stakingToPeer) getPeerAccount(key []byte) (state.PeerAccountHandler, error) {
	account, err := stp.peerState.LoadAccount(key)
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
func (stp *stakingToPeer) UpdateProtocol(body *block.Body, nonce uint64) error {
	affectedStates, err := stp.getAllModifiedStates(body)
	if err != nil {
		return err
	}

	for _, key := range affectedStates {
		if len(key) != stp.pubkeyConv.Len() {
			continue
		}

		blsPubKey := []byte(key)
		log.Trace("get on StakingScAddress called", "blsKey", blsPubKey)

		query := process.SCQuery{
			ScAddress: vm.StakingSCAddress,
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
			err = stp.peerState.RemoveAccount(blsPubKey)
			log.LogIfError(err, "staking to protocol RemoveAccount blsPubKey", blsPubKey)

			continue
		}

		var stakingData systemSmartContracts.StakedData
		err = stp.marshalizer.Unmarshal(&stakingData, data)
		if err != nil {
			return err
		}

		err = stp.updatePeerState(stakingData, blsPubKey, nonce)
		if err != nil {
			return err
		}
	}

	return nil
}

func (stp *stakingToPeer) updatePeerState(
	stakingData systemSmartContracts.StakedData,
	blsPubKey []byte,
	nonce uint64,
) error {
	account, err := stp.getPeerAccount(blsPubKey)
	if err != nil {
		return err
	}

	isUnJailForInactive := len(account.GetBLSPublicKey()) > 0 && !stakingData.Staked &&
		stakingData.UnJailedNonce == nonce && account.GetList() == string(core.JailedList)
	if isUnJailForInactive {
		account.SetListAndIndex(account.GetShardId(), string(core.InactiveList), uint32(stakingData.UnJailedNonce))
		account.SetTempRating(stp.jailRating)

		return stp.peerState.SaveAccount(account)
	}

	if stakingData.StakedNonce == math.MaxUint64 {
		return nil
	}

	if !bytes.Equal(account.GetRewardAddress(), stakingData.RewardAddress) {
		err = account.SetRewardAddress(stakingData.RewardAddress)
		if err != nil {
			return err
		}
	}

	if !bytes.Equal(account.GetBLSPublicKey(), blsPubKey) {
		err = account.SetBLSPublicKey(blsPubKey)
		if err != nil {
			return err
		}
	}

	isValidator := account.GetList() == string(core.EligibleList) || account.GetList() == string(core.WaitingList)
	if !stakingData.Jailed {
		if stakingData.StakedNonce == nonce && !isValidator {
			account.SetListAndIndex(account.GetShardId(), string(core.NewList), uint32(stakingData.StakedNonce))
			account.SetTempRating(stp.startRating)
			account.SetUnStakedEpoch(core.DefaultUnstakedEpoch)
		}

		if stakingData.UnStakedNonce == nonce && account.GetList() != string(core.InactiveList) {
			account.SetListAndIndex(account.GetShardId(), string(core.LeavingList), uint32(stakingData.UnStakedNonce))
			account.SetUnStakedEpoch(stakingData.UnStakedEpoch)
		}
	}

	if stakingData.UnJailedNonce == nonce {
		if account.GetTempRating() < stp.unJailRating {
			account.SetTempRating(stp.unJailRating)
		}

		isNewValidator := !isValidator && account.GetUnStakedEpoch() == core.DefaultUnstakedEpoch && stakingData.Staked
		if isNewValidator {
			account.SetListAndIndex(account.GetShardId(), string(core.NewList), uint32(stakingData.UnJailedNonce))
		}
	}

	if stakingData.JailedNonce == nonce && account.GetList() != string(core.InactiveList) {
		account.SetListAndIndex(account.GetShardId(), string(core.LeavingList), uint32(stakingData.JailedNonce))
		account.SetTempRating(stp.jailRating)
	}

	err = stp.peerState.SaveAccount(account)
	if err != nil {
		return err
	}

	return nil
}

func (stp *stakingToPeer) getAllModifiedStates(body *block.Body) ([]string, error) {
	affectedStates := make([]string, 0)

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}
		if miniBlock.SenderShardID != core.MetachainShardId || miniBlock.ReceiverShardID != core.MetachainShardId {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			tx, err := stp.currTxs.GetTx(txHash)
			if err != nil {
				continue
			}

			if !bytes.Equal(tx.GetRcvAddr(), vm.StakingSCAddress) {
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
