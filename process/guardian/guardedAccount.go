package guardian

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/guardians"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var guardianKey = []byte(core.ElrondProtectedKeyPrefix + core.GuardiansKeyIdentifier)

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
	if err != nil {
		return false
	}
	return true
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
	if err != nil {
		return false
	}
	return true
}

// SetGuardian sets a guardian for an account
func (agc *guardedAccount) SetGuardian(uah vmcommon.UserAccountHandler, guardianAddress []byte, txGuardianAddress []byte) error {
	stateUserAccount, ok := uah.(state.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if len(txGuardianAddress) > 0 {
		return agc.instantSetGuardian(stateUserAccount, guardianAddress, txGuardianAddress)
	}

	guardian := &guardians.Guardian{
		Address:         guardianAddress,
		ActivationEpoch: agc.currentEpoch + agc.guardianActivationEpochsDelay,
	}

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
	guardian := &guardians.Guardian{
		Address:         guardianAddress,
		ActivationEpoch: agc.currentEpoch,
	}

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

	if bytes.Equal(activeGuardian.Address, newGuardian.Address) {
		accountGuardians.Slice = []*guardians.Guardian{activeGuardian}
	} else {
		accountGuardians.Slice = []*guardians.Guardian{activeGuardian, newGuardian}
	}

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
	guardiansMarshalled, err := uah.RetrieveValueFromDataTrieTracker(guardianKey)
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
		if guardian.ActivationEpoch < agc.currentEpoch {
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
