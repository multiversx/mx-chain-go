package metachain

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
)

// ArgsNewEpochStartSystemSCProcessing defines the arguments structure for the end of epoch system sc processor
type ArgsNewEpochStartSystemSCProcessing struct {
	SystemVM             vmcommon.VMExecutionHandler
	UserAccountsDB       state.AccountsAdapter
	PeerAccountsDB       state.AccountsAdapter
	Marshalizer          marshal.Marshalizer
	StartRating          uint32
	ValidatorInfoCreator epochStart.ValidatorInfoCreator
	ChanceComputer       sharding.ChanceComputer
	ShardCoordinator     sharding.Coordinator

	EndOfEpochCallerAddress []byte
	StakingSCAddress        []byte

	SwitchJailWaitingEnableEpoch           uint32
	SwitchHysteresisForMinNodesEnableEpoch uint32
	DelegationEnableEpoch                  uint32
	StakingV2EnableEpoch                   uint32
	MaxNodesEnableConfig                   []config.MaxNodesChangeConfig

	GenesisNodesConfig  sharding.GenesisNodesSetupHandler
	EpochNotifier       process.EpochNotifier
	NodesConfigProvider epochStart.NodesConfigProvider
	StakingDataProvider epochStart.StakingDataProvider
}

type systemSCProcessor struct {
	systemVM                  vmcommon.VMExecutionHandler
	userAccountsDB            state.AccountsAdapter
	marshalizer               marshal.Marshalizer
	peerAccountsDB            state.AccountsAdapter
	chanceComputer            sharding.ChanceComputer
	shardCoordinator          sharding.Coordinator
	startRating               uint32
	validatorInfoCreator      epochStart.ValidatorInfoCreator
	genesisNodesConfig        sharding.GenesisNodesSetupHandler
	nodesConfigProvider       epochStart.NodesConfigProvider
	stakingDataProvider       epochStart.StakingDataProvider
	endOfEpochCallerAddress   []byte
	stakingSCAddress          []byte
	switchEnableEpoch         uint32
	hystNodesEnableEpoch      uint32
	delegationEnableEpoch     uint32
	stakingV2EnableEpoch      uint32
	maxNodesEnableConfig      []config.MaxNodesChangeConfig
	maxNodes                  uint32
	flagSwitchJailedWaiting   atomic.Flag
	flagHystNodesEnabled      atomic.Flag
	flagDelegationEnabled     atomic.Flag
	flagSetOwnerEnabled       atomic.Flag
	flagChangeMaxNodesEnabled atomic.Flag
	flagStakingV2Enabled      atomic.Flag
	mapNumSwitchedPerShard    map[uint32]uint32
	mapNumSwitchablePerShard  map[uint32]uint32
}

type validatorList []*state.ValidatorInfo

// Len will return the length of the validatorList
func (v validatorList) Len() int { return len(v) }

// Swap will interchange the objects on input indexes
func (v validatorList) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// Less will return true if object on index i should appear before object in index j
// Sorting of validators should be by index and public key
func (v validatorList) Less(i, j int) bool {
	if v[i].TempRating == v[j].TempRating {
		if v[i].Index == v[j].Index {
			return bytes.Compare(v[i].PublicKey, v[j].PublicKey) < 0
		}
		return v[i].Index < v[j].Index
	}
	return v[i].TempRating < v[j].TempRating
}

