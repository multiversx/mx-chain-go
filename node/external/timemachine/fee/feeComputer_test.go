package fee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewFeeComputer(t *testing.T) {
	t.Run("NilBuiltInFunctionsCostHandler", func(t *testing.T) {
		arguments := ArgsNewFeeComputer{
			BuiltInFunctionsCostHandler: nil,
			EconomicsConfig:             testscommon.GetEconomicsConfig(),
		}

		_, err := NewFeeComputer(arguments)
		require.Equal(t, process.ErrNilBuiltInFunctionsCostHandler, err)
	})

	t.Run("NilEconomicsConfig", func(t *testing.T) {
		arguments := ArgsNewFeeComputer{
			BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
			EconomicsConfig:             nil,
		}

		_, err := NewFeeComputer(arguments)
		require.Equal(t, ErrNilEconomicsConfig, err)
	})
}

func TestFeeComputer_ComputeTransactionFeeShouldWorkForDifferentEpochs(t *testing.T) {
	arguments := ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler:    &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:                testscommon.GetEconomicsConfig(),
		PenalizedTooMuchGasEnableEpoch: 124,
		GasPriceModifierEnableEpoch:    180,
	}

	contract, _ := hex.DecodeString("000000000000000000010000000000000000000000000000000000000000abba")

	computer, err := NewFeeComputer(arguments)
	require.Nil(t, err)
	require.NotNil(t, computer)

	checkComputedFee(t, "50000000000000", computer, 0, 80000, 1000000000, "", nil)
	checkComputedFee(t, "57500000000000", computer, 0, 80000, 1000000000, "hello", nil)
	checkComputedFee(t, "57500000000000", computer, 0, 1000000, 1000000000, "hello", contract)
	// (for review) >>> the following scenarios pass, as well. Is this all right?
	checkComputedFee(t, "80000000000000", computer, 124, 80000, 1000000000, "", nil)
	checkComputedFee(t, "1000000000000000", computer, 124, 1000000, 1000000000, "hello", contract)
	checkComputedFee(t, "404000000000000", computer, 124, 404000, 1000000000, "ESDTTransfer@464f4f2d653962643235@0a", contract)
	// <<< (for review)
	checkComputedFee(t, "57500010000000", computer, 180, 57501, 1000000000, "hello", contract)
	checkComputedFee(t, "66925000000000", computer, 180, 1000000, 1000000000, "hello", contract)
	checkComputedFee(t, "66925000000000", computer, 180, 1000000, 1000000000, "hello", contract)
	checkComputedFee(t, "107000000000000", computer, 180, 404000, 1000000000, "ESDTTransfer@464f4f2d653962643235@0a", contract)
}

func checkComputedFee(t *testing.T, expectedFee string, computer *feeComputer, epoch int, gasLimit uint64, gasPrice uint64, data string, receiver []byte) {
	tx := &dummyTx{
		gasLimit: gasLimit,
		gasPrice: gasPrice,
		data:     []byte(data),
		receiver: receiver,
	}

	fee, err := computer.ComputeTransactionFee(tx, epoch)
	require.Nil(t, err)
	require.Equal(t, expectedFee, fee.String())
}

type dummyTx struct {
	gasLimit uint64
	gasPrice uint64
	data     []byte
	receiver []byte
}

func (tx *dummyTx) GetGasLimit() uint64 {
	return tx.gasLimit
}

func (tx *dummyTx) GetGasPrice() uint64 {
	return tx.gasPrice
}

func (tx *dummyTx) GetData() []byte {
	return tx.data
}

func (tx *dummyTx) GetRcvAddr() []byte {
	return tx.receiver
}

func (tx *dummyTx) GetValue() *big.Int {
	return big.NewInt(0)
}
