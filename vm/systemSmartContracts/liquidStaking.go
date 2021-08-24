//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. liquidStaking.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const tokenIDKey = "tokenID"
const nonceAttributesPrefix = "n"
const attributesNoncePrefix = "a"

type liquidStaking struct {
	eei                      vm.SystemEI
	sigVerifier              vm.MessageSignVerifier
	delegationMgrSCAddress   []byte
	liquidStakingSCAddress   []byte
	endOfEpochAddr           []byte
	gasCost                  vm.GasCost
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	mutExecution             sync.RWMutex
	liquidStakingEnableEpoch uint32
	flagLiquidStaking        atomic.Flag
}

// ArgsNewLiquidStaking defines the arguments to create the liquid staking smart contract
type ArgsNewLiquidStaking struct {
	EpochConfig            config.EpochConfig
	Eei                    vm.SystemEI
	DelegationMgrSCAddress []byte
	LiquidStakingSCAddress []byte
	EndOfEpochAddress      []byte
	GasCost                vm.GasCost
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	EpochNotifier          vm.EpochNotifier
}

// NewLiquidStakingSystemSC creates a new liquid staking system SC
func NewLiquidStakingSystemSC(args ArgsNewLiquidStaking) (*liquidStaking, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.DelegationMgrSCAddress) < 1 {
		return nil, fmt.Errorf("%w for delegation manager sc address", vm.ErrInvalidAddress)
	}
	if len(args.EndOfEpochAddress) < 1 {
		return nil, fmt.Errorf("%w for end of epoch address", vm.ErrInvalidAddress)
	}
	if len(args.LiquidStakingSCAddress) < 1 {
		return nil, fmt.Errorf("%w for liquid staking sc address", vm.ErrInvalidAddress)
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

	l := &liquidStaking{
		eei:                      args.Eei,
		delegationMgrSCAddress:   args.DelegationMgrSCAddress,
		endOfEpochAddr:           args.EndOfEpochAddress,
		liquidStakingSCAddress:   args.LiquidStakingSCAddress,
		gasCost:                  args.GasCost,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		liquidStakingEnableEpoch: args.EpochConfig.EnableEpochs.BuiltInFunctionOnMetaEnableEpoch,
	}
	log.Debug("liquid staking: enable epoch", "epoch", l.liquidStakingEnableEpoch)

	args.EpochNotifier.RegisterNotifyHandler(l)

	return l, nil
}

// Execute calls one of the functions from the delegation contract and runs the code according to the input
func (l *liquidStaking) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	l.mutExecution.RLock()
	defer l.mutExecution.RUnlock()

	err := CheckIfNil(args)
	if err != nil {
		l.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if !l.flagLiquidStaking.IsSet() {
		l.eei.AddReturnMessage("liquid staking contract is not enabled")
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return l.init(args)
	case "claimDelegatedPosition":
		return l.claimDelegatedPosition(args)
	case "claimRewardsFromPosition":
		return l.claimRewardsFromDelegatedPosition(args)
	case "reDelegateRewardsFromPosition":
		return l.reDelegateRewardsFromPosition(args)
	case "unDelegateWithPosition":
		return l.unDelegateWithPosition(args)
	case "returnPosition":
		return l.returnPosition(args)
	}

	l.eei.AddReturnMessage(args.Function + " is an unknown function")
	return vmcommon.UserError
}

func (l *liquidStaking) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, l.liquidStakingSCAddress) {
		l.eei.AddReturnMessage("invalid caller")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		l.eei.AddReturnMessage("not a payable function")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		l.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.UserError
	}
	tokenID := args.Arguments[0]
	l.eei.SetStorage([]byte(tokenIDKey), tokenID)

	return vmcommon.Ok
}

func (l *liquidStaking) getTokenID() []byte {
	return l.eei.GetStorage([]byte(tokenIDKey))
}

func (l *liquidStaking) checkArgumentsWhenPositionIsInput(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.ESDTTransfers) < 1 {
		l.eei.AddReturnMessage("function requires liquid staking input")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		l.eei.AddReturnMessage("function is not payable in eGLD")
		return vmcommon.UserError
	}
	definedTokenID := l.getTokenID()
	for _, esdtTransfer := range args.ESDTTransfers {
		if !bytes.Equal(esdtTransfer.ESDTTokenName, definedTokenID) {
			l.eei.AddReturnMessage("wrong tokenID input")
			return vmcommon.UserError
		}
	}
	err := l.eei.UseGas(uint64(len(args.ESDTTransfers)) * l.gasCost.MetaChainSystemSCsCost.LiquidStakingOps)
	if err != nil {
		l.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	return vmcommon.Ok
}

