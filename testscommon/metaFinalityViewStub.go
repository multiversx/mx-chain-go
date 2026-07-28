package testscommon

import "github.com/multiversx/mx-chain-core-go/data"

// MetaFinalityViewStub -
type MetaFinalityViewStub struct {
	IsMetaHeaderHeldFinalCalled          func(header data.HeaderHandler, headerHash []byte) bool
	IsIncludedInHeldFinalMetaBlockCalled func(shardID uint32, headerHash []byte, nonce uint64, lowMetaNonceAnchor uint64) bool
	IsDeadMetaBlockCalled                func(headerHash []byte, nonce uint64) bool
}

// IsMetaHeaderHeldFinal -
func (stub *MetaFinalityViewStub) IsMetaHeaderHeldFinal(header data.HeaderHandler, headerHash []byte) bool {
	if stub.IsMetaHeaderHeldFinalCalled != nil {
		return stub.IsMetaHeaderHeldFinalCalled(header, headerHash)
	}

	return false
}

// IsIncludedInHeldFinalMetaBlock -
func (stub *MetaFinalityViewStub) IsIncludedInHeldFinalMetaBlock(shardID uint32, headerHash []byte, nonce uint64, lowMetaNonceAnchor uint64) bool {
	if stub.IsIncludedInHeldFinalMetaBlockCalled != nil {
		return stub.IsIncludedInHeldFinalMetaBlockCalled(shardID, headerHash, nonce, lowMetaNonceAnchor)
	}

	return false
}

// IsDeadMetaBlock -
func (stub *MetaFinalityViewStub) IsDeadMetaBlock(headerHash []byte, nonce uint64) bool {
	if stub.IsDeadMetaBlockCalled != nil {
		return stub.IsDeadMetaBlockCalled(headerHash, nonce)
	}

	return false
}

// IsInterfaceNil -
func (stub *MetaFinalityViewStub) IsInterfaceNil() bool {
	return stub == nil
}
