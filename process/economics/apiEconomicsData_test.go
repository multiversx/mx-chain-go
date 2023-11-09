package economics_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIEconomicsData(t *testing.T) {
	t.Parallel()

	t.Run("nil economics data", func(t *testing.T) {
		t.Parallel()

		apiEconomicsData, err := economics.NewAPIEconomicsData(nil)

		assert.True(t, check.IfNil(apiEconomicsData))
		assert.Equal(t, process.ErrNilEconomicsData, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgsForEconomicsData(1)
		economicsData, _ := economics.NewEconomicsData(args)
		apiEconomicsData, err := economics.NewAPIEconomicsData(economicsData)

		assert.False(t, check.IfNil(apiEconomicsData))
		assert.Nil(t, err)
	})
}

func TestApiEconomicsData_CheckValidityTxValues(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := uint64(42)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMiniBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaMiniBlock = fmt.Sprintf("%d", maxGasLimitPerBlock*10)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerTx = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	apiEconomicsData, _ := economics.NewAPIEconomicsData(economicsData)

	t.Run("maximum gas limit as defined should work", func(t *testing.T) {
		t.Parallel()

		tx := &transaction.Transaction{
			GasPrice: minGasPrice + 1,
			GasLimit: maxGasLimitPerBlock,
			Value:    big.NewInt(0),
		}
		err := apiEconomicsData.CheckValidityTxValues(tx)
		require.Nil(t, err)
	})
	t.Run("maximum gas limit + 1 as defined should error", func(t *testing.T) {
		t.Parallel()

		tx := &transaction.Transaction{
			GasPrice: minGasPrice + 1,
			GasLimit: maxGasLimitPerBlock + 1,
			Value:    big.NewInt(0),
		}
		err := apiEconomicsData.CheckValidityTxValues(tx)
		require.Equal(t, process.ErrMoreGasThanGasLimitPerMiniBlockForSafeCrossShard, err)
	})
	t.Run("maximum gas limit - 1 as defined should work", func(t *testing.T) {
		t.Parallel()

		tx := &transaction.Transaction{
			GasPrice: minGasPrice + 1,
			GasLimit: maxGasLimitPerBlock - 1,
			Value:    big.NewInt(0),
		}
		err := apiEconomicsData.CheckValidityTxValues(tx)
		require.Nil(t, err)
	})
}
