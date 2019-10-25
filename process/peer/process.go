package peer

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgValidatorStatisticsProcessor holds all dependencies for the validatorStatistics
type ArgValidatorStatisticsProcessor struct {
	InitialNodes       []*sharding.InitialNode
	Marshalizer        marshal.Marshalizer
	ShardHeaderStorage storage.Storer
	MetaHeaderStorage  storage.Storer
	NodesCoordinator   sharding.NodesCoordinator
	ShardCoordinator   sharding.Coordinator
	AdrConv            state.AddressConverter
	PeerAdapter        state.AccountsAdapter
}

type validatorStatistics struct {
	marshalizer        marshal.Marshalizer
	shardHeaderStorage storage.Storer
	metaHeaderStorage  storage.Storer
	nodesCoordinator   sharding.NodesCoordinator
	shardCoordinator   sharding.Coordinator
	adrConv            state.AddressConverter
	peerAdapter        state.AccountsAdapter
}

// NewValidatorStatisticsProcessor instantiates a new validatorStatistics structure responsible of keeping account of
//  each validator actions in the consensus process
func NewValidatorStatisticsProcessor(arguments ArgValidatorStatisticsProcessor) (*validatorStatistics, error) {
	if arguments.PeerAdapter == nil || arguments.PeerAdapter.IsInterfaceNil() {
		return nil, process.ErrNilPeerAccountsAdapter
	}
	if arguments.AdrConv == nil || arguments.AdrConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if arguments.NodesCoordinator == nil || arguments.NodesCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilNodesCoordinator
	}
	if arguments.ShardCoordinator == nil || arguments.ShardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if arguments.ShardHeaderStorage == nil || arguments.ShardHeaderStorage.IsInterfaceNil() {
		return nil, process.ErrNilShardHeaderStorage
	}
	if arguments.MetaHeaderStorage == nil || arguments.MetaHeaderStorage.IsInterfaceNil() {
		return nil, process.ErrNilMetaHeaderStorage
	}
	if arguments.Marshalizer == nil || arguments.Marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}

	vs := &validatorStatistics{
		peerAdapter:        arguments.PeerAdapter,
		adrConv:            arguments.AdrConv,
		nodesCoordinator:   arguments.NodesCoordinator,
		shardCoordinator:   arguments.ShardCoordinator,
		shardHeaderStorage: arguments.ShardHeaderStorage,
		metaHeaderStorage:  arguments.MetaHeaderStorage,
		marshalizer:        arguments.Marshalizer,
	}

	err := vs.SaveInitialState(arguments.InitialNodes)
	if err != nil {
		return nil, err
	}

	return vs, nil
}

// SaveInitialState takes an initial peer list, validates it and sets up the initial state for each of the peers
func (p *validatorStatistics) SaveInitialState(in []*sharding.InitialNode) error {
	for _, node := range in {
		err := p.initializeNode(node)
		if err != nil {
			return err
		}
	}

	_, err := p.peerAdapter.Commit()
	if err != nil {
		return err
	}

	return nil
}

// IsNodeValid calculates if a node that's present in the initial validator list
//  contains all the required information in order to be able to participate in consensus
func (p *validatorStatistics) IsNodeValid(node *sharding.InitialNode) bool {
	if len(node.PubKey) == 0 {
		return false
	}
	if len(node.Address) == 0 {
		return false
	}

	return true
}

// UpdatePeerState takes the ca header, updates the peer state for all of the
//  consensus members and returns the new root hash
func (p *validatorStatistics) UpdatePeerState(header data.HeaderHandler) ([]byte, error) {
	if header.GetNonce() == 0 {
		return p.peerAdapter.RootHash()
	}

	consensusGroup, err := p.nodesCoordinator.ComputeValidatorsGroup(header.GetPrevRandSeed(), header.GetRound(), header.GetShardID())
	if err != nil {
		return nil, err
	}

	err = p.updateValidatorInfo(consensusGroup, header.GetShardID())
	if err != nil {
		return nil, err
	}

	previousHeader, err := p.getHeader(header.GetPrevHash())
	if err != nil {
		return nil, err
	}

	err = p.checkForMissedBlocks(header, previousHeader)
	if err != nil {
		return nil, err
	}

	err = p.updateShardDataPeerState(header)
	if err != nil {
		return nil, err
	}

	return p.peerAdapter.Commit()
}

