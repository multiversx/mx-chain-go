package metachain

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
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
}

type systemSCProcessor struct {
	systemVM                vmcommon.VMExecutionHandler
	userAccountsDB          state.AccountsAdapter
	marshalizer             marshal.Marshalizer
	peerAccountsDB          state.AccountsAdapter
	chanceComputer          sharding.ChanceComputer
	startRating             uint32
	validatorInfoCreator    epochStart.ValidatorInfoCreator
	endOfEpochCallerAddress []byte
	stakingSCAddress        []byte
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

	return &systemSCProcessor{
		systemVM:                args.SystemVM,
		userAccountsDB:          args.UserAccountsDB,
		peerAccountsDB:          args.PeerAccountsDB,
		marshalizer:             args.Marshalizer,
		startRating:             args.StartRating,
		validatorInfoCreator:    args.ValidatorInfoCreator,
		endOfEpochCallerAddress: args.EndOfEpochCallerAddress,
		stakingSCAddress:        args.StakingSCAddress,
		chanceComputer:          args.ChanceComputer,
	}, nil
}

// ProcessSystemSmartContract does all the processing at end of epoch in case of system smart contract
func (s *systemSCProcessor) ProcessSystemSmartContract(validatorInfos map[uint32][]*state.ValidatorInfo, epoch uint32) error {
	err := s.swapJailedWithWaiting(validatorInfos, epoch)
	if err != nil {
		return err
	}
	return nil
}

func (s *systemSCProcessor) auctionSelection() error {

	return nil
}

func (s *systemSCProcessor) swapJailedWithWaiting(validatorInfos map[uint32][]*state.ValidatorInfo, epoch uint32) error {
	jailedValidators := s.getSortedJailedNodes(validatorInfos)

	log.Debug("number of jailed validators", "num", len(jailedValidators))

	newValidators := make(map[string]struct{})
	for _, jailedValidator := range jailedValidators {
		if _, ok := newValidators[string(jailedValidator.PublicKey)]; ok {
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

		newValidator, err := s.stakingToValidatorStatistics(validatorInfos, jailedValidator, vmOutput, epoch)
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
	_ uint32,
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

	var stakingData systemSmartContracts.StakedData
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

	newValidatorInfo := s.validatorInfoCreator.PeerAccountToValidatorInfo(account)
	switchJailedWithNewValidatorInMap(validatorInfos, jailedValidator, newValidatorInfo)

	return blsPubKey, nil
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
	jailedValidators := make([]*state.ValidatorInfo, 0)
	minChance := s.chanceComputer.GetChance(0)
	for _, listValidators := range validatorInfos {
		for _, validatorInfo := range listValidators {
			if s.chanceComputer.GetChance(validatorInfo.TempRating) < minChance {
				jailedValidators = append(jailedValidators, validatorInfo)
			}
		}
	}

	sort.Slice(jailedValidators, func(i, j int) bool {
		if jailedValidators[i].TempRating == jailedValidators[i].TempRating {
			return jailedValidators[i].Index < jailedValidators[j].Index
		}
		return jailedValidators[i].TempRating < jailedValidators[j].TempRating
	})

	return jailedValidators
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

// IsInterfaceNil returns true if underlying object is nil
func (s *systemSCProcessor) IsInterfaceNil() bool {
	return s == nil
}
