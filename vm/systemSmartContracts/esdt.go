//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. esdt.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
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
const canChangeOwner = "canChangeOwner"
const upgradable = "uppgradable"

const conversionBase = 10

type esdt struct {
	eei             vm.SystemEI
	gasCost         vm.GasCost
	baseIssuingCost *big.Int
	ownerAddress    []byte
	eSDTSCAddress   []byte
	marshalizer     marshal.Marshalizer
	hasher          hashing.Hasher
	disabled        bool
}

// ArgsNewESDTSmartContract defines the arguments needed for the esdt contract
type ArgsNewESDTSmartContract struct {
	Eei           vm.SystemEI
	GasCost       vm.GasCost
	ESDTSCConfig  config.ESDTSystemSCConfig
	ESDTSCAddress []byte
	Marshalizer   marshal.Marshalizer
	Hasher        hashing.Hasher
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

	baseIssuingCost, ok := big.NewInt(0).SetString(args.ESDTSCConfig.BaseIssuingCost, conversionBase)
	if !ok || baseIssuingCost.Cmp(big.NewInt(0)) < 0 {
		return nil, vm.ErrInvalidBaseIssuingCost
	}

	return &esdt{
		eei:             args.Eei,
		gasCost:         args.GasCost,
		baseIssuingCost: baseIssuingCost,
		ownerAddress:    []byte(args.ESDTSCConfig.OwnerAddress),
		eSDTSCAddress:   args.ESDTSCAddress,
		hasher:          args.Hasher,
		marshalizer:     args.Marshalizer,
		disabled:        args.ESDTSCConfig.Disabled,
	}, nil
}

// Execute calls one of the functions from the esdt smart contract and runs the code according to the input
func (e *esdt) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	if args.Function == core.SCDeployInitFunctionName {
		return e.init(args)
	}

	if e.disabled {
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
		return e.freezeUnFreeze(args, core.BuiltInFunctionESDTFreeze)
	case "unFreeze":
		return e.freezeUnFreeze(args, core.BuiltInFunctionESDTUnFreeze)
	case "wipe":
		return e.wipe(args)
	case "pause":
		return e.pauseUnPause(args, core.BuiltInFunctionESDTPause)
	case "unPause":
		return e.pauseUnPause(args, core.BuiltInFunctionESDTUnPause)
	case "claim":
		return e.claim(args)
	case "configChange":
		return e.configChange(args)
	case "esdtControlChanges":
		return e.esdtControlChanges(args)
	case "changeOwner":
		return e.changeOwner(args)
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
		OwnerAddress: owner,
		TokenName:    tokenName,
		MintedValue:  initialSupply,
		BurntValue:   big.NewInt(0),
	}
	e.upgradeProperties(newESDTToken, arguments[2:])

	err := e.saveToken(newESDTToken)
	if err != nil {
		return err
	}

	esdtTransferData := core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(initialSupply.Bytes())
	err = e.eei.Transfer(owner, e.eSDTSCAddress, big.NewInt(0), []byte(esdtTransferData), 0)
	if err != nil {
		return err
	}

	return nil
}

func (e *esdt) upgradeProperties(token *ESDTData, args [][]byte) {
	for _, arg := range args {
		optionalArg := string(arg)
		switch optionalArg {
		case burnable:
			token.Burnable = true
		case mintable:
			token.Mintable = true
		case canPause:
			token.CanPause = true
		case canFreeze:
			token.CanFreeze = true
		case canWipe:
			token.CanWipe = true
		case upgradable:
			token.Upgradable = true
		case canChangeOwner:
			token.CanChangeOwner = true
		}
	}
}

