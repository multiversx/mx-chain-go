package economics_test

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
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
			RewardsConfigByEpoch: []config.EpochRewardSettings{
				{
					LeaderPercentage:                 0.1,
					DeveloperPercentage:              0.1,
					ProtocolSustainabilityPercentage: 0.1,
					ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
					TopUpGradientPoint:               "300000000000000000000",
					TopUpFactor:                      0.25,
					EpochEnable:                      0,
				},
			},
		},
		FeeSettings: feeSettings,
	}
}

func feeSettingsDummy(gasModifier float64) config.FeeSettings {
	return config.FeeSettings{
		GasLimitSettings: []config.GasLimitSetting{
			{
				MaxGasLimitPerBlock:         "100000",
				MaxGasLimitPerMiniBlock:     "100000",
				MaxGasLimitPerMetaBlock:     "1000000",
				MaxGasLimitPerMetaMiniBlock: "1000000",
				MaxGasLimitPerTx:            "100000",
				MinGasLimit:                 "500",
				ExtraGasLimitGuardedTx:      "50000",
			},
		},
		MinGasPrice:            "18446744073709551615",
		GasPerDataByte:         "1",
		GasPriceModifier:       gasModifier,
		MaxGasPriceSetGuardian: "200000",
	}
}

func feeSettingsReal() config.FeeSettings {
	return config.FeeSettings{
		GasLimitSettings: []config.GasLimitSetting{
			{
				MaxGasLimitPerBlock:         "1500000000",
				MaxGasLimitPerMiniBlock:     "1500000000",
				MaxGasLimitPerMetaBlock:     "15000000000",
				MaxGasLimitPerMetaMiniBlock: "15000000000",
				MaxGasLimitPerTx:            "1500000000",
				MinGasLimit:                 "50000",
				ExtraGasLimitGuardedTx:      "50000",
			},
		},
		MinGasPrice:            "1000000000",
		GasPerDataByte:         "1500",
		GasPriceModifier:       0.01,
		MaxGasPriceSetGuardian: "200000",
	}
}

func createArgsForEconomicsData(gasModifier float64) economics.ArgsNewEconomicsData {
	feeSettings := feeSettingsDummy(gasModifier)
	args := economics.ArgsNewEconomicsData{
		Economics:     createDummyEconomicsConfig(feeSettings),
		EpochNotifier: &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.GasPriceModifierFlag
			},
		},
		BuiltInFunctionsCostHandler: &mock.BuiltInCostHandlerStub{},
		TxVersionChecker:            &testscommon.TxVersionCheckerStub{},
	}
	return args
}

func createArgsForEconomicsDataRealFees(handler economics.BuiltInFunctionsCostHandler) economics.ArgsNewEconomicsData {
	feeSettings := feeSettingsReal()
	args := economics.ArgsNewEconomicsData{
		Economics:     createDummyEconomicsConfig(feeSettings),
		EpochNotifier: &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.GasPriceModifierFlag
			},
		},
		BuiltInFunctionsCostHandler: handler,
		TxVersionChecker:            &testscommon.TxVersionCheckerStub{},
	}
	return args
}

func TestNewEconomicsData_NilOrEmptyEpochRewardsConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.RewardsSettings.RewardsConfigByEpoch = nil

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrEmptyEpochRewardsConfig, err)

	args.Economics.RewardsSettings.RewardsConfigByEpoch = make([]config.EpochRewardSettings, 0)
	_, err = economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrEmptyEpochRewardsConfig, err)
}

func TestNewEconomicsData_NilOrEmptyYearSettingsShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.GlobalSettings.YearSettings = nil

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrEmptyYearSettings, err)

	args.Economics.GlobalSettings.YearSettings = make([]*config.YearSetting, 0)
	_, err = economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrEmptyYearSettings, err)
}

func TestNewEconomicsData_NilOrEmptyGasLimitSettingsShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.FeeSettings.GasLimitSettings = nil

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrEmptyGasLimitSettings, err)

	args.Economics.FeeSettings.GasLimitSettings = make([]config.GasLimitSetting, 0)
	_, err = economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrEmptyGasLimitSettings, err)
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
		args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerBlock = gasLimitPerBlock
		_, err := economics.NewEconomicsData(args)
		assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerBlock))
	}
}

func TestNewEconomicsData_InvalidMaxGasLimitPerMiniBlockShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	badGasLimitPerMiniBlock := []string{
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

	for _, gasLimitPerMiniBlock := range badGasLimitPerMiniBlock {
		args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMiniBlock = gasLimitPerMiniBlock
		_, err := economics.NewEconomicsData(args)
		assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerMiniBlock))
	}
}

