//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. staking.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("vm/systemsmartcontracts")

const ownerKey = "owner"
const nodesConfigKey = "nodesConfig"

type stakingSC struct {
	eei                      vm.SystemEI
	unBondPeriod             uint64
	stakeAccessAddr          []byte // TODO add a viewAddress field and use it on all system SC view functions
	jailAccessAddr           []byte
	endOfEpochAccessAddr     []byte
	numRoundsWithoutBleed    uint64
	bleedPercentagePerRound  float64
	maximumPercentageToBleed float64
	gasCost                  vm.GasCost
	minNumNodes              uint64
	maxNumNodes              uint64
	marshalizer              marshal.Marshalizer
	stakeValue               *big.Int
	walletAddressLen         int
	mutExecution             sync.RWMutex
	minNodePrice             *big.Int
	enableEpochsHandler      common.EnableEpochsHandler
}

// ArgsNewStakingSmartContract holds the arguments needed to create a StakingSmartContract
type ArgsNewStakingSmartContract struct {
	StakingSCConfig      config.StakingSystemSCConfig
	MinNumNodes          uint64
	Eei                  vm.SystemEI
	StakingAccessAddr    []byte
	JailAccessAddr       []byte
	EndOfEpochAccessAddr []byte
	GasCost              vm.GasCost
	Marshalizer          marshal.Marshalizer
	EnableEpochsHandler  common.EnableEpochsHandler
}

// NewStakingSmartContract creates a staking smart contract
func NewStakingSmartContract(
	args ArgsNewStakingSmartContract,
) (*stakingSC, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.StakingAccessAddr) < 1 {
		return nil, vm.ErrInvalidStakingAccessAddress
	}
	if len(args.JailAccessAddr) < 1 {
		return nil, vm.ErrInvalidJailAccessAddress
	}
	if len(args.EndOfEpochAccessAddr) < 1 {
		return nil, vm.ErrInvalidEndOfEpochAccessAddress
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if args.StakingSCConfig.BleedPercentagePerRound < 0 {
		return nil, vm.ErrNegativeBleedPercentagePerRound
	}
	if args.StakingSCConfig.MaximumPercentageToBleed < 0 {
		return nil, vm.ErrNegativeMaximumPercentageToBleed
	}
	if args.MinNumNodes > args.StakingSCConfig.MaxNumberOfNodesForStake {
		return nil, vm.ErrInvalidMaxNumberOfNodes
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, vm.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.CorrectFirstQueuedFlag,
		common.ValidatorToDelegationFlag,
		common.StakingV2Flag,
		common.CorrectLastUnJailedFlag,
		common.CorrectJailedNotUnStakedEmptyQueueFlag,
		common.StakeFlag,
	})
	if err != nil {
		return nil, err
	}

	minStakeValue, okValue := big.NewInt(0).SetString(args.StakingSCConfig.MinStakeValue, conversionBase)
	if !okValue || minStakeValue.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, args.StakingSCConfig.MinStakeValue)
	}

	reg := &stakingSC{
		eei:                      args.Eei,
		unBondPeriod:             args.StakingSCConfig.UnBondPeriod,
		stakeAccessAddr:          args.StakingAccessAddr,
		jailAccessAddr:           args.JailAccessAddr,
		numRoundsWithoutBleed:    args.StakingSCConfig.NumRoundsWithoutBleed,
		bleedPercentagePerRound:  args.StakingSCConfig.BleedPercentagePerRound,
		maximumPercentageToBleed: args.StakingSCConfig.MaximumPercentageToBleed,
		gasCost:                  args.GasCost,
		minNumNodes:              args.MinNumNodes,
		maxNumNodes:              args.StakingSCConfig.MaxNumberOfNodesForStake,
		marshalizer:              args.Marshalizer,
		endOfEpochAccessAddr:     args.EndOfEpochAccessAddr,
		walletAddressLen:         len(args.StakingAccessAddr),
		minNodePrice:             minStakeValue,
		enableEpochsHandler:      args.EnableEpochsHandler,
	}

	var conversionOk bool
	reg.stakeValue, conversionOk = big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, conversionBase)
	if !conversionOk || reg.stakeValue.Cmp(zero) < 0 {
		return nil, vm.ErrNegativeInitialStakeValue
	}

	return reg, nil
}

