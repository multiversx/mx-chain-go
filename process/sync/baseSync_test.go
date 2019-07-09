package sync_test

//
//import (
//	"github.com/ElrondNetwork/elrond-go/data"
//	"github.com/ElrondNetwork/elrond-go/dataRetriever"
//)
//
//type StorageBootstrapperMock struct{
//	GetHeaderCalled func(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error)
//	GetBlockBodyCalled func(data.HeaderHandler) (data.BodyHandler, error)
//	RemoveBlockBodyCalled func(nonce uint64, blockUnit dataRetriever.UnitType, hdrNonceHashDataUnit dataRetriever.UnitType) error
//	GetNonceWithLastNotarizedCalled func(currentNonce uint64) (startNonce uint64, finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64)
//	ApplyNotarizedBlocksCalled func(finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64) error
//	CleanupNotarizedStorageCalled func(lastNotarized map[uint32]uint64)
//}
//
//func (sbm *StorageBootstrapperMock) getHeader(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
//	return sbm.GetHeaderCalled(shardId, nonce)
//}
//
//func (sbm *StorageBootstrapperMock) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
//	return sbm.GetBlockBodyCalled(headerHandler)
//}
//
//func (sbm *StorageBootstrapperMock) removeBlockBody(
//	nonce uint64,
//	blockUnit dataRetriever.UnitType,
//	hdrNonceHashDataUnit dataRetriever.UnitType,
//) error {
//
//	return sbm.RemoveBlockBodyCalled(nonce, blockUnit, hdrNonceHashDataUnit)
//}
//
//func (sbm *StorageBootstrapperMock) getNonceWithLastNotarized(
//	currentNonce uint64,
//) (startNonce uint64, finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64) {
//
//	return sbm.GetNonceWithLastNotarizedCalled(currentNonce)
//}
//
//func (sbm *StorageBootstrapperMock) applyNotarizedBlocks(
//	finalNotarized map[uint32]uint64,
//	lastNotarized map[uint32]uint64,
//) error {
//
//	return sbm.ApplyNotarizedBlocksCalled(finalNotarized, lastNotarized)
//}
//
//func (sbm *StorageBootstrapperMock) cleanupNotarizedStorage(lastNotarized map[uint32]uint64) {
//	sbm.CleanupNotarizedStorageCalled(lastNotarized)
//}
//
