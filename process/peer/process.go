package peer

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"sort"
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

type validatorActionType uint8

const (
	unknownAction    validatorActionType = 0
	leaderSuccess    validatorActionType = 1
	leaderFail       validatorActionType = 2
	validatorSuccess validatorActionType = 3
	validatorFail    validatorActionType = 4
)

// ArgValidatorStatisticsProcessor holds all dependencies for the validatorStatistics
type ArgValidatorStatisticsProcessor struct {
	StakeValue          *big.Int
	Marshalizer         marshal.Marshalizer
	NodesCoordinator    sharding.NodesCoordinator
	ShardCoordinator    sharding.Coordinator
	DataPool            DataPool
	StorageService      dataRetriever.StorageService
	AdrConv             state.AddressConverter
	PeerAdapter         state.AccountsAdapter
	Rater               sharding.RaterHandler
	RewardsHandler      process.RewardsHandler
	MaxComputableRounds uint64
	StartEpoch          uint32
}

type validatorStatistics struct {
	marshalizer             marshal.Marshalizer
	dataPool                DataPool
	storageService          dataRetriever.StorageService
	nodesCoordinator        sharding.NodesCoordinator
	shardCoordinator        sharding.Coordinator
	adrConv                 state.AddressConverter
	peerAdapter             state.AccountsAdapter
	rater                   sharding.RaterHandler
	rewardsHandler          process.RewardsHandler
	maxComputableRounds     uint64
	missedBlocksCounters    validatorRoundCounters
	mutMissedBlocksCounters sync.RWMutex
}

