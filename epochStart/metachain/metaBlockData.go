package metachain

import "math/big"

type metaBlockHandler interface {
	GetEpoch() uint32
	GetRound() uint64
	GetDevFeesInEpoch() *big.Int
	GetAccumulatedFeesInEpoch() *big.Int
}

type metaBlockData struct {
	epoch                  uint32
	round                  uint64
	accumulatedFeesInEpoch *big.Int
	devFeesInEpoch         *big.Int
}

// GetEpoch returns epoch field
func (m *metaBlockData) GetEpoch() uint32 {
	return m.epoch
}

// GetRound returns round field
func (m *metaBlockData) GetRound() uint64 {
	return m.round
}

// GetDevFeesInEpoch returns dev fees in epoch field
func (m *metaBlockData) GetDevFeesInEpoch() *big.Int {
	return m.devFeesInEpoch
}

// GetAccumulatedFeesInEpoch returns accumulated fees in epoch field
func (m *metaBlockData) GetAccumulatedFeesInEpoch() *big.Int {
	return m.accumulatedFeesInEpoch
}