func TestNewEconomicsData_InvalidMaxGasLimitPerMetaBlockShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	badGasLimitPerMetaBlock := []string{
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

	for _, gasLimitPerMetaBlock := range badGasLimitPerMetaBlock {
		args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaBlock = gasLimitPerMetaBlock
		_, err := economics.NewEconomicsData(args)
		assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerMetaBlock))
	}
}

func TestNewEconomicsData_InvalidMaxGasLimitPerMetaMiniBlockShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	badGasLimitPerMetaMiniBlock := []string{
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

	for _, gasLimitPerMetaMiniBlock := range badGasLimitPerMetaMiniBlock {
		args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaMiniBlock = gasLimitPerMetaMiniBlock
		_, err := economics.NewEconomicsData(args)
		assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerMetaMiniBlock))
	}
}

func TestNewEconomicsData_InvalidMaxGasLimitPerTxShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	badGasLimitPerTx := []string{
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

	for _, gasLimitPerTx := range badGasLimitPerTx {
		args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerTx = gasLimitPerTx
		_, err := economics.NewEconomicsData(args)
		assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerTx))
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
		args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = minGasLimit
		_, err := economics.NewEconomicsData(args)
		assert.Equal(t, process.ErrInvalidMinimumGasLimitForTx, err)
	}
}

func TestNewEconomicsData_InvalidLeaderPercentageShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.RewardsSettings.RewardsConfigByEpoch[0].LeaderPercentage = -0.1

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrInvalidRewardsPercentages, err)
}

func TestNewEconomicsData_InvalidMinimumInflationShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.GlobalSettings.MinimumInflation = -0.1

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrInvalidInflationPercentages, err)
}

func TestNewEconomicsData_InvalidMaximumInflationShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.GlobalSettings.YearSettings[0].MaximumInflation = -0.1

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrInvalidInflationPercentages, err)
}

func TestNewEconomicsData_InvalidGasPriceModifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.FeeSettings.GasPriceModifier = 1.1

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrInvalidGasModifier, err)
}

func TestNewEconomicsData_InvalidExtraGasLimitGuardedTxShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	badExtraGasLimitGuardedTx := []string{
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

	for _, gasLimitGuardedTx := range badExtraGasLimitGuardedTx {
		args.Economics.FeeSettings.GasLimitSettings[0].ExtraGasLimitGuardedTx = gasLimitGuardedTx
		_, err := economics.NewEconomicsData(args)
		assert.True(t, errors.Is(err, process.ErrInvalidExtraGasLimitGuardedTx))
	}
}

func TestNewEconomicsData_MaxGasLimitPerBlockLowerThanMinGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerBlock = "1"
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = "2"

	_, err := economics.NewEconomicsData(args)
	assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerBlock))
}

func TestNewEconomicsData_MaxGasLimitPerMiniBlockLowerThanMinGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMiniBlock = "1"
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = "2"

	_, err := economics.NewEconomicsData(args)
	assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerMiniBlock))
}

func TestNewEconomicsData_MaxGasLimitPerMetaBlockLowerThanMinGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaBlock = "1"
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = "2"

	_, err := economics.NewEconomicsData(args)
	assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerMetaBlock))
}

func TestNewEconomicsData_MaxGasLimitPerMetaMiniBlockLowerThanMinGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaMiniBlock = "1"
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = "2"

	_, err := economics.NewEconomicsData(args)
	assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerMetaMiniBlock))
}

func TestNewEconomicsData_MaxGasLimitPerTxLowerThanMinGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerTx = "1"
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = "2"

	_, err := economics.NewEconomicsData(args)
	assert.True(t, errors.Is(err, process.ErrInvalidMaxGasLimitPerTx))
}

func TestNewEconomicsData_InvalidGasPerDataByteShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	badGasPerDataByte := []string{
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

	for _, gasPerDataByte := range badGasPerDataByte {
		args.Economics.FeeSettings.GasPerDataByte = gasPerDataByte
		_, err := economics.NewEconomicsData(args)
		assert.True(t, errors.Is(err, process.ErrInvalidGasPerDataByte))
	}
}

func TestNewEconomicsData_InvalidMaxGasPriceSetGuardianShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	badMaxGasPriceSetGuardian := []string{
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

	for _, maxGasPerSetGuardian := range badMaxGasPriceSetGuardian {
		args.Economics.FeeSettings.MaxGasPriceSetGuardian = maxGasPerSetGuardian
		_, err := economics.NewEconomicsData(args)
		assert.True(t, errors.Is(err, process.ErrInvalidMaxGasPriceSetGuardian))
	}
}

