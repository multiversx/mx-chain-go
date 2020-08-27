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
	"github.com/ElrondNetwork/elrond-go/vm"
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

	EndOfEpochCallerAddress []byte
	StakingSCAddress        []byte
}

type systemSCProcessor struct {
	systemVM                vmcommon.VMExecutionHandler
	userAccountsDB          state.AccountsAdapter
	marshalizer             marshal.Marshalizer
	peerAccountsDB          state.AccountsAdapter
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

	return &systemSCProcessor{
		systemVM:                args.SystemVM,
		userAccountsDB:          args.UserAccountsDB,
		peerAccountsDB:          args.PeerAccountsDB,
		marshalizer:             args.Marshalizer,
		startRating:             args.StartRating,
		validatorInfoCreator:    args.ValidatorInfoCreator,
		endOfEpochCallerAddress: args.EndOfEpochCallerAddress,
		stakingSCAddress:        args.StakingSCAddress,
	}, nil
}

// ProcessSystemSmartContract does all the processing at end of epoch in case of system smart contract
func (s *systemSCProcessor) ProcessSystemSmartContract(validatorInfos map[uint32][]*state.ValidatorInfo) error {
	err := s.swapJailedWithWaiting(validatorInfos)
	if err != nil {
		return err
	}
	return nil
}

func (s *systemSCProcessor) auctionSelection() error {

	return nil
}

func (s *systemSCProcessor) swapJailedWithWaiting(validatorInfos map[uint32][]*state.ValidatorInfo) error {
	jailedValidators := s.getSortedJailedNodes(validatorInfos)

	log.Warn("number of jailed validators", "num", len(jailedValidators))
	for _, jailedValidator := range jailedValidators {

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

		log.Warn("swtichJailedWithWaiting", "key", jailedValidator.PublicKey, "code", vmOutput.ReturnCode.String(), "err", err)
		if vmOutput.ReturnCode != vmcommon.Ok || vmOutput.ReturnMessage == vm.ErrBLSPublicKeyAlreadyJailed.Error() {
			continue
		}

		err = s.processSCOutputAccounts(vmOutput)
		if err != nil {
			return err
		}

		err = s.stakingToValidatorStatistics(validatorInfos, jailedValidator, vmOutput)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *systemSCProcessor) stakingToValidatorStatistics(
	validatorInfos map[uint32][]*state.ValidatorInfo,
	jailedValidator *state.ValidatorInfo,
	vmOutput *vmcommon.VMOutput,
) error {
	stakingSCOutput, ok := vmOutput.OutputAccounts[string(s.stakingSCAddress)]
	if !ok {
		return epochStart.ErrStakingSCOutputAccountNotFound
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
	log.Warn("noone in waiting suitable for switch")
	if activeStorageUpdate == nil {
		return nil
	}

	var stakingData systemSmartContracts.StakedData
	err := s.marshalizer.Unmarshal(&stakingData, activeStorageUpdate.Data)
	if err != nil {
		return err
	}

	blsPubKey := activeStorageUpdate.Offset
	log.Warn("staking validator key ", "blsKey", blsPubKey)
	account, err := s.getPeerAccount(blsPubKey)
	if err != nil {
		return err
	}

	if !bytes.Equal(account.GetRewardAddress(), stakingData.RewardAddress) {
		err = account.SetRewardAddress(stakingData.RewardAddress)
		if err != nil {
			return err
		}
	}

	if !bytes.Equal(account.GetBLSPublicKey(), blsPubKey) {
		err = account.SetBLSPublicKey(blsPubKey)
		if err != nil {
			return err
		}
	}

	account.SetListAndIndex(jailedValidator.ShardId, string(core.NewList), uint32(stakingData.StakedNonce))
	account.SetTempRating(s.startRating)
	account.SetUnStakedEpoch(core.DefaultUnstakedEpoch)

	err = s.peerAccountsDB.SaveAccount(account)
	if err != nil {
		return err
	}

	err = s.peerAccountsDB.RemoveAccount(jailedValidator.PublicKey)
	if err != nil {
		return err
	}

	newValidatorInfo := s.validatorInfoCreator.PeerAccountToValidatorInfo(account)
	switchJailedWithNewValidatorInMap(validatorInfos, jailedValidator, newValidatorInfo)

	return nil
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
	for _, listValidators := range validatorInfos {
		for _, validatorInfo := range listValidators {
			if validatorInfo.List == string(core.JailedList) {
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
