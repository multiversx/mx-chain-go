package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

// CoreComponentsMock -
type CoreComponentsMock struct {
	IntMarsh                 marshal.Marshalizer
	TxMarsh                  marshal.Marshalizer
	Hash                     hashing.Hasher
	TxSignHasherField        hashing.Hasher
	UInt64ByteSliceConv      typeConverters.Uint64ByteSliceConverter
	AddrPubKeyConv           core.PubkeyConverter
	Chain                    string
	MinTxVersion             uint32
	StatHandler              core.AppStatusHandler
	EnableEpochsHandlerField common.EnableEpochsHandler
	TxVersionCheck           process.TxVersionCheckerHandler
}

// InternalMarshalizer -
func (ccm *CoreComponentsMock) InternalMarshalizer() marshal.Marshalizer {
	return ccm.IntMarsh
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

// ChainID -
func (ccm *CoreComponentsMock) ChainID() string {
	return ccm.Chain
}

// MinTransactionVersion -
func (ccm *CoreComponentsMock) MinTransactionVersion() uint32 {
	return ccm.MinTxVersion
}

// EnableEpochsHandler -
func (ccm *CoreComponentsMock) EnableEpochsHandler() common.EnableEpochsHandler {
	return ccm.EnableEpochsHandlerField
}

// TxVersionChecker -
func (ccm *CoreComponentsMock) TxVersionChecker() process.TxVersionCheckerHandler {
	return ccm.TxVersionCheck
}

// IsInterfaceNil -
func (ccm *CoreComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