func TestNewEconomicsData_InvalidGenesisTotalSupplyShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.GlobalSettings.GenesisTotalSupply = "invalid"

	_, err := economics.NewEconomicsData(args)
	assert.True(t, errors.Is(err, process.ErrInvalidGenesisTotalSupply))
}

func TestNewEconomicsData_InvalidProtocolSustainabilityAddressShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.RewardsSettings.RewardsConfigByEpoch[0].ProtocolSustainabilityAddress = ""

	_, err := economics.NewEconomicsData(args)
	assert.True(t, errors.Is(err, process.ErrNilProtocolSustainabilityAddress))
}

func TestNewEconomicsData_InvalidTopUpGradientPointShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.Economics.RewardsSettings.RewardsConfigByEpoch[0].TopUpGradientPoint = "invalid"

	_, err := economics.NewEconomicsData(args)
	assert.True(t, errors.Is(err, process.ErrInvalidRewardsTopUpGradientPoint))
}

func TestNewEconomicsData_NilBuiltInFunctionsCostHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.BuiltInFunctionsCostHandler = nil

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrNilBuiltInFunctionsCostHandler, err)
}

func TestNewEconomicsData_NilTxVersionCheckerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.TxVersionChecker = nil

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrNilTransactionVersionChecker, err)
}

func TestNewEconomicsData_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.EpochNotifier = nil

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrNilEpochNotifier, err)
}

func TestNewEconomicsData_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	args.EnableEpochsHandler = nil

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
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
	args.Economics.RewardsSettings.RewardsConfigByEpoch[0].LeaderPercentage = leaderPercentage
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.LeaderPercentage()
	assert.Equal(t, leaderPercentage, value)
}

func TestEconomicsData_ComputeMoveBalanceFeeNoTxData(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	gasPrice := uint64(500)
	minGasLimit := uint64(12)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = strconv.FormatUint(minGasLimit, 10)
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
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = strconv.FormatUint(minGasLimit, 10)
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
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	args.Economics.FeeSettings.GasPriceModifier = 0.01
	args.EpochNotifier = forking.NewGenericEpochNotifier()
	args.EnableEpochsHandler, _ = enablers.NewEnableEpochsHandler(config.EnableEpochs{
		PenalizedTooMuchGasEnableEpoch: 1,
		GasPriceModifierEnableEpoch:    2,
	}, args.EpochNotifier)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	cost := economicsData.ComputeTxFee(tx)
	expectedCost := core.SafeMul(minGasLimit, gasPrice)
	assert.Equal(t, expectedCost, cost)

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{EpochField: 1, TimestampField: 0})

	cost = economicsData.ComputeTxFee(tx)
	expectedCost = core.SafeMul(gasLimit, gasPrice)
	assert.Equal(t, expectedCost, cost)

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{EpochField: 2, TimestampField: 0})
	cost = economicsData.ComputeTxFee(tx)
	assert.Equal(t, big.NewInt(5050), cost)
}

func TestEconomicsData_ConfirmedEpochRewardsSettingsChangeOrderedConfigs(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	rs := []config.EpochRewardSettings{
		{
			LeaderPercentage:                 0.1,
			DeveloperPercentage:              0.1,
			ProtocolSustainabilityPercentage: 0.1,
			ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			TopUpGradientPoint:               "300000000000000000000",
			TopUpFactor:                      0.25,
			EpochEnable:                      0,
		},
		{
			LeaderPercentage:                 0.2,
			DeveloperPercentage:              0.2,
			ProtocolSustainabilityPercentage: 0.2,
			ProtocolSustainabilityAddress:    "erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49",
			TopUpGradientPoint:               "200000000000000000000",
			TopUpFactor:                      0.5,
			EpochEnable:                      2,
		},
	}

	args.Economics.RewardsSettings = config.RewardsSettings{RewardsConfigByEpoch: rs}
	economicsData, _ := economics.NewEconomicsData(args)

	rewardsActiveConfig := economicsData.GetRewardsActiveConfig(1)
	require.NotNil(t, rewardsActiveConfig)
	require.Equal(t, rs[0], *rewardsActiveConfig)

	rewardsActiveConfig = economicsData.GetRewardsActiveConfig(2)
	require.NotNil(t, rewardsActiveConfig)
	require.Equal(t, rs[0], *rewardsActiveConfig)

	rewardsActiveConfig = economicsData.GetRewardsActiveConfig(3)
	require.NotNil(t, rewardsActiveConfig)
	require.Equal(t, rs[1], *rewardsActiveConfig)
}

