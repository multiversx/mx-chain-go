package factory

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ ComponentHandler = (*managedCoreComponents)(nil)
var _ CoreComponentsHolder = (*managedCoreComponents)(nil)
var _ CoreComponentsHandler = (*managedCoreComponents)(nil)

// CoreComponentsHandlerArgs holds the arguments required to create a core components handler
type CoreComponentsHandlerArgs CoreComponentsFactoryArgs

// managedCoreComponents is an implementation of core components handler that can create, close and access the core components
type managedCoreComponents struct {
	coreComponentsFactory *coreComponentsFactory
	*coreComponents
	mutCoreComponents sync.RWMutex
}

// NewManagedCoreComponents creates a new core components handler implementation
func NewManagedCoreComponents(args CoreComponentsHandlerArgs) (*managedCoreComponents, error) {
	ccf := NewCoreComponentsFactory(CoreComponentsFactoryArgs(args))
	mcc := &managedCoreComponents{
		coreComponents:        nil,
		coreComponentsFactory: ccf,
	}
	return mcc, nil
}

// Create creates the core components
func (mcc *managedCoreComponents) Create() error {
	cc, err := mcc.coreComponentsFactory.Create()
	if err != nil {
		return err
	}

	mcc.mutCoreComponents.Lock()
	mcc.coreComponents = cc
	mcc.mutCoreComponents.Unlock()

	return nil
}

// Close closes the managed core components
func (mcc *managedCoreComponents) Close() error {
	mcc.mutCoreComponents.Lock()
	defer mcc.mutCoreComponents.Unlock()

	if mcc.coreComponents != nil {
		err := mcc.coreComponents.Close()
		if err != nil {
			return err
		}
		mcc.coreComponents = nil
	}

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mcc *managedCoreComponents) CheckSubcomponents() error {
	mcc.mutCoreComponents.Lock()
	defer mcc.mutCoreComponents.Unlock()

	if mcc.coreComponents == nil {
		return ErrNilCoreComponentsHolder
	}
	if check.IfNil(mcc.internalMarshalizer) {
		return ErrNilInternalMarshalizer
	}
	if check.IfNil(mcc.txSignMarshalizer) {
		return ErrNilTxSignMarshalizer
	}
	if check.IfNil(mcc.vmMarshalizer) {
		return ErrNilVmMarshalizer
	}
	if check.IfNil(mcc.hasher) {
		return ErrNilHasher
	}
	if check.IfNil(mcc.uint64ByteSliceConverter) {
		return ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(mcc.addressPubKeyConverter) {
		return ErrNilAddressPublicKeyConverter
	}
	if check.IfNil(mcc.validatorPubKeyConverter) {
		return ErrNilValidatorPublicKeyConverter
	}
	if check.IfNil(mcc.statusHandler) {
		return ErrNilStatusHandler
	}
	if check.IfNil(mcc.pathHandler) {
		return ErrNilPathHandler
	}
	if check.IfNil(mcc.watchdog) {
		return ErrNilWatchdog
	}
	if check.IfNil(mcc.alarmScheduler) {
		return ErrNilAlarmScheduler
	}
	if check.IfNil(mcc.syncTimer) {
		return ErrNilSyncTimer
	}
	if check.IfNil(mcc.rounder) {
		return ErrNilRounder
	}
	if check.IfNil(mcc.economicsData) {
		return ErrNilEconomicsHandler
	}
	if check.IfNil(mcc.ratingsData) {
		return ErrNilRatingsInfoHandler
	}
	if check.IfNil(mcc.rater) {
		return ErrNilRater
	}
	if check.IfNil(mcc.nodesSetupHandler) {
		return ErrNilNodesConfig
	}
	if len(mcc.chainID) == 0 {
		return ErrInvalidChainID
	}
	if mcc.minTransactionVersion == 0 {
		return ErrInvalidTransactionVersion
	}

	return nil
}

// InternalMarshalizer returns the core components internal marshalizer
func (mcc *managedCoreComponents) InternalMarshalizer() marshal.Marshalizer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.internalMarshalizer
}