// NewSystemSCProcessor creates the end of epoch system smart contract processor
func NewSystemSCProcessor(args ArgsNewEpochStartSystemSCProcessing) (*systemSCProcessor, error) {
	if check.IfNilReflect(args.SystemVM) {
		return nil, epochStart.ErrNilSystemVM
	}
	if check.IfNil(args.UserAccountsDB) {
		return nil, epochStart.ErrNilAccountsDB
	}
	if check.IfNil(args.PeerAccountsDB) {
		return nil, epochStart.ErrNilAccountsDB
	}
	if check.IfNil(args.Marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(args.ValidatorInfoCreator) {
		return nil, epochStart.ErrNilValidatorInfoProcessor
	}
	if len(args.EndOfEpochCallerAddress) == 0 {
		return nil, epochStart.ErrNilEndOfEpochCallerAddress
	}
	if len(args.StakingSCAddress) == 0 {
		return nil, epochStart.ErrNilStakingSCAddress
	}
	if check.IfNil(args.ChanceComputer) {
		return nil, epochStart.ErrNilChanceComputer
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, epochStart.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.GenesisNodesConfig) {
		return nil, epochStart.ErrNilGenesisNodesConfig
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, epochStart.ErrNilNodesConfigProvider
	}
	if check.IfNil(args.StakingDataProvider) {
		return nil, epochStart.ErrNilStakingDataProvider
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, epochStart.ErrNilShardCoordinator
	}

	s := &systemSCProcessor{
		systemVM:                 args.SystemVM,
		userAccountsDB:           args.UserAccountsDB,
		peerAccountsDB:           args.PeerAccountsDB,
		marshalizer:              args.Marshalizer,
		startRating:              args.StartRating,
		validatorInfoCreator:     args.ValidatorInfoCreator,
		genesisNodesConfig:       args.GenesisNodesConfig,
		endOfEpochCallerAddress:  args.EndOfEpochCallerAddress,
		stakingSCAddress:         args.StakingSCAddress,
		chanceComputer:           args.ChanceComputer,
		mapNumSwitchedPerShard:   make(map[uint32]uint32),
		mapNumSwitchablePerShard: make(map[uint32]uint32),
		switchEnableEpoch:        args.SwitchJailWaitingEnableEpoch,
		hystNodesEnableEpoch:     args.SwitchHysteresisForMinNodesEnableEpoch,
		delegationEnableEpoch:    args.DelegationEnableEpoch,
		stakingV2EnableEpoch:     args.StakingV2EnableEpoch,
		stakingDataProvider:      args.StakingDataProvider,
		nodesConfigProvider:      args.NodesConfigProvider,
		shardCoordinator:         args.ShardCoordinator,
	}

	s.maxNodesEnableConfig = make([]config.MaxNodesChangeConfig, len(args.MaxNodesEnableConfig))
	copy(s.maxNodesEnableConfig, args.MaxNodesEnableConfig)
	sort.Slice(s.maxNodesEnableConfig, func(i, j int) bool {
		return s.maxNodesEnableConfig[i].EpochEnable < s.maxNodesEnableConfig[j].EpochEnable
	})

	args.EpochNotifier.RegisterNotifyHandler(s)
	return s, nil
}

// ProcessSystemSmartContract does all the processing at end of epoch in case of system smart contract
func (s *systemSCProcessor) ProcessSystemSmartContract(
	validatorInfos map[uint32][]*state.ValidatorInfo,
	nonce uint64,
	epoch uint32,
) error {
	if s.flagHystNodesEnabled.IsSet() {
		err := s.updateSystemSCConfigMinNodes()
		if err != nil {
			return err
		}
	}

	if s.flagSetOwnerEnabled.IsSet() {
		err := s.updateOwnersForBlsKeys()
		if err != nil {
			return err
		}
	}

	if s.flagChangeMaxNodesEnabled.IsSet() {
		err := s.updateMaxNodes(validatorInfos, nonce)
		if err != nil {
			return err
		}
	}

	if s.flagDelegationEnabled.IsSet() {
		err := s.initDelegationSystemSC()
		if err != nil {
			return err
		}
	}

	if s.flagSwitchJailedWaiting.IsSet() {
		err := s.computeNumWaitingPerShard(validatorInfos)
		if err != nil {
			return err
		}

		err = s.swapJailedWithWaiting(validatorInfos)
		if err != nil {
			return err
		}
	}

	if s.flagStakingV2Enabled.IsSet() {
		err := s.prepareRewardsData(validatorInfos)
		if err != nil {
			return err
		}

		err = s.fillStakingDataForNonEligible(validatorInfos)
		if err != nil {
			return err
		}

		numUnStaked, err := s.unStakeNodesWithNotEnoughFunds(validatorInfos, epoch)
		if err != nil {
			return err
		}

		err = s.stakeNodesFromQueue(validatorInfos, numUnStaked, nonce)
		if err != nil {
			return err
		}
	}

	return nil
}

// ToggleUnStakeUnBond will pause/unPause the unStake/unBond functions on the validator system sc
func (s *systemSCProcessor) ToggleUnStakeUnBond(value bool) error {
	if !s.flagStakingV2Enabled.IsSet() {
		return nil
	}

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: s.endOfEpochCallerAddress,
			Arguments:  nil,
			CallValue:  big.NewInt(0),
		},
		RecipientAddr: vm.ValidatorSCAddress,
		Function:      "unPauseUnStakeUnBond",
	}

	if value {
		vmInput.Function = "pauseUnStakeUnBond"
	}

	vmOutput, err := s.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return epochStart.ErrSystemValidatorSCCall
	}

	err = s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return err
	}

	return nil
}

func (s *systemSCProcessor) unStakeNodesWithNotEnoughFunds(
	validatorInfos map[uint32][]*state.ValidatorInfo,
	epoch uint32,
) (uint32, error) {
	nodesToUnStake, mapOwnersKeys, err := s.stakingDataProvider.ComputeUnQualifiedNodes(validatorInfos)
	if err != nil {
		return 0, err
	}

	log.Debug("unStake nodes with not enough funds", "num", len(nodesToUnStake))
	for _, blsKey := range nodesToUnStake {
		err = s.unStakeOneNode(blsKey, epoch)
		if err != nil {
			return 0, err
		}

		validatorInfo := getValidatorInfoWithBLSKey(validatorInfos, blsKey)
		if validatorInfo == nil {
			return 0, epochStart.ErrNilValidatorInfo
		}

		validatorInfo.List = string(core.LeavingList)
	}

	err = s.updateDelegationContracts(mapOwnersKeys)
	if err != nil {
		return 0, err
	}

	return uint32(len(nodesToUnStake)), nil
}