func TestEconomicsData_ConfirmedGasLimitSettingsChangeOrderedConfigs(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	gls := []config.GasLimitSetting{
		{
			EnableEpoch:                 0,
			MaxGasLimitPerBlock:         "1500000000",
			MaxGasLimitPerMiniBlock:     "1500000000",
			MaxGasLimitPerMetaBlock:     "15000000000",
			MaxGasLimitPerMetaMiniBlock: "15000000000",
			MaxGasLimitPerTx:            "1500000000",
			MinGasLimit:                 "50000",
			ExtraGasLimitGuardedTx:      "50000",
		},
		{
			EnableEpoch:                 2,
			MaxGasLimitPerBlock:         "1500000000",
			MaxGasLimitPerMiniBlock:     "500000000",
			MaxGasLimitPerMetaBlock:     "15000000000",
			MaxGasLimitPerMetaMiniBlock: "5000000000",
			MaxGasLimitPerTx:            "500000000",
			MinGasLimit:                 "50000",
			ExtraGasLimitGuardedTx:      "50000",
		},
	}

	args.Economics.FeeSettings.GasLimitSettings = gls
	economicsData, _ := economics.NewEconomicsData(args)

	gasLimitSetting := economicsData.GetGasLimitSetting(1)
	require.NotNil(t, gasLimitSetting)
	require.Equal(t, gls[0], *gasLimitSetting)

	gasLimitSetting = economicsData.GetGasLimitSetting(2)
	require.NotNil(t, gasLimitSetting)
	require.Equal(t, gls[1], *gasLimitSetting)

	gasLimitSetting = economicsData.GetGasLimitSetting(3)
	require.NotNil(t, gasLimitSetting)
	require.Equal(t, gls[1], *gasLimitSetting)
}

func TestEconomicsData_ConfirmedEpochRewardsSettingsChangeUnOrderedConfigs(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	rs := []config.EpochRewardSettings{
		{
			LeaderPercentage:                 0.2,
			DeveloperPercentage:              0.2,
			ProtocolSustainabilityPercentage: 0.2,
			ProtocolSustainabilityAddress:    "erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49",
			TopUpGradientPoint:               "200000000000000000000",
			TopUpFactor:                      0.5,
			EpochEnable:                      2,
		},
		{
			LeaderPercentage:                 0.1,
			DeveloperPercentage:              0.1,
			ProtocolSustainabilityPercentage: 0.1,
			ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			TopUpGradientPoint:               "300000000000000000000",
			TopUpFactor:                      0.25,
			EpochEnable:                      0,
		},
	}

	args.Economics.RewardsSettings = config.RewardsSettings{RewardsConfigByEpoch: rs}
	economicsData, _ := economics.NewEconomicsData(args)

	rewardsActiveConfig := economicsData.GetRewardsActiveConfig(1)
	require.NotNil(t, rewardsActiveConfig)
	require.Equal(t, rs[1], *rewardsActiveConfig)

	rewardsActiveConfig = economicsData.GetRewardsActiveConfig(2)
	require.NotNil(t, rewardsActiveConfig)
	require.Equal(t, rs[1], *rewardsActiveConfig)

	rewardsActiveConfig = economicsData.GetRewardsActiveConfig(3)
	require.NotNil(t, rewardsActiveConfig)
	require.Equal(t, rs[0], *rewardsActiveConfig)
}

func TestEconomicsData_ConfirmedGasLimitSettingsChangeUnOrderedConfigs(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	gls := []config.GasLimitSetting{
		{
			EnableEpoch:                 2,
			MaxGasLimitPerBlock:         "1500000000",
			MaxGasLimitPerMiniBlock:     "500000000",
			MaxGasLimitPerMetaBlock:     "15000000000",
			MaxGasLimitPerMetaMiniBlock: "5000000000",
			MaxGasLimitPerTx:            "500000000",
			MinGasLimit:                 "50000",
			ExtraGasLimitGuardedTx:      "50000",
		},
		{
			EnableEpoch:                 0,
			MaxGasLimitPerBlock:         "1500000000",
			MaxGasLimitPerMiniBlock:     "1500000000",
			MaxGasLimitPerMetaBlock:     "15000000000",
			MaxGasLimitPerMetaMiniBlock: "15000000000",
			MaxGasLimitPerTx:            "1500000000",
			MinGasLimit:                 "50000",
			ExtraGasLimitGuardedTx:      "50000",
		},
	}

	args.Economics.FeeSettings.GasLimitSettings = gls
	economicsData, _ := economics.NewEconomicsData(args)

	gasLimitSetting := economicsData.GetGasLimitSetting(1)
	require.NotNil(t, gasLimitSetting)
	require.Equal(t, gls[1], *gasLimitSetting)

	gasLimitSetting = economicsData.GetGasLimitSetting(2)
	require.NotNil(t, gasLimitSetting)
	require.Equal(t, gls[0], *gasLimitSetting)

	gasLimitSetting = economicsData.GetGasLimitSetting(3)
	require.NotNil(t, gasLimitSetting)
	require.Equal(t, gls[0], *gasLimitSetting)
}

