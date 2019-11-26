package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.GetOrCreate("vm/systemsmartcontracts")

const ownerKey = "owner"
const initialStakeKey = "initialStake"

type StakingData struct {
	StartNonce    uint64   `json:"StartNonce"`
	Staked        bool     `json:"Staked"`
	UnStakedNonce uint64   `json:"UnStakedNonce"`
	BlsPubKey     []byte   `json:"BlsPubKey"`
	StakeValue    *big.Int `json:"StakeValue"`
}

type stakingSC struct {
	eei           vm.SystemEI
	stakeValue    *big.Int
	unBoundPeriod uint64
}

// NewStakingSmartContract creates a staking smart contract
func NewStakingSmartContract(stakeValue *big.Int, unBoundPeriod uint64, eei vm.SystemEI) (*stakingSC, error) {
	if stakeValue == nil {
		return nil, vm.ErrNilInitialStakeValue
	}
	if stakeValue.Cmp(big.NewInt(0)) < 1 {
		return nil, vm.ErrNegativeInitialStakeValue
	}
	if eei == nil || eei.IsInterfaceNil() {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	reg := &stakingSC{
		stakeValue:    big.NewInt(0).Set(stakeValue),
		eei:           eei,
		unBoundPeriod: unBoundPeriod,
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
		return r.stake(args)
	case "unStake":
		return r.unStake(args)
	case "unBound":
		return r.unBound(args)
	case "slash":
		return r.slash(args)
	case "get":
		return r.get(args)
	}

	return vmcommon.UserError
}

func (r *stakingSC) get(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		return vmcommon.UserError
	}

	value := r.eei.GetStorage(args.Arguments[0].Bytes())
	r.eei.Finish(value)

	return vmcommon.Ok
}

func (r *stakingSC) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(ownerKey))
	if ownerAddress != nil {
		log.Error("smart contract was already initialized")
		return vmcommon.UserError
	}

	r.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	r.eei.SetStorage(args.CallerAddr, big.NewInt(0).Bytes())
	r.eei.SetStorage([]byte(initialStakeKey), r.stakeValue.Bytes())
	return vmcommon.Ok
}

func (r *stakingSC) stake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	stakeValueBytes := r.eei.GetStorage([]byte(initialStakeKey))
	stakeValue := big.NewInt(0).SetBytes(stakeValueBytes)

	if args.CallValue.Cmp(stakeValue) != 0 || args.CallValue.Sign() <= 0 {
		return vmcommon.UserError
	}

	registrationData := StakingData{
		StartNonce:    0,
		Staked:        false,
		BlsPubKey:     nil,
		UnStakedNonce: 0,
		StakeValue:    big.NewInt(0).Set(stakeValue),
	}
	data := r.eei.GetStorage(args.CallerAddr)

	if data != nil {
		err := json.Unmarshal(data, &registrationData)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return vmcommon.UserError
		}
	}

	if registrationData.Staked == true {
		log.Debug("account already staked, re-staking is invalid")
		return vmcommon.UserError
	}

	registrationData.Staked = true

	if len(args.Arguments) < 1 {
		log.Debug("not enough arguments to process stake function")
		return vmcommon.UserError
	}

	registrationData.StartNonce = r.eei.BlockChainHook().CurrentNonce()
	registrationData.BlsPubKey = args.Arguments[0].Bytes()
	//TODO: verify if blsPubKey is valid

	data, err := json.Marshal(registrationData)
	if err != nil {
		log.Debug("marshal error on staking SC stake function ",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.CallerAddr, data)

	return vmcommon.Ok
}

func (r *stakingSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	var registrationData StakingData
	data := r.eei.GetStorage(args.CallerAddr)
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

	if registrationData.Staked == false {
		log.Error("unStake is not possible for address with is already unStaked")
		return vmcommon.UserError
	}

	registrationData.Staked = false
	registrationData.UnStakedNonce = r.eei.BlockChainHook().CurrentNonce()

	data, err = json.Marshal(registrationData)
	if err != nil {
		log.Debug("marshal error in unStake function of staking SC",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.CallerAddr, data)

	return vmcommon.Ok
}

func (r *stakingSC) unBound(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	var registrationData StakingData
	data := r.eei.GetStorage(args.CallerAddr)
	if data == nil {
		log.Error("unBound is not possible for address which is not staked")
		return vmcommon.UserError
	}

	err := json.Unmarshal(data, &registrationData)
	if err != nil {
		log.Debug("unmarshal error on finalize unstake function",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	if registrationData.Staked || registrationData.UnStakedNonce <= registrationData.StartNonce {
		log.Debug("unBound is not possible for address which is staked or is not in unbound period")
		return vmcommon.UserError
	}

	currentNonce := r.eei.BlockChainHook().CurrentNonce()
	if currentNonce-registrationData.UnStakedNonce < r.unBoundPeriod {
		log.Debug("unBound is not possible for address because unbound period did not pass")
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.CallerAddr, nil)

	ownerAddress := r.eei.GetStorage([]byte(ownerKey))
	err = r.eei.Transfer(args.CallerAddr, ownerAddress, registrationData.StakeValue, nil)
	if err != nil {
		log.Debug("transfer error on finalizeUnStake function",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

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

	var registrationData StakingData
	stakerAddress := args.Arguments[0].Bytes()
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
	slashValue := args.Arguments[1]
	registrationData.StakeValue = registrationData.StakeValue.Sub(stakedValue, slashValue)

	data, err = json.Marshal(registrationData)
	if err != nil {
		log.Debug("marshal error in slash function of staking smart contract",
			"error", err.Error(),
		)
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.CallerAddr, data)

	return vmcommon.Ok
}

// ValueOf returns the value of a selected key
func (r *stakingSC) ValueOf(key interface{}) interface{} {
	return nil
}

// IsInterfaceNil verifies if the underlying object is nil or not
func (r *stakingSC) IsInterfaceNil() bool {
	if r == nil {
		return true
	}
	return false
}
