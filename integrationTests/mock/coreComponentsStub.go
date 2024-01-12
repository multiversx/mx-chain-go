package mock

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
)

// CoreComponentsStub -
type CoreComponentsStub struct {
	InternalMarshalizerField           marshal.Marshalizer
	TxMarshalizerField                 marshal.Marshalizer
	VmMarshalizerField                 marshal.Marshalizer
	HasherField                        hashing.Hasher
	TxSignHasherField                  hashing.Hasher
	Uint64ByteSliceConverterField      typeConverters.Uint64ByteSliceConverter
	AddressPubKeyConverterField        core.PubkeyConverter
	ValidatorPubKeyConverterField      core.PubkeyConverter
	PathHandlerField                   storage.PathManagerHandler
	ChainIdCalled                      func() string
	MinTransactionVersionCalled        func() uint32
	StatusHandlerField                 core.AppStatusHandler
	WatchdogField                      core.WatchdogTimer
	AlarmSchedulerField                core.TimersScheduler
	SyncTimerField                     ntp.SyncTimer
	RoundHandlerField                  consensus.RoundHandler
	EconomicsDataField                 process.EconomicsDataHandler
	APIEconomicsHandler                process.EconomicsDataHandler
	RatingsDataField                   process.RatingsInfoHandler
	RaterField                         sharding.PeerAccountListAndRatingHandler
	GenesisNodesSetupField             sharding.GenesisNodesSetupHandler
	NodesShufflerField                 nodesCoordinator.NodesShuffler
	EpochNotifierField                 process.EpochNotifier
	RoundNotifierField                 process.RoundNotifier
	EnableRoundsHandlerField           process.EnableRoundsHandler
	EpochStartNotifierWithConfirmField factory.EpochStartNotifierWithConfirm
	ChanStopNodeProcessField           chan endProcess.ArgEndProcess
	GenesisTimeField                   time.Time
	TxVersionCheckField                process.TxVersionCheckerHandler
	NodeTypeProviderField              core.NodeTypeProviderHandler
	WasmVMChangeLockerInternal         common.Locker
	ProcessStatusHandlerInternal       common.ProcessStatusHandler
	HardforkTriggerPubKeyField         []byte
	EnableEpochsHandlerField           common.EnableEpochsHandler
}

// Create -
func (ccs *CoreComponentsStub) Create() error {
	return nil
}

