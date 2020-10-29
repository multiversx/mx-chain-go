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
	IntMarsh                    marshal.Marshalizer
	TxMarsh                     marshal.Marshalizer
	VmMarsh                     marshal.Marshalizer
	Hash                        hashing.Hasher
	UInt64ByteSliceConv         typeConverters.Uint64ByteSliceConverter
	AddrPubKeyConv              core.PubkeyConverter
	ValPubKeyConv               core.PubkeyConverter
	PathHdl                     storage.PathManagerHandler
	ChainIdCalled               func() string
	MinTransactionVersionCalled func() uint32
	StatusHdlUtils              nodeFactory.StatusHandlersUtils
	AppStatusHdl                core.AppStatusHandler
	WDTimer                     core.WatchdogTimer
	Alarm                       core.TimersScheduler
	NtpTimer                    ntp.SyncTimer
	RoundHandler                consensus.Rounder
	EconomicsHandler            process.EconomicsHandler
	RatingsConfig               process.RatingsInfoHandler
	RatingHandler               sharding.PeerAccountListAndRatingHandler
	NodesConfig                 sharding.GenesisNodesSetupHandler
	Shuffler                    sharding.NodesShuffler
	EpochChangeNotifier         factory.EpochNotifier
	EpochNotifierWithConfirm    factory.EpochStartNotifierWithConfirm
	ChanStopProcess             chan endProcess.ArgEndProcess
	StartTime                   time.Time
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
	return ccs.VmMarsh
}

// StatusHandlerUtils -
func (ccs *CoreComponentsStub) StatusHandlerUtils() nodeFactory.StatusHandlersUtils {
	return ccs.StatusHdlUtils
}

// StatusHandler -
func (ccs *CoreComponentsStub) StatusHandler() core.AppStatusHandler {
	return ccs.AppStatusHdl
}

// Watchdog -
func (ccs *CoreComponentsStub) Watchdog() core.WatchdogTimer {
	return ccs.WDTimer
}

// AlarmScheduler -
func (ccs *CoreComponentsStub) AlarmScheduler() core.TimersScheduler {
	return ccs.Alarm
}

// SyncTimer -
func (ccs *CoreComponentsStub) SyncTimer() ntp.SyncTimer {
	return ccs.NtpTimer
}

// Rounder -
func (ccs *CoreComponentsStub) Rounder() consensus.Rounder {
	return ccs.RoundHandler
}

// EconomicsData -
func (ccs *CoreComponentsStub) EconomicsData() process.EconomicsHandler {
	return ccs.EconomicsHandler
}

// RatingsData -
func (ccs *CoreComponentsStub) RatingsData() process.RatingsInfoHandler {
	return ccs.RatingsConfig
}

// Rater -
func (ccs *CoreComponentsStub) Rater() sharding.PeerAccountListAndRatingHandler {
	return ccs.RatingHandler

}

// GenesisNodesSetup -
func (ccs *CoreComponentsStub) GenesisNodesSetup() sharding.GenesisNodesSetupHandler {
	return ccs.NodesConfig
}

// NodesShuffler -
func (ccs *CoreComponentsStub) NodesShuffler() sharding.NodesShuffler {
	return ccs.Shuffler
}

// EpochNotifier -
func (ccm *CoreComponentsStub) EpochNotifier() factory.EpochNotifier {
	return ccm.EpochChangeNotifier
}

// EpochStartNotifierWithConfirm -
func (ccm *CoreComponentsStub) EpochStartNotifierWithConfirm() factory.EpochStartNotifierWithConfirm {
	return ccm.EpochNotifierWithConfirm
}

// GenesisTime -
func (ccs *CoreComponentsStub) GenesisTime() time.Time {
	return ccs.StartTime
}

// InternalMarshalizer -
func (ccs *CoreComponentsStub) InternalMarshalizer() marshal.Marshalizer {
	return ccs.IntMarsh
}

// SetInternalMarshalizer -
func (ccs *CoreComponentsStub) SetInternalMarshalizer(m marshal.Marshalizer) error {
	ccs.IntMarsh = m
	return nil
}

// TxMarshalizer -
func (ccs *CoreComponentsStub) TxMarshalizer() marshal.Marshalizer {
	return ccs.TxMarsh
}

// Hasher -
func (ccs *CoreComponentsStub) Hasher() hashing.Hasher {
	return ccs.Hash
}

// Uint64ByteSliceConverter -
func (ccs *CoreComponentsStub) Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter {
	return ccs.UInt64ByteSliceConv
}

// AddressPubKeyConverter -
func (ccs *CoreComponentsStub) AddressPubKeyConverter() core.PubkeyConverter {
	return ccs.AddrPubKeyConv
}

// ValidatorPubKeyConverter -
func (ccs *CoreComponentsStub) ValidatorPubKeyConverter() core.PubkeyConverter {
	return ccs.ValPubKeyConv
}

// PathHandler -
func (ccs *CoreComponentsStub) PathHandler() storage.PathManagerHandler {
	return ccs.PathHdl
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
	return ccs.ChanStopProcess
}

// IsInterfaceNil -
func (ccs *CoreComponentsStub) IsInterfaceNil() bool {
	return ccs == nil
}