func (s *systemSCProcessor) unStakeOneNode(blsKey []byte, epoch uint32) error {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: s.endOfEpochCallerAddress,
			Arguments:  [][]byte{blsKey},
			CallValue:  big.NewInt(0),
		},
		RecipientAddr: s.stakingSCAddress,
		Function:      "unStakeAtEndOfEpoch",
	}

	vmOutput, err := s.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		log.Debug("unStakeOneNode", "returnMessage", vmOutput.ReturnMessage, "returnCode", vmOutput.ReturnCode.String())
		return epochStart.ErrUnStakeExecuteError
	}

	err = s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return err
	}

	peerAccount, errGet := s.getPeerAccount(blsKey)
	if errGet != nil {
		return errGet
	}

	peerAccount.SetListAndIndex(peerAccount.GetShardId(), string(core.LeavingList), peerAccount.GetIndexInList())
	peerAccount.SetUnStakedEpoch(epoch)
	err = s.peerAccountsDB.SaveAccount(peerAccount)
	if err != nil {
		return err
	}

	return nil
}

func (s *systemSCProcessor) updateDelegationContracts(mapOwnerKeys map[string][][]byte) error {
	sortedDelegationsSCs := make([]string, 0, len(mapOwnerKeys))
	for address := range mapOwnerKeys {
		if s.shardCoordinator.ComputeId([]byte(address)) != core.MetachainShardId {
			continue
		}
		sortedDelegationsSCs = append(sortedDelegationsSCs, address)
	}

	sort.Slice(sortedDelegationsSCs, func(i, j int) bool {
		return sortedDelegationsSCs[i] < sortedDelegationsSCs[j]
	})

	for _, address := range sortedDelegationsSCs {
		vmInput := &vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallerAddr: s.endOfEpochCallerAddress,
				Arguments:  mapOwnerKeys[address],
				CallValue:  big.NewInt(0),
			},
			RecipientAddr: []byte(address),
			Function:      "unStakeAtEndOfEpoch",
		}

		vmOutput, err := s.systemVM.RunSmartContractCall(vmInput)
		if err != nil {
			return err
		}
		if vmOutput.ReturnCode != vmcommon.Ok {
			log.Debug("unStakeAtEndOfEpoch", "returnMessage", vmOutput.ReturnMessage, "returnCode", vmOutput.ReturnCode.String())
			return epochStart.ErrUnStakeExecuteError
		}

		err = s.processSCOutputAccounts(vmOutput)
		if err != nil {
			return err
		}
	}

	return nil
}

func getValidatorInfoWithBLSKey(validatorInfos map[uint32][]*state.ValidatorInfo, blsKey []byte) *state.ValidatorInfo {
	for _, validatorsInfoSlice := range validatorInfos {
		for _, validatorInfo := range validatorsInfoSlice {
			if bytes.Equal(validatorInfo.PublicKey, blsKey) {
				return validatorInfo
			}
		}
	}
	return nil
}

