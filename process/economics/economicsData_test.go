package economics_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDummyEconomicsConfig(feeSettings config.FeeSettings) *config.EconomicsConfig {
	return &config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			GenesisTotalSupply: "2000000000000000000000",
			MinimumInflation:   0,
			YearSettings: []*config.YearSetting{
				{
					Year:             0,
					MaximumInflation: 0.01,
				},
			},
		},
		RewardsSettings: config.RewardsSettings{
			LeaderPercentage:                 0.1,
			ProtocolSustainabilityPercentage: 0.1,
			ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			TopUpGradientPoint:               "300000000000000000000",
			TopUpFactor:                      0.25,
		},
		FeeSettings: feeSettings,
	}
}

func feeSettingsDummy(gasModifier float64) config.FeeSettings {
	return config.FeeSettings{
		MaxGasLimitPerBlock:     "100000",
		MaxGasLimitPerMetaBlock: "1000000",
		MinGasPrice:             "18446744073709551615",
		MinGasLimit:             "500",
		GasPerDataByte:          "1",
		GasPriceModifier:        gasModifier,
	}
}

func feeSettingsReal() config.FeeSettings {
	return config.FeeSettings{
		MaxGasLimitPerBlock:     "1500000000",
		MaxGasLimitPerMetaBlock: "15000000000",
		MinGasPrice:             "1000000000",
		MinGasLimit:             "50000",
		GasPerDataByte:          "1500",
		GasPriceModifier:        0.01,
	}
}

func createArgsForEconomicsData(gasModifier float64) economics.ArgsNewEconomicsData {
	feeSettings := feeSettingsDummy(gasModifier)
	args := economics.ArgsNewEconomicsData{
		Economics:                      createDummyEconomicsConfig(feeSettings),
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &mock.EpochNotifierStub{},
		BuiltInFunctionsCostHandler:    &mock.BuiltInCostHandlerStub{},
	}
	return args
}

func createArgsForEconomicsDataRealFees() economics.ArgsNewEconomicsData {
	feeSettings := feeSettingsReal()
	args := economics.ArgsNewEconomicsData{
		Economics:                      createDummyEconomicsConfig(feeSettings),
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &mock.EpochNotifierStub{},
		BuiltInFunctionsCostHandler:    &mock.BuiltInCostHandlerStub{},
	}
	return args
}

func TestNewEconomicsData_InvalidMaxGasLimitPerBlockShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	badGasLimitPerBlock := []string{
		"-1",
		"-100000000000000000000",
		"badValue",
		"",
		"#########",
		"11112S",
		"1111O0000",
		"10ERD",
		"10000000000000000000000000000000000000000000000000000000000000",
	}

	for _, gasLimitPerBlock := range badGasLimitPerBlock {
		args.Economics.FeeSettings.MaxGasLimitPerBlock = gasLimitPerBlock
		_, err := economics.NewEconomicsData(args)
		assert.Equal(t, process.ErrInvalidMaxGasLimitPerBlock, err)
	}

}

func TestNewEconomicsData_InvalidMinGasPriceShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	badGasPrice := []string{
		"-1",
		"-100000000000000000000",
		"badValue",
		"",
		"#########",
		"11112S",
		"1111O0000",
		"10ERD",
		"10000000000000000000000000000000000000000000000000000000000000",
	}

	for _, gasPrice := range badGasPrice {
		args.Economics.FeeSettings.MinGasPrice = gasPrice
		_, err := economics.NewEconomicsData(args)
		assert.Equal(t, process.ErrInvalidMinimumGasPrice, err)
	}

}

func TestNewEconomicsData_InvalidMinGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	bagMinGasLimit := []string{
		"-1",
		"-100000000000000000000",
		"badValue",
		"",
		"#########",
		"11112S",
		"1111O0000",
		"10ERD",
		"10000000000000000000000000000000000000000000000000000000000000",
	}

	for _, minGasLimit := range bagMinGasLimit {
		args.Economics.FeeSettings.MinGasLimit = minGasLimit
		_, err := economics.NewEconomicsData(args)
		assert.Equal(t, process.ErrInvalidMinimumGasLimitForTx, err)
	}

}