// Execute calls one of the functions from the staking smart contract and runs the code according to the input
func (s *stakingSC) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	s.mutExecution.RLock()
	defer s.mutExecution.RUnlock()
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	if len(args.ESDTTransfers) > 0 {
		s.eei.AddReturnMessage("cannot transfer ESDT to system SCs")
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return s.init(args)
	case "stake":
		return s.stake(args, false)
	case "register":
		return s.stake(args, true)
	case "unStake":
		return s.unStake(args)
	case "unBond":
		return s.unBond(args)
	case "get":
		return s.get(args)
	case "isStaked":
		return s.isStaked(args)
	case "slash":
		return s.slash(args)
	case "jail":
		return s.jail(args)
	case "unJail":
		return s.unJail(args)
	case "changeRewardAddress":
		return s.changeRewardAddress(args)
	case "changeValidatorKeys":
		return s.changeValidatorKey(args)
	case "switchJailedWithWaiting":
		return s.switchJailedWithWaiting(args)
	case "getQueueIndex":
		return s.getWaitingListIndex(args)
	case "getQueueSize":
		return s.getWaitingListSize(args)
	case "getRewardAddress":
		return s.getRewardAddress(args)
	case "getBLSKeyStatus":
		return s.getBLSKeyStatus(args)
	case "getRemainingUnBondPeriod":
		return s.getRemainingUnbondPeriod(args)
	case "getQueueRegisterNonceAndRewardAddress":
		return s.getWaitingListRegisterNonceAndRewardAddress(args)
	case "updateConfigMinNodes":
		return s.updateConfigMinNodes(args)
	case "setOwnersOnAddresses":
		return s.setOwnersOnAddresses(args)
	case "getOwner":
		return s.getOwner(args)
	case "updateConfigMaxNodes":
		return s.updateConfigMaxNodes(args)
	case "stakeNodesFromQueue":
		return s.stakeNodesFromQueue(args)
	case "unStakeAtEndOfEpoch":
		return s.unStakeAtEndOfEpoch(args)
	case "getTotalNumberOfRegisteredNodes":
		return s.getTotalNumberOfRegisteredNodes(args)
	case "resetLastUnJailedFromQueue":
		return s.resetLastUnJailedFromQueue(args)
	case "cleanAdditionalQueue":
		return s.cleanAdditionalQueue(args)
	case "changeOwnerAndRewardAddress":
		return s.changeOwnerAndRewardAddress(args)
	case "fixWaitingListQueueSize":
		return s.fixWaitingListQueueSize(args)
	case "addMissingNodeToQueue":
		return s.addMissingNodeToQueue(args)
	case "unStakeAllNodesFromQueue":
		return s.unStakeAllNodesFromQueue(args)
	}

	return vmcommon.UserError
}

func (s *stakingSC) addToStakedNodes(value int64) {
	stakeConfig := s.getConfig()
	stakeConfig.StakedNodes += value
	s.setConfig(stakeConfig)
}

func (s *stakingSC) removeFromStakedNodes() {
	stakeConfig := s.getConfig()
	if stakeConfig.StakedNodes > 0 {
		stakeConfig.StakedNodes--
	}
	s.setConfig(stakeConfig)
}

func (s *stakingSC) numSpareNodes() int64 {
	stakeConfig := s.getConfig()
	return stakeConfig.StakedNodes - stakeConfig.JailedNodes - stakeConfig.MinNumNodes
}

func (s *stakingSC) canStake() bool {
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) {
		return true
	}

	stakeConfig := s.getConfig()
	return stakeConfig.StakedNodes < stakeConfig.MaxNumNodes
}

