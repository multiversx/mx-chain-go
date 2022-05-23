package metachain

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/vm"
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
	ChanceComputer       nodesCoordinator.ChanceComputer
	ShardCoordinator     sharding.Coordinator
	EpochConfig          config.EpochConfig

	EndOfEpochCallerAddress []byte
	StakingSCAddress        []byte
	ESDTOwnerAddressBytes   []byte

	GenesisNodesConfig           sharding.GenesisNodesSetupHandler
	EpochNotifier                process.EpochNotifier
	NodesConfigProvider          epochStart.NodesConfigProvider
	StakingDataProvider          epochStart.StakingDataProvider
	AuctionListSelector          epochStart.AuctionListSelector
	MaxNodesChangeConfigProvider epochStart.MaxNodesChangeConfigProvider
}

type systemSCProcessor struct {
	*legacySystemSCProcessor
	auctionListSelector epochStart.AuctionListSelector

	governanceEnableEpoch    uint32
	builtInOnMetaEnableEpoch uint32
	stakingV4EnableEpoch     uint32

	flagGovernanceEnabled    atomic.Flag
	flagBuiltInOnMetaEnabled atomic.Flag
	flagInitStakingV4Enabled atomic.Flag
	flagStakingV4Enabled     atomic.Flag
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

	s := &systemSCProcessor{
		legacySystemSCProcessor:  legacy,
		governanceEnableEpoch:    args.EpochConfig.EnableEpochs.GovernanceEnableEpoch,
		builtInOnMetaEnableEpoch: args.EpochConfig.EnableEpochs.BuiltInFunctionOnMetaEnableEpoch,
		stakingV4EnableEpoch:     args.EpochConfig.EnableEpochs.StakingV4EnableEpoch,
		auctionListSelector:      args.AuctionListSelector,
	}

	log.Debug("systemSC: enable epoch for governanceV2 init", "epoch", s.governanceEnableEpoch)
	log.Debug("systemSC: enable epoch for create NFT on meta", "epoch", s.builtInOnMetaEnableEpoch)
	log.Debug("systemSC: enable epoch for staking v4", "epoch", s.stakingV4EnableEpoch)

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
	if s.flagGovernanceEnabled.IsSet() {
		err := s.updateToGovernanceV2()
		if err != nil {
			return err
		}
	}

	if s.flagBuiltInOnMetaEnabled.IsSet() {
		tokenID, err := s.initTokenOnMeta()
		if err != nil {
			return err
		}

		err = s.initLiquidStakingSC(tokenID)
		if err != nil {
			return err
		}
	}

	if s.flagInitStakingV4Enabled.IsSet() {
		err := s.stakeNodesFromQueue(validatorsInfoMap, math.MaxUint32, header.GetNonce(), common.AuctionList)
		if err != nil {
			return err
		}
	}

	if s.flagStakingV4Enabled.IsSet() {
		err := s.prepareStakingDataForEligibleNodes(validatorsInfoMap)
		if err != nil {
			return err
		}

		err = s.fillStakingDataForNonEligible(validatorsInfoMap)
		if err != nil {
			return err
		}

		unqualifiedOwners, err := s.unStakeNodesWithNotEnoughFundsWithStakingV4(validatorsInfoMap, header.GetEpoch())
		if err != nil {
			return err
		}

		err = s.auctionListSelector.SelectNodesFromAuctionList(validatorsInfoMap, unqualifiedOwners, header.GetPrevRandSeed())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *systemSCProcessor) unStakeNodesWithNotEnoughFundsWithStakingV4(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	epoch uint32,
) (map[string]struct{}, error) {
	nodesToUnStake, mapOwnersKeys, err := s.stakingDataProvider.ComputeUnQualifiedNodes(validatorsInfoMap)
	if err != nil {
		return nil, err
	}

	log.Debug("unStake nodes with not enough funds", "num", len(nodesToUnStake))
	for _, blsKey := range nodesToUnStake {
		log.Debug("unStake at end of epoch for node", "blsKey", blsKey)
		err = s.unStakeOneNode(blsKey, epoch)
		if err != nil {
			return nil, err
		}

		validatorInfo := validatorsInfoMap.GetValidator(blsKey)
		if validatorInfo == nil {
			return nil, fmt.Errorf(
				"%w in systemSCProcessor.unStakeNodesWithNotEnoughFundsWithStakingV4 because validator might be in additional queue after staking v4",
				epochStart.ErrNilValidatorInfo)
		}

		validatorLeaving := validatorInfo.ShallowClone()
		validatorLeaving.SetList(string(common.LeavingList))
		err = validatorsInfoMap.Replace(validatorInfo, validatorLeaving)
		if err != nil {
			return nil, err
		}
	}
	err = s.updateDelegationContracts(mapOwnersKeys)
	if err != nil {

	}

	return copyOwnerKeysInMap(mapOwnersKeys), nil
}

func copyOwnerKeysInMap(mapOwnersKeys map[string][][]byte) map[string]struct{} {
	ret := make(map[string]struct{})
	for owner := range mapOwnersKeys {
		ret[owner] = struct{}{}
	}

	return ret
}

func (s *systemSCProcessor) prepareStakingDataForAllNodes(validatorsInfoMap state.ShardValidatorsInfoMapHandler) error {
	allNodes := GetAllNodeKeys(validatorsInfoMap)
	return s.prepareStakingData(allNodes)
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

	s.flagGovernanceEnabled.SetValue(epoch == s.governanceEnableEpoch)
	log.Debug("systemProcessor: governanceV2", "enabled", s.flagGovernanceEnabled.IsSet())

	s.flagBuiltInOnMetaEnabled.SetValue(epoch == s.builtInOnMetaEnableEpoch)
	log.Debug("systemProcessor: create NFT on meta", "enabled", s.flagBuiltInOnMetaEnabled.IsSet())

	s.flagInitStakingV4Enabled.SetValue(epoch == s.stakingV4InitEnableEpoch)
	log.Debug("systemProcessor: init staking v4", "enabled", s.flagInitStakingV4Enabled.IsSet())

	s.flagStakingV4Enabled.SetValue(epoch >= s.stakingV4EnableEpoch)
	log.Debug("systemProcessor: staking v4", "enabled", s.flagStakingV4Enabled.IsSet())
}
