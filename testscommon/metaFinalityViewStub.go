package testscommon

import "github.com/multiversx/mx-chain-core-go/data"

// MetaFinalityViewStub -
type MetaFinalityViewStub struct {
	IsMetaHeaderHeldFinalCalled          func(header data.HeaderHandler, headerHash []byte) bool
	IsMetaHeaderSettlementReadyCalled    func(header data.HeaderHandler, headerHash []byte) bool
	IsShardHeaderIncludedCalled          func(metaHeader data.MetaHeaderHandler, shardID uint32, headerHash []byte, nonce uint64) bool
	IsIncludedInHeldFinalMetaBlockCalled func(shardID uint32, headerHash []byte, nonce uint64, ascendingFrom uint64, ascendingTo uint64) bool
	IsDeadMetaBlockCalled                func(headerHash []byte, nonce uint64) bool
}

// IsShardHeaderIncluded -
func (stub *MetaFinalityViewStub) IsShardHeaderIncluded(metaHeader data.MetaHeaderHandler, shardID uint32, headerHash []byte, nonce uint64) bool {
	if stub.IsShardHeaderIncludedCalled != nil {
		return stub.IsShardHeaderIncludedCalled(metaHeader, shardID, headerHash, nonce)
	}

	return false
}

// IsMetaHeaderSettlementReady -
func (stub *MetaFinalityViewStub) IsMetaHeaderSettlementReady(header data.HeaderHandler, headerHash []byte) bool {
	if stub.IsMetaHeaderSettlementReadyCalled != nil {
		return stub.IsMetaHeaderSettlementReadyCalled(header, headerHash)
	}

	return false
}

// IsMetaHeaderHeldFinal -
func (stub *MetaFinalityViewStub) IsMetaHeaderHeldFinal(header data.HeaderHandler, headerHash []byte) bool {
	if stub.IsMetaHeaderHeldFinalCalled != nil {
		return stub.IsMetaHeaderHeldFinalCalled(header, headerHash)
	}

	return false
}

// IsIncludedInHeldFinalMetaBlock -
func (stub *MetaFinalityViewStub) IsIncludedInHeldFinalMetaBlock(shardID uint32, headerHash []byte, nonce uint64, ascendingFrom uint64, ascendingTo uint64) bool {
	if stub.IsIncludedInHeldFinalMetaBlockCalled != nil {
		return stub.IsIncludedInHeldFinalMetaBlockCalled(shardID, headerHash, nonce, ascendingFrom, ascendingTo)
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