func TestEconomicsData_TxWithLowerGasPriceShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = fmt.Sprintf("%d", minGasLimit)
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
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit - 1,
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Equal(t, process.ErrInsufficientGasLimitInTx, err)
}

// This test should not be modified due to backwards compatibility reasons
func TestEconomicsData_TxWithHigherGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaMiniBlock = fmt.Sprintf("%d", maxGasLimitPerBlock*10)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = fmt.Sprintf("%d", minGasLimit)
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
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit,
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Nil(t, err)
}

func TestEconomicsData_TxWithWithMoreGasLimitThanMaximumPerMiniBlockForSafeCrossShardShouldNotWork(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := uint64(42)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaMiniBlock = fmt.Sprintf("%d", maxGasLimitPerBlock*10)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)

	t.Run("maximum gas limit as defined should work", func(t *testing.T) {
		// do not change this behavior: backwards compatibility reasons
		tx := &transaction.Transaction{
			GasPrice: minGasPrice + 1,
			GasLimit: maxGasLimitPerBlock,
			Value:    big.NewInt(0),
		}
		err := economicsData.CheckValidityTxValues(tx)
		require.Equal(t, process.ErrMoreGasThanGasLimitPerBlock, err)
	})
	t.Run("maximum gas limit + 1 as defined should error", func(t *testing.T) {
		tx := &transaction.Transaction{
			GasPrice: minGasPrice + 1,
			GasLimit: maxGasLimitPerBlock + 1,
			Value:    big.NewInt(0),
		}
		err := economicsData.CheckValidityTxValues(tx)
		require.Equal(t, process.ErrMoreGasThanGasLimitPerBlock, err)
	})
	t.Run("maximum gas limit - 1 as defined should work", func(t *testing.T) {
		tx := &transaction.Transaction{
			GasPrice: minGasPrice + 1,
			GasLimit: maxGasLimitPerBlock - 1,
			Value:    big.NewInt(0),
		}
		err := economicsData.CheckValidityTxValues(tx)
		require.Nil(t, err)
	})
}

func TestEconomicsData_TxWithWithMoreValueThanGenesisSupplyShouldError(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 42
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = fmt.Sprintf("%d", minGasLimit)
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
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = fmt.Sprintf("%d", minGasLimit)
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

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsDataRealFees(&mock.BuiltInCostHandlerStub{}))
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

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsDataRealFees(&mock.BuiltInCostHandlerStub{}))
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

func TestEconomicsData_ComputeGasUsedAndFeeBasedOnRefundValueSpecialBuiltIn_ToMuchGasProvided(t *testing.T) {
	t.Parallel()

	builtInCostHandler, _ := economics.NewBuiltInFunctionsCost(&economics.ArgsBuiltInFunctionCost{
		GasSchedule: testscommon.NewGasScheduleNotifierMock(defaults.FillGasMapInternal(map[string]map[string]uint64{}, 1)),
		ArgsParser:  smartContract.NewArgumentParser(),
	})
	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsDataRealFees(builtInCostHandler))

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 1200000,
		Data:     []byte("ESDTTransfer@54474e2d383862383366@0a"),
	}

	expectedGasUsed := uint64(1200000)
	expectedFee, _ := big.NewInt(0).SetString("114960000000000", 10)

	refundValue, _ := big.NewInt(0).SetString("0", 10)
	gasUsed, fee := economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, expectedGasUsed, gasUsed)
	require.Equal(t, expectedFee, fee)
}

func TestEconomicsData_ComputeGasUsedAndFeeBasedOnRefundValueStakeTx(t *testing.T) {
	builtInCostHandler, _ := economics.NewBuiltInFunctionsCost(&economics.ArgsBuiltInFunctionCost{
		GasSchedule: testscommon.NewGasScheduleNotifierMock(defaults.FillGasMapInternal(map[string]map[string]uint64{}, 1)),
		ArgsParser:  smartContract.NewArgumentParser(),
	})

	txStake := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 250000000,
		Data:     []byte("stake"),
	}

	expectedGasUsed := uint64(39378847)
	expectedFee, _ := big.NewInt(0).SetString("39378847000000000", 10)

	args := createArgsForEconomicsDataRealFees(builtInCostHandler)
	args.EpochNotifier = forking.NewGenericEpochNotifier()
	args.EnableEpochsHandler, _ = enablers.NewEnableEpochsHandler(config.EnableEpochs{
		PenalizedTooMuchGasEnableEpoch: 1000,
		GasPriceModifierEnableEpoch:    1000,
	}, args.EpochNotifier)
	economicData, _ := economics.NewEconomicsData(args)

	refundValueStake, _ := big.NewInt(0).SetString("210621153000000000", 10)
	gasUsedStake, feeStake := economicData.ComputeGasUsedAndFeeBasedOnRefundValue(txStake, refundValueStake)
	require.Equal(t, expectedGasUsed, gasUsedStake)
	require.Equal(t, expectedFee, feeStake)
}