func (p *validatorStatistics) checkForMissedBlocks(currentHeader, previousHeader data.HeaderHandler) error {
	if currentHeader.GetRound()-previousHeader.GetRound() <= 1 {
		return nil
	}

	for i := previousHeader.GetRound() + 1; i < currentHeader.GetRound(); i++ {
		consensusGroup, err := p.nodesCoordinator.ComputeValidatorsGroup(previousHeader.GetPrevRandSeed(), i, previousHeader.GetShardID())
		if err != nil {
			return err
		}

		leaderPeerAcc, err := p.getPeerAccount(consensusGroup[0].Address())
		if err != nil {
			return err
		}

		err = leaderPeerAcc.DecreaseLeaderSuccessRateWithJournal()
		if err != nil {
			return err
		}
	}

	return nil
}

// RevertPeerState takes the current and previous headers and undos the peer state
//  for all of the consensus members
func (p *validatorStatistics) RevertPeerState(header data.HeaderHandler) error {
	_ = p.peerAdapter.RecreateTrie(header.GetValidatorStatsRootHash())
	return nil
}

func (p *validatorStatistics) updateShardDataPeerState(header data.HeaderHandler) error {
	metaHeader, ok := header.(*block.MetaBlock)
	if !ok {
		return process.ErrInvalidMetaHeader
	}

	for _, h := range metaHeader.ShardInfo {
		shardHeader, err := p.getHeader(h.HeaderHash)
		if err != nil {
			return err
		}

		shardConsensus, err := p.nodesCoordinator.ComputeValidatorsGroup(shardHeader.GetPrevRandSeed(), shardHeader.GetRound(), shardHeader.GetShardID())
		if err != nil {
			return err
		}

		err = p.updateValidatorInfo(shardConsensus, shardHeader.GetShardID())
		if err != nil {
			return err
		}

		previousHeader, err := p.getMetaHeader(shardHeader.GetPrevHash())
		if err != nil {
			return err
		}

		err = p.checkForMissedBlocks(shardHeader, previousHeader)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *validatorStatistics) initializeNode(node *sharding.InitialNode) error {
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

func (p *validatorStatistics) generatePeerAccount(node *sharding.InitialNode) (*state.PeerAccount, error) {
	address, err := p.adrConv.CreateAddressFromHex(node.Address)
	if err != nil {
		return nil, err
	}

	acc, err := p.peerAdapter.GetAccountWithJournal(address)
	if err != nil {
		return nil, err
	}

	peerAccount, ok := acc.(*state.PeerAccount)
	if !ok {
		return nil, process.ErrInvalidPeerAccount
	}

	return peerAccount, nil
}

func (p *validatorStatistics) savePeerAccountData(peerAccount *state.PeerAccount, data *sharding.InitialNode) error {
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

func (p *validatorStatistics) updateValidatorInfo(validatorList []sharding.Validator, shardId uint32) error {
	lenValidators := len(validatorList)
	for i := 0; i < lenValidators; i++ {
		peerAcc, err := p.getPeerAccount(validatorList[i].Address())
		if err != nil {
			return err
		}

		isLeader := i == 0
		if isLeader {
			err = peerAcc.IncreaseLeaderSuccessRateWithJournal()
		} else {
			err = peerAcc.IncreaseValidatorSuccessRateWithJournal()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *validatorStatistics) getPeerAccount(address []byte) (state.PeerAccountHandler, error) {
	addressContainer, err := p.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	account, err := p.peerAdapter.GetExistingAccount(addressContainer)
	if err != nil {
		return nil, err
	}

	peerAccount, ok := account.(state.PeerAccountHandler)
	if !ok {
		return nil, process.ErrInvalidPeerAccount
	}

	return peerAccount, nil
}

func (p *validatorStatistics) getHeader(headerHash []byte) (data.HeaderHandler, error) {
	shardHeaderBytes, err := p.shardHeaderStorage.Get(headerHash)
	if err != nil {
		return nil, err
	}

	header := &block.Header{}
	err = p.marshalizer.Unmarshal(header, shardHeaderBytes)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (p *validatorStatistics) getMetaHeader(headerHash []byte) (data.HeaderHandler, error) {
	metaHeaderBytes, err := p.metaHeaderStorage.Get(headerHash)
	if err != nil {
		return nil, err
	}

	header := &block.MetaBlock{}
	err = p.marshalizer.Unmarshal(header, metaHeaderBytes)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *validatorStatistics) IsInterfaceNil() bool {
	if p == nil {
		return true
	}
	return false
}