func (s *systemSCProcessor) fillStakingDataForNonEligible(validatorInfos map[uint32][]*state.ValidatorInfo) error {
	for _, validatorsInfoSlice := range validatorInfos {
		for _, validatorInfo := range validatorsInfoSlice {
			if validatorInfo.List == string(core.EligibleList) {
				continue
			}

			err := s.stakingDataProvider.FillValidatorInfo(validatorInfo.PublicKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *systemSCProcessor) prepareRewardsData(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) error {
	eligibleNodesKeys := s.getEligibleNodesKeyMapOfType(validatorsInfo)
	err := s.prepareStakingDataForRewards(eligibleNodesKeys)
	if err != nil {
		return err
	}

	return nil
}

func (s *systemSCProcessor) prepareStakingDataForRewards(eligibleNodesKeys map[uint32][][]byte) error {
	sw := core.NewStopWatch()
	sw.Start("prepareStakingDataForRewards")
	defer func() {
		sw.Stop("prepareStakingDataForRewards")
		log.Debug("systemSCProcessor.prepareStakingDataForRewards time measurements", sw.GetMeasurements()...)
	}()

	return s.stakingDataProvider.PrepareStakingDataForRewards(eligibleNodesKeys)
}

func (s *systemSCProcessor) getEligibleNodesKeyMapOfType(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) map[uint32][][]byte {
	eligibleNodesKeys := make(map[uint32][][]byte)
	for shardID, validatorsInfoSlice := range validatorsInfo {
		eligibleNodesKeys[shardID] = make([][]byte, 0, s.nodesConfigProvider.ConsensusGroupSize(shardID))
		for _, validatorInfo := range validatorsInfoSlice {
			if validatorInfo.List == string(core.EligibleList) {
				eligibleNodesKeys[shardID] = append(eligibleNodesKeys[shardID], validatorInfo.PublicKey)
			}
		}
	}

	return eligibleNodesKeys
}

func getRewardsMiniBlockForMeta(miniBlocks block.MiniBlockSlice) *block.MiniBlock {
	for _, miniBlock := range miniBlocks {
		if miniBlock.Type != block.RewardsBlock {
			continue
		}
		if miniBlock.ReceiverShardID != core.MetachainShardId {
			continue
		}
		return miniBlock
	}
	return nil
}

// ProcessDelegationRewards will process the rewards which are directed towards the delegation system smart contracts
func (s *systemSCProcessor) ProcessDelegationRewards(
	miniBlocks block.MiniBlockSlice,
	txCache epochStart.TransactionCacher,
) error {
	if txCache == nil {
		return epochStart.ErrNilLocalTxCache
	}

	rwdMb := getRewardsMiniBlockForMeta(miniBlocks)
	if rwdMb == nil {
		return nil
	}

	for _, txHash := range rwdMb.TxHashes {
		rwdTx, err := txCache.GetTx(txHash)
		if err != nil {
			return err
		}

		err = s.executeRewardTx(rwdTx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *systemSCProcessor) executeRewardTx(rwdTx data.TransactionHandler) error {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: s.endOfEpochCallerAddress,
			Arguments:  nil,
			CallValue:  rwdTx.GetValue(),
		},
		RecipientAddr: rwdTx.GetRcvAddr(),
		Function:      "updateRewards",
	}

	vmOutput, err := s.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return epochStart.ErrSystemDelegationCall
	}

	err = s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return err
	}

	return nil
}

// updates the configuration of the system SC if the flags permit
func (s *systemSCProcessor) updateSystemSCConfigMinNodes() error {
	minNumberOfNodesWithHysteresis := s.genesisNodesConfig.MinNumberOfNodesWithHysteresis()
	err := s.setMinNumberOfNodes(minNumberOfNodesWithHysteresis)

	return err
}

// updates the configuration of the system SC if the flags permit
func (s *systemSCProcessor) updateMaxNodes(validatorInfos map[uint32][]*state.ValidatorInfo, nonce uint64) error {
	sw := core.NewStopWatch()
	sw.Start("total")
	defer func() {
		sw.Stop("total")
		log.Info("systemSCProcessor.updateMaxNodes", sw.GetMeasurements()...)
	}()

	maxNumberOfNodes := s.maxNodes
	sw.Start("setMaxNumberOfNodes")
	prevMaxNumberOfNodes, err := s.setMaxNumberOfNodes(maxNumberOfNodes)
	sw.Stop("setMaxNumberOfNodes")
	if err != nil {
		return err
	}

	if maxNumberOfNodes < prevMaxNumberOfNodes {
		return epochStart.ErrInvalidMaxNumberOfNodes
	}

	sw.Start("stakeNodesFromQueue")
	err = s.stakeNodesFromQueue(validatorInfos, maxNumberOfNodes-prevMaxNumberOfNodes, nonce)
	sw.Stop("stakeNodesFromQueue")
	if err != nil {
		return err
	}
	return nil
}

func (s *systemSCProcessor) computeNumWaitingPerShard(validatorInfos map[uint32][]*state.ValidatorInfo) error {
	for shardID, validatorInfoList := range validatorInfos {
		totalInWaiting := uint32(0)
		for _, validatorInfo := range validatorInfoList {
			switch validatorInfo.List {
			case string(core.WaitingList):
				totalInWaiting++
			}
		}
		s.mapNumSwitchablePerShard[shardID] = totalInWaiting
		s.mapNumSwitchedPerShard[shardID] = 0
	}
	return nil
}

func (s *systemSCProcessor) swapJailedWithWaiting(validatorInfos map[uint32][]*state.ValidatorInfo) error {
	jailedValidators := s.getSortedJailedNodes(validatorInfos)

	log.Debug("number of jailed validators", "num", len(jailedValidators))

	newValidators := make(map[string]struct{})
	for _, jailedValidator := range jailedValidators {
		if _, ok := newValidators[string(jailedValidator.PublicKey)]; ok {
			continue
		}
		if isValidator(jailedValidator) && s.mapNumSwitchablePerShard[jailedValidator.ShardId] <= s.mapNumSwitchedPerShard[jailedValidator.ShardId] {
			log.Debug("cannot switch in this epoch anymore for this shard as switched num waiting",
				"shardID", jailedValidator.ShardId,
				"numSwitched", s.mapNumSwitchedPerShard[jailedValidator.ShardId])
			continue
		}

		vmInput := &vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallerAddr: s.endOfEpochCallerAddress,
				Arguments:  [][]byte{jailedValidator.PublicKey},
				CallValue:  big.NewInt(0),
			},
			RecipientAddr: s.stakingSCAddress,
			Function:      "switchJailedWithWaiting",
		}

		vmOutput, err := s.systemVM.RunSmartContractCall(vmInput)
		if err != nil {
			return err
		}

		log.Debug("switchJailedWithWaiting called for",
			"key", jailedValidator.PublicKey,
			"returnMessage", vmOutput.ReturnMessage)
		if vmOutput.ReturnCode != vmcommon.Ok {
			continue
		}

		newValidator, err := s.stakingToValidatorStatistics(validatorInfos, jailedValidator, vmOutput)
		if err != nil {
			return err
		}

		if len(newValidator) != 0 {
			newValidators[string(newValidator)] = struct{}{}
		}
	}

	return nil
}

