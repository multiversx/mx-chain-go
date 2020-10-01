//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. esdt.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const minLengthForTokenName = 10
const maxLengthForTokenName = 20
const configKeyPrefix = "esdtConfig"
const burnable = "burnable"
const mintable = "mintable"
const canPause = "canPause"
const canFreeze = "canFreeze"
const canWipe = "canWipe"

const conversionBase = 10

type esdt struct {
	eei             vm.SystemEI
	gasCost         vm.GasCost
	baseIssuingCost *big.Int
	ownerAddress    []byte
	eSDTSCAddress   []byte
	marshalizer     marshal.Marshalizer
	hasher          hashing.Hasher
	enabledEpoch    uint32
	flagEnabled     atomic.Flag
}

// ArgsNewESDTSmartContract defines the arguments needed for the esdt contract
type ArgsNewESDTSmartContract struct {
	Eei           vm.SystemEI
	GasCost       vm.GasCost
	ESDTSCConfig  config.ESDTSystemSCConfig
	ESDTSCAddress []byte
	Marshalizer   marshal.Marshalizer
	Hasher        hashing.Hasher
	EpochNotifier vm.EpochNotifier
}

// NewESDTSmartContract creates the esdt smart contract, which controls the issuing of tokens
func NewESDTSmartContract(args ArgsNewESDTSmartContract) (*esdt, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, vm.ErrNilHasher
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, vm.ErrNilEpochNotifier
	}

	baseIssuingCost, okConvert := big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, conversionBase)
	if !okConvert || baseIssuingCost.Cmp(big.NewInt(0)) < 0 {
		return nil, vm.ErrInvalidBaseIssuingCost
	}

	e := &esdt{
		eei:             args.Eei,
		gasCost:         args.GasCost,
		baseIssuingCost: baseIssuingCost,
		ownerAddress:    []byte(args.ESDTSCConfig.OwnerAddress),
		eSDTSCAddress:   args.ESDTSCAddress,
		hasher:          args.Hasher,
		marshalizer:     args.Marshalizer,
		enabledEpoch:    args.ESDTSCConfig.EnabledEpoch,
	}
	args.EpochNotifier.RegisterNotifyHandler(e)

	return e, nil
}

// Execute calls one of the functions from the esdt smart contract and runs the code according to the input
func (e *esdt) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	if args.Function == core.SCDeployInitFunctionName {
		return e.init(args)
	}

	if !e.flagEnabled.IsSet() {
		e.eei.AddReturnMessage("ESDT SC disabled")
		return vmcommon.UserError
	}

	switch args.Function {
	case "issue":
		return e.issue(args)
	case "issueProtected":
		return e.issueProtected(args)
	case "burn":
		return e.burn(args)
	case "mint":
		return e.mint(args)
	case "freeze":
		return e.freeze(args)
	case "wipe":
		return e.wipe(args)
	case "pause":
		return e.pause(args)
	case "unPause":
		return e.unpause(args)
	case "claim":
		return e.claim(args)
	case "configChange":
		return e.configChange(args)
	case "esdtControlChanges":
		return e.esdtControlChanges(args)
	}

	e.eei.AddReturnMessage("invalid method to call")
	return vmcommon.FunctionNotFound
}

func (e *esdt) init(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	scConfig := &ESDTConfig{
		OwnerAddress:       e.ownerAddress,
		BaseIssuingCost:    e.baseIssuingCost,
		MinTokenNameLength: minLengthForTokenName,
		MaxTokenNameLength: maxLengthForTokenName,
	}
	marshaledData, err := e.marshalizer.Marshal(scConfig)
	log.LogIfError(err, "marshal error on esdt init function")

	e.eei.SetStorage([]byte(configKeyPrefix), marshaledData)
	return vmcommon.Ok
}

func (e *esdt) issueProtected(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, e.ownerAddress) {
		e.eei.AddReturnMessage("issueProtected can be called by whitelisted address only")
		return vmcommon.UserError
	}
	if len(args.Arguments) < 3 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments[0]) < len(args.CallerAddr) {
		e.eei.AddReturnMessage("token name length not in parameters")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(e.baseIssuingCost) != 0 {
		e.eei.AddReturnMessage("callValue not equals with baseIssuingCost")
		return vmcommon.OutOfFunds
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.ESDTIssue)
	if err != nil {
		e.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}

	err = e.issueToken(args.Arguments[0], args.Arguments[1:])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *esdt) issue(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 2 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments[0]) < minLengthForTokenName || len(args.Arguments[0]) > maxLengthForTokenName {
		e.eei.AddReturnMessage("token name length not in parameters")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(e.baseIssuingCost) != 0 {
		e.eei.AddReturnMessage("callValue not equals with baseIssuingCost")
		return vmcommon.OutOfFunds
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.ESDTIssue)
	if err != nil {
		e.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}

	err = e.issueToken(args.CallerAddr, args.Arguments)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func isTokenNameHumanReadable(tokenName []byte) bool {
	for _, ch := range tokenName {
		isSmallCharacter := ch >= 'a' && ch <= 'z'
		isBigCharacter := ch >= 'A' && ch <= 'Z'
		isNumber := ch >= '0' && ch <= '9'
		isReadable := isSmallCharacter || isBigCharacter || isNumber
		if !isReadable {
			return false
		}
	}
	return true
}

func (e *esdt) issueToken(owner []byte, arguments [][]byte) error {
	tokenName := arguments[0]
	initialSupply := big.NewInt(0).SetBytes(arguments[1])
	if initialSupply.Cmp(big.NewInt(0)) <= 0 {
		return vm.ErrNegativeOrZeroInitialSupply
	}

	data := e.eei.GetStorage(tokenName)
	if len(data) > 0 {
		return vm.ErrTokenAlreadyRegistered
	}

	if !isTokenNameHumanReadable(tokenName) {
		return vm.ErrTokenNameNotHumanReadable
	}

	newESDTToken := &ESDTData{
		IssuerAddress: owner,
		TokenName:     tokenName,
		Mintable:      false,
		Burnable:      false,
		CanPause:      false,
		Paused:        false,
		CanFreeze:     false,
		CanWipe:       false,
		MintedValue:   initialSupply,
		BurntValue:    big.NewInt(0),
	}
	for i := 2; i < len(arguments); i++ {
		optionalArg := string(arguments[i])
		switch optionalArg {
		case burnable:
			newESDTToken.Burnable = true
		case mintable:
			newESDTToken.Mintable = true
		case canPause:
			newESDTToken.CanPause = true
		case canFreeze:
			newESDTToken.CanFreeze = true
		case canWipe:
			newESDTToken.CanWipe = true
		}
	}

	marshaledData, err := e.marshalizer.Marshal(newESDTToken)
	if err != nil {
		return err
	}

	e.eei.SetStorage(tokenName, marshaledData)

	esdtTransferData := core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(initialSupply.Bytes())
	err = e.eei.Transfer(owner, e.eSDTSCAddress, big.NewInt(0), []byte(esdtTransferData), 0)
	if err != nil {
		return err
	}

	return nil
}

func (e *esdt) burn(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *esdt) mint(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *esdt) freeze(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *esdt) wipe(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *esdt) pause(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *esdt) unpause(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *esdt) configChange(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *esdt) claim(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *esdt) esdtControlChanges(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (e *esdt) EpochConfirmed(epoch uint32) {
	e.flagEnabled.Toggle(epoch >= e.enabledEpoch)
	log.Debug("esdt contract", "enabled", e.flagEnabled.IsSet())
}

// IsInterfaceNil returns true if underlying object is nil
func (e *esdt) IsInterfaceNil() bool {
	return e == nil
}