func (s *stakingSC) canStakeIfOneRemoved() bool {
	stakeConfig := s.getConfig()
	return stakeConfig.StakedNodes <= stakeConfig.MaxNumNodes
}

func (s *stakingSC) canUnStake() bool {
	return s.numSpareNodes() > 0
}

func (s *stakingSC) canUnBond() bool {
	return s.numSpareNodes() >= 0
}

func (s *stakingSC) changeValidatorKey(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("changeValidatorKey function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	s.eei.AddReturnMessage("function is deprecated")
	return vmcommon.UserError
}

func (s *stakingSC) changeRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("stake function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 2, len(args.Arguments)))
		return vmcommon.UserError
	}

	newRewardAddress := args.Arguments[0]
	if len(newRewardAddress) != s.walletAddressLen {
		s.eei.AddReturnMessage("invalid reward address")
		return vmcommon.UserError
	}

	for _, blsKey := range args.Arguments[1:] {
		stakedData, err := s.getOrCreateRegisteredData(blsKey)
		if err != nil {
			s.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			continue
		}

		stakedData.RewardAddress = newRewardAddress
		err = s.saveStakingData(blsKey, stakedData)
		if err != nil {
			s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (s *stakingSC) removeFromJailedNodes() {
	stakeConfig := s.getConfig()
	if stakeConfig.JailedNodes > 0 {
		stakeConfig.JailedNodes--
	}
	s.setConfig(stakeConfig)
}

func (s *stakingSC) unJailV1(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("unJail function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}

	for _, argument := range args.Arguments {
		stakedData, err := s.getOrCreateRegisteredData(argument)
		if err != nil {
			s.eei.AddReturnMessage("cannot get or created registered data: error " + err.Error())
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			s.eei.AddReturnMessage("cannot unJail a key that is not registered")
			return vmcommon.UserError
		}

		if stakedData.UnJailedNonce <= stakedData.JailedNonce {
			s.removeFromJailedNodes()
		}

		stakedData.JailedRound = math.MaxUint64
		stakedData.UnJailedNonce = s.eei.BlockChainHook().CurrentNonce()

		err = s.saveStakingData(argument, stakedData)
		if err != nil {
			s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (s *stakingSC) unJail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.StakeFlag) {
		return s.unJailV1(args)
	}

	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("unJail function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("wrong number of arguments, wanted 1")
		return vmcommon.UserError
	}

	stakedData, err := s.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage("cannot get or created registered data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(stakedData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("cannot unJail a key that is not registered")
		return vmcommon.UserError
	}
	if !stakedData.Jailed && !s.eei.CanUnJail(args.Arguments[0]) {
		s.eei.AddReturnMessage("cannot unJail a node which is not jailed")
		return vmcommon.UserError
	}

	stakedData.JailedRound = math.MaxUint64
	stakedData.UnJailedNonce = s.eei.BlockChainHook().CurrentNonce()
	stakedData.Jailed = false

	err = s.processStake(args.Arguments[0], stakedData, stakedData.NumJailed == 1)
	if err != nil {
		return vmcommon.UserError
	}

	err = s.saveStakingData(args.Arguments[0], stakedData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingSC) jail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.jailAccessAddr) {
		return vmcommon.UserError
	}

	for _, argument := range args.Arguments {
		stakedData, err := s.getOrCreateRegisteredData(argument)
		if err != nil {
			s.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			s.eei.AddReturnMessage("cannot jail a key that is not registered")
			return vmcommon.UserError
		}

		stakedData.JailedRound = s.eei.BlockChainHook().CurrentRound()
		stakedData.JailedNonce = s.eei.BlockChainHook().CurrentNonce()
		stakedData.Jailed = true
		stakedData.NumJailed++
		err = s.saveStakingData(argument, stakedData)
		if err != nil {
			s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (s *stakingSC) get(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
		s.eei.AddReturnMessage("function deprecated")
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	value := s.eei.GetStorage(args.Arguments[0])
	s.eei.Finish(value)

	return vmcommon.Ok
}

func (s *stakingSC) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := s.eei.GetStorage([]byte(ownerKey))
	if ownerAddress != nil {
		s.eei.AddReturnMessage("smart contract was already initialized")
		return vmcommon.UserError
	}

	s.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	s.eei.SetStorage(args.CallerAddr, big.NewInt(0).Bytes())

	stakeConfig := &StakingNodesConfig{
		MinNumNodes: int64(s.minNumNodes),
		MaxNumNodes: int64(s.maxNumNodes),
	}
	s.setConfig(stakeConfig)

	epoch := s.eei.BlockChainHook().CurrentEpoch()
	epochData := fmt.Sprintf("epoch_%d", epoch)

	s.eei.SetStorage([]byte(epochData), s.stakeValue.Bytes())

	return vmcommon.Ok
}

func (s *stakingSC) stake(args *vmcommon.ContractCallInput, onlyRegister bool) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("stake function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) < 3 {
		s.eei.AddReturnMessage("not enough arguments, needed BLS key, reward address and owner address")
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}
	if s.isNodeJailedOrWithBadRating(registrationData, args.Arguments[0]) {
		s.eei.AddReturnMessage("cannot stake node which is jailed or with bad rating")
		return vmcommon.UserError
	}

	registrationData.RewardAddress = args.Arguments[1]
	registrationData.OwnerAddress = args.Arguments[2]
	registrationData.StakeValue.Set(s.stakeValue)
	if !onlyRegister {
		err = s.processStake(args.Arguments[0], registrationData, false)
		if err != nil {
			return vmcommon.UserError
		}
	}

	err = s.saveStakingData(args.Arguments[0], registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save staking registered data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingSC) unStakeAtEndOfEpoch(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		// backward compatibility - no need for return message
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("not enough arguments, needed the BLS key")
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("cannot unStake a key that is not registered")
		return vmcommon.UserError
	}
	if registrationData.Jailed && !registrationData.Staked {
		s.eei.AddReturnMessage("already unStaked at switchJailedToWaiting")
		return vmcommon.Ok
	}

	if !registrationData.Staked && !registrationData.Waiting {
		log.Debug("stakingSC.unStakeAtEndOfEpoch: cannot unStake node which was already unStaked", "blsKey", hex.EncodeToString(args.Arguments[0]))
		return vmcommon.Ok
	}

	if registrationData.Staked {
		s.removeFromStakedNodes()
	}

	if registrationData.Waiting {
		err = s.removeFromWaitingList(args.Arguments[0])
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	registrationData.Staked = false
	registrationData.UnStakedEpoch = s.eei.BlockChainHook().CurrentEpoch()
	registrationData.UnStakedNonce = s.eei.BlockChainHook().CurrentNonce()
	registrationData.Waiting = false

	err = s.saveStakingData(args.Arguments[0], registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingSC) activeStakingFor(stakingData *StakedDataV2_0) {
	stakingData.RegisterNonce = s.eei.BlockChainHook().CurrentNonce()
	stakingData.Staked = true
	stakingData.StakedNonce = s.eei.BlockChainHook().CurrentNonce()
	stakingData.UnStakedEpoch = common.DefaultUnstakedEpoch
	stakingData.UnStakedNonce = 0
	stakingData.Waiting = false
}

func (s *stakingSC) processStake(blsKey []byte, registrationData *StakedDataV2_0, addFirst bool) error {
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) {
		return s.processStakeV2(registrationData)
	}

	return s.processStakeV1(blsKey, registrationData, addFirst)
}

func (s *stakingSC) processStakeV2(registrationData *StakedDataV2_0) error {
	if registrationData.Staked {
		return nil
	}

	registrationData.RegisterNonce = s.eei.BlockChainHook().CurrentNonce()
	s.addToStakedNodes(1)
	s.activeStakingFor(registrationData)

	return nil
}

func (s *stakingSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) {
		return s.unStakeV2(args)
	}

	return s.unStakeV1(args)
}

func (s *stakingSC) unStakeV2(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, retCode := s.checkUnStakeArgs(args)
	if retCode != vmcommon.Ok {
		return retCode
	}

	if !registrationData.Staked {
		s.eei.AddReturnMessage(vm.ErrWaitingListDisabled.Error())
		return vmcommon.ExecutionFailed
	}

	return s.tryUnStake(args.Arguments[0], registrationData)
}

func (s *stakingSC) checkUnStakeArgs(args *vmcommon.ContractCallInput) (*StakedDataV2_0, vmcommon.ReturnCode) {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("unStake function not allowed to be called by address " + string(args.CallerAddr))
		return nil, vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		s.eei.AddReturnMessage("not enough arguments, needed BLS key and reward address")
		return nil, vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return nil, vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("cannot unStake a key that is not registered")
		return nil, vmcommon.UserError
	}
	if !bytes.Equal(args.Arguments[1], registrationData.RewardAddress) {
		s.eei.AddReturnMessage("unStake possible only from staker caller")
		return nil, vmcommon.UserError
	}
	if s.isNodeJailedOrWithBadRating(registrationData, args.Arguments[0]) {
		s.eei.AddReturnMessage("cannot unStake node which is jailed or with bad rating")
		return nil, vmcommon.UserError
	}

	if !registrationData.Staked && !registrationData.Waiting {
		s.eei.AddReturnMessage("cannot unStake node which was already unStaked")
		return nil, vmcommon.UserError
	}

	return registrationData, vmcommon.Ok
}

func (s *stakingSC) tryUnStake(key []byte, registrationData *StakedDataV2_0) vmcommon.ReturnCode {
	if !s.canUnStake() {
		s.eei.AddReturnMessage("unStake is not possible as too many left")
		return vmcommon.UserError
	}

	s.removeFromStakedNodes()

	return s.doUnStake(key, registrationData)
}

func (s *stakingSC) doUnStake(key []byte, registrationData *StakedDataV2_0) vmcommon.ReturnCode {
	registrationData.Staked = false
	registrationData.UnStakedEpoch = s.eei.BlockChainHook().CurrentEpoch()
	registrationData.UnStakedNonce = s.eei.BlockChainHook().CurrentNonce()
	registrationData.Waiting = false

	err := s.saveStakingData(key, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingSC) unBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("unBond function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		s.eei.AddReturnMessage("not enough arguments, needed BLS key")
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}

	blsKeyBytes := make([]byte, 2*len(args.Arguments[0]))
	_ = hex.Encode(blsKeyBytes, args.Arguments[0])
	encodedBlsKey := string(blsKeyBytes)

	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage(fmt.Sprintf("cannot unBond key %s that is not registered", encodedBlsKey))
		return vmcommon.UserError
	}
	if registrationData.Staked {
		s.eei.AddReturnMessage(fmt.Sprintf("unBond is not possible for key %s which is staked", encodedBlsKey))
		return vmcommon.UserError
	}
	if s.isNodeJailedOrWithBadRating(registrationData, args.Arguments[0]) {
		s.eei.AddReturnMessage("cannot unBond node which is jailed or with bad rating " + encodedBlsKey)
		return vmcommon.UserError
	}
	if registrationData.Waiting {
		s.eei.AddReturnMessage(fmt.Sprintf("unBond in not possible for key %s which is in waiting list", encodedBlsKey))
		return vmcommon.UserError
	}

	currentNonce := s.eei.BlockChainHook().CurrentNonce()
	if registrationData.UnStakedNonce > 0 && currentNonce-registrationData.UnStakedNonce < s.unBondPeriod {
		s.eei.AddReturnMessage(fmt.Sprintf("unBond is not possible for key %s because unBond period did not pass", encodedBlsKey))
		return vmcommon.UserError
	}

	if !s.canUnBond() {
		s.eei.AddReturnMessage("unBond is currently unavailable: number of total validators in the network is at minimum")
		return vmcommon.UserError
	}
	if s.eei.IsValidator(args.Arguments[0]) {
		s.eei.AddReturnMessage("unBond is not possible: the node with key " + encodedBlsKey + " is still a validator")
		return vmcommon.UserError
	}

	s.eei.SetStorage(args.Arguments[0], nil)
	return vmcommon.Ok
}

// backward compatibility
func (s *stakingSC) slash(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	s.eei.AddReturnMessage("slash function called by not the owners address")
	return vmcommon.UserError
}

func (s *stakingSC) isStaked(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("key is not registered")
		return vmcommon.UserError
	}
	if registrationData.Staked {
		return vmcommon.Ok
	}

	s.eei.AddReturnMessage("account not staked")
	return vmcommon.UserError
}

func (s *stakingSC) tryRemoveJailedNodeFromStaked(registrationData *StakedDataV2_0) {
	if !s.enableEpochsHandler.IsFlagEnabled(common.CorrectJailedNotUnStakedEmptyQueueFlag) {
		s.removeAndSetUnstaked(registrationData)
		return
	}

	if s.canUnStake() {
		s.removeAndSetUnstaked(registrationData)
		return
	}

	s.eei.AddReturnMessage("did not switch as not enough validators remaining")
}

func (s *stakingSC) removeAndSetUnstaked(registrationData *StakedDataV2_0) {
	s.removeFromStakedNodes()
	registrationData.Staked = false
	registrationData.UnStakedEpoch = s.eei.BlockChainHook().CurrentEpoch()
	registrationData.UnStakedNonce = s.eei.BlockChainHook().CurrentNonce()
	registrationData.StakedNonce = math.MaxUint64
}

func (s *stakingSC) updateConfigMinNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		s.eei.AddReturnMessage("updateConfigMinNodes function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}

	stakeConfig := s.getConfig()
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("number of arguments must be 1")
		return vmcommon.UserError
	}

	newMinNodes := big.NewInt(0).SetBytes(args.Arguments[0]).Int64()
	if newMinNodes <= 0 {
		s.eei.AddReturnMessage("new minimum number of nodes zero or negative")
		return vmcommon.UserError
	}

	if newMinNodes > int64(s.maxNumNodes) {
		s.eei.AddReturnMessage("new minimum number of nodes greater than maximum number of nodes")
		return vmcommon.UserError
	}

	stakeConfig.MinNumNodes = newMinNodes
	s.setConfig(stakeConfig)

	return vmcommon.Ok
}

func (s *stakingSC) updateConfigMaxNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		s.eei.AddReturnMessage("updateConfigMaxNodes function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}

	stakeConfig := s.getConfig()
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("number of arguments must be 1")
		return vmcommon.UserError
	}

	newMaxNodes := big.NewInt(0).SetBytes(args.Arguments[0]).Int64()
	if newMaxNodes <= 0 {
		s.eei.AddReturnMessage("new max number of nodes zero or negative")
		return vmcommon.UserError
	}

	if newMaxNodes < int64(s.minNumNodes) {
		s.eei.AddReturnMessage("new max number of nodes less than min number of nodes")
		return vmcommon.UserError
	}

	prevMaxNumNodes := big.NewInt(stakeConfig.MaxNumNodes)
	s.eei.Finish(prevMaxNumNodes.Bytes())
	stakeConfig.MaxNumNodes = newMaxNodes
	s.setConfig(stakeConfig)

	return vmcommon.Ok
}