// SetInternalMarshalizer sets the internal marshalizer to the one given as parameter
func (mcc *managedCoreComponents) SetInternalMarshalizer(m marshal.Marshalizer) error {
	mcc.mutCoreComponents.Lock()
	defer mcc.mutCoreComponents.Unlock()

	if mcc.coreComponents == nil {
		return ErrNilCoreComponents
	}

	mcc.coreComponents.internalMarshalizer = m

	return nil
}

// TxMarshalizer returns the core components tx marshalizer
func (mcc *managedCoreComponents) TxMarshalizer() marshal.Marshalizer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.txSignMarshalizer
}

// VmMarshalizer returns the core components vm marshalizer
func (mcc *managedCoreComponents) VmMarshalizer() marshal.Marshalizer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.vmMarshalizer
}

// Hasher returns the core components Hasher
func (mcc *managedCoreComponents) Hasher() hashing.Hasher {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.hasher
}

// Uint64ByteSliceConverter returns the core component converter between a byte slice and uint64
func (mcc *managedCoreComponents) Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.uint64ByteSliceConverter
}

// AddressPubKeyConverter returns the address to public key converter
func (mcc *managedCoreComponents) AddressPubKeyConverter() core.PubkeyConverter {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.addressPubKeyConverter
}

// ValidatorPubKeyConverter returns the validator public key converter
func (mcc *managedCoreComponents) ValidatorPubKeyConverter() core.PubkeyConverter {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.validatorPubKeyConverter
}

// StatusHandler returns the core components status handler
func (mcc *managedCoreComponents) StatusHandler() core.AppStatusHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.statusHandler
}

// SetStatusHandler allows the change of the status handler
func (mcc *managedCoreComponents) SetStatusHandler(statusHandler core.AppStatusHandler) error {
	if check.IfNil(statusHandler) {
		return ErrNilStatusHandler
	}

	mcc.mutCoreComponents.Lock()
	defer mcc.mutCoreComponents.Unlock()

	if mcc.coreComponents == nil {
		return ErrNilCoreComponents
	}

	mcc.coreComponents.statusHandler = statusHandler

	return nil
}

// PathHandler returns the core components path handler
func (mcc *managedCoreComponents) PathHandler() storage.PathManagerHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.pathHandler
}

// ChainID returns the core components chainID
func (mcc *managedCoreComponents) ChainID() string {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return ""
	}

	return mcc.coreComponents.chainID
}

// MinTransactionVersion returns the minimum transaction version
func (mcc *managedCoreComponents) MinTransactionVersion() uint32 {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return 0
	}

	return mcc.coreComponents.minTransactionVersion
}

// AlarmScheduler returns the alarm scheduler
func (mcc *managedCoreComponents) AlarmScheduler() core.TimersScheduler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.alarmScheduler
}

// SyncTimer returns the ntp synchronization timer
func (mcc *managedCoreComponents) SyncTimer() ntp.SyncTimer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.syncTimer
}

// GenesisTime returns the time of the genesis block
func (mcc *managedCoreComponents) GenesisTime() time.Time {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return time.Time{}
	}

	return mcc.coreComponents.genesisTime
}

// Watchdog returns the minimum watchdog
func (mcc *managedCoreComponents) Watchdog() core.WatchdogTimer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.watchdog
}

// EconomicsData returns the configured economics data
func (mcc *managedCoreComponents) EconomicsData() process.EconomicsHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.economicsData
}

// RatingsData returns the configured ratings data
func (mcc *managedCoreComponents) RatingsData() process.RatingsInfoHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.ratingsData
}

// Rater returns the rater
func (mcc *managedCoreComponents) Rater() sharding.PeerAccountListAndRatingHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.rater
}

// GenesisNodesSetup returns the genesis nodes setup
func (mcc *managedCoreComponents) GenesisNodesSetup() NodesSetupHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.nodesSetupHandler
}

// Rounder returns the rounder
func (mcc *managedCoreComponents) Rounder() consensus.Rounder {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.coreComponents.rounder
}

// IsInterfaceNil returns true if there is no value under the interface
func (mcc *managedCoreComponents) IsInterfaceNil() bool {
	return mcc == nil
}