func (e *esdt) burn(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 2 {
		e.eei.AddReturnMessage("number of arguments must be equal with 2")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage("callValue must be 0")
		return vmcommon.OutOfFunds
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.ESDTOperations)
	if err != nil {
		e.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}
	burntValue := big.NewInt(0).SetBytes(args.Arguments[1])
	if burntValue.Cmp(big.NewInt(0)) <= 0 {
		e.eei.AddReturnMessage(vm.ErrNegativeOrZeroInitialSupply.Error())
		return vmcommon.UserError
	}
	token, err := e.getExistingToken(args.Arguments[0])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if !token.Burnable {
		e.eei.AddReturnMessage("token is not burnable")
		return vmcommon.UserError
	}
	token.BurntValue.Add(token.BurntValue, burntValue)

	err = e.saveToken(token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *esdt) mint(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 2 || len(args.Arguments) > 3 {
		e.eei.AddReturnMessage("accepted arguments number 2/3")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnerChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	mintValue := big.NewInt(0).SetBytes(args.Arguments[1])
	if mintValue.Cmp(big.NewInt(0)) <= 0 {
		e.eei.AddReturnMessage(vm.ErrNegativeOrZeroInitialSupply.Error())
		return vmcommon.UserError
	}
	if !token.Mintable {
		e.eei.AddReturnMessage("token is not burnable")
		return vmcommon.UserError
	}

	token.MintedValue.Add(token.MintedValue, mintValue)
	err := e.saveToken(token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	destination := token.OwnerAddress
	if len(args.Arguments) == 3 {
		if len(args.Arguments[2]) != len(args.CallerAddr) {
			e.eei.AddReturnMessage("destination address of invalid length")
			return vmcommon.UserError
		}
		destination = args.Arguments[2]
	}

	esdtTransferData := core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString(token.TokenName) + "@" + hex.EncodeToString(mintValue.Bytes())
	err = e.eei.Transfer(destination, e.eSDTSCAddress, big.NewInt(0), []byte(esdtTransferData), 0)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *esdt) freezeUnFreeze(args *vmcommon.ContractCallInput, builtInFunc string) vmcommon.ReturnCode {
	if len(args.Arguments) < 2 {
		e.eei.AddReturnMessage("invalid number of arguments, wanted 2")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnerChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.CanFreeze {
		e.eei.AddReturnMessage("cannot freeze")
		return vmcommon.UserError
	}
	if len(args.Arguments[1]) != len(args.CallerAddr) {
		e.eei.AddReturnMessage("invalid arguments")
		return vmcommon.UserError
	}

	esdtTransferData := builtInFunc + "@" + hex.EncodeToString(token.TokenName)
	err := e.eei.Transfer(args.Arguments[1], e.eSDTSCAddress, big.NewInt(0), []byte(esdtTransferData), 0)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *esdt) wipe(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 2 {
		e.eei.AddReturnMessage("invalid number of arguments, wanted 2")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnerChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.CanWipe {
		e.eei.AddReturnMessage("cannot freeze")
		return vmcommon.UserError
	}
	if len(args.Arguments[1]) != len(args.CallerAddr) {
		e.eei.AddReturnMessage("invalid arguments")
		return vmcommon.UserError
	}

	esdtTransferData := core.BuiltInFunctionESDTWipe + "@" + hex.EncodeToString(token.TokenName)
	err := e.eei.Transfer(args.Arguments[1], e.eSDTSCAddress, big.NewInt(0), []byte(esdtTransferData), 0)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *esdt) pauseUnPause(args *vmcommon.ContractCallInput, builtInFunc string) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		e.eei.AddReturnMessage("invalid number of arguments, wanted 2")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnerChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.CanPause {
		e.eei.AddReturnMessage("cannot freeze")
		return vmcommon.UserError
	}
	if len(args.Arguments[1]) != len(args.CallerAddr) {
		e.eei.AddReturnMessage("invalid arguments")
		return vmcommon.UserError
	}
	if token.IsPaused && builtInFunc == core.BuiltInFunctionESDTPause {
		e.eei.AddReturnMessage("cannot pause an already paused contract")
		return vmcommon.UserError
	}
	if !token.IsPaused && builtInFunc == core.BuiltInFunctionESDTUnPause {
		e.eei.AddReturnMessage("cannot unPause an already paused contract")
		return vmcommon.UserError
	}

	token.IsPaused = !token.IsPaused
	err := e.saveToken(token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	esdtTransferData := builtInFunc + "@" + hex.EncodeToString(token.TokenName)
	e.eei.TransferToAll(e.eSDTSCAddress, []byte(esdtTransferData))

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

func (e *esdt) basicOwnerChecks(args *vmcommon.ContractCallInput) (*ESDTData, vmcommon.ReturnCode) {
	if args.CallValue.Cmp(zero) != 0 {
		e.eei.AddReturnMessage("callValue must be 0")
		return nil, vmcommon.OutOfFunds
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.ESDTOperations)
	if err != nil {
		e.eei.AddReturnMessage("not enough gas")
		return nil, vmcommon.OutOfGas
	}
	token, err := e.getExistingToken(args.Arguments[0])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}
	if !bytes.Equal(token.OwnerAddress, args.CallerAddr) {
		e.eei.AddReturnMessage("mint can be called by owner only")
		return nil, vmcommon.UserError
	}

	return token, vmcommon.Ok
}

func (e *esdt) changeOwner(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 2 {
		e.eei.AddReturnMessage("expected num of arguments 2")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnerChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.CanChangeOwner {
		e.eei.AddReturnMessage("token is not upgradable")
		return vmcommon.UserError
	}
	if len(args.Arguments[1]) != len(args.CallerAddr) {
		e.eei.AddReturnMessage("destination address of invalid length")
		return vmcommon.UserError
	}

	token.OwnerAddress = args.Arguments[1]
	err := e.saveToken(token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *esdt) esdtControlChanges(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 2 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	token, returnCode := e.basicOwnerChecks(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if !token.Upgradable {
		e.eei.AddReturnMessage("token is not upgradable")
		return vmcommon.UserError
	}

	e.upgradeProperties(token, args.Arguments[1:])
	err := e.saveToken(token)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *esdt) saveToken(token *ESDTData) error {
	marshaledData, err := e.marshalizer.Marshal(token)
	if err != nil {
		return err
	}

	e.eei.SetStorage(token.TokenName, marshaledData)
	return nil
}

func (e *esdt) getExistingToken(tokenName []byte) (*ESDTData, error) {
	marshalledData := e.eei.GetStorage(tokenName)
	if len(marshalledData) == 0 {
		return nil, vm.ErrNoTokenWithGivenName
	}

	token := &ESDTData{}
	err := e.marshalizer.Unmarshal(token, marshalledData)
	return token, err
}

// IsInterfaceNil returns true if underlying object is nil
func (e *esdt) IsInterfaceNil() bool {
	return e == nil
}
