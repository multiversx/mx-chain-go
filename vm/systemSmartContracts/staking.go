package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("vm/systemsmartcontracts")

const ownerKey = "owner"

type stakingSC struct {
	eei           vm.SystemEI
	minStakeValue *big.Int
	unBondPeriod  uint64
	accessAddr    []byte
}

// StakedData represents the data which is saved for the selected nodes
type StakedData struct {
	RegisterNonce uint64   `json:"RegisterNonce"`
	Staked        bool     `json:"Staked"`
	UnStakedNonce uint64   `json:"UnStakedNonce"`
	UnStakedEpoch uint32   `json:"UnStakedEpoch"`
	RewardAddress []byte   `json:"RewardAddress"`
	StakeValue    *big.Int `json:"StakeValue"`
}

// NewStakingSmartContract creates a staking smart contract
func NewStakingSmartContract(
	minStakeValue *big.Int,
	unBondPeriod uint64,
	eei vm.SystemEI,
	accessAddr []byte,
) (*stakingSC, error) {
	if minStakeValue == nil {
		return nil, vm.ErrNilInitialStakeValue
	}
	if minStakeValue.Cmp(big.NewInt(0)) < 1 {
		return nil, vm.ErrNegativeInitialStakeValue
	}
	if check.IfNil(eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	reg := &stakingSC{
		minStakeValue: big.NewInt(0).Set(minStakeValue),
		eei:           eei,
		unBondPeriod:  unBondPeriod,
		accessAddr:    accessAddr,
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
	}

	return vmcommon.UserError
}

func (r *stakingSC) unJail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (r *stakingSC) jail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (r *stakingSC) get(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		return vmcommon.UserError
	}

	value := r.eei.GetStorage(args.Arguments[0])
	r.eei.Finish(value)

	return vmcommon.Ok
}

func (r *stakingSC) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(ownerKey))
	if ownerAddress != nil {
		log.Debug("smart contract was already initialized")
		return vmcommon.UserError
	}

	r.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	r.eei.SetStorage(args.CallerAddr, big.NewInt(0).Bytes())

	epoch := r.eei.BlockChainHook().CurrentEpoch()
	epochData := fmt.Sprintf("epoch_%d", epoch)

	r.eei.SetStorage([]byte(epochData), r.minStakeValue.Bytes())

	return vmcommon.Ok
}

func (r *stakingSC) setStakeValueForCurrentEpoch(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.accessAddr) {
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
	if !bytes.Equal(args.CallerAddr, r.accessAddr) {
		log.Debug("stake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		log.Debug("not enough arguments, needed BLS key and reward address")
		return vmcommon.UserError
	}

	stakeValue := r.getStakeValueForCurrentEpoch()
	registrationData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		RewardAddress: nil,
		UnStakedNonce: 0,
		UnStakedEpoch: 0,
		StakeValue:    big.NewInt(0).Set(stakeValue),
	}
	data := r.eei.GetStorage(args.Arguments[0])

	if data != nil {
		err := json.Unmarshal(data, &registrationData)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return vmcommon.UserError
		}

		if registrationData.StakeValue.Cmp(stakeValue) < 0 {
			registrationData.StakeValue.Set(stakeValue)
		}
	}

	if !onlyRegister {
		registrationData.Staked = true
	}

	registrationData.RegisterNonce = r.eei.BlockChainHook().CurrentNonce()
	registrationData.RewardAddress = args.Arguments[1]

	data, err := json.Marshal(registrationData)
	if err != nil {
		log.Debug("marshal error on staking SC stake function ",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.Arguments[0], data)

	return vmcommon.Ok
}

func (r *stakingSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.accessAddr) {
		log.Debug("unStake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		log.Debug("not enough arguments, needed BLS key and reward address")
		return vmcommon.UserError
	}

	var registrationData StakedData
	data := r.eei.GetStorage(args.Arguments[0])
	if data == nil {
		log.Debug("unStake is not possible for address which is not staked")
		return vmcommon.UserError
	}

	err := json.Unmarshal(data, &registrationData)
	if err != nil {
		log.Debug("unmarshal error in unStake function of staking SC",
			"error", err.Error(),
		)
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

	registrationData.Staked = false
	registrationData.UnStakedEpoch = r.eei.BlockChainHook().CurrentEpoch()
	registrationData.UnStakedNonce = r.eei.BlockChainHook().CurrentNonce()

	data, err = json.Marshal(registrationData)
	if err != nil {
		log.Debug("marshal error in unStake function of staking SC",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.Arguments[0], data)

	return vmcommon.Ok
}

func (r *stakingSC) unBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.accessAddr) {
		log.Debug("unStake function not allowed to be called by", "address", args.CallerAddr)
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		log.Debug("not enough arguments, needed BLS key and reward address")
		return vmcommon.UserError
	}

	var registrationData StakedData
	data := r.eei.GetStorage(args.Arguments[0])
	if data == nil {
		log.Error("unBond is not possible for address which is not staked")
		return vmcommon.UserError
	}

	err := json.Unmarshal(data, &registrationData)
	if err != nil {
		log.Debug("unmarshal error on unBond function",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	if registrationData.Staked || registrationData.UnStakedNonce <= registrationData.RegisterNonce {
		log.Debug("unBond is not possible for address which is staked or is not in unbond period")
		return vmcommon.UserError
	}

	currentNonce := r.eei.BlockChainHook().CurrentNonce()
	if currentNonce-registrationData.UnStakedNonce < r.unBondPeriod {
		log.Debug("unBond is not possible for address because unbond period did not pass")
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.Arguments[0], nil)
	r.eei.Finish(registrationData.StakeValue.Bytes())
	r.eei.Finish(big.NewInt(0).SetUint64(uint64(registrationData.UnStakedEpoch)).Bytes())

	return vmcommon.Ok
}

func (r *stakingSC) slash(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		log.Debug("slash function called by not the owners address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 2 {
		log.Debug("slash function called by wrong number of arguments")
		return vmcommon.UserError
	}

	var registrationData StakedData
	stakerAddress := args.Arguments[0]
	data := r.eei.GetStorage(stakerAddress)
	if data == nil {
		return vmcommon.UserError
	}
	err := json.Unmarshal(data, &registrationData)
	if err != nil {
		log.Debug("unmarshal error on slash function",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	if !registrationData.Staked {
		log.Debug("cannot slash already unstaked or user not staked")
		return vmcommon.UserError
	}

	stakedValue := big.NewInt(0).Set(registrationData.StakeValue)
	slashValue := big.NewInt(0).SetBytes(args.Arguments[1])
	registrationData.StakeValue = registrationData.StakeValue.Sub(stakedValue, slashValue)

	data, err = json.Marshal(registrationData)
	if err != nil {
		log.Debug("marshal error in slash function of staking smart contract",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.Arguments[0], data)

	return vmcommon.Ok
}

func (r *stakingSC) isStaked(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		return vmcommon.UserError
	}

	data := r.eei.GetStorage(args.Arguments[0])
	registrationData := StakedData{}
	if data != nil {
		err := json.Unmarshal(data, &registrationData)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return vmcommon.UserError
		}
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
