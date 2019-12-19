package peer

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("process/peer")

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
	Rater            sharding.RaterHandler
}

type validatorStatistics struct {
	marshalizer      marshal.Marshalizer
	dataPool         DataPool
	storageService   dataRetriever.StorageService
	nodesCoordinator sharding.NodesCoordinator
	shardCoordinator sharding.Coordinator
	adrConv          state.AddressConverter
	peerAdapter      state.AccountsAdapter
	prevShardInfo    map[string]block.ShardData
	mutPrevShardInfo sync.RWMutex
	mediator         shardMetaMediator
	rater            sharding.RaterHandler
	initialNodes     []*sharding.InitialNode
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
	if check.IfNil(arguments.Rater) {
		return nil, process.ErrNilRater
	}

	vs := &validatorStatistics{
		peerAdapter:      arguments.PeerAdapter,
		adrConv:          arguments.AdrConv,
		nodesCoordinator: arguments.NodesCoordinator,
		shardCoordinator: arguments.ShardCoordinator,
		dataPool:         arguments.DataPool,
		storageService:   arguments.StorageService,
		marshalizer:      arguments.Marshalizer,
		prevShardInfo:    make(map[string]block.ShardData),
		rater:            arguments.Rater,
	}
	vs.mediator = vs.createMediator()

	rater := arguments.Rater
	ratingReaderSetter, ok := rater.(sharding.RatingReaderSetter)

	if !ok {
		return nil, process.ErrNilRatingReader
	}
	log.Debug("setting ratingReader")

	rr := &RatingReader{
		getRating: vs.getRating,
	}

	ratingReaderSetter.SetRatingReader(rr)

	vs.initialNodes = arguments.InitialNodes

	err := vs.saveInitialState(vs.initialNodes, arguments.StakeValue, rater.GetStartRating())
	if err != nil {
		return nil, err
	}

	return vs, nil
}

