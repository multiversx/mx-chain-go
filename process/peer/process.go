package peer

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ShardedPeerAdapters represents a list of peer account adapters organized by the shard id
//  each validator account is alocated to
type ShardedPeerAdapters map[uint32]state.AccountsAdapter

type peerProcess struct {
	marshalizer marshal.Marshalizer
	shardHeaderStorage storage.Storer
	nodesCoordinator sharding.NodesCoordinator
	shardCoordinator sharding.Coordinator
	adrConv state.AddressConverter
	peerAdapters ShardedPeerAdapters
	mutPeerAdapters *sync.RWMutex
}

// NewPeerProcessor instantiates a new peerProcess structure responsible of keeping account of
//  each validator actions in the consensus process
func NewPeerProcessor(
	peerAdapters ShardedPeerAdapters,
	adrConv state.AddressConverter,
	nodesCoordinator sharding.NodesCoordinator,
	shardCoordinator sharding.Coordinator,
	shardHeaderStorage storage.Storer,
) (*peerProcess, error) {
	if peerAdapters == nil {
		return nil, process.ErrNilPeerAccountsAdapter
	}
	if adrConv == nil {
		return nil, process.ErrNilAddressConverter
	}
	if nodesCoordinator == nil {
		return nil, process.ErrNilNodesCoordinator
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if shardHeaderStorage == nil {
		return nil, process.ErrNilShardHeaderStorage
	}

	return &peerProcess{
		mutPeerAdapters: &sync.RWMutex{},
		peerAdapters: peerAdapters,
		adrConv: adrConv,
		nodesCoordinator: nodesCoordinator,
		shardCoordinator: shardCoordinator,
		shardHeaderStorage: shardHeaderStorage,
	}, nil
}

// LoadInitialState takes an initial peer list, validates it and sets up the initial state for each of the peers
func (p *peerProcess) LoadInitialState(in []*sharding.InitialNode) error {
	for _, node := range in {
		err := p.processNode(node)
		if err != nil {
			return err
		}
	}

	for _, peerAdapter := range p.peerAdapters {
		_, err := peerAdapter.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// IsNodeValid calculates if a node that's present in the initial validator list
//  contains all the required information in order to be able to participate in consensus
func (p *peerProcess) IsNodeValid(node *sharding.InitialNode) bool {
	if len(node.PubKey) == 0 {
		return false
	}
	if len(node.Address) == 0 {
		return false
	}

	return true
}

// UpdatePeerState takes the current and previous headers and updates the peer state
//  for all of the consensus members
func (p *peerProcess) UpdatePeerState(header, previousHeader data.HeaderHandler) error {
	consensusGroup, err := p.nodesCoordinator.ComputeValidatorsGroup(previousHeader.GetRandSeed(), previousHeader.GetRound(), previousHeader.GetShardID())
	if err != nil {
		return err
	}

	err  = p.updateValidatorInfo(consensusGroup, previousHeader.GetPubKeysBitmap(), previousHeader.GetShardID())
	if err != nil {
		return err
	}

	if header.GetShardID() == sharding.MetachainShardId {
		err = p.updateShardDataPeerState(header)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *peerProcess) updateShardDataPeerState(header data.HeaderHandler) error {
	metaHeader, ok := header.(*block.MetaBlock)
	if !ok {
		return process.ErrInvalidMetaHeader
	}

	for _, h := range metaHeader.ShardInfo {
		shardHeaderBytes, err := p.shardHeaderStorage.Get(h.HeaderHash)
		if err != nil {
			return err
		}

		shardHeader := &block.Header{}
		err = p.marshalizer.Unmarshal(shardHeader, shardHeaderBytes)
		if err != nil {
			return err
		}

		shardConsensus, err := p.nodesCoordinator.ComputeValidatorsGroup(shardHeader.GetRandSeed(), shardHeader.GetRound(), shardHeader.GetShardID())
		if err != nil {
			return err
		}

		err = p.updateValidatorInfo(shardConsensus, shardHeader.GetPubKeysBitmap(), shardHeader.GetShardID())
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *peerProcess) processNode(node *sharding.InitialNode) error {
	if !p.IsNodeValid(node) {
		return process.ErrInvalidInitialNodesState
	}

	peerAccount, err := p.generatePeerAccount(node)
	if err != nil {
		return err
	}

	err = p.savePeerAccountData(peerAccount, node)
	if err != nil {
		return err
	}

	return nil
}

func (p *peerProcess) generatePeerAccount(node *sharding.InitialNode) (*state.PeerAccount, error) {
	address, err := p.adrConv.CreateAddressFromHex(node.Address)
	if err != nil {
		return nil, err
	}

	peerAdapter, err := p.getPeerAdapter(node.AssignedShard())
	if err != nil {
		return nil, err
	}

	acc, err := peerAdapter.GetAccountWithJournal(address)
	if err != nil {
		return nil, err
	}

	peerAccount, ok := acc.(*state.PeerAccount)
	if !ok {
		return nil, process.ErrInvalidPeerAccount
	}

	return peerAccount, nil
}

func (p *peerProcess) savePeerAccountData(peerAccount *state.PeerAccount, data *sharding.InitialNode) error {
	err := peerAccount.SetAddressWithJournal([]byte(data.Address))
	if err != nil {
		return err
	}

	err = peerAccount.SetSchnorrPublicKeyWithJournal([]byte(data.Address))
	if err != nil {
		return err
	}

	err = peerAccount.SetBLSPublicKeyWithJournal([]byte(data.PubKey))
	if err != nil {
		return err
	}

	return nil
}

func (p *peerProcess) updateValidatorInfo(validatorList []sharding.Validator, signingBitmap []byte, shardId uint32) error {
	lenValidators := len(validatorList)
	for i := 0; i < lenValidators; i++ {
		peerAcc, err := p.getPeerAccount(validatorList[i].Address(), shardId)
		if err != nil {
			return err
		}

		isLeader := i == 0
		validatorSigned := (signingBitmap[i/8] & (1 << (uint16(i) % 8))) != 0
		if isLeader && validatorSigned {
			err = peerAcc.IncreaseLeaderSuccessRateWithJournal()
		}
		if isLeader && !validatorSigned {
			err = peerAcc.DecreaseLeaderSuccessRateWithJournal()
		}

		if !isLeader && validatorSigned {
			err = peerAcc.IncreaseValidatorSuccessRateWithJournal()
		}
		if !isLeader && !validatorSigned {
			err = peerAcc.DecreaseValidatorSuccessRateWithJournal()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *peerProcess) getPeerAccount(address []byte, shardId uint32) (*state.PeerAccount, error) {
	peerAdapter, err := p.getPeerAdapter(shardId)
	if err != nil {
		return nil, err
	}

	addressContainer, err := p.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	account, err := peerAdapter.GetExistingAccount(addressContainer)
	if err != nil {
		return nil, err
	}

	peerAccount, ok := account.(*state.PeerAccount)
	if !ok {
		return nil, process.ErrInvalidPeerAccount
	}

	return peerAccount, nil
}

func (p *peerProcess) getPeerAdapter(shardId uint32) (state.AccountsAdapter, error) {
	p.mutPeerAdapters.RLock()
	defer p.mutPeerAdapters.RUnlock()

	adapter, exists := p.peerAdapters[shardId]
	if !exists {
		return nil, errors.New(fmt.Sprintf("peer account adapter for shard id %d is not initialized", shardId))
	}

	return adapter, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *peerProcess) IsInterfaceNil() bool {
	if p == nil {
		return true
	}
	return false
}
