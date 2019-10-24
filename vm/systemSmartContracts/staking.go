package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.DefaultLogger()

const ownerKey = "owner"

type stakingData struct {
	StartNonce    uint64   `json:"StartNonce"`
	Staked        bool     `json:"Staked"`
	UnStakedNonce uint64   `json:"UnStakedNonce"`
	BlsPubKey     []byte   `json:"BlsPubKey"`
	StakeValue    *big.Int `json:"StakeValue"`
}

type stakingSC struct {
	eei        vm.SystemEI
	stakeValue *big.Int
}

// NewStakingSmartContract creates a staking smart contract
func NewStakingSmartContract(stakeValue *big.Int, eei vm.SystemEI) (*stakingSC, error) {
	if stakeValue == nil {
		return nil, vm.ErrNilInitialStakeValue
	}
	if eei == nil || eei.IsInterfaceNil() {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	reg := &stakingSC{
		stakeValue: big.NewInt(0).Set(stakeValue),
		eei:        eei,
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
	case "finalizeUnStake":
		return r.finalizeUnStake(args)
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
	r.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	r.eei.SetStorage(args.CallerAddr, big.NewInt(0).Bytes())
	return vmcommon.Ok
}

func (r *stakingSC) stake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(r.stakeValue) != 0 {
		return vmcommon.UserError
	}

	registrationData := stakingData{
		StartNonce:    0,
		Staked:        false,
		BlsPubKey:     nil,
		UnStakedNonce: 0,
		StakeValue:    big.NewInt(0),
	}
	data := r.eei.GetStorage(args.CallerAddr)

	if data != nil {
		err := json.Unmarshal(data, registrationData)
		if err != nil {
			log.Error("unmarshal error on staking smart contract stake function " + err.Error())
			return vmcommon.UserError
		}
	}

	if registrationData.Staked == true {
		log.Error("account already staked, re-staking is invalid")
		return vmcommon.UserError
	}

	registrationData.Staked = true

	if len(args.Arguments) < 1 {
		log.Error("not enough arguments to process stake function")
		return vmcommon.UserError
	}

	registrationData.StartNonce = args.Header.Number.Uint64()
	registrationData.BlsPubKey = args.Arguments[0].Bytes()
	//TODO: verify if blsPubKey is valid

	data, err := json.Marshal(registrationData)
	if err != nil {
		log.Error("marshal error on staking smart contract stake function " + err.Error())
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.CallerAddr, data)

	err = r.eei.Transfer(args.RecipientAddr, args.CallerAddr, args.CallValue, nil)
	if err != nil {
		log.Error("transfer error on stake function " + err.Error())
	}

	return vmcommon.Ok
}

func (r *stakingSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	var registrationData stakingData
	data := r.eei.GetStorage(args.CallerAddr)
	if data == nil {
		log.Error("unStake is not possible for address which is not staked")
		return vmcommon.UserError
	}

	err := json.Unmarshal(data, registrationData)
	if err != nil {
		log.Error("unmarshal error in unStake function of staking smart contract " + err.Error())
		return vmcommon.UserError
	}

	registrationData.Staked = false
	registrationData.UnStakedNonce = args.Header.Number.Uint64()

	data, err = json.Marshal(registrationData)
	if err != nil {
		log.Error("marshal error in unStake function of staking smart contract" + err.Error())
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.CallerAddr, data)

	return vmcommon.Ok
}

func (r *stakingSC) finalizeUnStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		return vmcommon.UserError
	}

	var registrationData stakingData
	for _, arg := range args.Arguments {
		data := r.eei.GetStorage(arg.Bytes())
		err := json.Unmarshal(data, registrationData)
		if err != nil {
			log.Error("unmarshal error on finalize unstake function" + err.Error())
			return vmcommon.UserError
		}

		if registrationData.UnStakedNonce == 0 {
			log.Error("validator did not unstaked yet")
			return vmcommon.UserError
		}

		r.eei.SetStorage(arg.Bytes(), nil)

		err = r.eei.Transfer(args.CallerAddr, arg.Bytes(), registrationData.StakeValue, nil)
		if err != nil {
			log.Error("transfer error on finalizeUnStake function " + err.Error())
			return vmcommon.UserError
		}
	}
	return vmcommon.Ok
}

func (r *stakingSC) slash(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		log.Error("slash function called by not the owners address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 2 {
		log.Error("slash function called by wrong number of arguments")
		return vmcommon.UserError
	}

	var registrationData stakingData
	data := r.eei.GetStorage(args.Arguments[0].Bytes())
	err := json.Unmarshal(data, registrationData)
	if err != nil {
		log.Error("unmarshal error on slash function" + err.Error())
		return vmcommon.UserError
	}

	if len(data) == 0 {
		log.Error("slash error: validator was not registered")
		return vmcommon.UserError
	}

	operation := big.NewInt(0).Set(registrationData.StakeValue)
	registrationData.StakeValue = registrationData.StakeValue.Sub(operation, args.Arguments[1])

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
