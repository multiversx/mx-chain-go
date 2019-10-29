package scToProtocol

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"math/big"
	"sort"
	"sync"
)

type ArgStakingToPeer struct {
	AdrConv     state.AddressConverter
	Hasher      hashing.Hasher
	Marshalizer marshal.Marshalizer
	PeerState   state.AccountsAdapter
	BaseState   state.AccountsAdapter

	ArgParser    process.ArgumentsParser
	CurrTxs      dataRetriever.TransactionCacher
	ScDataGetter external.ScDataGetter
}

// stakingToPeer defines the component which will translate changes from staking SC state
// to validator statistics trie
type stakingToPeer struct {
	adrConv     state.AddressConverter
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
	peerState   state.AccountsAdapter
	baseState   state.AccountsAdapter

	argParser    process.ArgumentsParser
	currTxs      dataRetriever.TransactionCacher
	scDataGetter external.ScDataGetter

	mutPeerChanges sync.Mutex
	peerChanges    map[string]block.PeerData
}

// NewStakingToPeer creates the component which moves from staking sc state to peer state
func NewStakingToPeer(args ArgStakingToPeer) (*stakingToPeer, error) {
	err := checkIfNil(args)
	if err != nil {
		return nil, err
	}

	st := &stakingToPeer{
		adrConv:        args.AdrConv,
		hasher:         args.Hasher,
		marshalizer:    args.Marshalizer,
		peerState:      args.PeerState,
		baseState:      args.BaseState,
		argParser:      args.ArgParser,
		currTxs:        args.CurrTxs,
		scDataGetter:   args.ScDataGetter,
		mutPeerChanges: sync.Mutex{},
		peerChanges:    make(map[string]block.PeerData),
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
	if args.ScDataGetter == nil || args.ScDataGetter.IsInterfaceNil() {
		return process.ErrNilSCDataGetter
	}

	return nil
}

func (stp *stakingToPeer) getPeerAccount(key []byte) (*state.PeerAccount, error) {
	adrSrc, err := stp.adrConv.CreateAddressFromPublicKeyBytes([]byte(key))
	if err != nil {
		return nil, err
	}

	account, err := stp.peerState.GetAccountWithJournal(adrSrc)
	if err != nil {
		return nil, err
	}

	peerAcc, ok := account.(*state.PeerAccount)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return peerAcc, nil
}

// UpdateProtocol applies changes from staking smart contract to peer state and creates the actual peer changes
func (stp *stakingToPeer) UpdateProtocol(body block.Body, nonce uint64) error {
	stp.mutPeerChanges.Lock()
	stp.peerChanges = make(map[string]block.PeerData)
	stp.mutPeerChanges.Unlock()

	affectedStates, err := stp.getAllModifiedStates(body)
	if err != nil {
		return err
	}

	for key := range affectedStates {
		peerAcc, err := stp.getPeerAccount([]byte(key))
		if err != nil {
			return err
		}

		data, err := stp.scDataGetter.Get(factory.StakingSCAddress, "get", []byte(key))
		if err != nil {
			return err
		}

		// no data under key -> peer can be deleted from trie
		if len(data) == 0 {
			err = stp.peerUnregistered(peerAcc, nonce)
			if err != nil {
				return err
			}

			adrSrc, err := stp.adrConv.CreateAddressFromPublicKeyBytes([]byte(key))
			if err != nil {
				return err
			}

			return stp.peerState.RemoveAccount(adrSrc)
		}

		var stakingData systemSmartContracts.StakingData
		err = stp.marshalizer.Unmarshal(&stakingData, data)
		if err != nil {
			return err
		}

		err = stp.createPeerChangeData(stakingData, peerAcc, nonce)
		if err != nil {
			return err
		}

		err = stp.updatePeerState(stakingData, peerAcc, nonce)
		if err != nil {
			return err
		}
	}

	return nil
}

func (stp *stakingToPeer) peerUnregistered(account *state.PeerAccount, nonce uint64) error {
	stp.mutPeerChanges.Lock()
	defer stp.mutPeerChanges.Unlock()

	actualPeerChange := block.PeerData{
		Address:     account.Address,
		PublicKey:   account.BLSPublicKey,
		Action:      block.PeerDeregistration,
		TimeStamp:   nonce,
		ValueChange: big.NewInt(0).Set(account.Stake),
	}

	peerHash, err := core.CalculateHash(stp.marshalizer, stp.hasher, actualPeerChange)
	if err != nil {
		return err
	}

	stp.peerChanges[string(peerHash)] = actualPeerChange
	return nil
}

func (stp *stakingToPeer) updatePeerState(
	stakingData systemSmartContracts.StakingData,
	account *state.PeerAccount,
	nonce uint64,
) error {
	if !bytes.Equal(stakingData.BlsPubKey, account.BLSPublicKey) {
		err := account.SetBLSPublicKeyWithJournal(stakingData.BlsPubKey)
		if err != nil {
			return err
		}
	}

	if stakingData.StakeValue.Cmp(account.Stake) != 0 {
		err := account.SetStakeWithJournal(stakingData.StakeValue)
		if err != nil {
			return err
		}
	}

	if stakingData.StartNonce != account.Nonce {
		err := account.SetNonceWithJournal(stakingData.StartNonce)
		if err != nil {
			return err
		}

		err = account.SetNodeInWaitingListWithJournal(true)
		if err != nil {
			return err
		}
	}

	if stakingData.UnStakedNonce != account.UnStakedNonce {
		err := account.SetUnStakedNonceWithJournal(stakingData.UnStakedNonce)
		if err != nil {
			return err
		}
	}

	return nil
}

func (stp *stakingToPeer) createPeerChangeData(
	stakingData systemSmartContracts.StakingData,
	account *state.PeerAccount,
	nonce uint64,
) error {
	stp.mutPeerChanges.Lock()
	defer stp.mutPeerChanges.Unlock()

	actualPeerChange := block.PeerData{
		Address:     account.Address,
		PublicKey:   account.BLSPublicKey,
		Action:      0,
		TimeStamp:   nonce,
		ValueChange: big.NewInt(0),
	}

	if len(account.BLSPublicKey) == 0 {
		actualPeerChange.Action = block.PeerRegistrantion
		actualPeerChange.TimeStamp = stakingData.StartNonce
		actualPeerChange.ValueChange.Set(stakingData.StakeValue)

		peerHash, err := core.CalculateHash(stp.marshalizer, stp.hasher, actualPeerChange)
		if err != nil {
			return err
		}

		stp.peerChanges[string(peerHash)] = actualPeerChange

		return nil
	}

	if account.Stake.Cmp(stakingData.StakeValue) != 0 {
		actualPeerChange.ValueChange.Sub(account.Stake, stakingData.StakeValue)
		if account.Stake.Cmp(stakingData.StakeValue) < 0 {
			actualPeerChange.Action = block.PeerSlashed
		} else {
			actualPeerChange.Action = block.PeerReStake
		}
	}

	if stakingData.StartNonce == nonce {
		actualPeerChange.Action = block.PeerRegistrantion
	}

	if stakingData.UnStakedNonce == nonce {
		actualPeerChange.Action = block.PeerUnstaking
	}
	//TODO what you do with actualPeerChange

	return nil
}

func (stp *stakingToPeer) getAllModifiedStates(body block.Body) (map[string]struct{}, error) {
	affectedStates := make(map[string]struct{})

	for _, miniBlock := range body {
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}

		for _, txHash := range miniBlock.TxHashes {
			tx, err := stp.currTxs.GetTx(txHash)
			if err != nil {
				return nil, err
			}

			if !bytes.Equal(tx.GetRecvAddress(), factory.StakingSCAddress) {
				continue
			}

			scr, ok := tx.(*smartContractResult.SmartContractResult)
			if !ok {
				return nil, process.ErrWrongTypeAssertion
			}

			storageUpdates, err := stp.argParser.GetStorageUpdates(scr.Data)
			if err != nil {
				return nil, err
			}

			for _, storageUpdate := range storageUpdates {
				affectedStates[string(storageUpdate.Offset)] = struct{}{}
			}
		}
	}

	return affectedStates, nil
}

