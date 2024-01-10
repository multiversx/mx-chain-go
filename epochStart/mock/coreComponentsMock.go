package mock

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
)

// CoreComponentsMock -
type CoreComponentsMock struct {
	IntMarsh                     marshal.Marshalizer
	Marsh                        marshal.Marshalizer
	Hash                         hashing.Hasher
	EpochNotifierField           process.EpochNotifier
	EnableEpochsHandlerField     common.EnableEpochsHandler
	TxSignHasherField            hashing.Hasher
	UInt64ByteSliceConv          typeConverters.Uint64ByteSliceConverter
	AddrPubKeyConv               core.PubkeyConverter
	ValPubKeyConv                core.PubkeyConverter
	PathHdl                      storage.PathManagerHandler
	ChainIdCalled                func() string
	MinTransactionVersionCalled  func() uint32
	GenesisNodesSetupCalled      func() sharding.GenesisNodesSetupHandler
	TxVersionCheckField          process.TxVersionCheckerHandler
	ChanStopNode                 chan endProcess.ArgEndProcess
	NodeTypeProviderField        core.NodeTypeProviderHandler
	ProcessStatusHandlerInstance common.ProcessStatusHandler
	HardforkTriggerPubKeyField   []byte
	PersisterFactoryField        storage.PersisterFactoryHandler
	mutCore                      sync.RWMutex
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

// EnableEpochsHandler -
func (ccm *CoreComponentsMock) EnableEpochsHandler() common.EnableEpochsHandler {
	return ccm.EnableEpochsHandlerField
}

// GenesisNodesSetup -
func (ccm *CoreComponentsMock) GenesisNodesSetup() sharding.GenesisNodesSetupHandler {
	if ccm.GenesisNodesSetupCalled != nil {
		return ccm.GenesisNodesSetupCalled()
	}
	return nil
}

// ProcessStatusHandler -
func (ccm *CoreComponentsMock) ProcessStatusHandler() common.ProcessStatusHandler {
	return ccm.ProcessStatusHandlerInstance
}

// HardforkTriggerPubKey -
func (ccm *CoreComponentsMock) HardforkTriggerPubKey() []byte {
	return ccm.HardforkTriggerPubKeyField
}

// PersisterFactory -
func (ccm *CoreComponentsMock) PersisterFactory() storage.PersisterFactoryHandler {
	return ccm.PersisterFactoryField
}

// IsInterfaceNil -
func (ccm *CoreComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
