package guardian

import (
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

const epochsForActivation = 10 // TODO: take from config

type guardedAccount struct {
	marshaller               marshal.Marshalizer
	epochNotifier            vmcommon.EpochNotifier
	mutEpoch                 sync.RWMutex
	currentEpoch             uint32
	guardianActivationEpochs uint32
}

// NewGuardedAccount creates a new guarded account
func NewGuardedAccount(marshaller marshal.Marshalizer, epochNotifier vmcommon.EpochNotifier) (*guardedAccount, error) {
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	agc := &guardedAccount{
		marshaller:               marshaller,
		epochNotifier:            epochNotifier,
		guardianActivationEpochs: epochsForActivation,
	}

	epochNotifier.RegisterNotifyHandler(agc)

	return agc, nil
}

// GetActiveGuardian returns the active guardian
func (agc *guardedAccount) GetActiveGuardian(uah vmcommon.UserAccountHandler) ([]byte, error) {
	stateUserAccount, ok := uah.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	configuredGuardians, err := agc.getConfiguredGuardians(stateUserAccount)
	if err != nil {
		return nil, err
	}
	if len(configuredGuardians.Data) == 0 {
		return nil, process.ErrAccountHasNoGuardianSet
	}

	guardian, err := agc.getActiveGuardian(*configuredGuardians)
	if err != nil {
		return nil, err
	}

	return guardian.Address, nil
}

// SetGuardian sets a guardian for an account
func (agc *guardedAccount) SetGuardian(uah vmcommon.UserAccountHandler, guardianAddress []byte) error {
	guardian := &guardians.Guardian{
		Address:         guardianAddress,
		ActivationEpoch: agc.currentEpoch + agc.guardianActivationEpochs,
	}

	stateUserAccount, ok := uah.(state.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	return agc.setAccountGuardian(stateUserAccount, guardian)
}

func (agc *guardedAccount) setAccountGuardian(uah state.UserAccountHandler, guardian *guardians.Guardian) error {
	configuredGuardians, err := agc.getConfiguredGuardians(uah)
	if err != nil {
		return err
	}
	newGuardians, err := agc.updateGuardians(guardian, *configuredGuardians)
	if err != nil {
		return err
	}

	accHandler, ok := uah.(vmcommon.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	return agc.saveAccountGuardians(accHandler, *newGuardians)
}

func (agc *guardedAccount) updateGuardians(newGuardian *guardians.Guardian, accountGuardians guardians.Guardians) (*guardians.Guardians, error) {
	numSetGuardians := len(accountGuardians.Data)

	if numSetGuardians == 0 {
		accountGuardians.Data = append(accountGuardians.Data, newGuardian)
		return &accountGuardians, nil
	}

	activeGuardian, err := agc.getActiveGuardian(accountGuardians)
	if err != nil {
		// no active guardian, do not replace the already pending guardian
		return nil, fmt.Errorf("%w in updateGuardians, with %d configured guardians", err, numSetGuardians)
	}

	if activeGuardian.Equal(newGuardian) {
		accountGuardians.Data = []*guardians.Guardian{activeGuardian}
	} else {
		accountGuardians.Data = []*guardians.Guardian{activeGuardian, newGuardian}
	}

	return &accountGuardians, nil
}

func (agc *guardedAccount) saveAccountGuardians(account vmcommon.UserAccountHandler, accountGuardians guardians.Guardians) error {
	marshalledData, err := agc.marshaller.Marshal(accountGuardians)
	if err != nil {
		return err
	}

	return account.AccountDataHandler().SaveKeyValue(guardianKey, marshalledData)
}

func (agc *guardedAccount) getConfiguredGuardians(uah state.UserAccountHandler) (*guardians.Guardians, error) {
	guardiansMarshalled, err := uah.RetrieveValueFromDataTrieTracker(guardianKey)
	if err != nil {
		return nil, err
	}
	if len(guardiansMarshalled) == 0 {
		return &guardians.Guardians{Data: make([]*guardians.Guardian, 0)}, nil
	}

	configuredGuardians := &guardians.Guardians{}
	err = agc.marshaller.Unmarshal(configuredGuardians, guardiansMarshalled)
	if err != nil {
		return nil, err
	}

	return configuredGuardians, nil
}

func (agc *guardedAccount) getActiveGuardian(gs guardians.Guardians) (*guardians.Guardian, error) {
	agc.mutEpoch.RLock()
	defer agc.mutEpoch.RUnlock()

	var selectedGuardian *guardians.Guardian
	for i, guardian := range gs.Data {
		if guardian == nil {
			continue
		}
		if guardian.ActivationEpoch > agc.currentEpoch {
			continue
		}
		if selectedGuardian == nil {
			selectedGuardian = gs.Data[i]
			continue
		}

		// get the most recent active guardian
		if selectedGuardian.ActivationEpoch < guardian.ActivationEpoch {
			selectedGuardian = gs.Data[i]
		}
	}

	if selectedGuardian == nil {
		return nil, process.ErrActiveHasNoActiveGuardian
	}

	return selectedGuardian, nil
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