// saveInitialState takes an initial peer list, validates it and sets up the initial state for each of the peers
func (p *validatorStatistics) saveInitialState(
	in []*sharding.InitialNode,
	stakeValue *big.Int,
	startRating uint32,
) error {
	for _, node := range in {
		err := p.initializeNode(node, stakeValue, startRating)
		if err != nil {
			return err
		}
	}

	hash, err := p.peerAdapter.Commit()
	if err != nil {
		return err
	}

	log.Trace("committed peer adapter", "root hash", core.ToHex(hash))

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

func (p *validatorStatistics) processPeerChanges(header data.HeaderHandler) error {
	if p.shardCoordinator.SelfId() == sharding.MetachainShardId {
		return nil
	}

	metaBlock, ok := header.(*block.MetaBlock)
	if !ok {
		return nil
	}

	for _, peerChange := range metaBlock.PeerInfo {
		err := p.updatePeerState(peerChange)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *validatorStatistics) updatePeerState(
	peerChange block.PeerData,
) error {
	adrSrc, err := p.adrConv.CreateAddressFromPublicKeyBytes(peerChange.Address)
	if err != nil {
		return err
	}

	if peerChange.Action == block.PeerDeregistration {
		return p.peerAdapter.RemoveAccount(adrSrc)
	}

	accHandler, err := p.peerAdapter.GetAccountWithJournal(adrSrc)
	if err != nil {
		return err
	}

	account, ok := accHandler.(*state.PeerAccount)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if !bytes.Equal(peerChange.PublicKey, account.BLSPublicKey) {
		err := account.SetBLSPublicKeyWithJournal(peerChange.PublicKey)
		if err != nil {
			return err
		}
	}

	zero := big.NewInt(0)
	if peerChange.ValueChange.Cmp(zero) != 0 {
		actualValue := zero.Add(account.Stake, peerChange.ValueChange)
		err := account.SetStakeWithJournal(actualValue)
		if err != nil {
			return err
		}
	}

	if peerChange.Action == block.PeerRegistrantion && peerChange.TimeStamp != account.Nonce {
		err := account.SetNonceWithJournal(peerChange.TimeStamp)
		if err != nil {
			return err
		}

		err = account.SetNodeInWaitingListWithJournal(true)
		if err != nil {
			return err
		}
	}

	if peerChange.Action == block.PeerUnstaking && peerChange.TimeStamp != account.UnStakedNonce {
		err := account.SetUnStakedNonceWithJournal(peerChange.TimeStamp)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdatePeerState takes a header, updates the peer state for all of the
//  consensus members and returns the new root hash
func (p *validatorStatistics) UpdatePeerState(header data.HeaderHandler) ([]byte, error) {
	if header.GetNonce() == 0 {
		return p.peerAdapter.RootHash()
	}

	err := p.processPeerChanges(header)
	if err != nil {
		return nil, err
	}

	consensusGroup, err := p.nodesCoordinator.ComputeValidatorsGroup(header.GetPrevRandSeed(), header.GetRound(), header.GetShardID())
	if err != nil {
		return nil, err
	}

	err = p.updateValidatorInfo(consensusGroup)
	if err != nil {
		return nil, err
	}

	// TODO: This should be removed when we have the genesis block in the storage also
	//  and make sure to calculate gaps for the first block also
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

	p.displayRatings()

	return p.peerAdapter.RootHash()
}

func (p *validatorStatistics) displayRatings() {
	for _, node := range p.initialNodes {
		log.Trace("ratings", "pk", node.Address, "tempRating", p.getTempRating(node.PubKey))
	}
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

	for i := previousHeaderRound + 1; i < currentHeaderRound; i++ {
		consensusGroup, err := p.nodesCoordinator.ComputeValidatorsGroup(prevRandSeed, i, shardId)
		if err != nil {
			return err
		}

		leaderPeerAcc, err := p.GetPeerAccount(consensusGroup[0].PubKey())
		if err != nil {
			return err
		}

		err = leaderPeerAcc.DecreaseLeaderSuccessRateWithJournal()
		if err != nil {
			return err
		}

		newRating := p.rater.ComputeDecreaseProposer(leaderPeerAcc.GetTempRating())
		err = leaderPeerAcc.SetTempRatingWithJournal(newRating)
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

		shardInfoErr = p.updateValidatorInfo(shardConsensus)
		if shardInfoErr != nil {
			return shardInfoErr
		}

		if h.Nonce == 1 {
			continue
		}

		sdKey := p.buildShardDataKey(h)
		p.mutPrevShardInfo.RLock()
		prevShardData, prevDataOk := p.prevShardInfo[sdKey]
		p.mutPrevShardInfo.RUnlock()
		if !prevDataOk {
			return process.ErrMissingPrevShardData
		}

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

func (p *validatorStatistics) initializeNode(node *sharding.InitialNode, stakeValue *big.Int,
	startRating uint32) error {
	if !p.IsNodeValid(node) {
		return process.ErrInvalidInitialNodesState
	}

	peerAccount, err := p.generatePeerAccount(node)
	if err != nil {
		return err
	}

	err = p.savePeerAccountData(peerAccount, node, stakeValue, startRating)
	if err != nil {
		return err
	}

	return nil
}

func (p *validatorStatistics) generatePeerAccount(node *sharding.InitialNode) (*state.PeerAccount, error) {
	address, err := p.adrConv.CreateAddressFromHex(node.PubKey)
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
	startRating uint32,
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

	err = peerAccount.SetRatingWithJournal(startRating)
	if err != nil {
		return err
	}

	err = peerAccount.SetTempRatingWithJournal(startRating)
	if err != nil {
		return err
	}

	return nil
}

func (p *validatorStatistics) updateValidatorInfo(validatorList []sharding.Validator) error {
	lenValidators := len(validatorList)
	for i := 0; i < lenValidators; i++ {
		peerAcc, err := p.GetPeerAccount(validatorList[i].PubKey())
		if err != nil {
			return err
		}

		var newRating uint32
		isLeader := i == 0
		if isLeader {
			err = peerAcc.IncreaseLeaderSuccessRateWithJournal()
			newRating = p.rater.ComputeIncreaseProposer(peerAcc.GetTempRating())
		} else {
			err = peerAcc.IncreaseValidatorSuccessRateWithJournal()
			newRating = p.rater.ComputeIncreaseValidator(peerAcc.GetTempRating())
		}

		err = peerAcc.SetTempRatingWithJournal(newRating)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetPeerAccount will return a PeerAccountHandler for a given address
func (p *validatorStatistics) GetPeerAccount(address []byte) (state.PeerAccountHandler, error) {
	addressContainer, err := p.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	account, err := p.peerAdapter.GetAccountWithJournal(addressContainer)
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

	missingPreviousShardData := p.loadExistingPrevShardData(currentHeader, previousHeader)
	missingPreviousShardData, err := p.loadMissingPrevShardDataFromStorage(missingPreviousShardData, previousHeader)
	if err != nil {
		return err
	}

	if len(missingPreviousShardData) > 0 {
		return process.ErrMissingShardDataInStorage
	}

	return nil
}

func (p *validatorStatistics) loadExistingPrevShardData(currentHeader, previousHeader *block.MetaBlock) map[string]block.ShardData {
	p.mutPrevShardInfo.Lock()
	defer p.mutPrevShardInfo.Unlock()

	p.prevShardInfo = make(map[string]block.ShardData, len(currentHeader.ShardInfo))
	missingPreviousShardData := make(map[string]block.ShardData, len(currentHeader.ShardInfo))

	for _, currentShardData := range currentHeader.ShardInfo {
		if currentShardData.Nonce == 1 {
			continue
		}

		sdKey := p.buildShardDataKey(currentShardData)
		prevShardData := p.getMatchingPrevShardData(currentShardData, currentHeader.ShardInfo)
		if prevShardData != nil {
			p.prevShardInfo[sdKey] = *prevShardData
			continue
		}

		prevShardData = p.getMatchingPrevShardData(currentShardData, previousHeader.ShardInfo)
		if prevShardData != nil {
			p.prevShardInfo[sdKey] = *prevShardData
			continue
		}

		missingPreviousShardData[sdKey] = currentShardData
	}

	return missingPreviousShardData
}

func (p *validatorStatistics) loadMissingPrevShardDataFromStorage(missingPreviousShardData map[string]block.ShardData, previousHeader *block.MetaBlock) (map[string]block.ShardData, error) {
	p.mutPrevShardInfo.Lock()
	defer p.mutPrevShardInfo.Unlock()

	searchHeader := &block.MetaBlock{}
	*searchHeader = *previousHeader
	for len(missingPreviousShardData) > 0 {
		if searchHeader.GetNonce() <= 1 {
			break
		}

		recursiveHeader, err := process.GetMetaHeader(searchHeader.GetPrevHash(), p.dataPool.MetaBlocks(), p.marshalizer, p.storageService)
		if err != nil {
			return nil, err
		}
		for key, shardData := range missingPreviousShardData {
			prevShardData := p.getMatchingPrevShardData(shardData, recursiveHeader.ShardInfo)
			if prevShardData == nil {
				continue
			}

			p.prevShardInfo[key] = *prevShardData
			delete(missingPreviousShardData, key)
		}
		*searchHeader = *recursiveHeader
	}

	return missingPreviousShardData, nil
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

		sdKey := p.buildShardDataKey(shardData)
		p.prevShardInfo[sdKey] = block.ShardData{
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

func (p *validatorStatistics) buildShardDataKey(sh block.ShardData) string {
	return fmt.Sprintf("%d_%d", sh.ShardID, sh.Nonce)
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

func (vs *validatorStatistics) getRating(s string) uint32 {
	peer, err := vs.GetPeerAccount([]byte(s))
	if err != nil {
		log.Debug("Error getting peer account", "error", err)
		return vs.rater.GetStartRating()
	}

	return peer.GetRating()
}

func (vs *validatorStatistics) getTempRating(s string) uint32 {
	peer, err := vs.GetPeerAccount([]byte(s))

	if err != nil {
		log.Debug("Error getting peer account", "error", err)
		return vs.rater.GetStartRating()
	}

	return peer.GetTempRating()
}
