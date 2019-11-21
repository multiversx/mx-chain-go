package peer

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.DefaultLogger()

// ArgValidatorStatisticsProcessor holds all dependencies for the validatorStatistics
type ArgValidatorStatisticsProcessor struct {
	InitialNodes     []*sharding.InitialNode
	StakeValue       *big.Int
	Marshalizer      marshal.Marshalizer
	NodesCoordinator sharding.NodesCoordinator
	ShardCoordinator sharding.Coordinator
	DataPool         DataPool
	StorageService   dataRetriever.StorageService
	AdrConv          state.AddressConverter
	PeerAdapter      state.AccountsAdapter
	Rater            sharding.Rater
}

type validatorStatistics struct {
	marshalizer      marshal.Marshalizer
	dataPool         DataPool
	storageService   dataRetriever.StorageService
	nodesCoordinator sharding.NodesCoordinator
	shardCoordinator sharding.Coordinator
	adrConv          state.AddressConverter
	peerAdapter      state.AccountsAdapter
	prevShardInfo    map[uint32]block.ShardData
	mutPrevShardInfo sync.RWMutex
	mediator         shardMetaMediator
	rater            sharding.Rater
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
	if arguments.StakeValue == nil {
		return nil, process.ErrNilEconomicsData
	}
	if arguments.Rater == nil {
		return nil, process.ErrNilRater
	}

	rater := arguments.Rater

	ratingReaderSetter, ok := rater.(sharding.RatingReaderSetter)

	vs := &validatorStatistics{
		peerAdapter:      arguments.PeerAdapter,
		adrConv:          arguments.AdrConv,
		nodesCoordinator: arguments.NodesCoordinator,
		shardCoordinator: arguments.ShardCoordinator,
		dataPool:         arguments.DataPool,
		storageService:   arguments.StorageService,
		marshalizer:      arguments.Marshalizer,
		prevShardInfo:    make(map[uint32]block.ShardData, 0),
		rater:            arguments.Rater,
	}
	vs.mediator = vs.createMediator()

	if ok {
		ratingReaderSetter.SetRatingReader(&RatingReader{
			getRating: func(s string) uint32 {
				peer, _ := vs.getPeerAccount([]byte(s))
				return peer.GetRating()
			},
		})
	}

	err := vs.saveInitialState(arguments.InitialNodes, arguments.StakeValue)
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

// UpdatePeerState takes a header, updates the peer state for all of the
//  consensus members and returns the new root hash
func (p *validatorStatistics) UpdatePeerState(header data.HeaderHandler) ([]byte, error) {
	currRh, _ := p.peerAdapter.RootHash()
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

	// TODO: Currently there is a bug that shards build a different meta genesis block.
	//  Instead of returning the root hash, when this bug is fixed we should go on
	//  with calculating missed blocks starting from genesis until
	//  the first block was actually proposed.
	if header.GetNonce() == 1 {
		return p.peerAdapter.RootHash()
	}

	previousHeader, err := process.GetMetaHeader(header.GetPrevHash(), p.dataPool.MetaBlocks(), p.marshalizer, p.storageService)
	if err != nil {
		return nil, err
	}

	err = p.checkForMissedBlocks(
		header.GetRound(),
		previousHeader.GetRound(),
		previousHeader.GetPrevRandSeed(),
		previousHeader.GetShardID(),
	)
	if err != nil {
		return nil, err
	}

	err = p.updateShardDataPeerState(header, previousHeader)
	if err != nil {
		return nil, err
	}

	rh, _ := p.peerAdapter.RootHash()
	log.Info(fmt.Sprintf("UPDATE PEER STATE FOR HEADER %d From %s TO %s", header.GetNonce(), core.ToB64(currRh), core.ToB64(rh)))

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

func (p *validatorStatistics) checkForMissedBlocks(
	currentHeaderRound,
	previousHeaderRound uint64,
	prevRandSeed []byte,
	shardId uint32,
) error {
	if currentHeaderRound-previousHeaderRound <= 1 {
		return nil
	}
	proposerDecreaseOption := p.rater.GetRatingOptionKeys()[2]

	for i := previousHeaderRound + 1; i < currentHeaderRound; i++ {
		consensusGroup, err := p.nodesCoordinator.ComputeValidatorsGroup(prevRandSeed, i, shardId)
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

		newRating := p.rater.ComputeRating(string(consensusGroup[0].Address()), proposerDecreaseOption, leaderPeerAcc.GetRating())

		err = leaderPeerAcc.SetRatingWithJournal(newRating)

		if err != nil {
			return err
		}

	}

	return nil
}

// RevertPeerState takes the current and previous headers and undos the peer state
//  for all of the consensus members
func (p *validatorStatistics) RevertPeerState(header data.HeaderHandler) error {
	return p.peerAdapter.RecreateTrie(header.GetValidatorStatsRootHash())
}

// RevertPeerStateToSnapshot reverts the applied changes to the peerAdapter
func (p *validatorStatistics) RevertPeerStateToSnapshot(snapshot int) error {
	return p.peerAdapter.RevertToSnapshot(snapshot)
}

func (p *validatorStatistics) updateShardDataPeerState(header, previousHeader data.HeaderHandler) error {
	metaHeader, ok := header.(*block.MetaBlock)
	if !ok {
		return process.ErrInvalidMetaHeader
	}
	prevMetaHeader, ok := previousHeader.(*block.MetaBlock)
	if !ok {
		return process.ErrInvalidMetaHeader
	}

	err := p.mediator.loadPreviousShardHeaders(metaHeader, prevMetaHeader)
	if err != nil {
		return err
	}

	for _, h := range metaHeader.ShardInfo {

		shardConsensus, shardInfoErr := p.nodesCoordinator.ComputeValidatorsGroup(h.PrevRandSeed, h.Round, h.ShardID)
		if shardInfoErr != nil {
			return shardInfoErr
		}

		shardInfoErr = p.updateValidatorInfo(shardConsensus, h.ShardID)
		if shardInfoErr != nil {
			return shardInfoErr
		}

		if h.Nonce == 1 {
			continue
		}

		p.mutPrevShardInfo.RLock()
		prevShardData, prevDataOk := p.prevShardInfo[h.ShardID]
		p.mutPrevShardInfo.RUnlock()
		if !prevDataOk {
			return process.ErrMissingPrevShardData
		}

		log.Info(fmt.Sprintf("CALCULATE MISSING BLOCKS SHR: %d, PSHR: %d, PSHRS: %s, PSHID: %d", h.Round, prevShardData.Round, core.ToB64(prevShardData.PrevRandSeed), h.ShardID))
		shardInfoErr = p.checkForMissedBlocks(
			h.Round,
			prevShardData.Round,
			prevShardData.PrevRandSeed,
			h.ShardID,
		)
		if shardInfoErr != nil {
			return shardInfoErr
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
		address := validatorList[i].Address()
		peerAcc, err := p.getPeerAccount(address)
		if err != nil {
			return err
		}

		isLeader := i == 0
		ratingOption := ""
		if isLeader {
			err = peerAcc.IncreaseLeaderSuccessRateWithJournal()
			ratingOption = p.rater.GetRatingOptionKeys()[0]
		} else {
			err = peerAcc.IncreaseValidatorSuccessRateWithJournal()
			ratingOption = p.rater.GetRatingOptionKeys()[1]
		}

		if err != nil {
			return err
		}

		newRating := p.rater.ComputeRating(string(address), ratingOption, peerAcc.GetRating())

		err = peerAcc.SetRatingWithJournal(newRating)

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

// loadPreviousShardHeaders loads the previous shard headers for a given metablock. For the metachain it's easy
//  since it has all the shard headers in its storage, but for the shard it's a bit trickier and we need
//  to iterate through past metachain headers until we find all the ShardData we are interested in
func (p *validatorStatistics) loadPreviousShardHeaders(currentHeader, previousHeader *block.MetaBlock) error {

	p.mutPrevShardInfo.Lock()
	defer p.mutPrevShardInfo.Unlock()

	p.prevShardInfo = make(map[uint32]block.ShardData)
	missingPreviousShardData := make(map[uint32]block.ShardData)

	for _, currentShardData := range currentHeader.ShardInfo {

		if currentShardData.Nonce == 1 {
			continue
		}

		prevShardData := p.getMatchingPrevShardData(currentShardData, currentHeader.ShardInfo)
		if prevShardData != nil {
			p.prevShardInfo[currentShardData.ShardID] = *prevShardData
			continue
		}

		prevShardData = p.getMatchingPrevShardData(currentShardData, previousHeader.ShardInfo)
		if prevShardData != nil {
			p.prevShardInfo[currentShardData.ShardID] = *prevShardData
			continue
		}

		missingPreviousShardData[currentShardData.ShardID] = currentShardData
	}

	searchHeader := &block.MetaBlock{}
	*searchHeader = *previousHeader
	for len(missingPreviousShardData) > 0 {
		if searchHeader.GetNonce() <= 1 {
			break
		}

		recursiveHeader, err := process.GetMetaHeader(searchHeader.GetPrevHash(), p.dataPool.MetaBlocks(), p.marshalizer, p.storageService)
		if err != nil {
			return err
		}
		for shardId, shardData := range missingPreviousShardData {
			prevShardData := p.getMatchingPrevShardData(shardData, recursiveHeader.ShardInfo)
			if prevShardData == nil {
				continue
			}

			p.prevShardInfo[shardId] = *prevShardData
			delete(missingPreviousShardData, shardId)
		}
		*searchHeader = *recursiveHeader
	}

	if len(missingPreviousShardData) > 0 {
		return process.ErrMissingShardDataInStorage
	}

	return nil
}

func (p *validatorStatistics) loadPreviousShardHeadersMeta(header *block.MetaBlock) error {
	p.mutPrevShardInfo.Lock()
	defer p.mutPrevShardInfo.Unlock()

	metaDataPool, ok := p.dataPool.(dataRetriever.MetaPoolsHolder)
	if !ok {
		return process.ErrInvalidMetaPoolHolder
	}

	for _, shardData := range header.ShardInfo {
		if shardData.Nonce == 1 {
			continue
		}

		previousHeader, err := process.GetShardHeader(
			shardData.PrevHash,
			metaDataPool.ShardHeaders(),
			p.marshalizer,
			p.storageService,
		)
		if err != nil {
			return err
		}

		p.prevShardInfo[shardData.ShardID] = block.ShardData{
			ShardID:      previousHeader.ShardId,
			Nonce:        previousHeader.Nonce,
			Round:        previousHeader.Round,
			PrevRandSeed: previousHeader.PrevRandSeed,
			PrevHash:     previousHeader.PrevHash,
		}
	}
	return nil
}

func (p *validatorStatistics) getMatchingPrevShardData(currentShardData block.ShardData, shardInfo []block.ShardData) *block.ShardData {
	for _, prevShardData := range shardInfo {
		if currentShardData.ShardID != prevShardData.ShardID {
			continue
		}
		if currentShardData.Nonce == prevShardData.Nonce+1 {
			return &prevShardData
		}
	}

	return nil
}

func (p *validatorStatistics) createMediator() shardMetaMediator {
	if p.shardCoordinator.SelfId() < sharding.MetachainShardId {
		return &shardMediator{p}
	}
	return &metaMediator{p}
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *validatorStatistics) IsInterfaceNil() bool {
	if p == nil {
		return true
	}
	return false
}