func (s *stakingSC) isNodeJailedOrWithBadRating(registrationData *StakedDataV2_0, blsKey []byte) bool {
	return registrationData.Jailed || s.eei.CanUnJail(blsKey) || s.eei.IsBadRating(blsKey)
}

func (s *stakingSC) getRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	stakedData, returnCode := s.getStakedDataIfExists(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	s.eei.Finish([]byte(hex.EncodeToString(stakedData.RewardAddress)))
	return vmcommon.Ok
}

func (s *stakingSC) getStakedDataIfExists(args *vmcommon.ContractCallInput) (*StakedDataV2_0, vmcommon.ReturnCode) {
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas")
		return nil, vmcommon.OutOfGas
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("number of arguments must be equal to 1")
		return nil, vmcommon.UserError
	}
	stakedData, err := s.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}
	if len(stakedData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("blsKey not registered in staking sc")
		return nil, vmcommon.UserError
	}

	return stakedData, vmcommon.Ok
}

func (s *stakingSC) getBLSKeyStatus(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	stakedData, returnCode := s.getStakedDataIfExists(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if stakedData.Jailed || s.eei.CanUnJail(args.Arguments[0]) {
		s.eei.Finish([]byte("jailed"))
		return vmcommon.Ok
	}
	if stakedData.Waiting {
		s.eei.Finish([]byte("queued"))
		return vmcommon.Ok
	}
	if stakedData.Staked {
		s.eei.Finish([]byte("staked"))
		return vmcommon.Ok
	}

	s.eei.Finish([]byte("unStaked"))
	return vmcommon.Ok
}

func (s *stakingSC) getTotalNumberOfRegisteredNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	waitingListHead, err := s.getWaitingListHead()
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	stakeConfig := s.getConfig()
	totalRegistered := stakeConfig.StakedNodes + stakeConfig.JailedNodes + int64(waitingListHead.Length)
	s.eei.Finish(big.NewInt(totalRegistered).Bytes())
	return vmcommon.Ok
}

