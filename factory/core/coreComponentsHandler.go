package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
)

var _ factory.ComponentHandler = (*managedCoreComponents)(nil)
var _ factory.CoreComponentsHolder = (*managedCoreComponents)(nil)
var _ factory.CoreComponentsHandler = (*managedCoreComponents)(nil)

// managedCoreComponents is an implementation of core components handler that can create, close and access the core components
type managedCoreComponents struct {
	coreComponentsFactory *coreComponentsFactory
	*coreComponents
	mutCoreComponents sync.RWMutex
}

// NewManagedCoreComponents creates a new core components handler implementation
func NewManagedCoreComponents(ccf *coreComponentsFactory) (*managedCoreComponents, error) {
	if ccf == nil {
		return nil, errors.ErrNilCoreComponentsFactory
	}

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
		return fmt.Errorf("%w: %v", errors.ErrCoreComponentsFactoryCreate, err)
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

	if mcc.coreComponents == nil {
		return nil
	}

	err := mcc.coreComponents.Close()
	if err != nil {
		return err
	}
	mcc.coreComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mcc *managedCoreComponents) CheckSubcomponents() error {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return errors.ErrNilCoreComponents
	}
	if check.IfNil(mcc.internalMarshalizer) {
		return errors.ErrNilInternalMarshalizer
	}
	if check.IfNil(mcc.txSignMarshalizer) {
		return errors.ErrNilTxSignMarshalizer
	}
	if check.IfNil(mcc.vmMarshalizer) {
		return errors.ErrNilVmMarshalizer
	}
	if check.IfNil(mcc.hasher) {
		return errors.ErrNilHasher
	}
	if check.IfNil(mcc.txSignHasher) {
		return errors.ErrNilTxSignHasher
	}
	if check.IfNil(mcc.uint64ByteSliceConverter) {
		return errors.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(mcc.addressPubKeyConverter) {
		return errors.ErrNilAddressPublicKeyConverter
	}
	if check.IfNil(mcc.validatorPubKeyConverter) {
		return errors.ErrNilValidatorPublicKeyConverter
	}
	if check.IfNil(mcc.pathHandler) {
		return errors.ErrNilPathHandler
	}
	if check.IfNil(mcc.watchdog) {
		return errors.ErrNilWatchdog
	}
	if check.IfNil(mcc.alarmScheduler) {
		return errors.ErrNilAlarmScheduler
	}
	if check.IfNil(mcc.syncTimer) {
		return errors.ErrNilSyncTimer
	}
	if check.IfNil(mcc.roundHandler) {
		return errors.ErrNilRoundHandler
	}
	if check.IfNil(mcc.economicsData) {
		return errors.ErrNilEconomicsHandler
	}
	if check.IfNil(mcc.ratingsData) {
		return errors.ErrNilRatingsInfoHandler
	}
	if check.IfNil(mcc.rater) {
		return errors.ErrNilRater
	}
	if check.IfNil(mcc.nodesSetupHandler) {
		return errors.ErrNilNodesConfig
	}
	if check.IfNil(mcc.epochNotifier) {
		return errors.ErrNilEpochNotifier
	}
	if check.IfNil(mcc.roundNotifier) {
		return errors.ErrNilRoundNotifier
	}
	if check.IfNil(mcc.processStatusHandler) {
		return errors.ErrNilProcessStatusHandler
	}
	if check.IfNil(mcc.enableEpochsHandler) {
		return errors.ErrNilEnableEpochsHandler
	}
	if check.IfNil(mcc.chainParametersHandler) {
		return errors.ErrNilChainParametersHandler
	}
	if check.IfNil(mcc.fieldsSizeChecker) {
		return errors.ErrNilFieldsSizeChecker
	}
	if len(mcc.chainID) == 0 {
		return errors.ErrInvalidChainID
	}
	if mcc.minTransactionVersion == 0 {
		return errors.ErrInvalidTransactionVersion
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

	return mcc.internalMarshalizer
}

// SetInternalMarshalizer sets the internal marshalizer to the one given as parameter
func (mcc *managedCoreComponents) SetInternalMarshalizer(m marshal.Marshalizer) error {
	mcc.mutCoreComponents.Lock()
	defer mcc.mutCoreComponents.Unlock()

	if mcc.coreComponents == nil {
		return errors.ErrNilCoreComponents
	}

	mcc.internalMarshalizer = m

	return nil
}

// TxMarshalizer returns the core components tx marshalizer
func (mcc *managedCoreComponents) TxMarshalizer() marshal.Marshalizer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.txSignMarshalizer
}

