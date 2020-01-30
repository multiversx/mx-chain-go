package peer

import (
	"bytes"
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

type validatorActionType uint8

const (
	unknownAction validatorActionType = 0
	leaderSuccess validatorActionType = 1
	leaderFail validatorActionType = 2
	validatorSuccess validatorActionType = 3
	validatorFail validatorActionType = 4
)

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
		err := vs.updatePeerData(peerChange)
		if err != nil {
			return err
		}
	}

	return nil
}

func (vs *validatorStatistics) updatePeerData(
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
		err = account.SetRewardAddressWithJournal(peerChange.Address)
		if err != nil {
			return err
		}
	}

	// TODO: Since the BLSPublicKey is the key in the peer accounts trie, this condition will never be met.
	//  In order to be able to change the BLSPublicKey of a validator, we need to provide in the peer
	//  change the old and the new public key. Also, important: the peer statistics data and the
	//  rating from the old account needs to be associated to the new account
	if !bytes.Equal(peerChange.PublicKey, account.BLSPublicKey) {
		err = account.SetBLSPublicKeyWithJournal(peerChange.PublicKey)
		if err != nil {
			return err
		}
	}

	zero := big.NewInt(0)
	if peerChange.ValueChange.Cmp(zero) != 0 {
		actualValue := zero.Add(account.Stake, peerChange.ValueChange)
		err = account.SetStakeWithJournal(actualValue)
		if err != nil {
			return err
		}
	}

	if peerChange.Action == block.PeerRegistration && peerChange.TimeStamp != account.Nonce {
		err = account.SetNonceWithJournal(peerChange.TimeStamp)
		if err != nil {
			return err
		}

		err = account.SetNodeInWaitingListWithJournal(true)
		if err != nil {
			return err
		}
	}

	if peerChange.Action == block.PeerUnstaking && peerChange.TimeStamp != account.UnStakedNonce {
		err = account.SetUnStakedNonceWithJournal(peerChange.TimeStamp)
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

	previousHeader, err := process.GetMetaHeader(header.GetPrevHash(), vs.dataPool.Headers(), vs.marshalizer, vs.storageService)
	if err != nil {
		log.Debug("UpdatePeerState after process.GetMetaHeader", "error", err.Error(), "hash", header.GetPrevHash(), "round", header.GetRound(), "nonce", header.GetNonce())
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

	err = vs.updateShardDataPeerState(header)
	if err != nil {
		return nil, err
	}

	if header.GetNonce() == 1 {
		return vs.peerAdapter.RootHash()
	}

	consensusGroup, err := vs.nodesCoordinator.ComputeValidatorsGroup(previousHeader.GetPrevRandSeed(), previousHeader.GetRound(), previousHeader.GetShardID())
	if err != nil {
		return nil, err
	}

	err = vs.updateValidatorInfo(consensusGroup, previousHeader.GetPubKeysBitmap(), previousHeader.GetShardID())
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
		log.Trace("measurements checkForMissedBlocks", sw.GetMeasurements()...)
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
	for j := 1; j < len(consensusGroup); j++ {
		validatorPeerAccount, verr := vs.GetPeerAccount(consensusGroup[j].PubKey())
		if verr != nil {
			return verr
		}

		verr = validatorPeerAccount.DecreaseValidatorSuccessRateWithJournal()
		if verr != nil {
			return verr
		}

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

// RevertPeerStateToSnapshot reverts the applied changes to the peerAdapter
func (vs *validatorStatistics) RevertPeerStateToSnapshot(snapshot int) error {
	return vs.peerAdapter.RevertToSnapshot(snapshot)
}

func (vs *validatorStatistics) updateShardDataPeerState(header data.HeaderHandler) error {
	metaHeader, ok := header.(*block.MetaBlock)
	if !ok {
		return process.ErrInvalidMetaHeader
	}

	for _, h := range metaHeader.ShardInfo {

		shardConsensus, shardInfoErr := vs.nodesCoordinator.ComputeValidatorsGroup(h.PrevRandSeed, h.Round, h.ShardID)
		if shardInfoErr != nil {
			return shardInfoErr
		}

		shardInfoErr = vs.updateValidatorInfo(shardConsensus, h.PubKeysBitmap, h.ShardID)
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

func (vs *validatorStatistics) updateValidatorInfo(validatorList []sharding.Validator, signingBitmap []byte,shardId uint32) error {
	lenValidators := len(validatorList)
	for i := 0; i < lenValidators; i++ {
		peerAcc, err := vs.GetPeerAccount(validatorList[i].PubKey())
		if err != nil {
			return err
		}

		var newRating uint32
		isLeader := i == 0
		validatorSigned := (signingBitmap[i/8] & (1 << (uint16(i) % 8))) != 0
		actionType :=  vs.computeValidatorActionType(isLeader, validatorSigned)

		switch actionType {
		case leaderSuccess:
			err = peerAcc.IncreaseLeaderSuccessRateWithJournal()
			newRating = vs.rater.ComputeIncreaseProposer(peerAcc.GetTempRating())
		case leaderFail:
			err = peerAcc.DecreaseLeaderSuccessRateWithJournal()
			newRating = vs.rater.ComputeDecreaseProposer(peerAcc.GetTempRating())
		case validatorSuccess:
			err = peerAcc.IncreaseValidatorSuccessRateWithJournal()
			newRating = vs.rater.ComputeIncreaseValidator(peerAcc.GetTempRating())
		case validatorFail:
			err = peerAcc.DecreaseValidatorSuccessRateWithJournal()
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