func (s *systemSCProcessor) stakingToValidatorStatistics(
	validatorInfos map[uint32][]*state.ValidatorInfo,
	jailedValidator *state.ValidatorInfo,
	vmOutput *vmcommon.VMOutput,
) ([]byte, error) {
	stakingSCOutput, ok := vmOutput.OutputAccounts[string(s.stakingSCAddress)]
	if !ok {
		return nil, epochStart.ErrStakingSCOutputAccountNotFound
	}

	var activeStorageUpdate *vmcommon.StorageUpdate
	for _, storageUpdate := range stakingSCOutput.StorageUpdates {
		isNewValidatorKey := len(storageUpdate.Offset) == len(jailedValidator.PublicKey) &&
			!bytes.Equal(storageUpdate.Offset, jailedValidator.PublicKey)
		if isNewValidatorKey {
			activeStorageUpdate = storageUpdate
			break
		}
	}
	if activeStorageUpdate == nil {
		log.Debug("no one in waiting suitable for switch")
		return nil, nil
	}

	err := s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return nil, err
	}

	var stakingData systemSmartContracts.StakedDataV2_0
	err = s.marshalizer.Unmarshal(&stakingData, activeStorageUpdate.Data)
	if err != nil {
		return nil, err
	}

	blsPubKey := activeStorageUpdate.Offset
	log.Debug("staking validator key who switches with the jailed one", "blsKey", blsPubKey)
	account, err := s.getPeerAccount(blsPubKey)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(account.GetRewardAddress(), stakingData.RewardAddress) {
		err = account.SetRewardAddress(stakingData.RewardAddress)
		if err != nil {
			return nil, err
		}
	}

	if !bytes.Equal(account.GetBLSPublicKey(), blsPubKey) {
		err = account.SetBLSPublicKey(blsPubKey)
		if err != nil {
			return nil, err
		}
	} else {
		// old jailed validator getting switched back after unJail with stake - must remove first from exported map
		deleteNewValidatorIfExistsFromMap(validatorInfos, blsPubKey, account.GetShardId())
	}

	account.SetListAndIndex(jailedValidator.ShardId, string(core.NewList), uint32(stakingData.StakedNonce))
	account.SetTempRating(s.startRating)
	account.SetUnStakedEpoch(core.DefaultUnstakedEpoch)

	err = s.peerAccountsDB.SaveAccount(account)
	if err != nil {
		return nil, err
	}

	jailedAccount, err := s.getPeerAccount(jailedValidator.PublicKey)
	if err != nil {
		return nil, err
	}

	jailedAccount.SetListAndIndex(jailedValidator.ShardId, string(core.JailedList), jailedValidator.Index)
	jailedAccount.ResetAtNewEpoch()
	err = s.peerAccountsDB.SaveAccount(jailedAccount)
	if err != nil {
		return nil, err
	}

	if isValidator(jailedValidator) {
		s.mapNumSwitchedPerShard[jailedValidator.ShardId]++
	}

	newValidatorInfo := s.validatorInfoCreator.PeerAccountToValidatorInfo(account)
	switchJailedWithNewValidatorInMap(validatorInfos, jailedValidator, newValidatorInfo)

	return blsPubKey, nil
}

func isValidator(validator *state.ValidatorInfo) bool {
	return validator.List == string(core.WaitingList) || validator.List == string(core.EligibleList)
}

func deleteNewValidatorIfExistsFromMap(
	validatorInfos map[uint32][]*state.ValidatorInfo,
	blsPubKey []byte,
	shardID uint32,
) {
	for index, validatorInfo := range validatorInfos[shardID] {
		if bytes.Equal(validatorInfo.PublicKey, blsPubKey) {
			length := len(validatorInfos[shardID])
			validatorInfos[shardID][index] = validatorInfos[shardID][length-1]
			validatorInfos[shardID][length-1] = nil
			validatorInfos[shardID] = validatorInfos[shardID][:length-1]
			break
		}
	}
}

