package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// HeadersCacherStub -
type HeadersCacherStub struct {
	AddCalled                           func(headerHash []byte, header data.HeaderHandler)
	RemoveHeaderByHashCalled            func(headerHash []byte)
	RemoveHeaderByNonceAndShardIdCalled func(hdrNonce uint64, shardId uint32)
	GetHeaderByNonceAndShardIdCalled    func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error)
	GetHeaderByHashCalled               func(hash []byte) (data.HeaderHandler, error)
	ClearCalled                         func()
	RegisterHandlerCalled               func(handler func(header data.HeaderHandler, shardHeaderHash []byte))
	NoncesCalled                        func(shardId uint32) []uint64
	LenCalled                           func() int
	MaxSizeCalled                       func() int
	GetNumHeadersCalled                 func(shardId uint32) int
}

// AddHeader -
func (hcs *HeadersCacherStub) AddHeader(headerHash []byte, header data.HeaderHandler) {
	if hcs.AddCalled != nil {
		hcs.AddCalled(headerHash, header)
	}
}

// RemoveHeaderByHash -
func (hcs *HeadersCacherStub) RemoveHeaderByHash(headerHash []byte) {
	if hcs.RemoveHeaderByHashCalled != nil {
		hcs.RemoveHeaderByHashCalled(headerHash)
	}
}

// RemoveHeaderByNonceAndShardId -
func (hcs *HeadersCacherStub) RemoveHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) {
	if hcs.RemoveHeaderByNonceAndShardIdCalled != nil {
		hcs.RemoveHeaderByNonceAndShardIdCalled(hdrNonce, shardId)
	}
}

// GetHeadersByNonceAndShardId -
func (hcs *HeadersCacherStub) GetHeadersByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
	if hcs.GetHeaderByNonceAndShardIdCalled != nil {
		return hcs.GetHeaderByNonceAndShardIdCalled(hdrNonce, shardId)
	}
	return nil, nil, nil
}

// GetHeaderByHash -
func (hcs *HeadersCacherStub) GetHeaderByHash(hash []byte) (data.HeaderHandler, error) {
	if hcs.GetHeaderByHashCalled != nil {
		return hcs.GetHeaderByHashCalled(hash)
	}
	return nil, nil
}

// Clear -
func (hcs *HeadersCacherStub) Clear() {
	if hcs.ClearCalled != nil {
		hcs.ClearCalled()
	}
}

// RegisterHandler -
func (hcs *HeadersCacherStub) RegisterHandler(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {
	if hcs.RegisterHandlerCalled != nil {
		hcs.RegisterHandlerCalled(handler)
	}
}

// Nonces -
func (hcs *HeadersCacherStub) Nonces(shardId uint32) []uint64 {
	if hcs.NoncesCalled != nil {
		return hcs.NoncesCalled(shardId)
	}
	return nil
}

// Len -
func (hcs *HeadersCacherStub) Len() int {
	return 0
}

// MaxSize -
func (hcs *HeadersCacherStub) MaxSize() int {
	return 1000
}

// IsInterfaceNil -
func (hcs *HeadersCacherStub) IsInterfaceNil() bool {
	return hcs == nil
}

// GetNumHeaders -
func (hcs *HeadersCacherStub) GetNumHeaders(shardId uint32) int {
	if hcs.GetNumHeadersCalled != nil {
		return hcs.GetNumHeadersCalled(shardId)
	}
	return 0
}