// VmMarshalizer returns the core components vm marshalizer
func (mcc *managedCoreComponents) VmMarshalizer() marshal.Marshalizer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.vmMarshalizer
}

// Hasher returns the core components Hasher
func (mcc *managedCoreComponents) Hasher() hashing.Hasher {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.hasher
}

// TxSignHasher returns the core components hasher to be used for signed transaction hashes
func (mcc *managedCoreComponents) TxSignHasher() hashing.Hasher {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.txSignHasher
}

// Uint64ByteSliceConverter returns the core component converter between a byte slice and uint64
func (mcc *managedCoreComponents) Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.uint64ByteSliceConverter
}

// AddressPubKeyConverter returns the address to public key converter
func (mcc *managedCoreComponents) AddressPubKeyConverter() core.PubkeyConverter {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.addressPubKeyConverter
}

// ValidatorPubKeyConverter returns the validator public key converter
func (mcc *managedCoreComponents) ValidatorPubKeyConverter() core.PubkeyConverter {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.validatorPubKeyConverter
}

// PathHandler returns the core components path handler
func (mcc *managedCoreComponents) PathHandler() storage.PathManagerHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.pathHandler
}

// ChainID returns the core components chainID
func (mcc *managedCoreComponents) ChainID() string {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return ""
	}

	return mcc.chainID
}

// MinTransactionVersion returns the minimum transaction version
func (mcc *managedCoreComponents) MinTransactionVersion() uint32 {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return 0
	}

	return mcc.minTransactionVersion
}

// TxVersionChecker returns the transaction version checker
func (mcc *managedCoreComponents) TxVersionChecker() process.TxVersionCheckerHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.txVersionChecker
}

// EncodedAddressLen returns the length of the encoded address
func (mcc *managedCoreComponents) EncodedAddressLen() uint32 {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return 0
	}

	return mcc.encodedAddressLen
}

// AlarmScheduler returns the alarm scheduler
func (mcc *managedCoreComponents) AlarmScheduler() core.TimersScheduler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.alarmScheduler
}

// SyncTimer returns the ntp synchronization timer
func (mcc *managedCoreComponents) SyncTimer() ntp.SyncTimer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.syncTimer
}

// GenesisTime returns the time of the genesis block
func (mcc *managedCoreComponents) GenesisTime() time.Time {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return time.Time{}
	}

	return mcc.genesisTime
}

// SupernovaGenesisTime returns the time for supernova round activation
func (mcc *managedCoreComponents) SupernovaGenesisTime() time.Time {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return time.Time{}
	}

	return mcc.supernovaGenesisTime
}

// Watchdog returns the minimum watchdog
func (mcc *managedCoreComponents) Watchdog() core.WatchdogTimer {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.watchdog
}

// EconomicsData returns the configured economics data
func (mcc *managedCoreComponents) EconomicsData() process.EconomicsDataHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.economicsData
}

// APIEconomicsData returns the configured economics data to be used on the REST API sub-system
func (mcc *managedCoreComponents) APIEconomicsData() process.EconomicsDataHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.apiEconomicsData
}

// RatingsData returns the configured ratings data
func (mcc *managedCoreComponents) RatingsData() process.RatingsInfoHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.ratingsData
}

// Rater returns the rater
func (mcc *managedCoreComponents) Rater() sharding.PeerAccountListAndRatingHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.rater
}

