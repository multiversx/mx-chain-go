package builtInFunctions

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
)

var _ process.BuiltinFunction = (*esdtPause)(nil)

type esdtPause struct {
	keyPrefix []byte
	pause     bool
	accounts  state.AccountsAdapter
}

// NewESDTPauseFunc returns the esdt pause/un-pause built-in function component
func NewESDTPauseFunc(
	accounts state.AccountsAdapter,
	pause bool,
) (*esdtPause, error) {
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}

	e := &esdtPause{
		keyPrefix: []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier),
		pause:     pause,
		accounts:  accounts,
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *esdtPause) SetNewGasConfig(_ *process.GasCost) {
}

// ProcessBuiltinFunction resolves ESDT pause function call
func (e *esdtPause) ProcessBuiltinFunction(
	_, _ state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	if vmInput == nil {
		return nil, process.ErrNilVmInput
	}
	if vmInput.CallValue.Cmp(zero) != 0 {
		return nil, process.ErrBuiltInFunctionCalledWithValue
	}
	if len(vmInput.Arguments) != 1 {
		return nil, process.ErrInvalidArguments
	}
	if !bytes.Equal(vmInput.CallerAddr, vm.ESDTSCAddress) {
		return nil, process.ErrAddressIsNotESDTSystemSC
	}
	if !core.IsSystemAccountAddress(vmInput.RecipientAddr) {
		return nil, process.ErrOnlySystemAccountAccepted
	}

	esdtTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	log.Trace(vmInput.Function, "sender", vmInput.CallerAddr, "receiver", vmInput.RecipientAddr, "token", esdtTokenKey)

	err := e.togglePause(esdtTokenKey)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{}
	return vmOutput, nil
}

func (e *esdtPause) togglePause(token []byte) error {
	systemSCAccount, err := e.getSystemAccount()
	if err != nil {
		return err
	}

	val, _ := systemSCAccount.DataTrieTracker().RetrieveValue(token)
	esdtMetaData := ESDTGlobalMetadataFromBytes(val)
	esdtMetaData.Paused = e.pause
	err = systemSCAccount.DataTrieTracker().SaveKeyValue(token, esdtMetaData.ToBytes())
	if err != nil {
		return err
	}

	return e.accounts.SaveAccount(systemSCAccount)
}

func (e *esdtPause) getSystemAccount() (state.UserAccountHandler, error) {
	systemSCAccount, err := e.accounts.LoadAccount(core.SystemAccountAddress)
	if err != nil {
		return nil, err
	}

	userAcc, ok := systemSCAccount.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return userAcc, nil
}

// IsPaused returns true if the token is paused
func (e *esdtPause) IsPaused(pauseKey []byte) bool {
	systemSCAccount, err := e.getSystemAccount()
	if err != nil {
		return false
	}

	val, _ := systemSCAccount.DataTrieTracker().RetrieveValue(pauseKey)
	if len(val) != lengthOfESDTMetadata {
		return false
	}
	esdtMetaData := ESDTGlobalMetadataFromBytes(val)

	return esdtMetaData.Paused
}

// IsInterfaceNil returns true if underlying object in nil
func (e *esdtPause) IsInterfaceNil() bool {
	return e == nil
}