// NewValidatorStatisticsProcessor instantiates a new validatorStatistics structure responsible of keeping account of
//  each validator actions in the consensus process
func NewValidatorStatisticsProcessor(arguments ArgValidatorStatisticsProcessor) (*validatorStatistics, error) {
	if check.IfNil(arguments.PeerAdapter) {
		return nil, process.ErrNilPeerAccountsAdapter
	}
	if check.IfNil(arguments.AdrConv) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(arguments.DataPool) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(arguments.StorageService) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(arguments.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(arguments.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if arguments.StakeValue == nil {
		return nil, process.ErrNilEconomicsData
	}
	if arguments.MaxComputableRounds == 0 {
		return nil, process.ErrZeroMaxComputableRounds
	}
	if check.IfNil(arguments.Rater) {
		return nil, process.ErrNilRater
	}
	if check.IfNil(arguments.RewardsHandler) {
		return nil, process.ErrNilRewardsHandler
	}

	vs := &validatorStatistics{
		peerAdapter:          arguments.PeerAdapter,
		adrConv:              arguments.AdrConv,
		nodesCoordinator:     arguments.NodesCoordinator,
		shardCoordinator:     arguments.ShardCoordinator,
		dataPool:             arguments.DataPool,
		storageService:       arguments.StorageService,
		marshalizer:          arguments.Marshalizer,
		missedBlocksCounters: make(validatorRoundCounters),
		rater:                arguments.Rater,
		rewardsHandler:       arguments.RewardsHandler,
		maxComputableRounds:  arguments.MaxComputableRounds,
	}

	rater := arguments.Rater
	ratingReaderSetter, ok := rater.(sharding.RatingReaderSetter)

	if !ok {
		return nil, process.ErrNilRatingReader
	}
	log.Debug("setting ratingReader")

	rr := &RatingReader{
		getRating:                  vs.getRating,
		updateRatingFromTempRating: vs.updateRatingFromTempRating,
	}

	ratingReaderSetter.SetRatingReader(rr)

	err := vs.saveInitialState(arguments.StakeValue, rater.GetStartRating(), arguments.StartEpoch)
	if err != nil {
		return nil, err
	}

	return vs, nil
}

// saveInitialState takes an initial peer list, validates it and sets up the initial state for each of the peers
func (vs *validatorStatistics) saveInitialState(
	stakeValue *big.Int,
	startRating uint32,
	startEpoch uint32,
) error {
	nodesMap, err := vs.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(startEpoch)
	if err != nil {
		return err
	}

	for _, pks := range nodesMap {
		for _, pk := range pks {
			node, _, err := vs.nodesCoordinator.GetValidatorWithPublicKey(pk, startEpoch)
			if err != nil {
				return err
			}

			err = vs.initializeNode(node, stakeValue, startRating)
			if err != nil {
				return err
			}
		}
	}

	hash, err := vs.peerAdapter.Commit()
	if err != nil {
		return err
	}

	log.Trace("committed peer adapter", "root hash", core.ToHex(hash))

	return nil
}

// UpdatePeerState takes a header, updates the peer state for all of the
//  consensus members and returns the new root hash
func (vs *validatorStatistics) UpdatePeerState(header data.HeaderHandler) ([]byte, error) {
	if header.GetNonce() == 0 {
		return vs.peerAdapter.RootHash()
	}

	vs.mutMissedBlocksCounters.Lock()
	vs.missedBlocksCounters.reset()
	vs.mutMissedBlocksCounters.Unlock()

	// TODO: remove if start of epoch block needs to be validated by the new epoch nodes
	epoch := header.GetEpoch()
	if header.IsStartOfEpochBlock() && epoch > 0 {
		epoch = epoch - 1
	}

	previousHeader, err := process.GetMetaHeader(header.GetPrevHash(), vs.dataPool.Headers(), vs.marshalizer, vs.storageService)
	if err != nil {
		log.Warn("UpdatePeerState could not get meta header from storage", "error", err.Error(), "hash", header.GetPrevHash(), "round", header.GetRound(), "nonce", header.GetNonce())
		return nil, err
	}

	err = vs.checkForMissedBlocks(
		header.GetRound(),
		previousHeader.GetRound(),
		previousHeader.GetPrevRandSeed(),
		previousHeader.GetShardID(),
		epoch,
	)
	if err != nil {
		return nil, err
	}

	err = vs.updateShardDataPeerState(header)
	if err != nil {
		return nil, err
	}

	err = vs.updateMissedBlocksCounters()
	if err != nil {
		return nil, err
	}

	if header.GetNonce() == 1 {
		return vs.peerAdapter.RootHash()
	}

	consensusGroup, err := vs.nodesCoordinator.ComputeConsensusGroup(previousHeader.GetPrevRandSeed(), previousHeader.GetRound(), previousHeader.GetShardID(), epoch)
	if err != nil {
		return nil, err
	}

	err = vs.updateValidatorInfo(consensusGroup, previousHeader.GetPubKeysBitmap(), previousHeader.GetAccumulatedFees())
	if err != nil {
		return nil, err
	}

	vs.displayRatings(header.GetEpoch())
	rootHash, err := vs.peerAdapter.RootHash()
	if err != nil {
		return nil, err
	}

	log.Trace("after updating validator stats", "rootHash", rootHash, "round", header.GetRound(), "selfId", vs.shardCoordinator.SelfId())

	return rootHash, nil
}

func (vs *validatorStatistics) displayRatings(epoch uint32) {
	validatorPKs, err := vs.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("could not get ValidatorPublicKeys", "epoch", epoch)
		return
	}
	log.Trace("started printing tempRatings")
	for shardId, list := range validatorPKs {
		for _, pk := range list {
			log.Trace("tempRating", "PK", pk, "tempRating", vs.getTempRating(string(pk)), "ShardID", shardId)
		}
	}
	log.Trace("finished printing tempRatings")
}

// Commit commits the validator statistics trie and returns the root hash
func (vs *validatorStatistics) Commit() ([]byte, error) {
	return vs.peerAdapter.Commit()
}

// RootHash returns the root hash of the validator statistics trie
func (vs *validatorStatistics) RootHash() ([]byte, error) {
	return vs.peerAdapter.RootHash()
}

func (vs *validatorStatistics) getValidatorDataFromLeaves(
	leaves map[string][]byte,
) (map[uint32][]*state.ValidatorInfo, error) {

	validators := make(map[uint32][]*state.ValidatorInfo, vs.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < vs.shardCoordinator.NumberOfShards(); i++ {
		validators[i] = make([]*state.ValidatorInfo, 0)
	}
	validators[core.MetachainShardId] = make([]*state.ValidatorInfo, 0)

	sliceLeaves := vs.convertMapToSortedSlice(leaves)

	sort.Slice(sliceLeaves, func(i, j int) bool {
		return bytes.Compare(sliceLeaves[i], sliceLeaves[j]) < 0
	})

	for _, pa := range sliceLeaves {
		peerAccount, err := vs.unmarshalPeer(pa)
		if err != nil {
			return nil, err
		}

		currentShardId := peerAccount.CurrentShardId
		validatorInfoData := vs.peerAccountToValidatorInfo(peerAccount)

		validators[currentShardId] = append(validators[currentShardId], validatorInfoData)
	}

	return validators, nil
}

func (vs *validatorStatistics) peerAccountToValidatorInfo(peerAccount *state.PeerAccount) *state.ValidatorInfo {
	return &state.ValidatorInfo{
		PublicKey:                  peerAccount.BLSPublicKey,
		ShardId:                    peerAccount.CurrentShardId,
		List:                       "list",
		Index:                      0,
		TempRating:                 peerAccount.TempRating,
		Rating:                     peerAccount.Rating,
		RewardAddress:              peerAccount.RewardAddress,
		LeaderSuccess:              peerAccount.LeaderSuccessRate.NrSuccess,
		LeaderFailure:              peerAccount.LeaderSuccessRate.NrFailure,
		ValidatorSuccess:           peerAccount.ValidatorSuccessRate.NrSuccess,
		ValidatorFailure:           peerAccount.ValidatorSuccessRate.NrFailure,
		NumSelectedInSuccessBlocks: peerAccount.NumSelectedInSuccessBlocks,
		AccumulatedFees:            big.NewInt(0).Set(peerAccount.AccumulatedFees),
	}
}

func (vs *validatorStatistics) unmarshalPeer(pa []byte) (*state.PeerAccount, error) {
	peerAccount := &state.PeerAccount{}
	err := vs.marshalizer.Unmarshal(peerAccount, pa)
	if err != nil {
		return nil, err
	}
	return peerAccount, nil
}

func (vs *validatorStatistics) convertMapToSortedSlice(leaves map[string][]byte) [][]byte {
	newLeaves := make([][]byte, len(leaves))
	i := 0
	for _, pa := range leaves {
		newLeaves[i] = pa
		i++
	}

	return newLeaves
}

// GetValidatorInfoForRootHash returns all the peer accounts from the trie with the given rootHash
func (vs *validatorStatistics) GetValidatorInfoForRootHash(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
	sw := core.NewStopWatch()
	sw.Start("GetValidatorInfoForRootHash")
	defer func() {
		sw.Stop("GetValidatorInfoForRootHash")
		log.Debug("GetValidatorInfoForRootHash", sw.GetMeasurements()...)
	}()

	allLeaves, err := vs.peerAdapter.GetAllLeaves(rootHash)
	if err != nil {
		return nil, err
	}

	vInfos, err := vs.getValidatorDataFromLeaves(allLeaves)
	if err != nil {
		return nil, err
	}

	return vInfos, err
}

// ResetValidatorStatisticsAtNewEpoch resets the validator info at the start of a new epoch
func (vs *validatorStatistics) ResetValidatorStatisticsAtNewEpoch(vInfos map[uint32][]*state.ValidatorInfo) error {
	sw := core.NewStopWatch()
	sw.Start("ResetValidatorStatisticsAtNewEpoch")
	defer func() {
		sw.Stop("ResetValidatorStatisticsAtNewEpoch")
		log.Debug("ResetValidatorStatisticsAtNewEpoch", sw.GetMeasurements()...)
	}()

	for _, validators := range vInfos {
		for _, validator := range validators {
			addrContainer, err := vs.adrConv.CreateAddressFromPublicKeyBytes(validator.GetPublicKey())
			if err != nil {
				return err
			}
			account, err := vs.peerAdapter.GetAccountWithJournal(addrContainer)
			if err != nil {
				return err
			}

			peerAccount, ok := account.(state.PeerAccountHandler)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			err = peerAccount.ResetAtNewEpoch()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vs *validatorStatistics) checkForMissedBlocks(
	currentHeaderRound,
	previousHeaderRound uint64,
	prevRandSeed []byte,
	shardId uint32,
	epoch uint32,
) error {
	missedRounds := currentHeaderRound - previousHeaderRound
	if missedRounds <= 1 {
		return nil
	}

	tooManyComputations := missedRounds > vs.maxComputableRounds
	if !tooManyComputations {
		return vs.computeDecrease(previousHeaderRound, currentHeaderRound, prevRandSeed, shardId, epoch)
	}

	return vs.decreaseAll(shardId, missedRounds-1, epoch)
}

func (vs *validatorStatistics) computeDecrease(previousHeaderRound uint64, currentHeaderRound uint64, prevRandSeed []byte, shardId uint32, epoch uint32) error {
	sw := core.NewStopWatch()
	sw.Start("checkForMissedBlocks")
	defer func() {
		sw.Stop("checkForMissedBlocks")
		log.Trace("measurements checkForMissedBlocks", sw.GetMeasurements()...)
	}()

	for i := previousHeaderRound + 1; i < currentHeaderRound; i++ {
		swInner := core.NewStopWatch()

		swInner.Start("ComputeValidatorsGroup")
		consensusGroup, err := vs.nodesCoordinator.ComputeConsensusGroup(prevRandSeed, i, shardId, epoch)
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

		vs.mutMissedBlocksCounters.Lock()
		vs.missedBlocksCounters.decreaseLeader(consensusGroup[0].PubKey())
		vs.mutMissedBlocksCounters.Unlock()

		swInner.Start("ComputeDecreaseProposer")
		newRating := vs.rater.ComputeDecreaseProposer(leaderPeerAcc.GetTempRating())
		swInner.Stop("ComputeDecreaseProposer")

		swInner.Start("SetTempRatingWithJournal")
		err = leaderPeerAcc.SetTempRatingWithJournal(newRating)
		swInner.Stop("SetTempRatingWithJournal")
		if err != nil {
			return err
		}

		swInner.Start("ComputeDecreaseAllValidators")
		err = vs.decreaseForConsensusValidators(consensusGroup)
		swInner.Stop("ComputeDecreaseAllValidators")
		if err != nil {
			return err
		}
		sw.Add(swInner)
	}
	return nil
}

func (vs *validatorStatistics) decreaseForConsensusValidators(consensusGroup []sharding.Validator) error {
	vs.mutMissedBlocksCounters.Lock()
	defer vs.mutMissedBlocksCounters.Unlock()

	for j := 1; j < len(consensusGroup); j++ {
		validatorPeerAccount, verr := vs.GetPeerAccount(consensusGroup[j].PubKey())
		if verr != nil {
			return verr
		}

		vs.missedBlocksCounters.decreaseValidator(consensusGroup[j].PubKey())

		newRating := vs.rater.ComputeDecreaseValidator(validatorPeerAccount.GetTempRating())
		verr = validatorPeerAccount.SetTempRatingWithJournal(newRating)
		if verr != nil {
			return verr
		}
	}

	return nil
}

// RevertPeerState takes the current and previous headers and undos the peer state
//  for all of the consensus members
func (vs *validatorStatistics) RevertPeerState(header data.HeaderHandler) error {
	return vs.peerAdapter.RecreateTrie(header.GetValidatorStatsRootHash())
}

func (vs *validatorStatistics) updateShardDataPeerState(header data.HeaderHandler) error {
	metaHeader, ok := header.(*block.MetaBlock)
	if !ok {
		return process.ErrInvalidMetaHeader
	}

	// TODO: remove if start of epoch block needs to be validated by the new epoch nodes
	epoch := header.GetEpoch()
	if header.IsStartOfEpochBlock() && epoch > 0 {
		epoch = epoch - 1
	}

	for _, h := range metaHeader.ShardInfo {
		shardConsensus, shardInfoErr := vs.nodesCoordinator.ComputeConsensusGroup(h.PrevRandSeed, h.Round, h.ShardID, epoch)
		if shardInfoErr != nil {
			return shardInfoErr
		}

		shardInfoErr = vs.updateValidatorInfo(shardConsensus, h.PubKeysBitmap, h.AccumulatedFees)
		if shardInfoErr != nil {
			return shardInfoErr
		}

		if h.Nonce == 1 {
			continue
		}

		prevShardData, shardInfoErr := process.GetShardHeader(
			h.PrevHash,
			vs.dataPool.Headers(),
			vs.marshalizer,
			vs.storageService,
		)
		if shardInfoErr != nil {
			return shardInfoErr
		}

		shardInfoErr = vs.checkForMissedBlocks(
			h.Round,
			prevShardData.Round,
			prevShardData.PrevRandSeed,
			h.ShardID,
			epoch,
		)
		if shardInfoErr != nil {
			return shardInfoErr
		}
	}

	return nil
}

func (vs *validatorStatistics) initializeNode(
	node sharding.Validator,
	stakeValue *big.Int,
	startRating uint32,
) error {
	peerAccount, err := vs.GetPeerAccount(node.PubKey())
	if err != nil {
		return err
	}

	err = vs.savePeerAccountData(peerAccount, node, stakeValue, startRating)
	if err != nil {
		return err
	}

	return nil
}

func (vs *validatorStatistics) savePeerAccountData(
	peerAccount state.PeerAccountHandler,
	data sharding.Validator,
	stakeValue *big.Int,
	startRating uint32,
) error {
	err := peerAccount.SetRewardAddressWithJournal(data.Address())
	if err != nil {
		return err
	}

	err = peerAccount.SetSchnorrPublicKeyWithJournal(data.Address())
	if err != nil {
		return err
	}

	err = peerAccount.SetBLSPublicKeyWithJournal(data.PubKey())
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

func (vs *validatorStatistics) updateValidatorInfo(validatorList []sharding.Validator, signingBitmap []byte, accumulatedFees *big.Int) error {
	lenValidators := len(validatorList)
	for i := 0; i < lenValidators; i++ {
		peerAcc, err := vs.GetPeerAccount(validatorList[i].PubKey())
		if err != nil {
			return err
		}

		err = peerAcc.IncreaseNumSelectedInSuccessBlocks()
		if err != nil {
			return err
		}

		var newRating uint32
		isLeader := i == 0
		validatorSigned := (signingBitmap[i/8] & (1 << (uint16(i) % 8))) != 0
		actionType := vs.computeValidatorActionType(isLeader, validatorSigned)

		switch actionType {
		case leaderSuccess:
			err = peerAcc.IncreaseLeaderSuccessRateWithJournal(1)
			newRating = vs.rater.ComputeIncreaseProposer(peerAcc.GetTempRating())
			if err != nil {
				return err
			}

			leaderAccumulatedFees := core.GetPercentageOfValue(accumulatedFees, vs.rewardsHandler.LeaderPercentage())
			err = peerAcc.AddToAccumulatedFees(leaderAccumulatedFees)
		case validatorSuccess:
			err = peerAcc.IncreaseValidatorSuccessRateWithJournal(1)
			newRating = vs.rater.ComputeIncreaseValidator(peerAcc.GetTempRating())
		case validatorFail:
			err = peerAcc.DecreaseValidatorSuccessRateWithJournal(1)
			newRating = vs.rater.ComputeDecreaseValidator(peerAcc.GetTempRating())
		}

		if err != nil {
			return err
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

func (vs *validatorStatistics) updateMissedBlocksCounters() error {
	vs.mutMissedBlocksCounters.Lock()
	defer func() {
		vs.missedBlocksCounters.reset()
		vs.mutMissedBlocksCounters.Unlock()
	}()

	for pubKey, roundCounters := range vs.missedBlocksCounters {
		peerAccount, err := vs.GetPeerAccount([]byte(pubKey))
		if err != nil {
			return err
		}

		if roundCounters.leaderDecreaseCount > 0 {
			err = peerAccount.DecreaseLeaderSuccessRateWithJournal(roundCounters.leaderDecreaseCount)
			if err != nil {
				return err
			}
		}

		if roundCounters.validatorDecreaseCount > 0 {
			err = peerAccount.DecreaseValidatorSuccessRateWithJournal(roundCounters.validatorDecreaseCount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vs *validatorStatistics) computeValidatorActionType(isLeader, validatorSigned bool) validatorActionType {
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

func (vs *validatorStatistics) updateRatingFromTempRating(pks []string) error {
	rootHash, err := vs.RootHash()
	if err != nil {
		log.Warn("updateRatingFromTempRating getRootHash failed", "error", err)
	}

	log.Trace("updateRatingFromTempRating before", "rootHash", rootHash)
	for _, pk := range pks {
		peer, getAccountErr := vs.GetPeerAccount([]byte(pk))
		if getAccountErr != nil {
			return getAccountErr
		}

		tempRating := vs.getTempRating(pk)
		rating := vs.getRating(pk)
		log.Trace("updateRatingFromTempRating", "pk", []byte(pk), "rating", rating, "tempRating", tempRating)
		err = peer.SetRatingWithJournal(vs.getTempRating(pk))
		if err != nil {
			return err
		}
	}
	rootHash, err = vs.RootHash()
	if err != nil {
		log.Warn("updateRatingFromTempRating.getting root hash failed", "error", err)
	}

	log.Trace("updateRatingFromTempRating after", "rootHash", rootHash)
	return nil
}

func (vs *validatorStatistics) display(validatorKey string) {
	peerAcc, err := vs.GetPeerAccount([]byte(validatorKey))

	if err != nil {
		log.Trace("display peer acc", "error", err)
		return
	}

	acc, ok := peerAcc.(*state.PeerAccount)

	if !ok {
		log.Trace("display", "error", "not a peeracc")
		return
	}

	log.Trace("validator statistics",
		"pk", acc.BLSPublicKey,
		"leader fail", acc.LeaderSuccessRate.NrFailure,
		"leader success", acc.LeaderSuccessRate.NrSuccess,
		"val fail", acc.ValidatorSuccessRate.NrFailure,
		"val success", acc.ValidatorSuccessRate.NrSuccess,
		"temp rating", acc.TempRating,
		"rating", acc.Rating,
	)
}

func (vs *validatorStatistics) decreaseAll(shardId uint32, missedRounds uint64, epoch uint32) error {

	log.Trace("ValidatorStatistics decreasing all", "shardId", shardId, "missedRounds", missedRounds)
	consensusGroupSize := vs.nodesCoordinator.ConsensusGroupSize(shardId)
	validators, err := vs.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		return err
	}
	shardValidators := validators[shardId]
	validatorsCount := len(shardValidators)
	percentageRoundMissedFromTotalValidators := float64(missedRounds) / float64(validatorsCount)
	leaderAppearances := uint32(percentageRoundMissedFromTotalValidators + 1 - math.SmallestNonzeroFloat64)
	consensusGroupAppearances := uint32(float64(consensusGroupSize)*percentageRoundMissedFromTotalValidators +
		1 - math.SmallestNonzeroFloat64)
	ratingDifference := uint32(0)
	for i, validator := range shardValidators {
		validatorPeerAccount, err := vs.GetPeerAccount(validator)
		if err != nil {
			return err
		}
		err = validatorPeerAccount.DecreaseLeaderSuccessRateWithJournal(leaderAppearances)
		if err != nil {
			return err
		}
		err = validatorPeerAccount.DecreaseValidatorSuccessRateWithJournal(consensusGroupAppearances)
		if err != nil {
			return err
		}

		currentTempRating := validatorPeerAccount.GetTempRating()
		for ct := uint32(0); ct < leaderAppearances; ct++ {
			currentTempRating = vs.rater.ComputeDecreaseProposer(currentTempRating)
		}

		for ct := uint32(0); ct < consensusGroupAppearances; ct++ {
			currentTempRating = vs.rater.ComputeDecreaseValidator(currentTempRating)
		}

		if i == 0 {
			ratingDifference = validatorPeerAccount.GetTempRating() - currentTempRating
		}

		err = validatorPeerAccount.SetTempRatingWithJournal(currentTempRating)
		if err != nil {
			return err
		}

		vs.display(string(validator))
	}

	log.Trace(fmt.Sprintf("Decrease leader: %v, decrease validator: %v, ratingDifference: %v", leaderAppearances, consensusGroupAppearances, ratingDifference))

	return nil
}

// Process - processes a validatorInfo and updates fields
func (vsp *validatorStatistics) Process(vid *state.ValidatorInfo) error {
	log.Trace("ValidatorInfoData", "pk", vid.GetPublicKey(), "rating", vid.GetRating(), "tempRating", vid.GetTempRating())

	pa, err := vsp.GetPeerAccount(vid.GetPublicKey())
	if err != nil {
		return err
	}

	err = pa.SetRatingWithJournal(vid.GetRating())
	if err != nil {
		return err
	}

	err = pa.SetTempRatingWithJournal(vid.GetTempRating())
	if err != nil {
		return err
	}

	err = pa.SetRatingWithJournal(vid.GetTempRating())
	if err != nil {
		return err
	}

	return nil
}
