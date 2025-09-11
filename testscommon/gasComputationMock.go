package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// GasComputationMock -
type GasComputationMock struct {
	CheckIncomingMiniBlocksCalled func(
		miniBlocks []data.MiniBlockHeaderHandler,
		transactions map[string][]data.TransactionHandler,
	) (int, int, error)
	CheckOutgoingTransactionsCalled func(
		txHashes [][]byte,
		transactions []data.TransactionHandler,
	) ([][]byte, error)
	GetLastMiniBlockIndexIncludedCalled func() int
	GetBandwidthForTransactionsCalled   func() uint64
	TotalGasConsumedCalled              func() uint64
	DecreaseIncomingLimitCalled         func()
	DecreaseOutgoingLimitCalled         func()
	ResetIncomingLimitCalled            func()
	ResetOutgoingLimitCalled            func()
	ResetCalled                         func()
}

// CheckIncomingMiniBlocks -
func (mock *GasComputationMock) CheckIncomingMiniBlocks(
	miniBlocks []data.MiniBlockHeaderHandler,
	transactions map[string][]data.TransactionHandler,
) (int, int, error) {
	if mock.CheckIncomingMiniBlocksCalled != nil {
		return mock.CheckIncomingMiniBlocksCalled(miniBlocks, transactions)
	}
	return 0, 0, nil
}

// CheckOutgoingTransactions -
func (mock *GasComputationMock) CheckOutgoingTransactions(
	txHashes [][]byte,
	transactions []data.TransactionHandler,
) ([][]byte, error) {
	if mock.CheckOutgoingTransactionsCalled != nil {
		return mock.CheckOutgoingTransactionsCalled(txHashes, transactions)
	}
	return nil, nil
}

// GetLastMiniBlockIndexIncluded -
func (mock *GasComputationMock) GetLastMiniBlockIndexIncluded() int {
	if mock.GetLastMiniBlockIndexIncludedCalled != nil {
		return mock.GetLastMiniBlockIndexIncludedCalled()
	}
	return 0
}

// GetBandwidthForTransactions -
func (mock *GasComputationMock) GetBandwidthForTransactions() uint64 {
	if mock.GetBandwidthForTransactionsCalled != nil {
		return mock.GetBandwidthForTransactionsCalled()
	}
	return 0
}

// TotalGasConsumed -
func (mock *GasComputationMock) TotalGasConsumed() uint64 {
	if mock.TotalGasConsumedCalled != nil {
		return mock.TotalGasConsumedCalled()
	}
	return 0
}

// DecreaseIncomingLimit -
func (mock *GasComputationMock) DecreaseIncomingLimit() {
	if mock.DecreaseIncomingLimitCalled != nil {
		mock.DecreaseIncomingLimitCalled()
	}
}

// DecreaseOutgoingLimit -
func (mock *GasComputationMock) DecreaseOutgoingLimit() {
	if mock.DecreaseOutgoingLimitCalled != nil {
		mock.DecreaseOutgoingLimitCalled()
	}
}

// ResetIncomingLimit -
func (mock *GasComputationMock) ResetIncomingLimit() {
	if mock.ResetIncomingLimitCalled != nil {
		mock.ResetIncomingLimitCalled()
	}
}

// ResetOutgoingLimit -
func (mock *GasComputationMock) ResetOutgoingLimit() {
	if mock.ResetOutgoingLimitCalled != nil {
		mock.ResetOutgoingLimitCalled()
	}
}

// Reset -
func (mock *GasComputationMock) Reset() {
	if mock.ResetCalled != nil {
		mock.ResetCalled()
	}
}

// IsInterfaceNil -
func (mock *GasComputationMock) IsInterfaceNil() bool {
	return mock == nil
}