func (l *liquidStaking) claimDelegatedPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if args.CallValue.Cmp(zero) != 0 {
		l.eei.AddReturnMessage("function is not payable in eGLD")
		return vmcommon.UserError
	}
	if len(args.Arguments) < 3 {
		l.eei.AddReturnMessage("not enough arguments")
		return vmcommon.UserError
	}
	if len(args.ESDTTransfers) > 0 {
		l.eei.AddReturnMessage("function is not payable in ESDT")
		return vmcommon.UserError
	}

	numOfCalls := big.NewInt(0).SetBytes(args.Arguments[0]).Int64()
	minNumArguments := numOfCalls*2 + 1
	if int64(len(args.Arguments)) < minNumArguments {
		l.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.UserError
	}

	listNonces := make([]uint64, 0)
	listValues := make([]*big.Int, 0)
	startIndex := int64(1)
	for i := int64(0); i < numOfCalls; i++ {
		scAddress := args.Arguments[startIndex+i*2]
		valueToClaim := big.NewInt(0).SetBytes(args.Arguments[startIndex+i*2+1])

		txData := "claimDelegatedPosition" + "@" + hex.EncodeToString(args.CallerAddr) + "@" + hex.EncodeToString(valueToClaim.Bytes())
		vmOutput, err := l.eei.ExecuteOnDestContext(scAddress, args.RecipientAddr, big.NewInt(0), []byte(txData))
		if err != nil {
			l.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			return vmOutput.ReturnCode
		}

		if len(vmOutput.ReturnData) != 1 {
			l.eei.AddReturnMessage("invalid return data")
			return vmcommon.UserError
		}

		rewardsCheckpoint := uint32(big.NewInt(0).SetBytes(vmOutput.ReturnData[0]).Uint64())
		nonce, err := l.createOrAddNFT(scAddress, rewardsCheckpoint, valueToClaim)
		if err != nil {
			l.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		listNonces = append(listNonces, nonce)
		listValues = append(listValues, valueToClaim)
	}

	var additionalArgs [][]byte
	if int64(len(args.Arguments)) > minNumArguments {
		additionalArgs = args.Arguments[minNumArguments:]
	}
	err := l.sendNFTMultiTransfer(args.CallerAddr, listNonces, listValues, additionalArgs)
	if err != nil {
		l.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (l *liquidStaking) executeOnDestinationSC(
	dstSCAddress []byte,
	functionToCall string,
	userAddress []byte,
	valueToSend *big.Int,
	rewardsCheckPoint uint32,
) ([][]byte, vmcommon.ReturnCode) {
	txData := functionToCall + "@" + hex.EncodeToString(userAddress) + "@" + hex.EncodeToString(valueToSend.Bytes())
	if rewardsCheckPoint > 0 {
		txData += "@" + hex.EncodeToString(big.NewInt(int64(rewardsCheckPoint)).Bytes())
	}
	vmOutput, err := l.eei.ExecuteOnDestContext(dstSCAddress, l.liquidStakingSCAddress, big.NewInt(0), []byte(txData))
	if err != nil {
		l.eei.AddReturnMessage(err.Error())
		return nil, vmcommon.UserError
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, vmOutput.ReturnCode
	}

	return vmOutput.ReturnData, vmcommon.Ok
}

func (l *liquidStaking) createOrAddNFT(
	delegationSCAddress []byte,
	rewardsCheckpoint uint32,
	value *big.Int,
) (uint64, error) {
	attributes := &LiquidStakingAttributes{
		ContractAddress:   delegationSCAddress,
		RewardsCheckpoint: rewardsCheckpoint,
	}

	marshalledData, err := l.marshalizer.Marshal(attributes)
	if err != nil {
		return 0, err
	}

	hash := l.hasher.Compute(string(marshalledData))
	attrNonceKey := append([]byte(attributesNoncePrefix), hash...)
	storageData := l.eei.GetStorage(attrNonceKey)
	if len(storageData) > 0 {
		nonce := big.NewInt(0).SetBytes(storageData).Uint64()
		err = l.addQuantityToNFT(nonce, value)
		if err != nil {
			return 0, err
		}

		return nonce, nil
	}

	nonce, err := l.createNewNFT(value)
	if err != nil {
		return 0, nil
	}

	nonceBytes := big.NewInt(0).SetUint64(nonce).Bytes()
	l.eei.SetStorage(attrNonceKey, nonceBytes)

	nonceKey := append([]byte(nonceAttributesPrefix), nonceBytes...)
	l.eei.SetStorage(nonceKey, marshalledData)

	return nonce, nil
}

func (l *liquidStaking) createNewNFT(value *big.Int) (uint64, error) {
	valuePlusOne := big.NewInt(0).Add(value, big.NewInt(1))

	args := make([][]byte, 7)
	args[0] = l.getTokenID()
	args[1] = valuePlusOne.Bytes()

	vmOutput, err := l.eei.ProcessBuiltInFunction(l.liquidStakingSCAddress, l.liquidStakingSCAddress, core.BuiltInFunctionESDTNFTCreate, args)
	if err != nil {
		return 0, err
	}
	if len(vmOutput.ReturnData) != 1 {
		return 0, vm.ErrNotEnoughReturnData
	}

	return big.NewInt(0).SetBytes(vmOutput.ReturnData[0]).Uint64(), nil
}

func (l *liquidStaking) addQuantityToNFT(nonce uint64, value *big.Int) error {
	args := make([][]byte, 3)
	args[0] = l.getTokenID()
	args[1] = big.NewInt(0).SetUint64(nonce).Bytes()
	args[2] = value.Bytes()

	_, err := l.eei.ProcessBuiltInFunction(l.liquidStakingSCAddress, l.liquidStakingSCAddress, core.BuiltInFunctionESDTNFTAddQuantity, args)
	if err != nil {
		return err
	}

	return nil
}

func (l *liquidStaking) getAttributesForNonce(nonce uint64) (*LiquidStakingAttributes, error) {
	nonceKey := append([]byte(nonceAttributesPrefix), big.NewInt(0).SetUint64(nonce).Bytes()...)
	marshalledData := l.eei.GetStorage(nonceKey)
	if len(marshalledData) == 0 {
		return nil, vm.ErrEmptyStorage
	}

	lAttr := &LiquidStakingAttributes{}
	err := l.marshalizer.Unmarshal(lAttr, marshalledData)
	if err != nil {
		return nil, err
	}

	return lAttr, nil
}

func (l *liquidStaking) sendNFTMultiTransfer(
	destinationAddress []byte,
	listNonces []uint64,
	listValue []*big.Int,
	additionalArgs [][]byte,
) error {

	numOfTransfer := int64(len(listNonces))
	args := make([][]byte, 0)
	args = append(args, destinationAddress)
	args = append(args, big.NewInt(numOfTransfer).Bytes())

	tokenID := l.getTokenID()
	for i := 0; i < len(listNonces); i++ {
		args = append(args, tokenID)
		args = append(args, big.NewInt(0).SetUint64(listNonces[i]).Bytes())
		args = append(args, listValue[i].Bytes())
	}

	if len(additionalArgs) > 0 {
		args = append(args, additionalArgs...)
	}

	_, err := l.eei.ProcessBuiltInFunction(l.liquidStakingSCAddress, l.liquidStakingSCAddress, core.BuiltInFunctionMultiESDTNFTTransfer, args)
	if err != nil {
		return err
	}

	return nil
}

func (l *liquidStaking) claimRewardsFromDelegatedPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := l.checkArgumentsWhenPositionIsInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	return vmcommon.Ok
}

func (l *liquidStaking) reDelegateRewardsFromPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := l.checkArgumentsWhenPositionIsInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

func (l *liquidStaking) unDelegateWithPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := l.checkArgumentsWhenPositionIsInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

func (l *liquidStaking) returnPosition(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	returnCode := l.checkArgumentsWhenPositionIsInput(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

// SetNewGasCost is called whenever a gas cost was changed
func (l *liquidStaking) SetNewGasCost(gasCost vm.GasCost) {
	l.mutExecution.Lock()
	l.gasCost = gasCost
	l.mutExecution.Unlock()
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (l *liquidStaking) EpochConfirmed(epoch uint32, _ uint64) {
	l.flagLiquidStaking.Toggle(epoch >= l.liquidStakingEnableEpoch)
	log.Debug("liquid staking system sc", "enabled", l.flagLiquidStaking.IsSet())
}

// CanUseContract returns true if contract can be used
func (l *liquidStaking) CanUseContract() bool {
	return l.flagLiquidStaking.IsSet()
}

// IsInterfaceNil returns true if underlying object is nil
func (l *liquidStaking) IsInterfaceNil() bool {
	return l == nil
}