// Close -
func (ccs *CoreComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (ccs *CoreComponentsStub) CheckSubcomponents() error {
	return nil
}

// VmMarshalizer -
func (ccs *CoreComponentsStub) VmMarshalizer() marshal.Marshalizer {
	return ccs.VmMarshalizerField
}

// Watchdog -
func (ccs *CoreComponentsStub) Watchdog() core.WatchdogTimer {
	return ccs.WatchdogField
}

// AlarmScheduler -
func (ccs *CoreComponentsStub) AlarmScheduler() core.TimersScheduler {
	return ccs.AlarmSchedulerField
}

// SyncTimer -
func (ccs *CoreComponentsStub) SyncTimer() ntp.SyncTimer {
	return ccs.SyncTimerField
}

// RoundHandler -
func (ccs *CoreComponentsStub) RoundHandler() consensus.RoundHandler {
	return ccs.RoundHandlerField
}

// EconomicsData -
func (ccs *CoreComponentsStub) EconomicsData() process.EconomicsDataHandler {
	return ccs.EconomicsDataField
}

// APIEconomicsData -
func (ccs *CoreComponentsStub) APIEconomicsData() process.EconomicsDataHandler {
	return ccs.APIEconomicsHandler
}

// RatingsData -
func (ccs *CoreComponentsStub) RatingsData() process.RatingsInfoHandler {
	return ccs.RatingsDataField
}

// Rater -
func (ccs *CoreComponentsStub) Rater() sharding.PeerAccountListAndRatingHandler {
	return ccs.RaterField

}

// GenesisNodesSetup -
func (ccs *CoreComponentsStub) GenesisNodesSetup() sharding.GenesisNodesSetupHandler {
	return ccs.GenesisNodesSetupField
}

// NodesShuffler -
func (ccs *CoreComponentsStub) NodesShuffler() nodesCoordinator.NodesShuffler {
	return ccs.NodesShufflerField
}

// EpochNotifier -
func (ccs *CoreComponentsStub) EpochNotifier() process.EpochNotifier {
	return ccs.EpochNotifierField
}

// RoundNotifier -
func (ccs *CoreComponentsStub) RoundNotifier() process.RoundNotifier {
	return ccs.RoundNotifierField
}

// EnableRoundsHandler -
func (ccs *CoreComponentsStub) EnableRoundsHandler() process.EnableRoundsHandler {
	return ccs.EnableRoundsHandlerField
}

// EpochStartNotifierWithConfirm -
func (ccs *CoreComponentsStub) EpochStartNotifierWithConfirm() factory.EpochStartNotifierWithConfirm {
	return ccs.EpochStartNotifierWithConfirmField
}

// GenesisTime -
func (ccs *CoreComponentsStub) GenesisTime() time.Time {
	return ccs.GenesisTimeField
}

// InternalMarshalizer -
func (ccs *CoreComponentsStub) InternalMarshalizer() marshal.Marshalizer {
	return ccs.InternalMarshalizerField
}

// SetInternalMarshalizer -
func (ccs *CoreComponentsStub) SetInternalMarshalizer(m marshal.Marshalizer) error {
	ccs.InternalMarshalizerField = m
	return nil
}

// TxMarshalizer -
func (ccs *CoreComponentsStub) TxMarshalizer() marshal.Marshalizer {
	return ccs.TxMarshalizerField
}

// Hasher -
func (ccs *CoreComponentsStub) Hasher() hashing.Hasher {
	return ccs.HasherField
}

// TxSignHasher -
func (ccs *CoreComponentsStub) TxSignHasher() hashing.Hasher {
	return ccs.TxSignHasherField
}

// Uint64ByteSliceConverter -
func (ccs *CoreComponentsStub) Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter {
	return ccs.Uint64ByteSliceConverterField
}

// AddressPubKeyConverter -
func (ccs *CoreComponentsStub) AddressPubKeyConverter() core.PubkeyConverter {
	return ccs.AddressPubKeyConverterField
}

// ValidatorPubKeyConverter -
func (ccs *CoreComponentsStub) ValidatorPubKeyConverter() core.PubkeyConverter {
	return ccs.ValidatorPubKeyConverterField
}

// PathHandler -
func (ccs *CoreComponentsStub) PathHandler() storage.PathManagerHandler {
	return ccs.PathHandlerField
}

// ChainID -
func (ccs *CoreComponentsStub) ChainID() string {
	if ccs.ChainIdCalled != nil {
		return ccs.ChainIdCalled()
	}
	return "undefined"
}

// MinTransactionVersion -
func (ccs *CoreComponentsStub) MinTransactionVersion() uint32 {
	if ccs.MinTransactionVersionCalled != nil {
		return ccs.MinTransactionVersionCalled()
	}
	return 1
}

// TxVersionChecker -
func (ccs *CoreComponentsStub) TxVersionChecker() process.TxVersionCheckerHandler {
	return ccs.TxVersionCheckField
}

// EncodedAddressLen -
func (ccs *CoreComponentsStub) EncodedAddressLen() uint32 {
	return uint32(ccs.AddressPubKeyConverter().Len() * 2)
}

// ChanStopNodeProcess -
func (ccs *CoreComponentsStub) ChanStopNodeProcess() chan endProcess.ArgEndProcess {
	return ccs.ChanStopNodeProcessField
}

// NodeTypeProvider -
func (ccs *CoreComponentsStub) NodeTypeProvider() core.NodeTypeProviderHandler {
	return ccs.NodeTypeProviderField
}

// WasmVMChangeLocker -
func (ccs *CoreComponentsStub) WasmVMChangeLocker() common.Locker {
	return ccs.WasmVMChangeLockerInternal
}

// ProcessStatusHandler -
func (ccs *CoreComponentsStub) ProcessStatusHandler() common.ProcessStatusHandler {
	return ccs.ProcessStatusHandlerInternal
}

// String -
func (ccs *CoreComponentsStub) String() string {
	return "CoreComponentsStub"
}

// HardforkTriggerPubKey -
func (ccs *CoreComponentsStub) HardforkTriggerPubKey() []byte {
	return ccs.HardforkTriggerPubKeyField
}

// EnableEpochsHandler -
func (ccs *CoreComponentsStub) EnableEpochsHandler() common.EnableEpochsHandler {
	return ccs.EnableEpochsHandlerField
}

// IsInterfaceNil -
func (ccs *CoreComponentsStub) IsInterfaceNil() bool {
	return ccs == nil
}
