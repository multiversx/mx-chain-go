package mock

import (
	"time"

	nodeFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// CoreComponentsStub -
type CoreComponentsStub struct {
	InternalMarshalizerField           marshal.Marshalizer
	TxMarshalizerField                 marshal.Marshalizer
	VmMarshalizerField                 marshal.Marshalizer
	HasherField                        hashing.Hasher
	Uint64ByteSliceConverterField      typeConverters.Uint64ByteSliceConverter
	AddressPubKeyConverterField        core.PubkeyConverter
	ValidatorPubKeyConverterField      core.PubkeyConverter
	PathHandlerField                   storage.PathManagerHandler
	ChainIdCalled                      func() string
	MinTransactionVersionCalled        func() uint32
	StatusHandlerUtilsField            nodeFactory.StatusHandlersUtils
	StatusHandlerField                 core.AppStatusHandler
	WatchdogField                      core.WatchdogTimer
	AlarmSchedulerField                core.TimersScheduler
	SyncTimerField                     ntp.SyncTimer
	RounderField                       consensus.Rounder
	EconomicsDataField                 process.EconomicsHandler
	RatingsDataField                   process.RatingsInfoHandler
	RaterField                         sharding.PeerAccountListAndRatingHandler
	GenesisNodesSetupField             sharding.GenesisNodesSetupHandler
	NodesShufflerField                 sharding.NodesShuffler
	EpochNotifierField                 factory.EpochNotifier
	EpochStartNotifierWithConfirmField factory.EpochStartNotifierWithConfirm
	ChanStopNodeProcessField           chan endProcess.ArgEndProcess
	GenesisTimeField                   time.Time
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

// StatusHandlerUtils -
func (ccs *CoreComponentsStub) StatusHandlerUtils() nodeFactory.StatusHandlersUtils {
	return ccs.StatusHandlerUtilsField
}

// StatusHandler -
func (ccs *CoreComponentsStub) StatusHandler() core.AppStatusHandler {
	return ccs.StatusHandlerField
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

// Rounder -
func (ccs *CoreComponentsStub) Rounder() consensus.Rounder {
	return ccs.RounderField
}

// EconomicsData -
func (ccs *CoreComponentsStub) EconomicsData() process.EconomicsHandler {
	return ccs.EconomicsDataField
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
func (ccs *CoreComponentsStub) NodesShuffler() sharding.NodesShuffler {
	return ccs.NodesShufflerField
}

// EpochNotifier -
func (ccs *CoreComponentsStub) EpochNotifier() factory.EpochNotifier {
	return ccs.EpochNotifierField
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

// ChanStopNodeProcess -
func (ccs *CoreComponentsStub) ChanStopNodeProcess() chan endProcess.ArgEndProcess {
	return ccs.ChanStopNodeProcessField
}

// IsInterfaceNil -
func (ccs *CoreComponentsStub) IsInterfaceNil() bool {
	return ccs == nil
}
