package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("vm/systemsmartcontracts")

const OwnerKey = "owner"

type stakingSC struct {
	eei                      vm.SystemEI
	minStakeValue            *big.Int
	unBondPeriod             uint64
	stakeAccessAddr          []byte
	jailAccessAddr           []byte
	numRoundsWithoutBleed    uint64
	bleedPercentagePerRound  float64
	maximumPercentageToBleed float64
	gasCost                  vm.GasCost
}

// StakedData represents the data which is saved for the selected nodes
type StakedData struct {
	RegisterNonce uint64   `json:"RegisterNonce"`
	Staked        bool     `json:"Staked"`
	UnStakedNonce uint64   `json:"UnStakedNonce"`
	UnStakedEpoch uint32   `json:"UnStakedEpoch"`
	RewardAddress []byte   `json:"RewardAddress"`
	StakeValue    *big.Int `json:"StakeValue"`
	JailedRound   uint64   `json:"JailedRound"`
}

// ArgsNewStakingSmartContract holds the arguments needed to create a StakingSmartContract
type ArgsNewStakingSmartContract struct {
	MinStakeValue            *big.Int
	UnBondPeriod             uint64
	Eei                      vm.SystemEI
	StakingAccessAddr        []byte
	JailAccessAddr           []byte
	NumRoundsWithoutBleed    uint64
	BleedPercentagePerRound  float64
	MaximumPercentageToBleed float64
	GasCost                  vm.GasCost
}

// NewStakingSmartContract creates a staking smart contract
func NewStakingSmartContract(
	args ArgsNewStakingSmartContract,
) (*stakingSC, error) {
	if args.MinStakeValue == nil {
		return nil, vm.ErrNilInitialStakeValue
	}
	if args.MinStakeValue.Cmp(big.NewInt(0)) < 1 {
		return nil, vm.ErrNegativeInitialStakeValue
	}
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.StakingAccessAddr) < 1 {
		return nil, vm.ErrInvalidStakingAccessAddress
	}
	if len(args.JailAccessAddr) < 1 {
		return nil, vm.ErrInvalidJailAccessAddress
	}

	reg := &stakingSC{
		minStakeValue:            big.NewInt(0).Set(args.MinStakeValue),
		eei:                      args.Eei,
		unBondPeriod:             args.UnBondPeriod,
		stakeAccessAddr:          args.StakingAccessAddr,
		jailAccessAddr:           args.JailAccessAddr,
		numRoundsWithoutBleed:    args.NumRoundsWithoutBleed,
		bleedPercentagePerRound:  args.BleedPercentagePerRound,
		maximumPercentageToBleed: args.MaximumPercentageToBleed,
		gasCost:                  args.GasCost,
	}
	return reg, nil
}

// Execute calls one of the functions from the staking smart contract and runs the code according to the input
func (r *stakingSC) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	switch args.Function {
	case "_init":
		return r.init(args)
	case "stake":
		return r.stake(args, false)
	case "register":
		return r.stake(args, true)
	case "unStake":
		return r.unStake(args)
	case "unBond":
		return r.unBond(args)
	case "slash":
		return r.slash(args)
	case "get":
		return r.get(args)
	case "isStaked":
		return r.isStaked(args)
	case "setStakeValue":
		return r.setStakeValueForCurrentEpoch(args)
	case "jail":
		return r.jail(args)
	case "unJail":
		return r.unJail(args)
	case "changeRewardAddress":
		return r.changeRewardAddress(args)
	case "changeValidatorKeys":
		return r.changeValidatorKey(args)
	}

	return vmcommon.UserError
}

func getPercentageOfValue(value *big.Int, percentage float64) *big.Int {
	x := new(big.Float).SetInt(value)
	y := big.NewFloat(percentage)

	z := new(big.Float).Mul(x, y)

	op := big.NewInt(0)
	result, _ := z.Int(op)

	return result
}

