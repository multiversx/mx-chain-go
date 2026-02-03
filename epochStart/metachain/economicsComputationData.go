package metachain

import "math/big"

type economicsComputationDataHandler interface {
	GetEpoch() uint32
	GetRound() uint64
	GetTimeStamp() uint64
	GetDevFeesInEpoch() *big.Int
	GetAccumulatedFeesInEpoch() *big.Int
}

type economicsComputationData struct {
	newEpoch               uint32
	round                  uint64
	timeStamp              uint64
	accumulatedFeesInEpoch *big.Int
	devFeesInEpoch         *big.Int
}

// GetEpoch returns epoch field
func (m *economicsComputationData) GetEpoch() uint32 {
	return m.newEpoch
}

// GetRound returns round field
func (m *economicsComputationData) GetRound() uint64 {
	return m.round
}

// GetTimeStamp returns time stamp field
func (m *economicsComputationData) GetTimeStamp() uint64 {
	return m.timeStamp
}

// GetDevFeesInEpoch returns dev fees in epoch field
func (m *economicsComputationData) GetDevFeesInEpoch() *big.Int {
	return m.devFeesInEpoch
}

// GetAccumulatedFeesInEpoch returns accumulated fees in epoch field
func (m *economicsComputationData) GetAccumulatedFeesInEpoch() *big.Int {
	return m.accumulatedFeesInEpoch
}