// PeerChanges returns peer changes created in current round
func (stp *stakingToPeer) PeerChanges() []block.PeerData {
	stp.mutPeerChanges.Lock()
	peersData := make([]block.PeerData, 0)
	for _, peerData := range stp.peerChanges {
		peersData = append(peersData, peerData)
	}
	stp.mutPeerChanges.Unlock()

	sort.Slice(peersData, func(i, j int) bool {
		return string(peersData[i].Address) < string(peersData[j].Address)
	})

	return peersData
}

// VerifyPeerChanges verifies if peer changes from header is the same as the one created while processing
func (stp *stakingToPeer) VerifyPeerChanges(peerChanges []block.PeerData) error {
	createdPeersData := stp.PeerChanges()
	createdHash, err := core.CalculateHash(stp.marshalizer, stp.hasher, createdPeersData)
	if err != nil {
		return err
	}

	receivedHash, err := core.CalculateHash(stp.marshalizer, stp.hasher, peerChanges)
	if err != nil {
		return err
	}

	if !bytes.Equal(createdHash, receivedHash) {
		return process.ErrPeerChangesHashDoesNotMatch
	}

	return nil
}

// IsInterfaceNil
func (stp *stakingToPeer) IsInterfaceNil() bool {
	if stp == nil {
		return true
	}
	return false
}
