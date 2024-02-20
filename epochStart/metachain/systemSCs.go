package metachain

import (
	"fmt"
	"math"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// ArgsNewEpochStartSystemSCProcessing defines the arguments structure for the end of epoch system sc processor
type ArgsNewEpochStartSystemSCProcessing struct {
	SystemVM             vmcommon.VMExecutionHandler
	UserAccountsDB       state.AccountsAdapter
	PeerAccountsDB       state.AccountsAdapter
	Marshalizer          marshal.Marshalizer
	StartRating          uint32
	ValidatorInfoCreator epochStart.ValidatorInfoCreator
	ChanceComputer       nodesCoordinator.ChanceComputer
	ShardCoordinator     sharding.Coordinator

	EndOfEpochCallerAddress []byte
	StakingSCAddress        []byte
	ESDTOwnerAddressBytes   []byte

	GenesisNodesConfig           sharding.GenesisNodesSetupHandler
	EpochNotifier                process.EpochNotifier
	NodesConfigProvider          epochStart.NodesConfigProvider
	StakingDataProvider          epochStart.StakingDataProvider
	AuctionListSelector          epochStart.AuctionListSelector
	MaxNodesChangeConfigProvider epochStart.MaxNodesChangeConfigProvider
	EnableEpochsHandler          common.EnableEpochsHandler
}

type systemSCProcessor struct {
	*legacySystemSCProcessor
	auctionListSelector epochStart.AuctionListSelector
	enableEpochsHandler common.EnableEpochsHandler
}

// NewSystemSCProcessor creates the end of epoch system smart contract processor
func NewSystemSCProcessor(args ArgsNewEpochStartSystemSCProcessing) (*systemSCProcessor, error) {
	if check.IfNil(args.EpochNotifier) {
		return nil, epochStart.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.AuctionListSelector) {
		return nil, epochStart.ErrNilAuctionListSelector
	}

	legacy, err := newLegacySystemSCProcessor(args)
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, epochStart.ErrNilEnableEpochsHandler
	}

	err = core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.SwitchHysteresisForMinNodesFlagInSpecificEpochOnly,
		common.StakingV2OwnerFlagInSpecificEpochOnly,
		common.CorrectLastUnJailedFlagInSpecificEpochOnly,
		common.DelegationSmartContractFlag,
		common.CorrectLastUnJailedFlag,
		common.SwitchJailWaitingFlag,
		common.StakingV2Flag,
		common.ESDTFlagInSpecificEpochOnly,
		common.GovernanceFlag,
		common.SaveJailedAlwaysFlag,
		common.StakingV4Step1Flag,
		common.StakingV4Step2Flag,
		common.StakingQueueFlag,
		common.StakingV4StartedFlag,
		common.DelegationSmartContractFlagInSpecificEpochOnly,
		common.GovernanceFlagInSpecificEpochOnly,
	})
	if err != nil {
		return nil, err
	}

	s := &systemSCProcessor{
		legacySystemSCProcessor: legacy,
		auctionListSelector:     args.AuctionListSelector,
		enableEpochsHandler:     args.EnableEpochsHandler,
	}

	args.EpochNotifier.RegisterNotifyHandler(s)
	return s, nil
}

// ProcessSystemSmartContract does all the processing at end of epoch in case of system smart contract
func (s *systemSCProcessor) ProcessSystemSmartContract(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	header data.HeaderHandler,
) error {
	err := checkNilInputValues(validatorsInfoMap, header)
	if err != nil {
		return err
	}

	err = s.processLegacy(validatorsInfoMap, header.GetNonce(), header.GetEpoch())
	if err != nil {
		return err
	}
	return s.processWithNewFlags(validatorsInfoMap, header)
}

func checkNilInputValues(validatorsInfoMap state.ShardValidatorsInfoMapHandler, header data.HeaderHandler) error {
	if check.IfNil(header) {
		return process.ErrNilHeaderHandler
	}
	if validatorsInfoMap == nil {
		return fmt.Errorf("systemSCProcessor.ProcessSystemSmartContract : %w, header nonce: %d ",
			errNilValidatorsInfoMap, header.GetNonce())
	}

	return nil
}

func (s *systemSCProcessor) processWithNewFlags(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	header data.HeaderHandler,
) error {
	if s.enableEpochsHandler.IsFlagEnabled(common.GovernanceFlagInSpecificEpochOnly) {
		err := s.updateToGovernanceV2()
		if err != nil {
			return err
		}
	}

	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4Step1Flag) {
		err := s.stakeNodesFromQueue(validatorsInfoMap, math.MaxUint32, header.GetNonce(), common.AuctionList)
		if err != nil {
			return err
		}
	}

	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4Step2Flag) {
		err := s.prepareStakingDataForEligibleNodes(validatorsInfoMap)
		if err != nil {
			return err
		}

		err = s.fillStakingDataForNonEligible(validatorsInfoMap)
		if err != nil {
			return err
		}

		err = s.unStakeNodesWithNotEnoughFundsWithStakingV4(validatorsInfoMap, header.GetEpoch())
		if err != nil {
			return err
		}

		err = s.auctionListSelector.SelectNodesFromAuctionList(validatorsInfoMap, header.GetPrevRandSeed())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *systemSCProcessor) unStakeNodesWithNotEnoughFundsWithStakingV4(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	epoch uint32,
) error {
	nodesToUnStake, mapOwnersKeys, err := s.stakingDataProvider.ComputeUnQualifiedNodes(validatorsInfoMap)
	if err != nil {
		return err
	}

	log.Debug("unStake nodes with not enough funds", "num", len(nodesToUnStake))
	for _, blsKey := range nodesToUnStake {
		log.Debug("unStake at end of epoch for node", "blsKey", blsKey)
		err = s.unStakeOneNode(blsKey, epoch)
		if err != nil {
			return err
		}

		validatorInfo := validatorsInfoMap.GetValidator(blsKey)
		if validatorInfo == nil {
			return fmt.Errorf(
				"%w in systemSCProcessor.unStakeNodesWithNotEnoughFundsWithStakingV4 because validator might be in additional queue after staking v4",
				epochStart.ErrNilValidatorInfo)
		}

		validatorLeaving := validatorInfo.ShallowClone()
		validatorLeaving.SetListAndIndex(string(common.LeavingList), validatorLeaving.GetIndex(), true)
		err = validatorsInfoMap.Replace(validatorInfo, validatorLeaving)
		if err != nil {
			return err
		}
	}

	return s.updateDelegationContracts(mapOwnersKeys)
}

func (s *systemSCProcessor) updateToGovernanceV2() error {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.GovernanceSCAddress,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{},
		},
		RecipientAddr: vm.GovernanceSCAddress,
		Function:      "initV2",
	}
	vmOutput, errRun := s.systemVM.RunSmartContractCall(vmInput)
	if errRun != nil {
		return fmt.Errorf("%w when updating to governanceV2", errRun)
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return fmt.Errorf("got return code %s when updating to governanceV2", vmOutput.ReturnCode)
	}

	err := s.processSCOutputAccounts(vmOutput)
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
func (s *systemSCProcessor) EpochConfirmed(epoch uint32, _ uint64) {
	s.legacyEpochConfirmed(epoch)
}
