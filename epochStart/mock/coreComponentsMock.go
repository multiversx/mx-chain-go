package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// CoreComponentsMock -
type CoreComponentsMock struct {
	IntMarsh                    marshal.Marshalizer
	Marsh                       marshal.Marshalizer
	Hash                        hashing.Hasher
	EpochNotifierField          process.EpochNotifier
	TxSignHasherField           hashing.Hasher
	UInt64ByteSliceConv         typeConverters.Uint64ByteSliceConverter
	AddrPubKeyConv              core.PubkeyConverter
	ValPubKeyConv               core.PubkeyConverter
	PathHdl                     storage.PathManagerHandler
	ChainIdCalled               func() string
	MinTransactionVersionCalled func() uint32
	StatusHandlerCalled         func() core.AppStatusHandler
	GenesisNodesSetupCalled     func() sharding.GenesisNodesSetupHandler
	TxVersionCheckField         process.TxVersionCheckerHandler
	ChanStopNode                chan endProcess.ArgEndProcess
	NodeTypeProviderField       core.NodeTypeProviderHandler
	mutCore                     sync.RWMutex
}

// ChanStopNodeProcess -
func (ccm *CoreComponentsMock) ChanStopNodeProcess() chan endProcess.ArgEndProcess {
	ccm.mutCore.RLock()
	defer ccm.mutCore.RUnlock()

	if ccm.ChanStopNode != nil {
		return ccm.ChanStopNode
	}

	return endProcess.GetDummyEndProcessChannel()
}

// NodeTypeProvider -
func (ccm *CoreComponentsMock) NodeTypeProvider() core.NodeTypeProviderHandler {
	return ccm.NodeTypeProviderField
}

// InternalMarshalizer -
func (ccm *CoreComponentsMock) InternalMarshalizer() marshal.Marshalizer {
	ccm.mutCore.RLock()
	defer ccm.mutCore.RUnlock()

	return ccm.IntMarsh
}

// SetInternalMarshalizer -
func (ccm *CoreComponentsMock) SetInternalMarshalizer(m marshal.Marshalizer) error {
	ccm.mutCore.Lock()
	ccm.IntMarsh = m
	ccm.mutCore.Unlock()

	return nil
}

// TxMarshalizer -
func (ccm *CoreComponentsMock) TxMarshalizer() marshal.Marshalizer {
	return ccm.Marsh
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
	return ccm.TxVersionCheckField
}

// EpochNotifier -
func (ccm *CoreComponentsMock) EpochNotifier() process.EpochNotifier {
	return ccm.EpochNotifierField
}

// StatusHandler -
func (ccm *CoreComponentsMock) StatusHandler() core.AppStatusHandler {
	if ccm.StatusHandlerCalled != nil {
		return ccm.StatusHandlerCalled()
	}
	return nil
}

// GenesisNodesSetup -
func (ccm *CoreComponentsMock) GenesisNodesSetup() sharding.GenesisNodesSetupHandler {
	if ccm.GenesisNodesSetupCalled != nil {
		return ccm.GenesisNodesSetupCalled()
	}
	return nil
}

// IsInterfaceNil -
func (ccm *CoreComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