func (s *stakingSC) getRemainingUnbondPeriod(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	stakedData, returnCode := s.getStakedDataIfExists(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if stakedData.UnStakedNonce == 0 {
		s.eei.AddReturnMessage("not in unbond period")
		return vmcommon.UserError
	}

	currentNonce := s.eei.BlockChainHook().CurrentNonce()
	passedNonce := currentNonce - stakedData.UnStakedNonce
	if passedNonce >= s.unBondPeriod {
		if s.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
			s.eei.Finish(zero.Bytes())
		} else {
			s.eei.Finish([]byte("0"))
		}
	} else {
		remaining := s.unBondPeriod - passedNonce
		if s.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
			s.eei.Finish(big.NewInt(0).SetUint64(remaining).Bytes())
		} else {
			s.eei.Finish([]byte(strconv.Itoa(int(remaining))))
		}
	}

	return vmcommon.Ok
}

func (s *stakingSC) setOwnersOnAddresses(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		s.eei.AddReturnMessage("setOwnersOnAddresses function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments)%2 != 0 {
		s.eei.AddReturnMessage("invalid number of arguments: expected an even number of arguments")
		return vmcommon.UserError
	}
	for i := 0; i < len(args.Arguments); i += 2 {
		stakedData, err := s.getOrCreateRegisteredData(args.Arguments[i])
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			s.eei.AddReturnMessage(fmt.Sprintf("process stopped at index %d, bls key %s", i, hex.EncodeToString(args.Arguments[i])))
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			log.Error("staking data does not exists",
				"bls key", hex.EncodeToString(args.Arguments[i]),
				"owner as hex", hex.EncodeToString(args.Arguments[i+1]))
			continue
		}

		stakedData.OwnerAddress = args.Arguments[i+1]
		err = s.saveStakingData(args.Arguments[i], stakedData)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			s.eei.AddReturnMessage(fmt.Sprintf("process stopped at index %d, bls key %s", i, hex.EncodeToString(args.Arguments[i])))
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (s *stakingSC) getOwner(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("this is only a view function")
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, len(args.Arguments)))
		return vmcommon.UserError
	}

	stakedData, errGet := s.getOrCreateRegisteredData(args.Arguments[0])
	if errGet != nil {
		s.eei.AddReturnMessage(errGet.Error())
		return vmcommon.UserError
	}
	if len(stakedData.OwnerAddress) == 0 {
		s.eei.AddReturnMessage("owner address is nil")
		return vmcommon.UserError
	}

	s.eei.Finish(stakedData.OwnerAddress)
	return vmcommon.Ok
}

