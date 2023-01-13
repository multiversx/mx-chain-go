package metachain

import (
	"fmt"
	"math"
	"math/big"

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

	governanceEnableEpoch    uint32
	builtInOnMetaEnableEpoch uint32
	stakingV4EnableEpoch     uint32

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
	err := s.processLegacy(validatorsInfoMap, header.GetNonce(), header.GetEpoch())
	if err != nil {
		return err
	}
	return s.processWithNewFlags(validatorsInfoMap, header)
}

func (s *systemSCProcessor) processWithNewFlags(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	header data.HeaderHandler,
) error {
	if s.enableEpochsHandler.IsGovernanceFlagEnabledForCurrentEpoch() {
		err := s.updateToGovernanceV2()
		if err != nil {
			return err
		}
	}

	if s.enableEpochsHandler.IsInitLiquidStakingEnabled() {
		tokenID, err := s.initTokenOnMeta()
		if err != nil {
			return err
		}

		err = s.initLiquidStakingSC(tokenID)
		if err != nil {
			return err
		}
	}

	if s.enableEpochsHandler.IsStakingV4InitEnabled() {
		err := s.stakeNodesFromQueue(validatorsInfoMap, math.MaxUint32, header.GetNonce(), common.AuctionList)
		if err != nil {
			return err
		}
	}

	if s.enableEpochsHandler.IsStakingV4Enabled() {
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
		validatorLeaving.SetList(string(common.LeavingList))
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

func (s *systemSCProcessor) initTokenOnMeta() ([]byte, error) {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  vm.ESDTSCAddress,
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{},
			GasProvided: math.MaxUint64,
		},
		RecipientAddr: vm.ESDTSCAddress,
		Function:      "initDelegationESDTOnMeta",
	}
	vmOutput, errRun := s.systemVM.RunSmartContractCall(vmInput)
	if errRun != nil {
		return nil, fmt.Errorf("%w when setting up NFTs on metachain", errRun)
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf("got return code %s, return message %s when setting up NFTs on metachain", vmOutput.ReturnCode, vmOutput.ReturnMessage)
	}
	if len(vmOutput.ReturnData) != 1 {
		return nil, fmt.Errorf("invalid return data on initDelegationESDTOnMeta")
	}

	err := s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return nil, err
	}

	return vmOutput.ReturnData[0], nil
}

func (s *systemSCProcessor) initLiquidStakingSC(tokenID []byte) error {
	codeMetaData := &vmcommon.CodeMetadata{
		Upgradeable: false,
		Payable:     false,
		Readable:    true,
	}

	vmInput := &vmcommon.ContractCreateInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.LiquidStakingSCAddress,
			Arguments:  [][]byte{tokenID},
			CallValue:  big.NewInt(0),
		},
		ContractCode:         vm.LiquidStakingSCAddress,
		ContractCodeMetadata: codeMetaData.ToBytes(),
	}

	vmOutput, err := s.systemVM.RunSmartContractCreate(vmInput)
	if err != nil {
		return err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return epochStart.ErrCouldNotInitLiquidStakingSystemSC
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

// IsInterfaceNil returns true if underlying object is nil
func (s *systemSCProcessor) IsInterfaceNil() bool {
	return s == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (s *systemSCProcessor) EpochConfirmed(epoch uint32, _ uint64) {
	s.legacyEpochConfirmed(epoch)
}
