package metachain

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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

	EndOfEpochCallerAddress []byte
	StakingSCAddress        []byte

	SwitchJailWaitingEnableEpoch           uint32
	SwitchHysteresisForMinNodesEnableEpoch uint32
	MaxNodesEnableConfig                   []config.MaxNodesChangeConfig

	GenesisNodesConfig sharding.GenesisNodesSetupHandler
	EpochNotifier      process.EpochNotifier
}

type systemSCProcessor struct {
	systemVM                vmcommon.VMExecutionHandler
	userAccountsDB          state.AccountsAdapter
	marshalizer             marshal.Marshalizer
	peerAccountsDB          state.AccountsAdapter
	chanceComputer          sharding.ChanceComputer
	startRating             uint32
	validatorInfoCreator    epochStart.ValidatorInfoCreator
	genesisNodesConfig      sharding.GenesisNodesSetupHandler
	endOfEpochCallerAddress []byte
	stakingSCAddress        []byte
	switchEnableEpoch       uint32
	hystNodesEnableEpoch    uint32
	maxNodesEnableConfig    []config.MaxNodesChangeConfig
	maxNodes                uint32

	flagSwitchEnabled         atomic.Flag
	flagHystNodesEnabled      atomic.Flag
	flagChangeMaxNodesEnabled atomic.Flag

	mapNumSwitchedPerShard   map[uint32]uint32
	mapNumSwitchablePerShard map[uint32]uint32
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
func (s *systemSCProcessor) ProcessSystemSmartContract(validatorInfos map[uint32][]*state.ValidatorInfo) error {
	if s.flagHystNodesEnabled.IsSet() {
		err := s.updateSystemSCConfigMinNodes()
		if err != nil {
			return err
		}
	}

	if s.flagChangeMaxNodesEnabled.IsSet() {
		err := s.updateSystemSCConfigMaxNodes()
		if err != nil {
			return err
		}
	}

	if s.flagSwitchEnabled.IsSet() {
		err := s.computeNumWaitingPerShard(validatorInfos)
		if err != nil {
			return err
		}

		err = s.swapJailedWithWaiting(validatorInfos)
		if err != nil {
			return err
		}
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
func (s *systemSCProcessor) updateSystemSCConfigMaxNodes() error {
	maxNumberOfNodes := s.maxNodes
	err := s.setMaxNumberOfNodes(maxNumberOfNodes)

	return err
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

	var stakingData systemSmartContracts.StakedDataV2
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

func (s *systemSCProcessor) getExistingAccount(address []byte) (state.UserAccountHandler, error) {
	acnt, err := s.userAccountsDB.GetExistingAccount(address)
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
		acc, err := s.getExistingAccount(outAcc.Address)
		if err != nil {
			return err
		}

		storageUpdates := process.GetSortedStorageUpdates(outAcc)
		for _, storeUpdate := range storageUpdates {
			acc.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
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

func (s *systemSCProcessor) setMaxNumberOfNodes(maxNumNodes uint32) error {
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
		return err
	}

	log.Debug("setMaxNumberOfNodes called with",
		"maxNumNodes", maxNumNodes,
		"returnMessage", vmOutput.ReturnMessage)

	if vmOutput.ReturnCode != vmcommon.Ok {
		return epochStart.ErrInvalidMaxNumberOfNodes
	}

	err = s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (s *systemSCProcessor) IsInterfaceNil() bool {
	return s == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (s *systemSCProcessor) EpochConfirmed(epoch uint32) {
	s.flagSwitchEnabled.Toggle(epoch >= s.switchEnableEpoch)

	// only toggle on exact epoch. In future epochs the config should have already been synchronized from peers
	s.flagHystNodesEnabled.Toggle(epoch == s.hystNodesEnableEpoch)

	for _, maxNodesConfig := range s.maxNodesEnableConfig {
		if epoch >= maxNodesConfig.EpochEnable {
			// to cover also rollbacks, we always set the maxNodes and set the Enabled flag
			s.flagChangeMaxNodesEnabled.Toggle(true)
			s.maxNodes = maxNodesConfig.MaxNumNodes
		}
	}

	log.Debug("systemSCProcessor: switch jail with waiting", "enabled", s.flagSwitchEnabled.IsSet())
	log.Debug("systemProcessor: consider also (minimum) hysteresis nodes for minimum number of nodes",
		"enabled", s.flagHystNodesEnabled.IsSet())
	log.Debug("systemProcessor:change of maximum number of nodes and/or shuffling percentage",
		"enabled", s.flagChangeMaxNodesEnabled.IsSet(),
		"epoch", epoch,
		"maxNodes", s.maxNodes,
	)
}