func (s *stakingSC) changeOwnerAndRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.ValidatorToDelegationFlag) {
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("change owner and reward address can be called by validator SC only")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		s.eei.AddReturnMessage("number of arguments is 2 at minimum")
		return vmcommon.UserError
	}

	newOwnerAddress := args.Arguments[0]
	if !core.IsSmartContractAddress(newOwnerAddress) {
		s.eei.AddReturnMessage("new address must be a smart contract address")
		return vmcommon.UserError
	}

	for i := 1; i < len(args.Arguments); i++ {
		blsKey := args.Arguments[i]
		registrationData, err := s.getOrCreateRegisteredData(blsKey)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		if len(registrationData.RewardAddress) == 0 {
			s.eei.AddReturnMessage("cannot change owner and reward address for a key which is not registered")
			return vmcommon.UserError
		}

		if registrationData.Jailed || s.eei.CanUnJail(blsKey) {
			s.eei.AddReturnMessage("can not migrate nodes while jailed nodes exists")
			return vmcommon.UserError
		}

		registrationData.OwnerAddress = newOwnerAddress
		registrationData.RewardAddress = newOwnerAddress
		err = s.saveStakingData(blsKey, registrationData)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

type validatorFundInfo struct {
	numNodesToUnstake uint32
}

func (s *stakingSC) checkValidatorFunds(
	mapCheckedOwners map[string]*validatorFundInfo,
	owner []byte,
	nodePrice *big.Int,
) (*validatorFundInfo, error) {
	validatorInfo, okInMap := mapCheckedOwners[string(owner)]
	if okInMap {
		return validatorInfo, nil
	}

	marshaledData := s.eei.GetStorageFromAddress(s.stakeAccessAddr, owner)
	if len(marshaledData) == 0 {
		validatorInfo = &validatorFundInfo{
			numNodesToUnstake: math.MaxUint32,
		}
		mapCheckedOwners[string(owner)] = validatorInfo
		return validatorInfo, nil
	}

	validatorData := &ValidatorDataV2{}
	err := s.marshalizer.Unmarshal(validatorData, marshaledData)
	if err != nil {
		return nil, err
	}

	numRegisteredKeys := int64(len(validatorData.BlsPubKeys))
	numQualified := big.NewInt(0).Div(validatorData.TotalStakeValue, nodePrice).Int64()

	if numQualified >= numRegisteredKeys {
		validatorInfo = &validatorFundInfo{
			numNodesToUnstake: 0,
		}
		mapCheckedOwners[string(owner)] = validatorInfo
		return validatorInfo, nil
	}

	numRegisteredNodesWithMinStake := int64(0)
	for _, blsKey := range validatorData.BlsPubKeys {
		stakedData, errGet := s.getOrCreateRegisteredData(blsKey)
		if errGet != nil {
			return nil, errGet
		}

		if stakedData.Staked || stakedData.Waiting || stakedData.Jailed {
			numRegisteredNodesWithMinStake++
		}
	}

	numToUnStake := uint32(0)
	if numRegisteredNodesWithMinStake > numQualified {
		numToUnStake = uint32(numRegisteredNodesWithMinStake - numQualified)
	}

	validatorInfo = &validatorFundInfo{
		numNodesToUnstake: numToUnStake,
	}
	mapCheckedOwners[string(owner)] = validatorInfo

	return validatorInfo, nil
}

// CanUseContract returns true if contract can be used
func (s *stakingSC) CanUseContract() bool {
	return true
}

// SetNewGasCost is called whenever a gas cost was changed
func (s *stakingSC) SetNewGasCost(gasCost vm.GasCost) {
	s.mutExecution.Lock()
	s.gasCost = gasCost
	s.mutExecution.Unlock()
}

// IsInterfaceNil verifies if the underlying object is nil or not
func (s *stakingSC) IsInterfaceNil() bool {
	return s == nil
}