func TestNewEconomicsData_InvalidLeaderPercentageShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.RewardsSettings.LeaderPercentage = -0.1

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrInvalidRewardsPercentages, err)

}

func TestNewEconomicsData_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.EpochNotifier = nil

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrNilEpochNotifier, err)

}

func TestNewEconomicsData_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	economicsData, _ := economics.NewEconomicsData(args)
	assert.NotNil(t, economicsData)
}

func TestEconomicsData_LeaderPercentage(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	leaderPercentage := 0.40
	args.Economics.RewardsSettings.LeaderPercentage = leaderPercentage
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.LeaderPercentage()
	assert.Equal(t, leaderPercentage, value)
}

func TestEconomicsData_ComputeMoveBalanceFeeNoTxData(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	gasPrice := uint64(500)
	minGasLimit := uint64(12)
	args.Economics.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: minGasLimit,
		Value:    big.NewInt(0),
	}

	cost := economicsData.ComputeMoveBalanceFee(tx)

	expectedCost := big.NewInt(0).SetUint64(gasPrice)
	expectedCost.Mul(expectedCost, big.NewInt(0).SetUint64(minGasLimit))
	assert.Equal(t, expectedCost, cost)
}

func TestEconomicsData_ComputeMoveBalanceFeeWithTxData(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	gasPrice := uint64(500)
	minGasLimit := uint64(12)
	txData := "text to be notarized"
	args.Economics.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: minGasLimit,
		Data:     []byte(txData),
		Value:    big.NewInt(0),
	}

	cost := economicsData.ComputeMoveBalanceFee(tx)

	expectedGasLimit := big.NewInt(0).SetUint64(minGasLimit)
	expectedGasLimit.Add(expectedGasLimit, big.NewInt(int64(len(txData))))
	expectedCost := big.NewInt(0).SetUint64(gasPrice)
	expectedCost.Mul(expectedCost, expectedGasLimit)
	assert.Equal(t, expectedCost, cost)
}

func TestEconomicsData_ComputeTxFeeShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	gasPrice := uint64(500)
	gasLimit := uint64(20)
	minGasLimit := uint64(10)
	args.Economics.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	args.Economics.FeeSettings.GasPriceModifier = 0.01
	args.PenalizedTooMuchGasEnableEpoch = 1
	args.GasPriceModifierEnableEpoch = 2
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	cost := economicsData.ComputeTxFee(tx)
	expectedCost := core.SafeMul(minGasLimit, gasPrice)
	assert.Equal(t, expectedCost, cost)

	economicsData.EpochConfirmed(1)

	cost = economicsData.ComputeTxFee(tx)
	expectedCost = core.SafeMul(gasLimit, gasPrice)
	assert.Equal(t, expectedCost, cost)

	economicsData.EpochConfirmed(2)
	cost = economicsData.ComputeTxFee(tx)
	assert.Equal(t, big.NewInt(5050), cost)
}

func TestEconomicsData_TxWithLowerGasPriceShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice - 1,
		GasLimit: minGasLimit,
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Equal(t, process.ErrInsufficientGasPriceInTx, err)
}

func TestEconomicsData_TxWithLowerGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit - 1,
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Equal(t, process.ErrInsufficientGasLimitInTx, err)
}

func TestEconomicsData_TxWithHigherGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit + 1,
		Data:     []byte("1"),
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Equal(t, process.ErrMoreGasThanGasLimitPerBlock, err)
}

func TestEconomicsData_TxWithWithMinGasPriceAndLimitShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 1
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit,
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Nil(t, err)
}

