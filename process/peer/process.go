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
func (vs *validatorStatistics) saveInitialState(
	in []*sharding.InitialNode,
	stakeValue *big.Int,
	startRating uint32,
) error {
	for _, node := range in {
		err := vs.initializeNode(node, stakeValue, startRating)
		if err != nil {
			return err
		}
	}

	hash, err := vs.peerAdapter.Commit()
	if err != nil {
		return err
	}

	log.Trace("committed peer adapter", "root hash", core.ToHex(hash))

	return nil
}

// IsNodeValid calculates if a node that's present in the initial validator list
//  contains all the required information in order to be able to participate in consensus
func (vs *validatorStatistics) IsNodeValid(node *sharding.InitialNode) bool {
	if len(node.PubKey) == 0 {
		return false
	}
	if len(node.Address) == 0 {
		return false
	}

	return true
}

func (vs *validatorStatistics) processPeerChanges(header data.HeaderHandler) error {
	if vs.shardCoordinator.SelfId() == sharding.MetachainShardId {
		return nil
	}

	metaBlock, ok := header.(*block.MetaBlock)
	if !ok {
		return nil
	}

	for _, peerChange := range metaBlock.PeerInfo {
		err := vs.updatePeerState(peerChange)
		if err != nil {
			return err
		}
	}

	return nil
}