// GenesisNodesSetup returns the genesis nodes setup
func (mcc *managedCoreComponents) GenesisNodesSetup() sharding.GenesisNodesSetupHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.nodesSetupHandler
}

// RoundHandler returns the roundHandler
func (mcc *managedCoreComponents) RoundHandler() consensus.RoundHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.roundHandler
}

// NodesShuffler returns the nodes shuffler
func (mcc *managedCoreComponents) NodesShuffler() nodesCoordinator.NodesShuffler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.nodesShuffler
}

// EpochNotifier returns the epoch notifier
func (mcc *managedCoreComponents) EpochNotifier() process.EpochNotifier {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.epochNotifier
}

// RoundNotifier returns the epoch notifier
func (mcc *managedCoreComponents) RoundNotifier() process.RoundNotifier {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.roundNotifier
}

// ChainParametersSubscriber returns the chain parameters subscriber
func (mcc *managedCoreComponents) ChainParametersSubscriber() process.ChainParametersSubscriber {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.chainParametersSubscriber
}

// EnableRoundsHandler returns the rounds activation handler
func (mcc *managedCoreComponents) EnableRoundsHandler() common.EnableRoundsHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.enableRoundsHandler
}

// EpochStartNotifierWithConfirm returns the epoch notifier with confirm
func (mcc *managedCoreComponents) EpochStartNotifierWithConfirm() factory.EpochStartNotifierWithConfirm {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.epochStartNotifierWithConfirm
}

// ChanStopNodeProcess returns the channel for stop node
func (mcc *managedCoreComponents) ChanStopNodeProcess() chan endProcess.ArgEndProcess {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.chanStopNodeProcess
}

// NodeTypeProvider returns the node type provider
func (mcc *managedCoreComponents) NodeTypeProvider() core.NodeTypeProviderHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.nodeTypeProvider
}

// WasmVMChangeLocker returns the wasm VM change locker
func (mcc *managedCoreComponents) WasmVMChangeLocker() common.Locker {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.wasmVMChangeLocker
}

// ProcessStatusHandler returns the process status handler
func (mcc *managedCoreComponents) ProcessStatusHandler() common.ProcessStatusHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.processStatusHandler
}

// HardforkTriggerPubKey returns the hardfork source public key
func (mcc *managedCoreComponents) HardforkTriggerPubKey() []byte {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.hardforkTriggerPubKey
}

// EnableEpochsHandler returns the enable epochs handler
func (mcc *managedCoreComponents) EnableEpochsHandler() common.EnableEpochsHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.enableEpochsHandler
}

// ChainParametersHandler returns the chain parameters handler
func (mcc *managedCoreComponents) ChainParametersHandler() process.ChainParametersHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.chainParametersHandler
}

// FieldsSizeChecker returns the fields size checker component
func (mcc *managedCoreComponents) FieldsSizeChecker() common.FieldsSizeChecker {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.fieldsSizeChecker
}

// EpochChangeGracePeriodHandler returns the epoch change grace period handler component
func (mcc *managedCoreComponents) EpochChangeGracePeriodHandler() common.EpochChangeGracePeriodHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.epochChangeGracePeriodHandler
}

// ProcessConfigsHandler returns the process configs handler component
func (mcc *managedCoreComponents) ProcessConfigsHandler() common.ProcessConfigsHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.processConfigsHandler
}

// CommonConfigsHandler returns the epoch start configs handler component
func (mcc *managedCoreComponents) CommonConfigsHandler() common.CommonConfigsHandler {
	mcc.mutCoreComponents.RLock()
	defer mcc.mutCoreComponents.RUnlock()

	if mcc.coreComponents == nil {
		return nil
	}

	return mcc.epochStartConfigsHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (mcc *managedCoreComponents) IsInterfaceNil() bool {
	return mcc == nil
}

// String returns the name of the component
func (mcc *managedCoreComponents) String() string {
	return factory.CoreComponentsName
}
