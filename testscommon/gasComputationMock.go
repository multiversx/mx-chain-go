package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// GasComputationMock -
type GasComputationMock struct {
	AddIncomingMiniBlocksCalled func(
		miniBlocks []data.MiniBlockHeaderHandler,
		transactions map[string][]data.TransactionHandler,
	) (int, int, error)
	AddOutgoingTransactionsCalled func(
		txHashes [][]byte,
		transactions []data.TransactionHandler,
	) ([][]byte, []data.MiniBlockHeaderHandler, error)
	GetBandwidthForTransactionsCalled func() uint64
	TotalGasConsumedCalled            func() uint64
	DecreaseIncomingLimitCalled       func()
	DecreaseOutgoingLimitCalled       func()
	ZeroIncomingLimitCalled           func()
	ZeroOutgoingLimitCalled           func()
	ResetIncomingLimitCalled          func()
	ResetOutgoingLimitCalled          func()
	ResetCalled                       func()
	RevertIncomingMiniBlocksCalled    func(miniBlockHashes [][]byte)
}

// AddIncomingMiniBlocks -
func (mock *GasComputationMock) AddIncomingMiniBlocks(
	miniBlocks []data.MiniBlockHeaderHandler,
	transactions map[string][]data.TransactionHandler,
) (int, int, error) {
	if mock.AddIncomingMiniBlocksCalled != nil {
		return mock.AddIncomingMiniBlocksCalled(miniBlocks, transactions)
	}
	return 0, 0, nil
}

// AddOutgoingTransactions -
func (mock *GasComputationMock) AddOutgoingTransactions(
	txHashes [][]byte,
	transactions []data.TransactionHandler,
) ([][]byte, []data.MiniBlockHeaderHandler, error) {
	if mock.AddOutgoingTransactionsCalled != nil {
		return mock.AddOutgoingTransactionsCalled(txHashes, transactions)
	}
	return nil, nil, nil
}

// RevertIncomingMiniBlocks -
func (mock *GasComputationMock) RevertIncomingMiniBlocks(miniBlockHashes [][]byte) {
	if mock.RevertIncomingMiniBlocksCalled != nil {
		mock.RevertIncomingMiniBlocksCalled(miniBlockHashes)
	}
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

// ZeroIncomingLimit -
func (mock *GasComputationMock) ZeroIncomingLimit() {
	if mock.ZeroIncomingLimitCalled != nil {
		mock.ZeroIncomingLimitCalled()
	}
}

// ZeroOutgoingLimit -
func (mock *GasComputationMock) ZeroOutgoingLimit() {
	if mock.ZeroOutgoingLimitCalled != nil {
		mock.ZeroOutgoingLimitCalled()
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
