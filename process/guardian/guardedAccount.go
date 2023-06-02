package guardian

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/guardians"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var guardianKey = []byte(core.ProtectedKeyPrefix + core.GuardiansKeyIdentifier)

type guardedAccount struct {
	marshaller                    marshal.Marshalizer
	epochNotifier                 vmcommon.EpochNotifier
	mutEpoch                      sync.RWMutex
	guardianActivationEpochsDelay uint32
	currentEpoch                  uint32
}

// NewGuardedAccount creates a new guarded account
func NewGuardedAccount(
	marshaller marshal.Marshalizer,
	epochNotifier vmcommon.EpochNotifier,
	setGuardianEpochsDelay uint32,
) (*guardedAccount, error) {
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if setGuardianEpochsDelay == 0 {
		return nil, process.ErrInvalidSetGuardianEpochsDelay
	}

	agc := &guardedAccount{
		marshaller:                    marshaller,
		epochNotifier:                 epochNotifier,
		guardianActivationEpochsDelay: setGuardianEpochsDelay,
	}

	epochNotifier.RegisterNotifyHandler(agc)

	return agc, nil
}

// GetActiveGuardian returns the active guardian
func (agc *guardedAccount) GetActiveGuardian(uah vmcommon.UserAccountHandler) ([]byte, error) {
	configuredGuardians, err := agc.getVmUserAccountConfiguredGuardian(uah)
	if err != nil {
		return nil, err
	}

	guardian, err := agc.getActiveGuardian(configuredGuardians)
	if err != nil {
		return nil, err
	}

	return guardian.Address, nil
}

// CleanOtherThanActive cleans the pending guardian or old/disabled guardian, if any
func (agc *guardedAccount) CleanOtherThanActive(uah vmcommon.UserAccountHandler) {
	configuredGuardians, err := agc.getVmUserAccountConfiguredGuardian(uah)
	if err != nil {
		return
	}

	activeGuardian, err := agc.getActiveGuardian(configuredGuardians)
	if err != nil {
		configuredGuardians.Slice = []*guardians.Guardian{}
	} else {
		configuredGuardians.Slice = []*guardians.Guardian{activeGuardian}
	}

	_ = agc.saveAccountGuardians(uah, configuredGuardians)
}

// HasActiveGuardian returns true if the account has an active guardian configured, false otherwise
func (agc *guardedAccount) HasActiveGuardian(uah state.UserAccountHandler) bool {
	if check.IfNil(uah) {
		return false
	}

	configuredGuardians, err := agc.getConfiguredGuardians(uah)
	if err != nil {
		return false
	}
	_, err = agc.getActiveGuardian(configuredGuardians)

	return err == nil
}

// HasPendingGuardian return true if the account has a pending guardian, false otherwise
func (agc *guardedAccount) HasPendingGuardian(uah state.UserAccountHandler) bool {
	if check.IfNil(uah) {
		return false
	}

	configuredGuardians, err := agc.getConfiguredGuardians(uah)
	if err != nil {
		return false
	}

	_, err = agc.getPendingGuardian(configuredGuardians)

	return err == nil
}

