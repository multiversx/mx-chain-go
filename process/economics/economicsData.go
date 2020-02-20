package economics

import (
	"math/big"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

// EconomicsData will store information about economics
type EconomicsData struct {
	leaderPercentage     float64
	maxGasLimitPerBlock  uint64
	gasPerDataByte       uint64
	dataLimitForBaseCalc uint64
	minGasPrice          uint64
	minGasLimit          uint64
	genesisNodePrice     *big.Int
	unBoundPeriod        uint64
	ratingsData          *RatingsData
	developerPercentage  float64
	genesisTotalSupply   *big.Int
	minInflation         float64
	maxInflation         float64
}

// NewEconomicsData will create and object with information about economics parameters
func NewEconomicsData(economics *config.EconomicsConfig) (*EconomicsData, error) {
	data, err := convertValues(economics)
	if err != nil {
		return nil, err
	}

	err = checkValues(economics)
	if err != nil {
		return nil, err
	}

	rd, err := NewRatingsData(economics.RatingSettings)
	if err != nil {
		return nil, err
	}

	if data.maxGasLimitPerBlock < data.minGasLimit {
		return nil, process.ErrInvalidMaxGasLimitPerBlock
	}

	return &EconomicsData{
		leaderPercentage:     economics.RewardsSettings.LeaderPercentage,
		maxGasLimitPerBlock:  data.maxGasLimitPerBlock,
		minGasPrice:          data.minGasPrice,
		minGasLimit:          data.minGasLimit,
		genesisNodePrice:     data.genesisNodePrice,
		unBoundPeriod:        data.unBoundPeriod,
		gasPerDataByte:       data.gasPerDataByte,
		dataLimitForBaseCalc: data.dataLimitForBaseCalc,
		ratingsData:          rd,
		developerPercentage:  economics.RewardsSettings.DeveloperPercentage,
		minInflation:         economics.GlobalSettings.MinimumInflation,
		maxInflation:         economics.GlobalSettings.MaximumInflation,
		genesisTotalSupply:   data.genesisTotalSupply,
	}, nil
}

func convertValues(economics *config.EconomicsConfig) (*EconomicsData, error) {
	conversionBase := 10
	bitConversionSize := 64

	minGasPrice, err := strconv.ParseUint(economics.FeeSettings.MinGasPrice, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidMinimumGasPrice
	}

	minGasLimit, err := strconv.ParseUint(economics.FeeSettings.MinGasLimit, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidMinimumGasLimitForTx
	}

	genesisNodePrice := new(big.Int)
	genesisNodePrice, ok := genesisNodePrice.SetString(economics.ValidatorSettings.GenesisNodePrice, conversionBase)
	if !ok {
		return nil, process.ErrInvalidRewardsValue
	}

	unBoundPeriod, err := strconv.ParseUint(economics.ValidatorSettings.UnBoundPeriod, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidUnboundPeriod
	}

	maxGasLimitPerBlock, err := strconv.ParseUint(economics.FeeSettings.MaxGasLimitPerBlock, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidMaxGasLimitPerBlock
	}

	gasPerDataByte, err := strconv.ParseUint(economics.FeeSettings.GasPerDataByte, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidGasPerDataByte
	}

	dataLimitForBaseCalc, err := strconv.ParseUint(economics.FeeSettings.DataLimitForBaseCalc, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidGasPerDataByte
	}

	genesisTotalSupply := new(big.Int)
	genesisTotalSupply, ok = genesisTotalSupply.SetString(economics.GlobalSettings.TotalSupply, conversionBase)
	if !ok {
		return nil, process.ErrInvalidGenesisTotalSupply
	}

	return &EconomicsData{
		minGasPrice:          minGasPrice,
		minGasLimit:          minGasLimit,
		genesisNodePrice:     genesisNodePrice,
		unBoundPeriod:        unBoundPeriod,
		maxGasLimitPerBlock:  maxGasLimitPerBlock,
		gasPerDataByte:       gasPerDataByte,
		dataLimitForBaseCalc: dataLimitForBaseCalc,
		genesisTotalSupply:   genesisTotalSupply,
	}, nil
}

func checkValues(economics *config.EconomicsConfig) error {
	if isPercentageInvalid(economics.RewardsSettings.LeaderPercentage) ||
		isPercentageInvalid(economics.RewardsSettings.DeveloperPercentage) ||
		isPercentageInvalid(economics.GlobalSettings.MaximumInflation) ||
		isPercentageInvalid(economics.GlobalSettings.MaximumInflation) {
		return process.ErrInvalidRewardsPercentages
	}

	return nil
}

func isPercentageInvalid(percentage float64) bool {
	isLessThanZero := percentage < 0.0
	isGreaterThanOne := percentage > 1.0
	if isLessThanZero || isGreaterThanOne {
		return true
	}
	return false
}

// LeaderPercentage will return leader reward percentage
func (ed *EconomicsData) LeaderPercentage() float64 {
	return ed.leaderPercentage
}

// MinInflationRate will return the minimum inflation rate
func (ed *EconomicsData) MinInflationRate() float64 {
	return ed.minInflation
}

// MaxInflationRate will return the maximum inflation rate
func (ed *EconomicsData) MaxInflationRate() float64 {
	return ed.maxInflation
}

// GenesisTotalSupply will return the genesis total supply
func (ed *EconomicsData) GenesisTotalSupply() *big.Int {
	return ed.genesisTotalSupply
}

// MinGasPrice will return min gas price
func (ed *EconomicsData) MinGasPrice() uint64 {
	return ed.minGasPrice
}

// ComputeFee computes the provided transaction's fee
func (ed *EconomicsData) ComputeFee(tx process.TransactionWithFeeHandler) *big.Int {
	gasPrice := big.NewInt(0).SetUint64(tx.GetGasPrice())
	gasLimit := big.NewInt(0).SetUint64(ed.ComputeGasLimit(tx))

	return gasPrice.Mul(gasPrice, gasLimit)
}

// CheckValidityTxValues checks if the provided transaction is economically correct
func (ed *EconomicsData) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	if ed.minGasPrice > tx.GetGasPrice() {
		return process.ErrInsufficientGasPriceInTx
	}

	requiredGasLimit := ed.ComputeGasLimit(tx)
	if requiredGasLimit > tx.GetGasLimit() {
		return process.ErrInsufficientGasLimitInTx
	}

	if requiredGasLimit > ed.maxGasLimitPerBlock {
		return process.ErrHigherGasLimitRequiredInTx
	}

	return nil
}

// MaxGasLimitPerBlock will return maximum gas limit allowed per block
func (ed *EconomicsData) MaxGasLimitPerBlock() uint64 {
	return ed.maxGasLimitPerBlock
}

// DeveloperPercentage will return the developer percentage value
func (ed *EconomicsData) DeveloperPercentage() float64 {
	return ed.developerPercentage
}

// ComputeGasLimit returns the gas limit need by the provided transaction in order to be executed
func (ed *EconomicsData) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	gasLimit := ed.minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * ed.gasPerDataByte

	return gasLimit
}

// GenesisNodePrice will return the minimum stake value
func (ed *EconomicsData) GenesisNodePrice() *big.Int {
	return ed.genesisNodePrice
}

// UnBoundPeriod will return the unbound period
func (ed *EconomicsData) UnBoundPeriod() uint64 {
	return ed.unBoundPeriod
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *EconomicsData) IsInterfaceNil() bool {
	return ed == nil
}

// RatingsData will return the ratingsDataObject
func (ed *EconomicsData) RatingsData() *RatingsData {
	return ed.ratingsData
}
