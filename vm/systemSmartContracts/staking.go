//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. staking.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
)

var log = logger.GetOrCreate("vm/systemsmartcontracts")

const ownerKey = "owner"
const nodesConfigKey = "nodesConfig"
const waitingListHeadKey = "waitingList"
const waitingElementPrefix = "w_"

type stakingSC struct {
	eei                      vm.SystemEI
	unBondPeriod             uint64
	stakeAccessAddr          []byte //TODO add a viewAddress field and use it on all system SC view functions
	jailAccessAddr           []byte
	endOfEpochAccessAddr     []byte
	numRoundsWithoutBleed    uint64
	bleedPercentagePerRound  float64
	maximumPercentageToBleed float64
	gasCost                  vm.GasCost
	minNumNodes              uint64
	maxNumNodes              uint64
	marshalizer              marshal.Marshalizer
	enableStakingEpoch       uint32
	stakeValue               *big.Int
	flagEnableStaking        atomic.Flag
	flagStakingV2            atomic.Flag
	stakingV2Epoch           uint32
	walletAddressLen         int
	mutExecution             sync.RWMutex
	minNodePrice             *big.Int
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
	EpochNotifier        vm.EpochNotifier
}

type waitingListReturnData struct {
	blsKeys         [][]byte
	stakedDataList  []*StakedDataV2_0
	lastKey         []byte
	afterLastjailed bool
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
	if args.StakingSCConfig.MaxNumberOfNodesForStake < 0 {
		return nil, vm.ErrNegativeWaitingNodesPercentage
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
	if check.IfNil(args.EpochNotifier) {
		return nil, vm.ErrNilEpochNotifier
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
		enableStakingEpoch:       args.StakingSCConfig.StakeEnableEpoch,
		stakingV2Epoch:           args.StakingSCConfig.StakingV2Epoch,
		walletAddressLen:         len(args.StakingAccessAddr),
		minNodePrice:             minStakeValue,
	}

	conversionOk := true
	reg.stakeValue, conversionOk = big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, conversionBase)
	if !conversionOk || reg.stakeValue.Cmp(zero) < 0 {
		return nil, vm.ErrNegativeInitialStakeValue
	}

	args.EpochNotifier.RegisterNotifyHandler(reg)

	return reg, nil
}