func TestEconomicsData_ComputeGasUsedAndFeeBasedOnRefundValueSpecialBuiltIn(t *testing.T) {
	t.Parallel()

	builtInCostHandler, _ := economics.NewBuiltInFunctionsCost(&economics.ArgsBuiltInFunctionCost{
		GasSchedule: testscommon.NewGasScheduleNotifierMock(defaults.FillGasMapInternal(map[string]map[string]uint64{}, 1)),
		ArgsParser:  smartContract.NewArgumentParser(),
	})
	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsDataRealFees(builtInCostHandler))

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 104009,
		Data:     []byte("ESDTTransfer@54474e2d383862383366@0a"),
	}

	expectedGasUsed := uint64(104001)
	expectedFee, _ := big.NewInt(0).SetString("104000010000000", 10)

	refundValue, _ := big.NewInt(0).SetString("0", 10)
	gasUsed, fee := economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, expectedGasUsed, gasUsed)
	require.Equal(t, expectedFee, fee)
}

func TestEconomicsData_ComputeGasUsedAndFeeBasedOnRefundValueSpecialBuiltInTooMuchGas(t *testing.T) {
	t.Parallel()

	builtInCostHandler, _ := economics.NewBuiltInFunctionsCost(&economics.ArgsBuiltInFunctionCost{
		GasSchedule: testscommon.NewGasScheduleNotifierMock(defaults.FillGasMapInternal(map[string]map[string]uint64{}, 1)),
		ArgsParser:  smartContract.NewArgumentParser(),
	})
	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsDataRealFees(builtInCostHandler))

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 104011,
		Data:     []byte("ESDTTransfer@54474e2d383862383366@0a"),
	}

	expectedGasUsed := uint64(104011)
	expectedFee, _ := big.NewInt(0).SetString("104000110000000", 10)

	refundValue, _ := big.NewInt(0).SetString("0", 10)
	gasUsed, fee := economicData.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, expectedGasUsed, gasUsed)
	require.Equal(t, expectedFee, fee)
}

func TestEconomicsData_ComputeGasLimitBasedOnBalance(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsDataRealFees(&mock.BuiltInCostHandlerStub{})
	args.EpochNotifier = forking.NewGenericEpochNotifier()
	args.EnableEpochsHandler, _ = enablers.NewEnableEpochsHandler(config.EnableEpochs{
		GasPriceModifierEnableEpoch: 1,
	}, args.EpochNotifier)
	economicData, _ := economics.NewEconomicsData(args)
	txData := []byte("0061736d0100000001150460037f7f7e017f60027f7f017e60017e0060000002420303656e7611696e74363473746f7261676553746f7265000003656e7610696e74363473746f726167654c6f6164000103656e760b696e74363466696e6973680002030504030303030405017001010105030100020608017f01419088040b072f05066d656d6f7279020004696e6974000309696e6372656d656e7400040964656372656d656e7400050367657400060a8a01041300418088808000410742011080808080001a0b2e01017e4180888080004107418088808000410710818080800042017c22001080808080001a20001082808080000b2e01017e41808880800041074180888080004107108180808000427f7c22001080808080001a20001082808080000b160041808880800041071081808080001082808080000b0b0f01004180080b08434f554e54455200@0500@0100")
	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 1200000,
		Data:     txData,
		Value:    big.NewInt(10),
	}

	senderBalance, _ := big.NewInt(0).SetString("1", 10)
	_, err := economicData.ComputeGasLimitBasedOnBalance(tx, senderBalance)
	require.Equal(t, process.ErrInsufficientFunds, err)

	senderBalance, _ = big.NewInt(0).SetString("1000", 10)
	_, err = economicData.ComputeGasLimitBasedOnBalance(tx, senderBalance)
	require.Equal(t, process.ErrInsufficientFunds, err)

	senderBalance, _ = big.NewInt(0).SetString("120000000000000010", 10)
	gasLimit, err := economicData.ComputeGasLimitBasedOnBalance(tx, senderBalance)
	require.Nil(t, err)
	require.Equal(t, uint64(120000000), gasLimit)

	senderBalance, _ = big.NewInt(0).SetString("120000000000000010", 10)

	args.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{EpochField: 10, TimestampField: 10})
	gasLimit, err = economicData.ComputeGasLimitBasedOnBalance(tx, senderBalance)
	require.Nil(t, err)
	require.Equal(t, uint64(11894070000), gasLimit)
}

