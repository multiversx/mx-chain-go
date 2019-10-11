package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var log = logger.DefaultLogger()

const ownerKey = "owner"

type registerData struct {
	StartNonce    uint64   `json:"StartNonce"`
	Staked        bool     `json:"Staked"`
	UnStakedNonce uint64   `json:"UnStakedNonce"`
	BlsPubKey     []byte   `json:"BlsPubKey"`
	StakeValue    *big.Int `json:"StakeValue"`
}

type registerSC struct {
	eei        vm.SystemEI
	stakeValue *big.Int
}

// NewRegisterSmartContract creates a register smart contract
func NewRegisterSmartContract(stakeValue *big.Int, eei vm.SystemEI) (*registerSC, error) {
	if stakeValue == nil {
		return nil, vm.ErrNilInitialStakeValue
	}
	if eei == nil || eei.IsInterfaceNil() {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}

	reg := &registerSC{stakeValue: big.NewInt(0).Set(stakeValue), eei: eei}
	return reg, nil
}

// Execute calls one of the functions from the register smart contract and runs the code according to the input
func (r *registerSC) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
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
	}

	return vmcommon.UserError
}

func (r *registerSC) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	r.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	r.eei.SetStorage(args.CallerAddr, big.NewInt(0).Bytes())
	return vmcommon.Ok
}

func (r *registerSC) slash(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	return vmcommon.Ok
}

func (r *registerSC) stake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(r.stakeValue) != 0 {
		return vmcommon.UserError
	}

	registrationData := registerData{
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
			log.Error("unmarshal error on register smart contract stake function " + err.Error())
			return vmcommon.UserError
		}
	}

	if registrationData.Staked == true {
		log.Error("account already registered, re-staking is invalid")
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
		log.Error("marshal error on register smart contract stake function " + err.Error())
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.CallerAddr, data)

	err = r.eei.Transfer(args.RecipientAddr, args.CallerAddr, args.CallValue, nil)
	if err != nil {
		log.Error("transfer error on stake function " + err.Error())
	}

	return vmcommon.Ok
}

func (r *registerSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	var registrationData registerData
	data := r.eei.GetStorage(args.CallerAddr)
	if data == nil {
		log.Error("unStake is not possible for address which is not staked")
		return vmcommon.UserError
	}

	err := json.Unmarshal(data, registrationData)
	if err != nil {
		log.Error("unmarshal error in unStake function of register smart contract " + err.Error())
		return vmcommon.UserError
	}

	registrationData.Staked = false
	registrationData.UnStakedNonce = args.Header.Number.Uint64()

	data, err = json.Marshal(registrationData)
	if err != nil {
		log.Error("marshal error in unStake function of register smart contract" + err.Error())
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.CallerAddr, data)

	return vmcommon.Ok
}

func (r *registerSC) finalizeUnStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		return vmcommon.UserError
	}

	var registrationData registerData
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

// ValueOf returns the value of a selected key
func (r *registerSC) ValueOf(key interface{}) interface{} {
	return nil
}

// IsInterfaceNil verifies if the underlying object is nil or not
func (r *registerSC) IsInterfaceNil() bool {
	if r == nil {
		return true
	}
	return false
}