func switchJailedWithNewValidatorInMap(
	validatorInfos map[uint32][]*state.ValidatorInfo,
	jailedValidator *state.ValidatorInfo,
	newValidator *state.ValidatorInfo,
) {
	for index, validatorInfo := range validatorInfos[jailedValidator.ShardId] {
		if bytes.Equal(validatorInfo.PublicKey, jailedValidator.PublicKey) {
			validatorInfos[jailedValidator.ShardId][index] = newValidator
			break
		}
	}
}

func (s *systemSCProcessor) getUserAccount(address []byte) (state.UserAccountHandler, error) {
	acnt, err := s.userAccountsDB.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	stAcc, ok := acnt.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return stAcc, nil
}

// save account changes in state from vmOutput - protected by VM - every output can be treated as is.
func (s *systemSCProcessor) processSCOutputAccounts(
	vmOutput *vmcommon.VMOutput,
) error {

	outputAccounts := process.SortVMOutputInsideData(vmOutput)
	for _, outAcc := range outputAccounts {
		acc, err := s.getUserAccount(outAcc.Address)
		if err != nil {
			return err
		}

		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		for _, storeUpdate := range storageUpdates {
			err = acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
			if err != nil {
				return err
			}
		}

		if outAcc.BalanceDelta != nil && outAcc.BalanceDelta.Cmp(zero) != 0 {
			err = acc.AddToBalance(outAcc.BalanceDelta)
			if err != nil {
				return err
			}
		}

		err = s.userAccountsDB.SaveAccount(acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *systemSCProcessor) getSortedJailedNodes(validatorInfos map[uint32][]*state.ValidatorInfo) []*state.ValidatorInfo {
	newJailedValidators := make([]*state.ValidatorInfo, 0)
	oldJailedValidators := make([]*state.ValidatorInfo, 0)

	minChance := s.chanceComputer.GetChance(0)
	for _, listValidators := range validatorInfos {
		for _, validatorInfo := range listValidators {
			if validatorInfo.List == string(core.JailedList) {
				oldJailedValidators = append(oldJailedValidators, validatorInfo)
			} else if s.chanceComputer.GetChance(validatorInfo.TempRating) < minChance {
				newJailedValidators = append(newJailedValidators, validatorInfo)
			}
		}
	}

	sort.Sort(validatorList(oldJailedValidators))
	sort.Sort(validatorList(newJailedValidators))

	return append(oldJailedValidators, newJailedValidators...)
}

func (s *systemSCProcessor) getPeerAccount(key []byte) (state.PeerAccountHandler, error) {
	account, err := s.peerAccountsDB.LoadAccount(key)
	if err != nil {
		return nil, err
	}

	peerAcc, ok := account.(state.PeerAccountHandler)
	if !ok {
		return nil, epochStart.ErrWrongTypeAssertion
	}

	return peerAcc, nil
}

func (s *systemSCProcessor) setMinNumberOfNodes(minNumNodes uint32) error {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: s.endOfEpochCallerAddress,
			Arguments:  [][]byte{big.NewInt(int64(minNumNodes)).Bytes()},
			CallValue:  big.NewInt(0),
		},
		RecipientAddr: s.stakingSCAddress,
		Function:      "updateConfigMinNodes",
	}

	vmOutput, err := s.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return err
	}

	log.Debug("setMinNumberOfNodes called with",
		"minNumNodes", minNumNodes,
		"returnMessage", vmOutput.ReturnMessage)

	if vmOutput.ReturnCode != vmcommon.Ok {
		return epochStart.ErrInvalidMinNumberOfNodes
	}

	err = s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return err
	}

	return nil
}

func (s *systemSCProcessor) setMaxNumberOfNodes(maxNumNodes uint32) (uint32, error) {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: s.endOfEpochCallerAddress,
			Arguments:  [][]byte{big.NewInt(int64(maxNumNodes)).Bytes()},
			CallValue:  big.NewInt(0),
		},
		RecipientAddr: s.stakingSCAddress,
		Function:      "updateConfigMaxNodes",
	}

	vmOutput, err := s.systemVM.RunSmartContractCall(vmInput)
	if err != nil {
		return 0, err
	}

	log.Debug("setMaxNumberOfNodes called with",
		"maxNumNodes", maxNumNodes,
		"returnMessage", vmOutput.ReturnMessage)

	if vmOutput.ReturnCode != vmcommon.Ok {
		return 0, epochStart.ErrInvalidMaxNumberOfNodes
	}
	if len(vmOutput.ReturnData) != 1 {
		return 0, epochStart.ErrInvalidSystemSCReturn
	}

	err = s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return 0, err
	}

	prevMaxNumNodes := big.NewInt(0).SetBytes(vmOutput.ReturnData[0]).Uint64()
	return uint32(prevMaxNumNodes), nil
}