// Execute calls one of the functions from the staking smart contract and runs the code according to the input
func (s *stakingSC) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	s.mutExecution.RLock()
	defer s.mutExecution.RUnlock()
	if CheckIfNil(args) != nil {
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
	stakeConfig := s.getConfig()
	return stakeConfig.StakedNodes < stakeConfig.MaxNumNodes
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
	if len(args.Arguments) < 2 {
		s.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 2, len(args.Arguments)))
		return vmcommon.UserError
	}

	oldKey := args.Arguments[0]
	newKey := args.Arguments[1]
	if len(oldKey) != len(newKey) {
		s.eei.AddReturnMessage("invalid bls key")
		return vmcommon.UserError
	}

	stakedData, err := s.getOrCreateRegisteredData(oldKey)
	if err != nil {
		s.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(stakedData.RewardAddress) == 0 {
		// if not registered this is not an error
		return vmcommon.Ok
	}

	s.eei.SetStorage(oldKey, nil)
	err = s.saveStakingData(newKey, stakedData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
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
	if !s.flagEnableStaking.IsSet() {
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

func (s *stakingSC) processStake(blsKey []byte, registrationData *StakedDataV2_0, addFirst bool) error {
	if registrationData.Staked {
		return nil
	}

	registrationData.RegisterNonce = s.eei.BlockChainHook().CurrentNonce()
	if !s.canStake() {
		s.eei.AddReturnMessage(fmt.Sprintf("staking is full key put into waiting list %s", hex.EncodeToString(blsKey)))
		err := s.addToWaitingList(blsKey, addFirst)
		if err != nil {
			s.eei.AddReturnMessage("error while adding to waiting")
			return err
		}
		registrationData.Waiting = true
		s.eei.Finish([]byte{waiting})
		return nil
	}

	err := s.removeFromWaitingList(blsKey)
	if err != nil {
		s.eei.AddReturnMessage("error while removing from waiting")
		return err
	}
	s.addToStakedNodes(1)
	s.activeStakingFor(registrationData)

	return nil
}

func (s *stakingSC) activeStakingFor(stakingData *StakedDataV2_0) {
	stakingData.RegisterNonce = s.eei.BlockChainHook().CurrentNonce()
	stakingData.Staked = true
	stakingData.StakedNonce = s.eei.BlockChainHook().CurrentNonce()
	stakingData.UnStakedEpoch = core.DefaultUnstakedEpoch
	stakingData.UnStakedNonce = 0
	stakingData.Waiting = false
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

func (s *stakingSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("unStake function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		s.eei.AddReturnMessage("not enough arguments, needed BLS key and reward address")
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
	if !bytes.Equal(args.Arguments[1], registrationData.RewardAddress) {
		s.eei.AddReturnMessage("unStake possible only from staker caller")
		return vmcommon.UserError
	}
	if s.isNodeJailedOrWithBadRating(registrationData, args.Arguments[0]) {
		s.eei.AddReturnMessage("cannot unStake node which is jailed or with bad rating")
		return vmcommon.UserError
	}

	if !registrationData.Staked && !registrationData.Waiting {
		s.eei.AddReturnMessage("cannot unStake node which was already unStaked")
		return vmcommon.UserError
	}

	if !registrationData.Staked {
		registrationData.Waiting = false
		err = s.removeFromWaitingList(args.Arguments[0])
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
		err = s.saveStakingData(args.Arguments[0], registrationData)
		if err != nil {
			s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
			return vmcommon.UserError
		}

		return vmcommon.Ok
	}

	_, err = s.moveFirstFromWaitingToStakedIfNeeded(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if !s.canUnStake() {
		s.eei.AddReturnMessage("unStake is not possible as too many left")
		return vmcommon.UserError
	}

	s.removeFromStakedNodes()
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

func (s *stakingSC) moveFirstFromWaitingToStakedIfNeeded(blsKey []byte) (bool, error) {
	waitingElementKey := s.createWaitingListKey(blsKey)
	elementInList, err := s.getWaitingListElement(waitingElementKey)
	if err == nil {
		// node in waiting - remove from it - and that's it
		return false, s.removeFromWaitingList(blsKey)
	}

	waitingList, err := s.getWaitingListHead()
	if err != nil {
		return false, err
	}
	if waitingList.Length == 0 {
		return false, nil
	}
	elementInList, err = s.getWaitingListElement(waitingList.FirstKey)
	if err != nil {
		return false, err
	}
	err = s.removeFromWaitingList(elementInList.BLSPublicKey)
	if err != nil {
		return false, err
	}

	nodeData, err := s.getOrCreateRegisteredData(elementInList.BLSPublicKey)
	if err != nil {
		return false, err
	}
	if len(nodeData.RewardAddress) == 0 || nodeData.Staked {
		return false, vm.ErrInvalidWaitingList
	}

	nodeData.Waiting = false
	nodeData.Staked = true
	nodeData.RegisterNonce = s.eei.BlockChainHook().CurrentNonce()
	nodeData.StakedNonce = s.eei.BlockChainHook().CurrentNonce()
	nodeData.UnStakedNonce = 0
	nodeData.UnStakedEpoch = core.DefaultUnstakedEpoch

	s.addToStakedNodes(1)
	return true, s.saveStakingData(elementInList.BLSPublicKey, nodeData)
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

func (s *stakingSC) addToWaitingList(blsKey []byte, addJailed bool) error {
	inWaitingListKey := s.createWaitingListKey(blsKey)
	marshaledData := s.eei.GetStorage(inWaitingListKey)
	if len(marshaledData) != 0 {
		return nil
	}

	waitingList, err := s.getWaitingListHead()
	if err != nil {
		return err
	}

	waitingList.Length += 1
	if waitingList.Length == 1 {
		waitingList.FirstKey = inWaitingListKey
		waitingList.LastKey = inWaitingListKey
		if addJailed {
			waitingList.LastJailedKey = inWaitingListKey
		}

		elementInWaiting := &ElementInList{
			BLSPublicKey: blsKey,
			PreviousKey:  waitingList.LastKey,
			NextKey:      make([]byte, 0),
		}
		return s.saveElementAndList(inWaitingListKey, elementInWaiting, waitingList)
	}

	if addJailed {
		return s.insertAfterLastJailed(waitingList, blsKey)
	}

	return s.addToEndOfTheList(waitingList, blsKey)
}

func (s *stakingSC) addToEndOfTheList(waitingList *WaitingList, blsKey []byte) error {
	inWaitingListKey := s.createWaitingListKey(blsKey)
	oldLastKey := make([]byte, len(waitingList.LastKey))
	copy(oldLastKey, waitingList.LastKey)

	lastElement, err := s.getWaitingListElement(waitingList.LastKey)
	if err != nil {
		return err
	}
	lastElement.NextKey = inWaitingListKey
	elementInWaiting := &ElementInList{
		BLSPublicKey: blsKey,
		PreviousKey:  oldLastKey,
		NextKey:      make([]byte, 0),
	}

	err = s.saveWaitingListElement(oldLastKey, lastElement)
	if err != nil {
		return err
	}

	waitingList.LastKey = inWaitingListKey
	return s.saveElementAndList(inWaitingListKey, elementInWaiting, waitingList)
}

func (s *stakingSC) insertAfterLastJailed(
	waitingList *WaitingList,
	blsKey []byte,
) error {
	inWaitingListKey := s.createWaitingListKey(blsKey)
	if len(waitingList.LastJailedKey) == 0 {
		nextKey := make([]byte, len(waitingList.FirstKey))
		copy(nextKey, waitingList.FirstKey)
		waitingList.FirstKey = inWaitingListKey
		waitingList.LastJailedKey = inWaitingListKey
		elementInWaiting := &ElementInList{
			BLSPublicKey: blsKey,
			PreviousKey:  inWaitingListKey,
			NextKey:      nextKey,
		}
		return s.saveElementAndList(inWaitingListKey, elementInWaiting, waitingList)
	}

	lastJailedElement, err := s.getWaitingListElement(waitingList.LastJailedKey)
	if err != nil {
		return err
	}

	if bytes.Equal(waitingList.LastKey, waitingList.LastJailedKey) {
		waitingList.LastJailedKey = inWaitingListKey
		return s.addToEndOfTheList(waitingList, blsKey)
	}

	firstNonJailedElement, err := s.getWaitingListElement(lastJailedElement.NextKey)
	if err != nil {
		return err
	}

	elementInWaiting := &ElementInList{
		BLSPublicKey: blsKey,
		PreviousKey:  make([]byte, len(inWaitingListKey)),
		NextKey:      make([]byte, len(inWaitingListKey)),
	}
	copy(elementInWaiting.PreviousKey, waitingList.LastJailedKey)
	copy(elementInWaiting.NextKey, lastJailedElement.NextKey)

	lastJailedElement.NextKey = inWaitingListKey
	firstNonJailedElement.PreviousKey = inWaitingListKey
	waitingList.LastJailedKey = inWaitingListKey

	err = s.saveWaitingListElement(elementInWaiting.PreviousKey, lastJailedElement)
	if err != nil {
		return err
	}
	err = s.saveWaitingListElement(elementInWaiting.NextKey, firstNonJailedElement)
	if err != nil {
		return err
	}
	err = s.saveWaitingListElement(inWaitingListKey, elementInWaiting)
	if err != nil {
		return err
	}
	return s.saveWaitingListHead(waitingList)
}

func (s *stakingSC) saveElementAndList(key []byte, element *ElementInList, waitingList *WaitingList) error {
	err := s.saveWaitingListElement(key, element)
	if err != nil {
		return err
	}

	return s.saveWaitingListHead(waitingList)
}

func (s *stakingSC) removeFromWaitingList(blsKey []byte) error {
	inWaitingListKey := s.createWaitingListKey(blsKey)
	marshaledData := s.eei.GetStorage(inWaitingListKey)
	if len(marshaledData) == 0 {
		return nil
	}
	s.eei.SetStorage(inWaitingListKey, nil)

	elementToRemove := &ElementInList{}
	err := s.marshalizer.Unmarshal(elementToRemove, marshaledData)
	if err != nil {
		return err
	}

	waitingList, err := s.getWaitingListHead()
	if err != nil {
		return err
	}
	if waitingList.Length == 0 {
		return vm.ErrInvalidWaitingList
	}
	waitingList.Length -= 1
	if waitingList.Length == 0 {
		s.eei.SetStorage([]byte(waitingListHeadKey), nil)
		return nil
	}

	if bytes.Equal(elementToRemove.PreviousKey, inWaitingListKey) {
		if bytes.Equal(inWaitingListKey, waitingList.LastJailedKey) {
			waitingList.LastJailedKey = make([]byte, 0)
		}

		nextElement, errGet := s.getWaitingListElement(elementToRemove.NextKey)
		if errGet != nil {
			return errGet
		}

		nextElement.PreviousKey = elementToRemove.NextKey
		waitingList.FirstKey = elementToRemove.NextKey
		return s.saveElementAndList(elementToRemove.NextKey, nextElement, waitingList)
	}

	waitingList.LastJailedKey = make([]byte, len(elementToRemove.PreviousKey))
	copy(waitingList.LastJailedKey, elementToRemove.PreviousKey)
	previousElement, err := s.getWaitingListElement(elementToRemove.PreviousKey)
	if err != nil {
		return err
	}
	if len(elementToRemove.NextKey) == 0 {
		waitingList.LastKey = elementToRemove.PreviousKey
		previousElement.NextKey = make([]byte, 0)
		return s.saveElementAndList(elementToRemove.PreviousKey, previousElement, waitingList)
	}

	nextElement, err := s.getWaitingListElement(elementToRemove.NextKey)
	if err != nil {
		return err
	}

	nextElement.PreviousKey = elementToRemove.PreviousKey
	previousElement.NextKey = elementToRemove.NextKey

	err = s.saveWaitingListElement(elementToRemove.NextKey, nextElement)
	if err != nil {
		return err
	}
	return s.saveElementAndList(elementToRemove.PreviousKey, previousElement, waitingList)
}

func (s *stakingSC) getWaitingListElement(key []byte) (*ElementInList, error) {
	marshaledData := s.eei.GetStorage(key)
	if len(marshaledData) == 0 {
		return nil, vm.ErrElementNotFound
	}

	element := &ElementInList{}
	err := s.marshalizer.Unmarshal(element, marshaledData)
	if err != nil {
		return nil, err
	}

	return element, nil
}

func (s *stakingSC) isInWaiting(blsKey []byte) bool {
	waitingKey := s.createWaitingListKey(blsKey)
	marshaledData := s.eei.GetStorage(waitingKey)
	return len(marshaledData) > 0
}

func (s *stakingSC) saveWaitingListElement(key []byte, element *ElementInList) error {
	marshaledData, err := s.marshalizer.Marshal(element)
	if err != nil {
		return err
	}

	s.eei.SetStorage(key, marshaledData)
	return nil
}

func (s *stakingSC) getWaitingListHead() (*WaitingList, error) {
	waitingList := &WaitingList{
		FirstKey:      make([]byte, 0),
		LastKey:       make([]byte, 0),
		Length:        0,
		LastJailedKey: make([]byte, 0),
	}
	marshaledData := s.eei.GetStorage([]byte(waitingListHeadKey))
	if len(marshaledData) == 0 {
		return waitingList, nil
	}

	err := s.marshalizer.Unmarshal(waitingList, marshaledData)
	if err != nil {
		return nil, err
	}

	return waitingList, nil
}

func (s *stakingSC) saveWaitingListHead(waitingList *WaitingList) error {
	marshaledData, err := s.marshalizer.Marshal(waitingList)
	if err != nil {
		return err
	}

	s.eei.SetStorage([]byte(waitingListHeadKey), marshaledData)
	return nil
}

func (s *stakingSC) createWaitingListKey(blsKey []byte) []byte {
	return []byte(waitingElementPrefix + string(blsKey))
}

func (s *stakingSC) switchJailedWithWaiting(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		s.eei.AddReturnMessage("switchJailedWithWaiting function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		return vmcommon.UserError
	}

	registrationData, err := s.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("no need to jail as not a validator")
		return vmcommon.UserError
	}
	if !registrationData.Staked {
		s.eei.AddReturnMessage("no need to jail as not a validator")
		return vmcommon.UserError
	}
	if registrationData.Jailed {
		s.eei.AddReturnMessage(vm.ErrBLSPublicKeyAlreadyJailed.Error())
		return vmcommon.UserError
	}
	switched, err := s.moveFirstFromWaitingToStakedIfNeeded(args.Arguments[0])
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	registrationData.NumJailed++
	registrationData.Jailed = true
	registrationData.JailedNonce = s.eei.BlockChainHook().CurrentNonce()
	if !switched {
		s.eei.AddReturnMessage("did not switch as nobody in waiting, but jailed")
	} else {
		s.removeFromStakedNodes()
		registrationData.Staked = false
		registrationData.UnStakedEpoch = s.eei.BlockChainHook().CurrentEpoch()
		registrationData.UnStakedNonce = s.eei.BlockChainHook().CurrentNonce()
		registrationData.StakedNonce = math.MaxUint64
	}

	err = s.saveStakingData(args.Arguments[0], registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
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
	if !s.flagStakingV2.IsSet() {
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

func (s *stakingSC) getWaitingListIndex(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("this is only a view function")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("number of arguments must be equal to 1")
		return vmcommon.UserError
	}

	waitingElementKey := s.createWaitingListKey(args.Arguments[0])
	_, err := s.getWaitingListElement(waitingElementKey)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	waitingListHead, err := s.getWaitingListHead()
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if bytes.Equal(waitingElementKey, waitingListHead.FirstKey) {
		s.eei.Finish([]byte(strconv.Itoa(1)))
		return vmcommon.Ok
	}
	if bytes.Equal(waitingElementKey, waitingListHead.LastKey) {
		s.eei.Finish([]byte(strconv.Itoa(int(waitingListHead.Length))))
		return vmcommon.Ok
	}

	prevElement, err := s.getWaitingListElement(waitingListHead.FirstKey)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	index := uint32(2)
	nextKey := make([]byte, len(waitingElementKey))
	copy(nextKey, prevElement.NextKey)
	for len(nextKey) != 0 && index <= waitingListHead.Length {
		if bytes.Equal(nextKey, waitingElementKey) {
			s.eei.Finish([]byte(strconv.Itoa(int(index))))
			return vmcommon.Ok
		}

		prevElement, err = s.getWaitingListElement(nextKey)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		index++
		copy(nextKey, prevElement.NextKey)
	}

	s.eei.AddReturnMessage("element in waiting list not found")
	return vmcommon.UserError
}

func (s *stakingSC) getWaitingListSize(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas")
		return vmcommon.OutOfGas
	}

	waitingListHead, err := s.getWaitingListHead()
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	s.eei.Finish([]byte(strconv.Itoa(int(waitingListHead.Length))))
	return vmcommon.Ok
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
		if s.flagStakingV2.IsSet() {
			s.eei.Finish(zero.Bytes())
		} else {
			s.eei.Finish([]byte("0"))
		}
	} else {
		remaining := s.unBondPeriod - passedNonce
		if s.flagStakingV2.IsSet() {
			s.eei.Finish(big.NewInt(0).SetUint64(remaining).Bytes())
		} else {
			s.eei.Finish([]byte(strconv.Itoa(int(remaining))))
		}
	}

	return vmcommon.Ok
}

func (s *stakingSC) getWaitingListRegisterNonceAndRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("this is only a view function")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 0 {
		s.eei.AddReturnMessage("number of arguments must be equal to 0")
		return vmcommon.UserError
	}

	waitingListData, err := s.getFirstElementsFromWaitingList(math.MaxUint32)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if len(waitingListData.stakedDataList) == 0 {
		s.eei.AddReturnMessage("no one in waitingList")
		return vmcommon.UserError
	}

	for _, stakedData := range waitingListData.stakedDataList {
		s.eei.Finish([]byte(hex.EncodeToString(stakedData.RewardAddress)))
		s.eei.Finish([]byte(strconv.Itoa(int(stakedData.RegisterNonce))))
	}

	return vmcommon.Ok
}

func (s *stakingSC) setOwnersOnAddresses(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.flagStakingV2.IsSet() {
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
	if !s.flagStakingV2.IsSet() {
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

	s.eei.Finish(stakedData.OwnerAddress)
	return vmcommon.Ok
}

func (s *stakingSC) getTotalNumberOfRegisteredNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.flagStakingV2.IsSet() {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	stakeConfig := s.getConfig()
	waitingListHead, err := s.getWaitingListHead()
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	totalRegistered := stakeConfig.StakedNodes + stakeConfig.JailedNodes + int64(waitingListHead.Length)
	s.eei.Finish(big.NewInt(totalRegistered).Bytes())
	return vmcommon.Ok
}

func (s *stakingSC) stakeNodesFromQueue(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.flagStakingV2.IsSet() {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		s.eei.AddReturnMessage("stake nodes from waiting list can be called by endOfEpochAccess address only")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("number of arguments must be equal to 1")
		return vmcommon.UserError
	}

	numNodesToStake := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	waitingListData, err := s.getFirstElementsFromWaitingList(math.MaxUint32)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if len(waitingListData.blsKeys) == 0 {
		s.eei.AddReturnMessage("no nodes in queue")
		return vmcommon.Ok
	}

	stakedNodes := uint64(0)
	mapCheckedOwners := make(map[string]bool)
	for i, blsKey := range waitingListData.blsKeys {
		stakedData := waitingListData.stakedDataList[i]
		if stakedNodes >= numNodesToStake {
			break
		}

		hasEnoughFunds, errCheck := s.checkValidatorFunds(mapCheckedOwners, stakedData.OwnerAddress)
		if errCheck != nil {
			s.eei.AddReturnMessage(errCheck.Error())
			return vmcommon.UserError
		}
		if !hasEnoughFunds {
			continue
		}

		s.activeStakingFor(stakedData)
		err = s.saveStakingData(blsKey, stakedData)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		// remove from waiting list
		err = s.removeFromWaitingList(blsKey)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		stakedNodes++
		// return the change key
		s.eei.Finish(blsKey)
		s.eei.Finish(stakedData.RewardAddress)
	}

	s.addToStakedNodes(int64(stakedNodes))

	return vmcommon.Ok
}

func (s *stakingSC) checkValidatorFunds(
	mapCheckedOwners map[string]bool,
	owner []byte,
) (bool, error) {
	hasFunds, okInMap := mapCheckedOwners[string(owner)]
	if okInMap {
		return hasFunds, nil
	}

	marshaledData := s.eei.GetStorageFromAddress(s.stakeAccessAddr, owner)
	if len(marshaledData) == 0 {
		mapCheckedOwners[string(owner)] = false
		return false, nil
	}

	validatorData := &ValidatorDataV2{}
	err := s.marshalizer.Unmarshal(validatorData, marshaledData)
	if err != nil {
		return false, err
	}

	numRegisteredKeys := int64(len(validatorData.BlsPubKeys))
	numQualified := big.NewInt(0).Div(validatorData.TotalStakeValue, s.minNodePrice).Int64()

	if numQualified < numRegisteredKeys {
		mapCheckedOwners[string(owner)] = false
		return false, nil
	}

	mapCheckedOwners[string(owner)] = true

	return true, nil
}

func (s *stakingSC) getFirstElementsFromWaitingList(numNodes uint32) (*waitingListReturnData, error) {
	waitingListData := &waitingListReturnData{}

	waitingListHead, err := s.getWaitingListHead()
	if err != nil {
		return nil, err
	}
	if waitingListHead.Length == 0 {
		return waitingListData, nil
	}

	blsKeysToStake := make([][]byte, 0)
	stakedDataList := make([]*StakedDataV2_0, 0)
	index := uint32(1)
	nextKey := make([]byte, len(waitingListHead.FirstKey))
	copy(nextKey, waitingListHead.FirstKey)
	for len(nextKey) != 0 && index <= waitingListHead.Length && index <= numNodes {
		element, errGet := s.getWaitingListElement(nextKey)
		if errGet != nil {
			return nil, errGet
		}

		if bytes.Equal(nextKey, waitingListHead.LastJailedKey) {
			waitingListData.afterLastjailed = true
		}

		stakedData, errGet := s.getOrCreateRegisteredData(element.BLSPublicKey)
		if errGet != nil {
			return nil, errGet
		}

		blsKeysToStake = append(blsKeysToStake, element.BLSPublicKey)
		stakedDataList = append(stakedDataList, stakedData)
		index++
		copy(nextKey, element.NextKey)
	}

	waitingListData.blsKeys = blsKeysToStake
	waitingListData.stakedDataList = stakedDataList
	waitingListData.lastKey = nextKey
	return waitingListData, nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (s *stakingSC) EpochConfirmed(epoch uint32) {
	s.flagEnableStaking.Toggle(epoch >= s.enableStakingEpoch)
	log.Debug("stakingSC: stake/unstake/unbond", "enabled", s.flagEnableStaking.IsSet())

	s.flagStakingV2.Toggle(epoch >= s.stakingV2Epoch)
	log.Debug("stakingSC: set owner", "enabled", s.flagStakingV2.IsSet())
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