func TestEconomicsData_MaxGasPriceSetGuardian(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsDataRealFees(&mock.BuiltInCostHandlerStub{})
	maxGasPriceSetGuardianString := "2000000"
	expectedMaxGasPriceSetGuardian, err := strconv.ParseUint(maxGasPriceSetGuardianString, 10, 64)
	require.Nil(t, err)
	args.Economics.FeeSettings.MaxGasPriceSetGuardian = maxGasPriceSetGuardianString
	economicData, _ := economics.NewEconomicsData(args)

	require.Equal(t, expectedMaxGasPriceSetGuardian, economicData.MaxGasPriceSetGuardian())
}

func TestEconomicsData_SetStatusHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil status handler should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsForEconomicsDataRealFees(&mock.BuiltInCostHandlerStub{})
		economicData, _ := economics.NewEconomicsData(args)

		err := economicData.SetStatusHandler(nil)
		require.Equal(t, core.ErrNilAppStatusHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgsForEconomicsDataRealFees(&mock.BuiltInCostHandlerStub{})
		economicData, _ := economics.NewEconomicsData(args)

		err := economicData.SetStatusHandler(&statusHandler.AppStatusHandlerStub{})
		require.NoError(t, err)
	})
}

func TestEconomicsData_MinInflationRate(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minInflationRate := 0.40
	args.Economics.GlobalSettings.MinimumInflation = minInflationRate
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MinInflationRate()
	assert.Equal(t, minInflationRate, value)
}

func TestEconomicsData_MaxInflationRate(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minInflationRate := 0.40
	maxInflationRate := 0.99
	args.Economics.GlobalSettings.MinimumInflation = minInflationRate
	args.Economics.GlobalSettings.YearSettings[0].MaximumInflation = maxInflationRate
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MaxInflationRate(0)
	assert.Equal(t, maxInflationRate, value)

	value = economicsData.MaxInflationRate(1) // missing from GlobalSettings
	assert.Equal(t, minInflationRate, value)
}

func TestEconomicsData_MinGasPrice(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(10000000000000000000)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MinGasPrice()
	assert.Equal(t, minGasPrice, value)
}

func TestEconomicsData_MinGasPriceForProcessing(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(10000000000000000000)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MinGasPriceForProcessing()
	assert.Equal(t, minGasPrice, value)
}

func TestEconomicsData_MinGasLimit(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	minGasPrice := uint64(100)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = fmt.Sprintf("%d", minGasPrice)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MinGasLimit()
	assert.Equal(t, minGasPrice, value)
}

func TestEconomicsData_ExtraGasLimitGuardedTx(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	extraGasLimitGuardedTx := uint64(100)
	args.Economics.FeeSettings.GasLimitSettings[0].ExtraGasLimitGuardedTx = fmt.Sprintf("%d", extraGasLimitGuardedTx)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.ExtraGasLimitGuardedTx()
	assert.Equal(t, extraGasLimitGuardedTx, value)
}

func TestEconomicsData_GasPerDataByte(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	gasPerDataByte := uint64(100)
	args.Economics.FeeSettings.GasPerDataByte = fmt.Sprintf("%d", gasPerDataByte)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.GasPerDataByte()
	assert.Equal(t, gasPerDataByte, value)
}

func TestEconomicsData_ComputeFeeForProcessing(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	gasPrice := uint64(500)
	gasLimit := uint64(20)
	minGasLimit := uint64(10)
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	args.Economics.FeeSettings.GasPriceModifier = 0.01
	args.EpochNotifier = forking.NewGenericEpochNotifier()
	args.EnableEpochsHandler, _ = enablers.NewEnableEpochsHandler(config.EnableEpochs{
		PenalizedTooMuchGasEnableEpoch: 1,
		GasPriceModifierEnableEpoch:    2,
	}, args.EpochNotifier)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	gasToUse := uint64(100)
	value := economicsData.ComputeFeeForProcessing(tx, gasToUse)
	require.Equal(t, fmt.Sprintf("%d", gasPrice*gasToUse), value.String())
}