func (r *stakingSC) calculateStakeAfterBleed(startRound uint64, endRound uint64, stake *big.Int) *big.Int {
	if startRound > endRound {
		return stake
	}
	if endRound-startRound < r.numRoundsWithoutBleed {
		return stake
	}

	totalRoundsToBleed := endRound - startRound - r.numRoundsWithoutBleed
	totalPercentageToBleed := float64(totalRoundsToBleed) * r.bleedPercentagePerRound
	totalPercentageToBleed = math.Min(totalPercentageToBleed, r.maximumPercentageToBleed)

	bleedValue := getPercentageOfValue(stake, totalPercentageToBleed)
	stakeAfterBleed := big.NewInt(0).Sub(stake, bleedValue)

	if stakeAfterBleed.Cmp(big.NewInt(0)) < 0 {
		stakeAfterBleed = big.NewInt(0)
	}

	return stakeAfterBleed
}

func (r *stakingSC) changeValidatorKey(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		log.Debug("stake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		return vmcommon.UserError
	}

	err := r.eei.UseGas(r.gasCost.MetaChainSystemSCsCost.ChangeValidatorKeys +
		r.gasCost.BaseOperationCost.DataCopyPerByte*uint64(len(args.Arguments[0])))
	if err != nil {
		return vmcommon.OutOfGas
	}

	oldKey := args.Arguments[0]
	newKey := args.Arguments[1]
	if len(oldKey) != len(newKey) {
		return vmcommon.UserError
	}

	stakedData, err := r.getOrCreateRegisteredData(oldKey)
	if err != nil {
		return vmcommon.UserError
	}
	if len(stakedData.RewardAddress) == 0 {
		// if not registered this is not an error
		return vmcommon.Ok
	}

	r.eei.SetStorage(oldKey, nil)
	err = r.saveStakingData(newKey, stakedData)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (r *stakingSC) changeRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		log.Debug("stake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		return vmcommon.UserError
	}

	newRewardAddress := args.Arguments[0]
	if len(newRewardAddress) != len(args.CallerAddr) {
		return vmcommon.UserError
	}

	for i := 1; i < len(args.Arguments); i++ {
		blsKey := args.Arguments[i]
		err := r.eei.UseGas(r.gasCost.BaseOperationCost.DataCopyPerByte * uint64(len(blsKey)))
		if err != nil {
			return vmcommon.OutOfGas
		}

		stakedData, err := r.getOrCreateRegisteredData(blsKey)
		if err != nil {
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			continue
		}

		err = r.eei.UseGas(r.gasCost.MetaChainSystemSCsCost.ChangeRewardAddress)
		if err != nil {
			return vmcommon.OutOfGas
		}

		stakedData.RewardAddress = newRewardAddress
		err = r.saveStakingData(blsKey, stakedData)
		if err != nil {
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (r *stakingSC) unJail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		log.Debug("stake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}

	for _, argument := range args.Arguments {
		stakedData, err := r.getOrCreateRegisteredData(argument)
		if err != nil {
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			return vmcommon.UserError
		}

		err = r.eei.UseGas(r.gasCost.MetaChainSystemSCsCost.UnJail)
		if err != nil {
			return vmcommon.OutOfGas
		}

		stakedData.StakeValue = r.calculateStakeAfterBleed(
			stakedData.JailedRound,
			r.eei.BlockChainHook().CurrentRound(),
			stakedData.StakeValue,
		)
		stakedData.JailedRound = math.MaxUint64

		err = r.saveStakingData(argument, stakedData)
		if err != nil {
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (r *stakingSC) getOrCreateRegisteredData(key []byte) (*StakedData, error) {
	registrationData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		UnStakedEpoch: 0,
		RewardAddress: nil,
		StakeValue:    big.NewInt(0),
		JailedRound:   math.MaxUint64,
	}

	data := r.eei.GetStorage(key)
	if data != nil {
		err := json.Unmarshal(data, &registrationData)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return nil, err
		}
	}

	return &registrationData, nil
}

func (r *stakingSC) saveStakingData(key []byte, stakedData *StakedData) error {
	data, err := json.Marshal(*stakedData)
	if err != nil {
		log.Debug("marshal error on staking SC stake function ",
			"error", err.Error(),
		)
		return err
	}

	r.eei.SetStorage(key, data)
	return nil
}

func (r *stakingSC) jail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.jailAccessAddr) {
		return vmcommon.UserError
	}

	for _, argument := range args.Arguments {
		stakedData, err := r.getOrCreateRegisteredData(argument)
		if err != nil {
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			return vmcommon.UserError
		}

		stakedData.JailedRound = r.eei.BlockChainHook().CurrentRound()
		err = r.saveStakingData(argument, stakedData)
		if err != nil {
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (r *stakingSC) get(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		return vmcommon.UserError
	}

	err := r.eei.UseGas(r.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		return vmcommon.OutOfGas
	}

	value := r.eei.GetStorage(args.Arguments[0])
	r.eei.Finish(value)

	return vmcommon.Ok
}

func (r *stakingSC) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(OwnerKey))
	if ownerAddress != nil {
		log.Debug("smart contract was already initialized")
		return vmcommon.UserError
	}

	r.eei.SetStorage([]byte(OwnerKey), args.CallerAddr)
	r.eei.SetStorage(args.CallerAddr, big.NewInt(0).Bytes())

	epoch := r.eei.BlockChainHook().CurrentEpoch()
	epochData := fmt.Sprintf("epoch_%d", epoch)

	r.eei.SetStorage([]byte(epochData), r.minStakeValue.Bytes())

	return vmcommon.Ok
}

func (r *stakingSC) setStakeValueForCurrentEpoch(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		log.Debug("stake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}

	if len(args.Arguments) < 1 {
		log.Debug("nil arguments to call setStakeValueForCurrentEpoch")
		return vmcommon.UserError
	}

	epoch := r.eei.BlockChainHook().CurrentEpoch()
	epochData := fmt.Sprintf("epoch_%d", epoch)

	inputStakeValue := big.NewInt(0).SetBytes(args.Arguments[0])
	if inputStakeValue.Cmp(r.minStakeValue) < 0 {
		inputStakeValue.Set(r.minStakeValue)
	}

	r.eei.SetStorage([]byte(epochData), inputStakeValue.Bytes())

	return vmcommon.Ok
}

func (r *stakingSC) getStakeValueForCurrentEpoch() *big.Int {
	stakeValue := big.NewInt(0)

	epoch := r.eei.BlockChainHook().CurrentEpoch()
	epochData := fmt.Sprintf("epoch_%d", epoch)

	stakeValueBytes := r.eei.GetStorage([]byte(epochData))
	stakeValue.SetBytes(stakeValueBytes)

	if stakeValue.Cmp(r.minStakeValue) < 0 {
		stakeValue.Set(r.minStakeValue)
	}

	return stakeValue
}

func (r *stakingSC) stake(args *vmcommon.ContractCallInput, onlyRegister bool) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		log.Debug("stake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		log.Debug("not enough arguments, needed BLS key and reward address")
		return vmcommon.UserError
	}

	stakeValue := r.getStakeValueForCurrentEpoch()
	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		return vmcommon.UserError
	}

	err = r.eei.UseGas(r.gasCost.MetaChainSystemSCsCost.Stake +
		r.gasCost.BaseOperationCost.StorePerByte*uint64(len(args.Arguments[0])) +
		r.gasCost.BaseOperationCost.StorePerByte*uint64(len(args.Arguments[1])))
	if err != nil {
		return vmcommon.OutOfGas
	}

	if registrationData.StakeValue.Cmp(stakeValue) < 0 {
		registrationData.StakeValue.Set(stakeValue)
	}

	if !onlyRegister {
		registrationData.Staked = true
	}

	registrationData.RegisterNonce = r.eei.BlockChainHook().CurrentNonce()
	registrationData.RewardAddress = args.Arguments[1]

	err = r.saveStakingData(args.Arguments[0], registrationData)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (r *stakingSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		log.Debug("unStake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		log.Debug("not enough arguments, needed BLS key and reward address")
		return vmcommon.UserError
	}

	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		return vmcommon.UserError
	}

	if !bytes.Equal(args.Arguments[1], registrationData.RewardAddress) {
		log.Debug("unStake possible only from staker",
			"caller", args.CallerAddr,
			"staker", registrationData.RewardAddress,
		)
		return vmcommon.UserError
	}

	if !registrationData.Staked {
		log.Error("unStake is not possible for address with is already unStaked")
		return vmcommon.UserError
	}
	if registrationData.JailedRound != math.MaxUint64 {
		log.Error("unStake is not possible for jailed nodes")
		return vmcommon.UserError
	}

	err = r.eei.UseGas(r.gasCost.MetaChainSystemSCsCost.UnStake)
	if err != nil {
		return vmcommon.OutOfGas
	}

	registrationData.Staked = false
	registrationData.UnStakedEpoch = r.eei.BlockChainHook().CurrentEpoch()
	registrationData.UnStakedNonce = r.eei.BlockChainHook().CurrentNonce()

	err = r.saveStakingData(args.Arguments[0], registrationData)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (r *stakingSC) unBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		log.Debug("unStake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		log.Debug("not enough arguments, needed BLS key and reward address")
		return vmcommon.UserError
	}

	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		return vmcommon.UserError
	}

	if registrationData.Staked || registrationData.UnStakedNonce <= registrationData.RegisterNonce {
		log.Debug("unBond is not possible for address which is staked or is not in unBond period")
		return vmcommon.UserError
	}

	currentNonce := r.eei.BlockChainHook().CurrentNonce()
	if currentNonce-registrationData.UnStakedNonce < r.unBondPeriod {
		log.Debug("unBond is not possible for address because unBond period did not pass")
		return vmcommon.UserError
	}
	if registrationData.JailedRound != math.MaxUint64 {
		log.Error("unBond is not possible for jailed nodes")
		return vmcommon.UserError
	}

	err = r.eei.UseGas(r.gasCost.MetaChainSystemSCsCost.UnBond)
	if err != nil {
		return vmcommon.OutOfGas
	}

	r.eei.SetStorage(args.Arguments[0], nil)
	r.eei.Finish(registrationData.StakeValue.Bytes())
	r.eei.Finish(big.NewInt(0).SetUint64(uint64(registrationData.UnStakedEpoch)).Bytes())

	return vmcommon.Ok
}