func (vs *validatorStatistics) updatePeerState(
	peerChange block.PeerData,
) error {
	adrSrc, err := vs.adrConv.CreateAddressFromPublicKeyBytes(peerChange.PublicKey)
	if err != nil {
		return err
	}

	if peerChange.Action == block.PeerDeregistration {
		return vs.peerAdapter.RemoveAccount(adrSrc)
	}

	accHandler, err := vs.peerAdapter.GetAccountWithJournal(adrSrc)
	if err != nil {
		return err
	}

	account, ok := accHandler.(*state.PeerAccount)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if !bytes.Equal(peerChange.Address, account.RewardAddress) {
		err := account.SetRewardAddressWithJournal(peerChange.Address)
		if err != nil {
			return err
		}
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

	if peerChange.Action == block.PeerRegistration && peerChange.TimeStamp != account.Nonce {
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
func (vs *validatorStatistics) UpdatePeerState(header data.HeaderHandler) ([]byte, error) {
	if header.GetNonce() == 0 {
		return vs.peerAdapter.RootHash()
	}

	err := vs.processPeerChanges(header)
	if err != nil {
		return nil, err
	}

	consensusGroup, err := vs.nodesCoordinator.ComputeValidatorsGroup(header.GetPrevRandSeed(), header.GetRound(), header.GetShardID())
	if err != nil {
		return nil, err
	}

	err = vs.updateValidatorInfo(consensusGroup)
	if err != nil {
		return nil, err
	}

	// TODO: This should be removed when we have the genesis block in the storage also
	//  and make sure to calculate gaps for the first block also
	if header.GetNonce() == 1 {
		return vs.peerAdapter.RootHash()
	}

	previousHeader, err := process.GetMetaHeader(header.GetPrevHash(), vs.dataPool.Headers(), vs.marshalizer, vs.storageService)
	if err != nil {
		return nil, err
	}

	err = vs.checkForMissedBlocks(
		header.GetRound(),
		previousHeader.GetRound(),
		previousHeader.GetPrevRandSeed(),
		previousHeader.GetShardID(),
	)
	if err != nil {
		return nil, err
	}

	err = vs.updateShardDataPeerState(header, previousHeader)
	if err != nil {
		return nil, err
	}

	vs.displayRatings()

	return vs.peerAdapter.RootHash()
}

func (vs *validatorStatistics) displayRatings() {
	for _, node := range vs.initialNodes {
		log.Trace("ratings", "pk", node.Address, "tempRating", vs.getTempRating(node.PubKey))
	}
}

// Commit commits the validator statistics trie and returns the root hash
func (vs *validatorStatistics) Commit() ([]byte, error) {
	return vs.peerAdapter.Commit()
}

// RootHash returns the root hash of the validator statistics trie
func (vs *validatorStatistics) RootHash() ([]byte, error) {
	return vs.peerAdapter.RootHash()
}

func (vs *validatorStatistics) checkForMissedBlocks(
	currentHeaderRound,
	previousHeaderRound uint64,
	prevRandSeed []byte,
	shardId uint32,
) error {
	if currentHeaderRound-previousHeaderRound <= 1 {
		return nil
	}

	sw := core.NewStopWatch()
	sw.Start("checkForMissedBlocks")
	defer func() {
		sw.Stop("checkForMissedBlocks")
		log.Debug("measurements checkForMissedBlocks", sw.GetMeasurements()...)
	}()

	for i := previousHeaderRound + 1; i < currentHeaderRound; i++ {
		swInner := core.NewStopWatch()

		swInner.Start("ComputeValidatorsGroup")
		consensusGroup, err := vs.nodesCoordinator.ComputeValidatorsGroup(prevRandSeed, i, shardId)
		swInner.Stop("ComputeValidatorsGroup")
		if err != nil {
			return err
		}

		swInner.Start("GetPeerAccount")
		leaderPeerAcc, err := vs.GetPeerAccount(consensusGroup[0].PubKey())
		swInner.Stop("GetPeerAccount")
		if err != nil {
			return err
		}

		swInner.Start("DecreaseLeaderSuccessRateWithJournal")
		err = leaderPeerAcc.DecreaseLeaderSuccessRateWithJournal()
		swInner.Stop("DecreaseLeaderSuccessRateWithJournal")
		if err != nil {
			return err
		}

		swInner.Start("ComputeDecreaseProposer")
		newRating := vs.rater.ComputeDecreaseProposer(leaderPeerAcc.GetTempRating())
		swInner.Stop("ComputeDecreaseProposer")

		swInner.Start("SetTempRatingWithJournal")
		err = leaderPeerAcc.SetTempRatingWithJournal(newRating)
		swInner.Stop("SetTempRatingWithJournal")
		if err != nil {
			return err
		}

		sw.Add(swInner)
	}

	return nil
}

// RevertPeerState takes the current and previous headers and undos the peer state
//  for all of the consensus members
func (vs *validatorStatistics) RevertPeerState(header data.HeaderHandler) error {
	return vs.peerAdapter.RecreateTrie(header.GetValidatorStatsRootHash())
}

// RevertPeerStateToSnapshot reverts the applied changes to the peerAdapter
func (vs *validatorStatistics) RevertPeerStateToSnapshot(snapshot int) error {
	return vs.peerAdapter.RevertToSnapshot(snapshot)
}

func (vs *validatorStatistics) updateShardDataPeerState(header, previousHeader data.HeaderHandler) error {
	metaHeader, ok := header.(*block.MetaBlock)
	if !ok {
		return process.ErrInvalidMetaHeader
	}
	prevMetaHeader, ok := previousHeader.(*block.MetaBlock)
	if !ok {
		return process.ErrInvalidMetaHeader
	}

	err := vs.mediator.loadPreviousShardHeaders(metaHeader, prevMetaHeader)
	if err != nil {
		return err
	}

	for _, h := range metaHeader.ShardInfo {

		shardConsensus, shardInfoErr := vs.nodesCoordinator.ComputeValidatorsGroup(h.PrevRandSeed, h.Round, h.ShardID)
		if shardInfoErr != nil {
			return shardInfoErr
		}

		shardInfoErr = vs.updateValidatorInfo(shardConsensus)
		if shardInfoErr != nil {
			return shardInfoErr
		}

		if h.Nonce == 1 {
			continue
		}

		sdKey := vs.buildShardDataKey(h)
		vs.mutPrevShardInfo.RLock()
		prevShardData, prevDataOk := vs.prevShardInfo[sdKey]
		vs.mutPrevShardInfo.RUnlock()
		if !prevDataOk {
			return process.ErrMissingPrevShardData
		}

		shardInfoErr = vs.checkForMissedBlocks(
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

func (vs *validatorStatistics) initializeNode(node *sharding.InitialNode, stakeValue *big.Int,
	startRating uint32) error {
	if !vs.IsNodeValid(node) {
		return process.ErrInvalidInitialNodesState
	}

	peerAccount, err := vs.generatePeerAccount(node)
	if err != nil {
		return err
	}

	err = vs.savePeerAccountData(peerAccount, node, stakeValue, startRating)
	if err != nil {
		return err
	}

	return nil
}

func (vs *validatorStatistics) generatePeerAccount(node *sharding.InitialNode) (*state.PeerAccount, error) {
	address, err := vs.adrConv.CreateAddressFromHex(node.PubKey)
	if err != nil {
		return nil, err
	}

	acc, err := vs.peerAdapter.GetAccountWithJournal(address)
	if err != nil {
		return nil, err
	}

	peerAccount, ok := acc.(*state.PeerAccount)
	if !ok {
		return nil, process.ErrInvalidPeerAccount
	}

	return peerAccount, nil
}

func (vs *validatorStatistics) savePeerAccountData(
	peerAccount *state.PeerAccount,
	data *sharding.InitialNode,
	stakeValue *big.Int,
	startRating uint32,
) error {
	err := peerAccount.SetRewardAddressWithJournal([]byte(data.Address))
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

func (vs *validatorStatistics) updateValidatorInfo(validatorList []sharding.Validator) error {
	lenValidators := len(validatorList)
	for i := 0; i < lenValidators; i++ {
		peerAcc, err := vs.GetPeerAccount(validatorList[i].PubKey())
		if err != nil {
			return err
		}

		var newRating uint32
		isLeader := i == 0
		if isLeader {
			err = peerAcc.IncreaseLeaderSuccessRateWithJournal()
			newRating = vs.rater.ComputeIncreaseProposer(peerAcc.GetTempRating())
		} else {
			err = peerAcc.IncreaseValidatorSuccessRateWithJournal()
			newRating = vs.rater.ComputeIncreaseValidator(peerAcc.GetTempRating())
		}

		err = peerAcc.SetTempRatingWithJournal(newRating)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetPeerAccount will return a PeerAccountHandler for a given address
func (vs *validatorStatistics) GetPeerAccount(address []byte) (state.PeerAccountHandler, error) {
	addressContainer, err := vs.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	account, err := vs.peerAdapter.GetAccountWithJournal(addressContainer)
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
func (vs *validatorStatistics) loadPreviousShardHeaders(currentHeader, previousHeader *block.MetaBlock) error {

	missingPreviousShardData := vs.loadExistingPrevShardData(currentHeader, previousHeader)
	missingPreviousShardData, err := vs.loadMissingPrevShardDataFromStorage(missingPreviousShardData, previousHeader)
	if err != nil {
		return err
	}

	if len(missingPreviousShardData) > 0 {
		return process.ErrMissingShardDataInStorage
	}

	return nil
}

func (vs *validatorStatistics) loadExistingPrevShardData(currentHeader, previousHeader *block.MetaBlock) map[string]block.ShardData {
	vs.mutPrevShardInfo.Lock()
	defer vs.mutPrevShardInfo.Unlock()

	vs.prevShardInfo = make(map[string]block.ShardData, len(currentHeader.ShardInfo))
	missingPreviousShardData := make(map[string]block.ShardData, len(currentHeader.ShardInfo))

	for _, currentShardData := range currentHeader.ShardInfo {
		if currentShardData.Nonce == 1 {
			continue
		}

		sdKey := vs.buildShardDataKey(currentShardData)
		prevShardData := vs.getMatchingPrevShardData(currentShardData, currentHeader.ShardInfo)
		if prevShardData != nil {
			vs.prevShardInfo[sdKey] = *prevShardData
			continue
		}

		prevShardData = vs.getMatchingPrevShardData(currentShardData, previousHeader.ShardInfo)
		if prevShardData != nil {
			vs.prevShardInfo[sdKey] = *prevShardData
			continue
		}

		missingPreviousShardData[sdKey] = currentShardData
	}

	return missingPreviousShardData
}

func (vs *validatorStatistics) loadMissingPrevShardDataFromStorage(missingPreviousShardData map[string]block.ShardData, previousHeader *block.MetaBlock) (map[string]block.ShardData, error) {
	vs.mutPrevShardInfo.Lock()
	defer vs.mutPrevShardInfo.Unlock()

	searchHeader := &block.MetaBlock{}
	*searchHeader = *previousHeader
	for len(missingPreviousShardData) > 0 {
		if searchHeader.GetNonce() <= 1 {
			break
		}

		recursiveHeader, err := process.GetMetaHeader(searchHeader.GetPrevHash(), vs.dataPool.Headers(), vs.marshalizer, vs.storageService)
		if err != nil {
			return nil, err
		}
		for key, shardData := range missingPreviousShardData {
			prevShardData := vs.getMatchingPrevShardData(shardData, recursiveHeader.ShardInfo)
			if prevShardData == nil {
				continue
			}

			vs.prevShardInfo[key] = *prevShardData
			delete(missingPreviousShardData, key)
		}
		*searchHeader = *recursiveHeader
	}

	return missingPreviousShardData, nil
}

func (vs *validatorStatistics) loadPreviousShardHeadersMeta(header *block.MetaBlock) error {
	vs.mutPrevShardInfo.Lock()
	defer vs.mutPrevShardInfo.Unlock()

	metaDataPool, ok := vs.dataPool.(dataRetriever.MetaPoolsHolder)
	if !ok {
		return process.ErrInvalidMetaPoolHolder
	}

	for _, shardData := range header.ShardInfo {
		if shardData.Nonce == 1 {
			continue
		}

		previousHeader, err := process.GetShardHeader(
			shardData.PrevHash,
			metaDataPool.Headers(),
			vs.marshalizer,
			vs.storageService,
		)
		if err != nil {
			return err
		}

		sdKey := vs.buildShardDataKey(shardData)
		vs.prevShardInfo[sdKey] = block.ShardData{
			ShardID:      previousHeader.ShardId,
			Nonce:        previousHeader.Nonce,
			Round:        previousHeader.Round,
			PrevRandSeed: previousHeader.PrevRandSeed,
			PrevHash:     previousHeader.PrevHash,
		}
	}
	return nil
}

func (vs *validatorStatistics) getMatchingPrevShardData(currentShardData block.ShardData, shardInfo []block.ShardData) *block.ShardData {
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

func (vs *validatorStatistics) buildShardDataKey(sh block.ShardData) string {
	return fmt.Sprintf("%d_%d", sh.ShardID, sh.Nonce)
}

func (vs *validatorStatistics) createMediator() shardMetaMediator {
	if vs.shardCoordinator.SelfId() < sharding.MetachainShardId {
		return &shardMediator{
			vs: vs,
		}
	}
	return &metaMediator{
		vs: vs,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (vs *validatorStatistics) IsInterfaceNil() bool {
	return vs == nil
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
