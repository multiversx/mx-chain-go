package factory

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

// CoreComponentsMock -
type CoreComponentsMock struct {
	IntMarsh                    marshal.Marshalizer
	TxMarsh                     marshal.Marshalizer
	VmMarsh                     marshal.Marshalizer
	Hash                        hashing.Hasher
	TxSignHasherField           hashing.Hasher
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
	RoundHandlerField           consensus.RoundHandler
	EconomicsHandler            process.EconomicsDataHandler
	RatingsConfig               process.RatingsInfoHandler
	RatingHandler               sharding.PeerAccountListAndRatingHandler
	NodesConfig                 sharding.GenesisNodesSetupHandler
	EpochChangeNotifier         process.EpochNotifier
	EpochNotifierWithConfirm    factory.EpochStartNotifierWithConfirm
	ChanStopProcess             chan endProcess.ArgEndProcess
	Shuffler                    sharding.NodesShuffler
	TxVersionCheckHandler       process.TxVersionCheckerHandler
	StartTime                   time.Time
}

// Create -
func (ccm *CoreComponentsMock) Create() error {
	return nil
}

// Close -
func (ccm *CoreComponentsMock) Close() error {
	return nil
}

// CheckSubcomponents -
func (ccm *CoreComponentsMock) CheckSubcomponents() error {
	return nil
}

// VmMarshalizer -
func (ccm *CoreComponentsMock) VmMarshalizer() marshal.Marshalizer {
	return ccm.VmMarsh
}

// StatusHandlerUtils -
func (ccm *CoreComponentsMock) StatusHandlerUtils() nodeFactory.StatusHandlersUtils {
	return ccm.StatusHdlUtils
}

// StatusHandler -
func (ccm *CoreComponentsMock) StatusHandler() core.AppStatusHandler {
	return ccm.AppStatusHdl
}

// Watchdog -
func (ccm *CoreComponentsMock) Watchdog() core.WatchdogTimer {
	return ccm.WDTimer
}

// AlarmScheduler -
func (ccm *CoreComponentsMock) AlarmScheduler() core.TimersScheduler {
	return ccm.Alarm
}

// SyncTimer -
func (ccm *CoreComponentsMock) SyncTimer() ntp.SyncTimer {
	return ccm.NtpTimer
}

// RoundHandler -
func (ccm *CoreComponentsMock) RoundHandler() consensus.RoundHandler {
	return ccm.RoundHandlerField
}

// EconomicsData -
func (ccm *CoreComponentsMock) EconomicsData() process.EconomicsDataHandler {
	return ccm.EconomicsHandler
}

// RatingsData -
func (ccm *CoreComponentsMock) RatingsData() process.RatingsInfoHandler {
	return ccm.RatingsConfig
}

// Rater -
func (ccm *CoreComponentsMock) Rater() sharding.PeerAccountListAndRatingHandler {
	return ccm.RatingHandler

}

// GenesisNodesSetup -
func (ccm *CoreComponentsMock) GenesisNodesSetup() sharding.GenesisNodesSetupHandler {
	return ccm.NodesConfig
}

// NodesShuffler -
func (ccm *CoreComponentsMock) NodesShuffler() sharding.NodesShuffler {
	return ccm.Shuffler
}

// EpochNotifier -
func (ccm *CoreComponentsMock) EpochNotifier() process.EpochNotifier {
	return ccm.EpochChangeNotifier
}

// EpochStartNotifierWithConfirm -
func (ccm *CoreComponentsMock) EpochStartNotifierWithConfirm() factory.EpochStartNotifierWithConfirm {
	return ccm.EpochNotifierWithConfirm
}

// GenesisTime -
func (ccm *CoreComponentsMock) GenesisTime() time.Time {
	return ccm.StartTime
}

// InternalMarshalizer -
func (ccm *CoreComponentsMock) InternalMarshalizer() marshal.Marshalizer {
	return ccm.IntMarsh
}

// SetInternalMarshalizer -
func (ccm *CoreComponentsMock) SetInternalMarshalizer(m marshal.Marshalizer) error {
	ccm.IntMarsh = m
	return nil
}

// TxMarshalizer -
func (ccm *CoreComponentsMock) TxMarshalizer() marshal.Marshalizer {
	return ccm.TxMarsh
}

// Hasher -
func (ccm *CoreComponentsMock) Hasher() hashing.Hasher {
	return ccm.Hash
}

// TxSignHasher -
func (ccm *CoreComponentsMock) TxSignHasher() hashing.Hasher {
	return ccm.TxSignHasherField
}

// Uint64ByteSliceConverter -
func (ccm *CoreComponentsMock) Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter {
	return ccm.UInt64ByteSliceConv
}

// AddressPubKeyConverter -
func (ccm *CoreComponentsMock) AddressPubKeyConverter() core.PubkeyConverter {
	return ccm.AddrPubKeyConv
}

// ValidatorPubKeyConverter -
func (ccm *CoreComponentsMock) ValidatorPubKeyConverter() core.PubkeyConverter {
	return ccm.ValPubKeyConv
}

// PathHandler -
func (ccm *CoreComponentsMock) PathHandler() storage.PathManagerHandler {
	return ccm.PathHdl
}

// ChainID -
func (ccm *CoreComponentsMock) ChainID() string {
	if ccm.ChainIdCalled != nil {
		return ccm.ChainIdCalled()
	}
	return "undefined"
}

// MinTransactionVersion -
func (ccm *CoreComponentsMock) MinTransactionVersion() uint32 {
	if ccm.MinTransactionVersionCalled != nil {
		return ccm.MinTransactionVersionCalled()
	}
	return 1
}

// TxVersionChecker -
func (ccm *CoreComponentsMock) TxVersionChecker() process.TxVersionCheckerHandler {
	return ccm.TxVersionCheckHandler
}

// EncodedAddressLen -
func (ccm *CoreComponentsMock) EncodedAddressLen() uint32 {
	return uint32(ccm.AddressPubKeyConverter().Len() * 2)
}

// ChanStopNodeProcess -
func (ccm *CoreComponentsMock) ChanStopNodeProcess() chan endProcess.ArgEndProcess {
	return ccm.ChanStopProcess
}

// IsInterfaceNil -
func (ccm *CoreComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
