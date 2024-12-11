package pool

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/data"
)

// HeadersPoolStub -
type HeadersPoolStub struct {
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
func (hps *HeadersPoolStub) AddHeader(headerHash []byte, header data.HeaderHandler) {
	if hps.AddCalled != nil {
		hps.AddCalled(headerHash, header)
	}
}

// RemoveHeaderByHash -
func (hps *HeadersPoolStub) RemoveHeaderByHash(headerHash []byte) {
	if hps.RemoveHeaderByHashCalled != nil {
		hps.RemoveHeaderByHashCalled(headerHash)
	}
}

// RemoveHeaderByNonceAndShardId -
func (hps *HeadersPoolStub) RemoveHeaderByNonceAndShardId(hdrNonce uint64, shardId uint32) {
	if hps.RemoveHeaderByNonceAndShardIdCalled != nil {
		hps.RemoveHeaderByNonceAndShardIdCalled(hdrNonce, shardId)
	}
}

// GetHeadersByNonceAndShardId -
func (hps *HeadersPoolStub) GetHeadersByNonceAndShardId(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
	if hps.GetHeaderByNonceAndShardIdCalled != nil {
		return hps.GetHeaderByNonceAndShardIdCalled(hdrNonce, shardId)
	}
	return nil, nil, errors.New("err")
}

// GetHeaderByHash -
func (hps *HeadersPoolStub) GetHeaderByHash(hash []byte) (data.HeaderHandler, error) {
	if hps.GetHeaderByHashCalled != nil {
		return hps.GetHeaderByHashCalled(hash)
	}
	return nil, nil
}

// Clear -
func (hps *HeadersPoolStub) Clear() {
	if hps.ClearCalled != nil {
		hps.ClearCalled()
	}
}

// RegisterHandler -
func (hps *HeadersPoolStub) RegisterHandler(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {
	if hps.RegisterHandlerCalled != nil {
		hps.RegisterHandlerCalled(handler)
	}
}

// Nonces -
func (hps *HeadersPoolStub) Nonces(shardId uint32) []uint64 {
	if hps.NoncesCalled != nil {
		return hps.NoncesCalled(shardId)
	}
	return nil
}

// Len -
func (hps *HeadersPoolStub) Len() int {
	return 0
}

// MaxSize -
func (hps *HeadersPoolStub) MaxSize() int {
	return 100
}

// IsInterfaceNil -
func (hps *HeadersPoolStub) IsInterfaceNil() bool {
	return hps == nil
}

// GetNumHeaders -
func (hps *HeadersPoolStub) GetNumHeaders(shardId uint32) int {
	if hps.GetNumHeadersCalled != nil {
		return hps.GetNumHeadersCalled(shardId)
	}

	return 0
}

func (hps *HeadersPoolStub) AddHeaderInShard(_ []byte, _ data.HeaderHandler, _ uint32) {
}
