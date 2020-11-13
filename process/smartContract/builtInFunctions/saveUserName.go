package builtInFunctions

import (
	"encoding/hex"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.BuiltinFunction = (*saveUserName)(nil)

type saveUserName struct {
	gasCost         uint64
	mapDnsAddresses map[string]struct{}
	enableChange    bool
	mutExecution    sync.RWMutex
}

// NewSaveUserNameFunc returns a username built in function implementation
func NewSaveUserNameFunc(
	gasCost uint64,
	mapDnsAddresses map[string]struct{},
	enableChange bool,
) (*saveUserName, error) {
	if mapDnsAddresses == nil {
		return nil, process.ErrNilDnsAddresses
	}

	s := &saveUserName{
		gasCost:      gasCost,
		enableChange: enableChange,
	}
	s.mapDnsAddresses = make(map[string]struct{}, len(mapDnsAddresses))
	for key := range mapDnsAddresses {
		s.mapDnsAddresses[key] = struct{}{}
	}

	return s, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (s *saveUserName) SetNewGasConfig(gasCost *process.GasCost) {
	s.mutExecution.Lock()
	s.gasCost = gasCost.BuiltInCost.SaveUserName
	s.mutExecution.Unlock()
}

// ProcessBuiltinFunction sets the username to the account if it is allowed
func (s *saveUserName) ProcessBuiltinFunction(
	_, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	s.mutExecution.RLock()
	defer s.mutExecution.RUnlock()

	if vmInput == nil {
		return nil, process.ErrNilVmInput
	}
	if vmInput.CallValue.Cmp(zero) != 0 {
		return nil, process.ErrBuiltInFunctionCalledWithValue
	}
	if vmInput.GasProvided < s.gasCost {
		return nil, process.ErrNotEnoughGas
	}
	_, ok := s.mapDnsAddresses[string(vmInput.CallerAddr)]
	if !ok {
		return nil, process.ErrCallerIsNotTheDNSAddress
	}
	if len(vmInput.Arguments) != 1 {
		return nil, process.ErrInvalidArguments
	}

	if check.IfNil(acntDst) {
		log.Trace("setUserName called dst not in shard")
		// cross-shard call, in sender shard only the gas is taken out
		vmOutput := &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}
		vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
		setUserNameTxData := core.BuiltInFunctionSetUserName + "@" + hex.EncodeToString(vmInput.Arguments[0])
		outTransfer := vmcommon.OutputTransfer{
			Value:    big.NewInt(0),
			GasLimit: vmInput.GasProvided,
			Data:     []byte(setUserNameTxData),
			CallType: vmcommon.AsynchronousCall,
		}
		vmOutput.OutputAccounts[string(vmInput.RecipientAddr)] = &vmcommon.OutputAccount{
			Address:         vmInput.RecipientAddr,
			OutputTransfers: []vmcommon.OutputTransfer{outTransfer},
		}
		return vmOutput, nil
	}

	log.Trace("setUserName called in shard")
	currentUserName := acntDst.GetUserName()
	if !s.enableChange && len(currentUserName) > 0 {
		return nil, process.ErrUserNameChangeIsDisabled
	}

	acntDst.SetUserName(vmInput.Arguments[0])

	return &vmcommon.VMOutput{GasRemaining: vmInput.GasProvided - s.gasCost, ReturnCode: vmcommon.Ok}, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (s *saveUserName) IsInterfaceNil() bool {
	return s == nil
}