func TestEconomicsData_TxWithWithMoreGasLimitThanMaximumPerBlockShouldNotWork(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := uint64(42)
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)

	tx := &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: maxGasLimitPerBlock,
		Value:    big.NewInt(0),
	}
	err := economicsData.CheckValidityTxValues(tx)
	require.Equal(t, process.ErrMoreGasThanGasLimitPerBlock, err)

	tx = &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: maxGasLimitPerBlock + 1,
		Value:    big.NewInt(0),
	}
	err = economicsData.CheckValidityTxValues(tx)
	require.Equal(t, process.ErrMoreGasThanGasLimitPerBlock, err)

	tx = &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: maxGasLimitPerBlock - 1,
		Value:    big.NewInt(0),
	}
	err = economicsData.CheckValidityTxValues(tx)
	require.Nil(t, err)
}

func TestEconomicsData_TxWithWithMoreValueThanGenesisSupplyShouldError(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 42
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: minGasLimit + 1,
		Value:    big.NewInt(0).Add(economicsData.GenesisTotalSupply(), big.NewInt(1)),
	}

	err := economicsData.CheckValidityTxValues(tx)
	assert.Equal(t, err, process.ErrTxValueTooBig)

	tx.Value.SetBytes([]byte("99999999999999999999999999999999999999999999999999999999999999999999999999999999999999"))
	err = economicsData.CheckValidityTxValues(tx)
	assert.Equal(t, err, process.ErrTxValueOutOfBounds)
}

func TestEconomicsData_SCRWithNotEnoughMoveBalanceShouldNotError(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 42
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	scr := &smartContractResult.SmartContractResult{
		GasPrice: minGasPrice + 1,
		GasLimit: minGasLimit + 1,
		Value:    big.NewInt(0).Add(economicsData.GenesisTotalSupply(), big.NewInt(1)),
	}

	err := economicsData.CheckValidityTxValues(scr)
	assert.Equal(t, err, process.ErrTxValueTooBig)

	scr.Value.SetBytes([]byte("99999999999999999999999999999999999999999999999999999999999999999999999999999999999999"))
	err = economicsData.CheckValidityTxValues(scr)
	assert.Equal(t, err, process.ErrTxValueOutOfBounds)

	scr.Value.Set(big.NewInt(10))
	scr.GasLimit = 1
	err = economicsData.CheckValidityTxValues(scr)
	assert.Nil(t, err)

	moveBalanceFee := economicsData.ComputeMoveBalanceFee(scr)
	assert.Equal(t, moveBalanceFee, big.NewInt(0))
}

func TestEconomicsData_MoveBalanceNoData(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 500,
	}

	gasUsed := economicData.ComputeGasLimit(tx)
	require.Equal(t, uint64(500), gasUsed)

	fee := economicData.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)
	require.Equal(t, big.NewInt(500000000000), fee)
}

func TestEconomicsData_MoveBalanceWithData(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 1000,
		Data:     []byte("hello"),
	}

	gasUsed := economicData.ComputeGasLimit(tx)
	require.Equal(t, uint64(505), gasUsed)

	fee := economicData.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)
	require.Equal(t, big.NewInt(505000000000), fee)
}

func TestEconomicsData_ComputeGasUsedAndFeeBasedOnRefundValue(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 1000,
		Data:     []byte("hello"),
	}

	refundValue := big.NewInt(42500000000)

	gasUsed, fee := economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, uint64(957), gasUsed)
	require.Equal(t, big.NewInt(957500000000), fee)
}

func TestEconomicsData_ComputeGasUsedAndFeeBasedOnRefundValueFeeMultiplier(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(0.5))

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 1000,
		Data:     []byte("hello"),
	}

	refundValue := big.NewInt(42500000000)

	gasUsed, fee := economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, uint64(915), gasUsed)
	require.Equal(t, big.NewInt(710000000000), fee)
}

func TestEconomicsData_ComputeTxFeeBasedOnGasUsed(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(0.5))

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 1000,
		Data:     []byte("hello"),
	}

	// expected fee should be equal with: moveBalance*gasPrice + (gasLimit-moveBalanceGas)*gasPrice*gasModifier
	expectedFee := 505*1000000000 + 495*500000000

	fee := economicData.ComputeTxFeeBasedOnGasUsed(tx, tx.GasLimit)
	require.Equal(t, big.NewInt(int64(expectedFee)), fee)
}