// SetGuardian sets a guardian for an account
func (agc *guardedAccount) SetGuardian(uah vmcommon.UserAccountHandler, guardianAddress []byte, txGuardianAddress []byte, guardianServiceUID []byte) error {
	stateUserAccount, ok := uah.(state.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if len(guardianServiceUID) == 0 {
		return process.ErrNilGuardianServiceUID
	}

	if len(txGuardianAddress) > 0 {
		return agc.instantSetGuardian(stateUserAccount, guardianAddress, txGuardianAddress, guardianServiceUID)
	}

	agc.mutEpoch.RLock()
	guardian := &guardians.Guardian{
		Address:         guardianAddress,
		ActivationEpoch: agc.currentEpoch + agc.guardianActivationEpochsDelay,
		ServiceUID:      guardianServiceUID,
	}
	agc.mutEpoch.RUnlock()

	return agc.setAccountGuardian(stateUserAccount, guardian)
}

func (agc *guardedAccount) getVmUserAccountConfiguredGuardian(uah vmcommon.UserAccountHandler) (*guardians.Guardians, error) {
	stateUserAccount, ok := uah.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	configuredGuardians, err := agc.getConfiguredGuardians(stateUserAccount)
	if err != nil {
		return nil, err
	}
	if len(configuredGuardians.Slice) == 0 {
		return nil, process.ErrAccountHasNoGuardianSet
	}

	return configuredGuardians, nil
}

func (agc *guardedAccount) setAccountGuardian(uah state.UserAccountHandler, guardian *guardians.Guardian) error {
	configuredGuardians, err := agc.getConfiguredGuardians(uah)
	if err != nil {
		return err
	}

	newGuardians, err := agc.updateGuardians(guardian, configuredGuardians)
	if err != nil {
		return err
	}

	accHandler, ok := uah.(vmcommon.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	return agc.saveAccountGuardians(accHandler, newGuardians)
}

func (agc *guardedAccount) instantSetGuardian(
	uah state.UserAccountHandler,
	guardianAddress []byte,
	txGuardianAddress []byte,
	guardianServiceUID []byte,
) error {
	accountGuardians, err := agc.getConfiguredGuardians(uah)
	if err != nil {
		return err
	}

	activeGuardian, err := agc.getActiveGuardian(accountGuardians)
	if err != nil {
		return err
	}

	if !bytes.Equal(activeGuardian.Address, txGuardianAddress) {
		return process.ErrTransactionAndAccountGuardianMismatch
	}

	// immediately set the new guardian
	agc.mutEpoch.RLock()
	guardian := &guardians.Guardian{
		Address:         guardianAddress,
		ActivationEpoch: agc.currentEpoch,
		ServiceUID:      guardianServiceUID,
	}
	agc.mutEpoch.RUnlock()

	accountGuardians.Slice = []*guardians.Guardian{guardian}
	accHandler, ok := uah.(vmcommon.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	return agc.saveAccountGuardians(accHandler, accountGuardians)
}

// TODO: add constraints on not co-signed txs on interceptor, for setGuardian
// 1. Gas price cannot exceed a preconfigured limit
// 2. If there is already one guardian pending, do not allow setting another one
func (agc *guardedAccount) updateGuardians(newGuardian *guardians.Guardian, accountGuardians *guardians.Guardians) (*guardians.Guardians, error) {
	numSetGuardians := len(accountGuardians.Slice)

	if numSetGuardians == 0 {
		accountGuardians.Slice = append(accountGuardians.Slice, newGuardian)
		return accountGuardians, nil
	}

	activeGuardian, err := agc.getActiveGuardian(accountGuardians)
	if err != nil {
		// no active guardian, do not replace the already pending guardian
		return nil, fmt.Errorf("%w in updateGuardians, with %d configured guardians", err, numSetGuardians)
	}
	accountGuardians.Slice = []*guardians.Guardian{activeGuardian, newGuardian}

	return accountGuardians, nil
}

func (agc *guardedAccount) saveAccountGuardians(account vmcommon.UserAccountHandler, accountGuardians *guardians.Guardians) error {
	marshalledData, err := agc.marshaller.Marshal(accountGuardians)
	if err != nil {
		return err
	}

	return account.AccountDataHandler().SaveKeyValue(guardianKey, marshalledData)
}

func (agc *guardedAccount) getConfiguredGuardians(uah state.UserAccountHandler) (*guardians.Guardians, error) {
	guardiansMarshalled, _, err := uah.RetrieveValue(guardianKey)
	if err != nil || len(guardiansMarshalled) == 0 {
		return &guardians.Guardians{Slice: make([]*guardians.Guardian, 0)}, nil
	}

	configuredGuardians := &guardians.Guardians{}
	err = agc.marshaller.Unmarshal(configuredGuardians, guardiansMarshalled)
	if err != nil {
		return nil, err
	}

	return configuredGuardians, nil
}

func (agc *guardedAccount) getActiveGuardian(gs *guardians.Guardians) (*guardians.Guardian, error) {
	agc.mutEpoch.RLock()
	defer agc.mutEpoch.RUnlock()

	var selectedGuardian *guardians.Guardian
	for _, guardian := range gs.Slice {
		if guardian == nil {
			continue
		}
		if guardian.ActivationEpoch > agc.currentEpoch {
			continue
		}
		if selectedGuardian == nil {
			selectedGuardian = guardian
			continue
		}

		// get the most recent active guardian
		if selectedGuardian.ActivationEpoch < guardian.ActivationEpoch {
			selectedGuardian = guardian
		}
	}

	if selectedGuardian == nil {
		return nil, process.ErrAccountHasNoActiveGuardian
	}

	return selectedGuardian, nil
}

// GetConfiguredGuardians returns the configured guardians for an account
func (agc *guardedAccount) GetConfiguredGuardians(uah state.UserAccountHandler) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
	configuredGuardians, err := agc.getConfiguredGuardians(uah)
	if err != nil {
		return nil, nil, err
	}

	active, _ = agc.getActiveGuardian(configuredGuardians)
	pending, _ = agc.getPendingGuardian(configuredGuardians)

	return
}

func (agc *guardedAccount) getPendingGuardian(gs *guardians.Guardians) (*guardians.Guardian, error) {
	if gs == nil {
		return nil, process.ErrAccountHasNoPendingGuardian
	}

	agc.mutEpoch.RLock()
	defer agc.mutEpoch.RUnlock()

	for _, guardian := range gs.Slice {
		if guardian == nil {
			continue
		}
		if guardian.ActivationEpoch <= agc.currentEpoch {
			continue
		}
		return guardian, nil
	}

	return nil, process.ErrAccountHasNoPendingGuardian
}

// EpochConfirmed is the registered callback function for the epoch change notifier
func (agc *guardedAccount) EpochConfirmed(epoch uint32, _ uint64) {
	agc.mutEpoch.Lock()
	agc.currentEpoch = epoch
	agc.mutEpoch.Unlock()
}

// IsInterfaceNil returns true if the receiver is nil
func (agc *guardedAccount) IsInterfaceNil() bool {
	return agc == nil
}