func TestEconomicsData_GasPriceForProcessing(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	gasPrice := uint64(500)
	gasLimit := uint64(20)
	minGasLimit := uint64(10)
	gasModifier := 0.01
	args.Economics.FeeSettings.GasLimitSettings[0].MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	args.Economics.FeeSettings.GasPriceModifier = gasModifier
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	value := economicsData.GasPriceForProcessing(tx)
	require.Equal(t, uint64(float64(gasPrice)*gasModifier), value)
}

func TestEconomicsData_MaxGasLimitPerBlock(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	maxGasLimitPerBlock := uint64(100000)
	maxGasLimitPerMetaBlock := uint64(1000000)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaBlock = fmt.Sprintf("%d", maxGasLimitPerMetaBlock)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MaxGasLimitPerBlock(0)
	assert.Equal(t, maxGasLimitPerBlock, value)

	value = economicsData.MaxGasLimitPerBlock(core.MetachainShardId)
	assert.Equal(t, maxGasLimitPerMetaBlock, value)
}

func TestEconomicsData_MaxGasLimitPerMiniBlock(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	maxGasLimitPerMiniBlock := uint64(100000)
	maxGasLimitPerMetaMiniBlock := uint64(1000000)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMiniBlock = fmt.Sprintf("%d", maxGasLimitPerMiniBlock)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaMiniBlock = fmt.Sprintf("%d", maxGasLimitPerMetaMiniBlock)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MaxGasLimitPerMiniBlock(0)
	assert.Equal(t, maxGasLimitPerMiniBlock, value)

	value = economicsData.MaxGasLimitPerMiniBlock(core.MetachainShardId)
	assert.Equal(t, maxGasLimitPerMetaMiniBlock, value)
}

func TestEconomicsData_MaxGasLimitPerBlockForSafeCrossShard(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	maxGasLimitPerBlock := uint64(100000)
	maxGasLimitPerMetaBlock := uint64(1000000)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaBlock = fmt.Sprintf("%d", maxGasLimitPerMetaBlock)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MaxGasLimitPerBlockForSafeCrossShard()
	assert.Equal(t, maxGasLimitPerBlock, value)
}

func TestEconomicsData_MaxGasLimitPerMiniBlockForSafeCrossShard(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	maxGasLimitPerMiniBlock := uint64(100000)
	maxGasLimitPerMetaMiniBlock := uint64(1000000)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMiniBlock = fmt.Sprintf("%d", maxGasLimitPerMiniBlock)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerMetaMiniBlock = fmt.Sprintf("%d", maxGasLimitPerMetaMiniBlock)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MaxGasLimitPerMiniBlockForSafeCrossShard()
	assert.Equal(t, maxGasLimitPerMiniBlock, value)
}

func TestEconomicsData_MaxGasLimitPerTx(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	maxGasLimitPerTx := uint64(100000)
	args.Economics.FeeSettings.GasLimitSettings[0].MaxGasLimitPerTx = fmt.Sprintf("%d", maxGasLimitPerTx)
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.MaxGasLimitPerTx()
	assert.Equal(t, maxGasLimitPerTx, value)
}

func TestEconomicsData_DeveloperPercentage(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	developerPercentage := 0.5
	args.Economics.RewardsSettings.RewardsConfigByEpoch[0].DeveloperPercentage = developerPercentage
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.DeveloperPercentage()
	assert.Equal(t, developerPercentage, value)
}

func TestEconomicsData_ProtocolSustainabilityPercentage(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	protocolSustainabilityPercentage := 0.5
	args.Economics.RewardsSettings.RewardsConfigByEpoch[0].ProtocolSustainabilityPercentage = protocolSustainabilityPercentage
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.ProtocolSustainabilityPercentage()
	assert.Equal(t, protocolSustainabilityPercentage, value)
}

func TestEconomicsData_ProtocolSustainabilityAddress(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	protocolSustainabilityAddress := "erd12345"
	args.Economics.RewardsSettings.RewardsConfigByEpoch[0].ProtocolSustainabilityAddress = protocolSustainabilityAddress
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.ProtocolSustainabilityAddress()
	assert.Equal(t, protocolSustainabilityAddress, value)
}

func TestEconomicsData_RewardsTopUpGradientPoint(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	topUpGradientPoint := "300000000000000000000"
	args.Economics.RewardsSettings.RewardsConfigByEpoch[0].TopUpGradientPoint = topUpGradientPoint
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.RewardsTopUpGradientPoint()
	assert.Equal(t, topUpGradientPoint, value.String())
}

func TestEconomicsData_RewardsTopUpFactor(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData(1)
	topUpFactor := 0.1
	args.Economics.RewardsSettings.RewardsConfigByEpoch[0].TopUpFactor = topUpFactor
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.RewardsTopUpFactor()
	assert.Equal(t, topUpFactor, value)
}