func TestEconomicsData_ComputeGasUsedAndFeeBasedOnRefundValueZero(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(0.5))

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 1000,
		Data:     []byte("hello"),
	}

	// expected fee should be equal with: moveBalance*gasPrice + (gasLimit-moveBalanceGas)*gasPrice*gasModifier
	gasUsed, fee := economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx, big.NewInt(0))
	require.Equal(t, uint64(1000), gasUsed)
	require.Equal(t, big.NewInt(752500000000), fee)
}

func TestEconomicsData_ComputeGasUsedAndFeeBasedOnRefundValueCheckGasUsedValue(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsDataRealFees())
	txData := []byte("0061736d0100000001150460037f7f7e017f60027f7f017e60017e0060000002420303656e7611696e74363473746f7261676553746f7265000003656e7610696e74363473746f726167654c6f6164000103656e760b696e74363466696e6973680002030504030303030405017001010105030100020608017f01419088040b072f05066d656d6f7279020004696e6974000309696e6372656d656e7400040964656372656d656e7400050367657400060a8a01041300418088808000410742011080808080001a0b2e01017e4180888080004107418088808000410710818080800042017c22001080808080001a20001082808080000b2e01017e41808880800041074180888080004107108180808000427f7c22001080808080001a20001082808080000b160041808880800041071081808080001082808080000b0b0f01004180080b08434f554e54455200@0500@0100")
	tx1 := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 3000000,
		Data:     txData,
	}

	tx2 := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 4000000,
		Data:     txData,
	}

	tx3 := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 5000000,
		Data:     txData,
	}

	tx4 := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 4123456,
		Data:     txData,
	}

	expectedGasUsed := uint64(1770511)
	expectedFee, _ := big.NewInt(0).SetString("1077005110000000", 10)

	refundValue, _ := big.NewInt(0).SetString("12294890000000", 10)
	gasUsed, fee := economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx1, refundValue)
	require.Equal(t, expectedGasUsed, gasUsed)
	require.Equal(t, expectedFee, fee)

	refundValue, _ = big.NewInt(0).SetString("22294890000000", 10)
	gasUsed, fee = economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx2, refundValue)
	require.Equal(t, expectedFee, fee)
	require.Equal(t, expectedGasUsed, gasUsed)

	refundValue, _ = big.NewInt(0).SetString("32294890000000", 10)
	gasUsed, fee = economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx3, refundValue)
	require.Equal(t, expectedFee, fee)
	require.Equal(t, expectedGasUsed, gasUsed)

	refundValue, _ = big.NewInt(0).SetString("23529450000000", 10)
	gasUsed, fee = economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx4, refundValue)
	require.Equal(t, expectedFee, fee)
	require.Equal(t, expectedGasUsed, gasUsed)
}

func TestEconomicsData_ComputeGasUsedAndFeeBasedOnRefundValueCheck(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsDataRealFees())
	txData := []byte("0061736d0100000001150460037f7f7e017f60027f7f017e60017e0060000002420303656e7611696e74363473746f7261676553746f7265000003656e7610696e74363473746f726167654c6f6164000103656e760b696e74363466696e6973680002030504030303030405017001010105030100020608017f01419088040b072f05066d656d6f7279020004696e6974000309696e6372656d656e7400040964656372656d656e7400050367657400060a8a01041300418088808000410742011080808080001a0b2e01017e4180888080004107418088808000410710818080800042017c22001080808080001a20001082808080000b2e01017e41808880800041074180888080004107108180808000427f7c22001080808080001a20001082808080000b160041808880800041071081808080001082808080000b0b0f01004180080b08434f554e54455200@0500@0100")
	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 1200000,
		Data:     txData,
	}

	expectedGasUsed := uint64(1200000)
	expectedFee, _ := big.NewInt(0).SetString("1071300000000000", 10)

	refundValue, _ := big.NewInt(0).SetString("0", 10)
	gasUsed, fee := economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, expectedGasUsed, gasUsed)
	require.Equal(t, expectedFee, fee)
}