func (s *systemSCProcessor) updateOwnersForBlsKeys() error {
	sw := core.NewStopWatch()
	sw.Start("systemSCProcessor")
	defer func() {
		sw.Stop("systemSCProcessor")
		log.Info("systemSCProcessor.updateOwnersForBlsKeys time measurements", sw.GetMeasurements()...)
	}()

	sw.Start("getValidatorSystemAccount")
	userValidatorAccount, err := s.getValidatorSystemAccount()
	sw.Stop("getValidatorSystemAccount")
	if err != nil {
		return err
	}

	sw.Start("getArgumentsForSetOwnerFunctionality")
	arguments, err := s.getArgumentsForSetOwnerFunctionality(userValidatorAccount)
	sw.Stop("getArgumentsForSetOwnerFunctionality")
	if err != nil {
		return err
	}

	sw.Start("callSetOwnersOnAddresses")
	err = s.callSetOwnersOnAddresses(arguments)
	sw.Stop("callSetOwnersOnAddresses")
	if err != nil {
		return err
	}

	return nil
}

func (s *systemSCProcessor) getValidatorSystemAccount() (state.UserAccountHandler, error) {
	validatorAccount, err := s.userAccountsDB.LoadAccount(vm.ValidatorSCAddress)
	if err != nil {
		return nil, fmt.Errorf("%w when loading validator account", err)
	}

	userValidatorAccount, ok := validatorAccount.(state.UserAccountHandler)
	if !ok {
		return nil, fmt.Errorf("%w when loading validator account", epochStart.ErrWrongTypeAssertion)
	}

	if check.IfNil(userValidatorAccount.DataTrie()) {
		return nil, epochStart.ErrNilDataTrie
	}

	return userValidatorAccount, nil
}

func (s *systemSCProcessor) getArgumentsForSetOwnerFunctionality(userValidatorAccount state.UserAccountHandler) ([][]byte, error) {
	arguments := make([][]byte, 0)

	rootHash, err := userValidatorAccount.DataTrie().Root()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	chLeaves, err := userValidatorAccount.DataTrie().GetAllLeavesOnChannel(rootHash, ctx)
	if err != nil {
		return nil, err
	}
	for leaf := range chLeaves {
		validatorData := &systemSmartContracts.ValidatorDataV2{}
		value, errTrim := leaf.ValueWithoutSuffix(append(leaf.Key(), vm.ValidatorSCAddress...))
		if errTrim != nil {
			return nil, fmt.Errorf("%w for validator key %s", errTrim, hex.EncodeToString(leaf.Key()))
		}

		err = s.marshalizer.Unmarshal(validatorData, value)
		if err != nil {
			continue
		}
		for _, blsKey := range validatorData.BlsPubKeys {
			arguments = append(arguments, blsKey)
			arguments = append(arguments, leaf.Key())
		}
	}

	return arguments, nil
}

func (s *systemSCProcessor) callSetOwnersOnAddresses(arguments [][]byte) error {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.EndOfEpochAddress,
			CallValue:  big.NewInt(0),
			Arguments:  arguments,
		},
		RecipientAddr: vm.StakingSCAddress,
		Function:      "setOwnersOnAddresses",
	}

	vmOutput, errRun := s.systemVM.RunSmartContractCall(vmInput)
	if errRun != nil {
		return fmt.Errorf("%w when calling setOwnersOnAddresses function", errRun)
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return fmt.Errorf("got return code %s when calling setOwnersOnAddresses", vmOutput.ReturnCode)
	}

	return s.processSCOutputAccounts(vmOutput)
}

func (s *systemSCProcessor) initDelegationSystemSC() error {
	codeMetaData := &vmcommon.CodeMetadata{
		Upgradeable: false,
		Payable:     false,
		Readable:    true,
	}

	vmInput := &vmcommon.ContractCreateInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.DelegationManagerSCAddress,
			Arguments:  [][]byte{},
			CallValue:  big.NewInt(0),
		},
		ContractCode:         vm.DelegationManagerSCAddress,
		ContractCodeMetadata: codeMetaData.ToBytes(),
	}

	vmOutput, err := s.systemVM.RunSmartContractCreate(vmInput)
	if err != nil {
		return err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return epochStart.ErrCouldNotInitDelegationSystemSC
	}

	err = s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return err
	}

	err = s.updateSystemSCContractsCode(vmInput.ContractCodeMetadata)
	if err != nil {
		return err
	}

	return nil
}

