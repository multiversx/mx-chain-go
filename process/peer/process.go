package peer

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"math/big"
)

// ArgValidatorStatisticsProcessor holds all dependencies for the validatorStatistics
type ArgValidatorStatisticsProcessor struct {
	InitialNodes     []*sharding.InitialNode
	Economics        *economics.EconomicsData
	Marshalizer      marshal.Marshalizer
	NodesCoordinator sharding.NodesCoordinator
	ShardCoordinator sharding.Coordinator
	DataPool         dataRetriever.MetaPoolsHolder
	StorageService   dataRetriever.StorageService
	AdrConv          state.AddressConverter
	PeerAdapter      state.AccountsAdapter
}

type validatorStatistics struct {
	marshalizer      marshal.Marshalizer
	dataPool         dataRetriever.MetaPoolsHolder
	storageService   dataRetriever.StorageService
	nodesCoordinator sharding.NodesCoordinator
	shardCoordinator sharding.Coordinator
	adrConv          state.AddressConverter
	peerAdapter      state.AccountsAdapter
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
	if arguments.DataPool == nil || arguments.DataPool.IsInterfaceNil() {
		return nil, process.ErrNilDataPoolHolder
	}
	if arguments.StorageService == nil || arguments.StorageService.IsInterfaceNil() {
		return nil, process.ErrNilStorage
	}
	if arguments.NodesCoordinator == nil || arguments.NodesCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilNodesCoordinator
	}
	if arguments.ShardCoordinator == nil || arguments.ShardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if arguments.Marshalizer == nil || arguments.Marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if arguments.Economics == nil {
		return nil, process.ErrNilEconomicsData
	}
	if arguments.Economics.StakeValue() == nil {
		return nil, process.ErrNilEconomicsData
	}

	vs := &validatorStatistics{
		peerAdapter:      arguments.PeerAdapter,
		adrConv:          arguments.AdrConv,
		nodesCoordinator: arguments.NodesCoordinator,
		shardCoordinator: arguments.ShardCoordinator,
		dataPool:         arguments.DataPool,
		storageService:   arguments.StorageService,
		marshalizer:      arguments.Marshalizer,
	}

	err := vs.saveInitialState(arguments.InitialNodes, arguments.Economics.StakeValue())
	if err != nil {
		return nil, err
	}

	return vs, nil
}

// saveInitialState takes an initial peer list, validates it and sets up the initial state for each of the peers
func (p *validatorStatistics) saveInitialState(in []*sharding.InitialNode, stakeValue *big.Int) error {
	for _, node := range in {
		err := p.initializeNode(node, stakeValue)
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

	//previousHeader, err := p.getMetaHeaderFromStorage(header.GetPrevHash())
	previousHeader, err := process.GetMetaHeader(header.GetPrevHash(), p.dataPool.MetaBlocks(), p.marshalizer, p.storageService)
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

	return p.peerAdapter.RootHash()
}

// Commit commits the validator statistics trie and returns the root hash
func (p *validatorStatistics) Commit() ([]byte, error) {
	return p.peerAdapter.Commit()
}

// RootHash returns the root hash of the validator statistics trie
func (p *validatorStatistics) RootHash() ([]byte, error) {
	return p.peerAdapter.RootHash()
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

// RevertPeerStateToSnapshot reverts the applied changes to the peerAdapter
func (p *validatorStatistics) RevertPeerStateToSnapshot(snapshot int) error {
	return p.peerAdapter.RevertToSnapshot(snapshot)
}

func (p *validatorStatistics) updateShardDataPeerState(header data.HeaderHandler) error {
	metaHeader, ok := header.(*block.MetaBlock)
	if !ok {
		return process.ErrInvalidMetaHeader
	}

	for _, h := range metaHeader.ShardInfo {
		shardHeader, err := process.GetShardHeaderFromPool(h.HeaderHash, p.dataPool.ShardHeaders())
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

		if shardHeader.GetNonce() == 1 {
			continue
		}

		//previousHeader, err :=  p.getHeaderFromStorage(shardHeader.GetPrevHash())
		previousHeader, err := process.GetShardHeader(shardHeader.GetPrevHash(), p.dataPool.ShardHeaders(), p.marshalizer, p.storageService)
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

func (p *validatorStatistics) initializeNode(node *sharding.InitialNode, stakeValue *big.Int) error {
	if !p.IsNodeValid(node) {
		return process.ErrInvalidInitialNodesState
	}

	peerAccount, err := p.generatePeerAccount(node)
	if err != nil {
		return err
	}

	err = p.savePeerAccountData(peerAccount, node, stakeValue)
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

func (p *validatorStatistics) savePeerAccountData(
	peerAccount *state.PeerAccount,
	data *sharding.InitialNode,
	stakeValue *big.Int,
) error {
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

	err = peerAccount.SetStakeWithJournal(stakeValue)
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

// IsInterfaceNil returns true if there is no value under the interface
func (p *validatorStatistics) IsInterfaceNil() bool {
	if p == nil {
		return true
	}
	return false
}