func (r *stakingSC) slash(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(OwnerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		log.Debug("slash function called by not the owners address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 2 {
		log.Debug("slash function called by wrong number of arguments")
		return vmcommon.UserError
	}

	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		return vmcommon.UserError
	}
	if !registrationData.Staked {
		log.Debug("cannot slash already unstaked or user not staked")
		return vmcommon.UserError
	}

	stakedValue := big.NewInt(0).Set(registrationData.StakeValue)
	slashValue := big.NewInt(0).SetBytes(args.Arguments[1])
	registrationData.StakeValue = registrationData.StakeValue.Sub(stakedValue, slashValue)
	registrationData.JailedRound = r.eei.BlockChainHook().CurrentRound()

	err = r.saveStakingData(args.Arguments[0], registrationData)
	if err != nil {
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (r *stakingSC) isStaked(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		return vmcommon.UserError
	}

	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		return vmcommon.UserError
	}

	err = r.eei.UseGas(r.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		return vmcommon.OutOfGas
	}

	if registrationData.Staked {
		log.Debug("account already staked, re-staking is invalid")
		return vmcommon.Ok
	}

	return vmcommon.UserError
}

// IsInterfaceNil verifies if the underlying object is nil or not
func (r *stakingSC) IsInterfaceNil() bool {
	return r == nil
}