func (s *systemSCProcessor) updateSystemSCContractsCode(contractMetadata []byte) error {
	contractsToUpdate := make([][]byte, 0)
	contractsToUpdate = append(contractsToUpdate, vm.StakingSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.ValidatorSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.GovernanceSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.ESDTSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.DelegationManagerSCAddress)
	contractsToUpdate = append(contractsToUpdate, vm.FirstDelegationSCAddress)

	for _, address := range contractsToUpdate {
		userAcc, err := s.getUserAccount(address)
		if err != nil {
			return err
		}

		userAcc.SetOwnerAddress(address)
		userAcc.SetCodeMetadata(contractMetadata)
		userAcc.SetCode(address)

		err = s.userAccountsDB.SaveAccount(userAcc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *systemSCProcessor) stakeNodesFromQueue(
	validatorInfos map[uint32][]*state.ValidatorInfo,
	nodesToStake uint32,
	nonce uint64,
) error {
	if nodesToStake == 0 {
		return nil
	}

	nodesToStakeAsBigInt := big.NewInt(0).SetUint64(uint64(nodesToStake))
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.EndOfEpochAddress,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{nodesToStakeAsBigInt.Bytes()},
		},
		RecipientAddr: vm.StakingSCAddress,
		Function:      "stakeNodesFromQueue",
	}
	vmOutput, errRun := s.systemVM.RunSmartContractCall(vmInput)
	if errRun != nil {
		return fmt.Errorf("%w when staking nodes from waiting list", errRun)
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return fmt.Errorf("got return code %s when staking nodes from waiting list", vmOutput.ReturnCode)
	}
	if len(vmOutput.ReturnData)%2 != 0 {
		return fmt.Errorf("%w return data must be divisible by 2 when staking nodes from waiting list", epochStart.ErrInvalidSystemSCReturn)
	}

	err := s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return err
	}

	err = s.addNewlyStakedNodesToValidatorTrie(validatorInfos, vmOutput.ReturnData, nonce)
	if err != nil {
		return err
	}

	return nil
}

func (s *systemSCProcessor) addNewlyStakedNodesToValidatorTrie(
	validatorInfos map[uint32][]*state.ValidatorInfo,
	returnData [][]byte,
	nonce uint64,
) error {
	for i := 0; i < len(returnData); i += 2 {
		blsKey := returnData[i]
		rewardAddress := returnData[i+1]

		peerAcc, err := s.getPeerAccount(blsKey)
		if err != nil {
			return err
		}

		err = peerAcc.SetRewardAddress(rewardAddress)
		if err != nil {
			return err
		}

		err = peerAcc.SetBLSPublicKey(blsKey)
		if err != nil {
			return err
		}

		peerAcc.SetListAndIndex(peerAcc.GetShardId(), string(core.NewList), uint32(nonce))
		peerAcc.SetTempRating(s.startRating)
		peerAcc.SetUnStakedEpoch(core.DefaultUnstakedEpoch)

		err = s.peerAccountsDB.SaveAccount(peerAcc)
		if err != nil {
			return err
		}

		validatorInfo := &state.ValidatorInfo{
			PublicKey:       blsKey,
			ShardId:         peerAcc.GetShardId(),
			List:            string(core.NewList),
			Index:           uint32(nonce),
			TempRating:      s.startRating,
			Rating:          s.startRating,
			RewardAddress:   rewardAddress,
			AccumulatedFees: big.NewInt(0),
		}
		validatorInfos[peerAcc.GetShardId()] = append(validatorInfos[peerAcc.GetShardId()], validatorInfo)
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (s *systemSCProcessor) IsInterfaceNil() bool {
	return s == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (s *systemSCProcessor) EpochConfirmed(epoch uint32) {
	s.flagSwitchJailedWaiting.Toggle(epoch >= s.switchEnableEpoch)
	log.Debug("systemSCProcessor: switch jail with waiting", "enabled", s.flagSwitchJailedWaiting.IsSet())

	// only toggle on exact epoch. In future epochs the config should have already been synchronized from peers
	s.flagHystNodesEnabled.Toggle(epoch == s.hystNodesEnableEpoch)

	s.flagChangeMaxNodesEnabled.Toggle(false)
	for _, maxNodesConfig := range s.maxNodesEnableConfig {
		if epoch == maxNodesConfig.EpochEnable {
			s.flagChangeMaxNodesEnabled.Toggle(true)
			s.maxNodes = maxNodesConfig.MaxNumNodes
			break
		}
	}

	log.Debug("systemSCProcessor: consider also (minimum) hysteresis nodes for minimum number of nodes",
		"enabled", epoch >= s.hystNodesEnableEpoch)

	// only toggle on exact epoch as init should be called only once
	s.flagDelegationEnabled.Toggle(epoch == s.delegationEnableEpoch)
	log.Debug("systemSCProcessor: delegation", "enabled", epoch >= s.delegationEnableEpoch)

	s.flagSetOwnerEnabled.Toggle(epoch == s.stakingV2EnableEpoch)
	s.flagStakingV2Enabled.Toggle(epoch >= s.stakingV2EnableEpoch)
	log.Debug("systemSCProcessor: stakingV2", "enabled", epoch >= s.stakingV2EnableEpoch)
	log.Debug("systemSCProcessor:change of maximum number of nodes and/or shuffling percentage",
		"enabled", s.flagChangeMaxNodesEnabled.IsSet(),
		"epoch", epoch,
		"maxNodes", s.maxNodes,
	)
}
