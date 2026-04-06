package txcachemocks

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
)

// MempoolHostMock -
type MempoolHostMock struct {
	minGasLimit             uint64
	minGasPrice             uint64
	gasPerDataByte          uint64
	gasPriceModifier        float64
	extraGasLimitForGuarded uint64
	extraGasLimitForRelayed uint64

	ComputeTxFeeCalled        func(tx data.TransactionWithFeeHandler) *big.Int
	GetTransferredValueCalled func(tx data.TransactionHandler) *big.Int
}

// NewMempoolHostMock -
func NewMempoolHostMock() *MempoolHostMock {
	return &MempoolHostMock{
		minGasLimit:             50_000,
		minGasPrice:             1_000_000_000,
		gasPerDataByte:          1500,
		gasPriceModifier:        0.01,
		extraGasLimitForGuarded: 50_000,
		extraGasLimitForRelayed: 50_000,
	}
}

// ComputeTxFee -
func (mock *MempoolHostMock) ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int {
	if mock.ComputeTxFeeCalled != nil {
		return mock.ComputeTxFeeCalled(tx)
	}

	dataLength := uint64(len(tx.GetData()))
	gasPriceForMovement := tx.GetGasPrice()
	gasPriceForProcessing := uint64(float64(gasPriceForMovement) * mock.gasPriceModifier)

	gasLimitForMovement := mock.minGasLimit + dataLength*mock.gasPerDataByte

	if txAsGuarded, ok := tx.(data.GuardedTransactionHandler); ok && len(txAsGuarded.GetGuardianAddr()) > 0 {
		gasLimitForMovement += mock.extraGasLimitForGuarded
	}

	if txAsRelayed, ok := tx.(data.RelayedTransactionHandler); ok && len(txAsRelayed.GetRelayerAddr()) > 0 {
		gasLimitForMovement += mock.extraGasLimitForRelayed
	}

	if tx.GetGasLimit() < gasLimitForMovement {
		panic("tx.GetGasLimit() < gasLimitForMovement")
	}

	gasLimitForProcessing := tx.GetGasLimit() - gasLimitForMovement
	feeForMovement := core.SafeMul(gasPriceForMovement, gasLimitForMovement)
	feeForProcessing := core.SafeMul(gasPriceForProcessing, gasLimitForProcessing)
	fee := big.NewInt(0).Add(feeForMovement, feeForProcessing)
	return fee
}

// GetTransferredValue -
func (mock *MempoolHostMock) GetTransferredValue(tx data.TransactionHandler) *big.Int {
	if mock.GetTransferredValueCalled != nil {
		return mock.GetTransferredValueCalled(tx)
	}

	return tx.GetValue()
}

// IsInterfaceNil -
func (mock *MempoolHostMock) IsInterfaceNil() bool {
	return mock == nil
}
