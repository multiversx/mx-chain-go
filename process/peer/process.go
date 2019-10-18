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

type validatorActionType uint8

const (
	unknownAction validatorActionType = 0
	leaderSuccess validatorActionType = 1
	leaderFail validatorActionType = 2
	validatorSuccess validatorActionType = 3
	validatorFail validatorActionType = 4
)

type validatorStatistics struct {
	marshalizer marshal.Marshalizer
	shardHeaderStorage storage.Storer
	nodesCoordinator sharding.NodesCoordinator
	shardCoordinator sharding.Coordinator
	adrConv state.AddressConverter
	peerAdapter state.AccountsAdapter
}

// NewValidatorStatisticsProcessor instantiates a new validatorStatistics structure responsible of keeping account of
//  each validator actions in the consensus process
func NewValidatorStatisticsProcessor(
	in []*sharding.InitialNode,
	peerAdapter state.AccountsAdapter,
	adrConv state.AddressConverter,
	nodesCoordinator sharding.NodesCoordinator,
	shardCoordinator sharding.Coordinator,
	shardHeaderStorage storage.Storer,
	marshalizer marshal.Marshalizer,
) (*validatorStatistics, error) {
	if peerAdapter == nil {
		return nil, process.ErrNilPeerAccountsAdapter
	}
	if adrConv == nil || adrConv.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if nodesCoordinator == nil || nodesCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilNodesCoordinator
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if shardHeaderStorage == nil || shardHeaderStorage.IsInterfaceNil() {
		return nil, process.ErrNilShardHeaderStorage
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}

	peerProcessor := &validatorStatistics{
		peerAdapter: peerAdapter,
		adrConv: adrConv,
		nodesCoordinator: nodesCoordinator,
		shardCoordinator: shardCoordinator,
		shardHeaderStorage: shardHeaderStorage,
		marshalizer: marshalizer,
	}

	err := peerProcessor.LoadInitialState(in)
	if err != nil {
		return nil, err
	}

	return peerProcessor, nil
}

// LoadInitialState takes an initial peer list, validates it and sets up the initial state for each of the peers
func (p *validatorStatistics) LoadInitialState(in []*sharding.InitialNode) error {
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

// UpdatePeerState takes the current and previous headers and updates the peer state
//  for all of the consensus members
func (p *validatorStatistics) UpdatePeerState(header, previousHeader data.HeaderHandler) error {
	consensusGroup, err := p.nodesCoordinator.ComputeValidatorsGroup(previousHeader.GetPrevRandSeed(), previousHeader.GetRound(), previousHeader.GetShardID())
	if err != nil {
		return err
	}

	err  = p.updateValidatorInfo(consensusGroup, previousHeader.GetPubKeysBitmap(), previousHeader.GetShardID())
	if err != nil {
		return err
	}

	err = p.checkForMissedBlocks(header, previousHeader)
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

func (p *validatorStatistics) checkForMissedBlocks(header, previousHeader data.HeaderHandler) error {
	if header.GetRound() - previousHeader.GetRound() <= 1 {
		return nil
	}

	for i := previousHeader.GetRound() + 1; i < header.GetRound(); i++ {
		consensusGroup, err := p.nodesCoordinator.ComputeValidatorsGroup(previousHeader.GetPrevRandSeed(), previousHeader.GetRound(), previousHeader.GetShardID())
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
func (p *validatorStatistics) RevertPeerState(header, previousHeader data.HeaderHandler) error {
	consensusGroup, err := p.nodesCoordinator.ComputeValidatorsGroup(previousHeader.GetPrevRandSeed(), previousHeader.GetRound(), previousHeader.GetShardID())
	if err != nil {
		return err
	}

	err  = p.revertValidatorInfo(consensusGroup, previousHeader.GetPubKeysBitmap(), previousHeader.GetShardID())
	if err != nil {
		return err
	}

	if header.GetShardID() == sharding.MetachainShardId {
		err = p.revertShardDataPeerState(header)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *validatorStatistics) updateShardDataPeerState(header data.HeaderHandler) error {
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

		shardConsensus, err := p.nodesCoordinator.ComputeValidatorsGroup(shardHeader.GetPrevRandSeed(), shardHeader.GetRound(), shardHeader.GetShardID())
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

func (p *validatorStatistics) revertShardDataPeerState(header data.HeaderHandler) error {
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

		shardConsensus, err := p.nodesCoordinator.ComputeValidatorsGroup(shardHeader.GetPrevRandSeed(), shardHeader.GetRound(), shardHeader.GetShardID())
		if err != nil {
			return err
		}

		err = p.revertValidatorInfo(shardConsensus, shardHeader.GetPubKeysBitmap(), shardHeader.GetShardID())
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

func (p *validatorStatistics) updateValidatorInfo(validatorList []sharding.Validator, signingBitmap []byte, shardId uint32) error {
	lenValidators := len(validatorList)
	for i := 0; i < lenValidators; i++ {
		peerAcc, err := p.getPeerAccount(validatorList[i].Address())
		if err != nil {
			return err
		}

		isLeader := i == 0
		validatorSigned := (signingBitmap[i/8] & (1 << (uint16(i) % 8))) != 0
		actionType :=  p.computeValidatorActionType(isLeader, validatorSigned)

		switch actionType {
		case leaderSuccess:
			err = peerAcc.IncreaseLeaderSuccessRateWithJournal()
		case leaderFail:
			err = peerAcc.DecreaseLeaderSuccessRateWithJournal()
		case validatorSuccess:
			err = peerAcc.IncreaseValidatorSuccessRateWithJournal()
		case validatorFail:
			err = peerAcc.DecreaseValidatorSuccessRateWithJournal()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *validatorStatistics) revertValidatorInfo(validatorList []sharding.Validator, signingBitmap []byte, shardId uint32) error {
	lenValidators := len(validatorList)
	for i := 0; i < lenValidators; i++ {
		peerAcc, err := p.getPeerAccount(validatorList[i].Address())
		if err != nil {
			return err
		}

		isLeader := i == 0
		validatorSigned := (signingBitmap[i/8] & (1 << (uint16(i) % 8))) != 0
		actionType :=  p.computeValidatorActionType(isLeader, validatorSigned)

		switch actionType {
		case leaderSuccess:
			err = peerAcc.DecreaseLeaderSuccessRateWithJournal()
		case leaderFail:
			err = peerAcc.IncreaseLeaderSuccessRateWithJournal()
		case validatorSuccess:
			err = peerAcc.DecreaseValidatorSuccessRateWithJournal()
		case validatorFail:
			err = peerAcc.IncreaseValidatorSuccessRateWithJournal()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *validatorStatistics) computeValidatorActionType(isLeader, validatorSigned bool) validatorActionType {
	if isLeader && validatorSigned {
		return leaderSuccess
	}
	if isLeader && !validatorSigned {
		return leaderFail
	}
	if !isLeader && validatorSigned {
		return validatorSuccess
	}
	if !isLeader && !validatorSigned {
		return validatorFail
	}

	return unknownAction
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
