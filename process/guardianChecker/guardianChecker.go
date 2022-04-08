package guardianChecker

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/guardians"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var guardianKey = []byte(core.ElrondProtectedKeyPrefix + core.GuardiansKeyIdentifier)

type guardianChecker struct {
	marshaller    marshal.Marshalizer
	epochNotifier vmcommon.EpochNotifier
	mutEpoch      sync.RWMutex
	currentEpoch  uint32
}

// NewAccountGuardianChecker creates a new account guardian checker
func NewAccountGuardianChecker(marshaller marshal.Marshalizer, epochNotifier vmcommon.EpochNotifier) (*guardianChecker, error) {
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	agc := &guardianChecker{
		marshaller:    marshaller,
		epochNotifier: epochNotifier,
	}

	epochNotifier.RegisterNotifyHandler(agc)

	return agc, nil
}

// GetActiveGuardian returns the active guardian
func (agc *guardianChecker) GetActiveGuardian(uah data.UserAccountHandler) ([]byte, error) {
	guardiansMarshalled, err := uah.RetrieveValueFromDataTrieTracker(guardianKey)
	if err != nil {
		return nil, err
	}

	if len(guardiansMarshalled) == 0 {
		return nil, process.ErrAccountHasNoGuardianSet
	}

	configuredGuardians := &guardians.Guardians{}
	err = agc.marshaller.Unmarshal(configuredGuardians, guardiansMarshalled)
	if err != nil {
		return nil, err
	}

	guardian, err := agc.getActiveGuardian(configuredGuardians)
	if err != nil {
		return nil, err
	}

	return guardian.Address, nil
}

func (agc *guardianChecker) getActiveGuardian(gs *guardians.Guardians) (*guardians.Guardian, error) {
	if gs == nil {
		return nil, process.ErrAccountHasNoGuardianSet
	}

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
func (agc *guardianChecker) EpochConfirmed(epoch uint32, _ uint64) {
	agc.mutEpoch.Lock()
	agc.currentEpoch = epoch
	agc.mutEpoch.Unlock()
}

// IsInterfaceNil returns true if the receiver is nil
func (agc *guardianChecker) IsInterfaceNil() bool {
	return agc == nil
}
