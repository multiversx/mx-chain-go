package economics

import (
	"math/big"
	"strconv"
	"sync"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.RewardsHandler = (*EconomicsData)(nil)
var _ process.FeeHandler = (*EconomicsData)(nil)

// EconomicsData will store information about economics
type EconomicsData struct {
	leaderPercentage                 float64
	protocolSustainabilityPercentage float64
	protocolSustainabilityAddress    string
	maxGasLimitPerBlock              uint64
	maxGasLimitPerMetaBlock          uint64
	gasPerDataByte                   uint64
	dataLimitForBaseCalc             uint64
	minGasPrice                      uint64
	minGasLimit                      uint64
	developerPercentage              float64
	genesisTotalSupply               *big.Int
	minInflation                     float64
	yearSettings                     map[uint32]*config.YearSetting
	mutYearSettings                  sync.RWMutex
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

	if data.maxGasLimitPerBlock < data.minGasLimit {
		return nil, process.ErrInvalidMaxGasLimitPerBlock
	}

	ed := &EconomicsData{
		leaderPercentage:                 economics.RewardsSettings.LeaderPercentage,
		protocolSustainabilityPercentage: economics.RewardsSettings.ProtocolSustainabilityPercentage,
		protocolSustainabilityAddress:    economics.RewardsSettings.ProtocolSustainabilityAddress,
		maxGasLimitPerBlock:              data.maxGasLimitPerBlock,
		maxGasLimitPerMetaBlock:          data.maxGasLimitPerMetaBlock,
		minGasPrice:                      data.minGasPrice,
		minGasLimit:                      data.minGasLimit,
		gasPerDataByte:                   data.gasPerDataByte,
		dataLimitForBaseCalc:             data.dataLimitForBaseCalc,
		developerPercentage:              economics.RewardsSettings.DeveloperPercentage,
		minInflation:                     economics.GlobalSettings.MinimumInflation,
		genesisTotalSupply:               data.genesisTotalSupply,
	}

	ed.yearSettings = make(map[uint32]*config.YearSetting)
	for _, yearSetting := range economics.GlobalSettings.YearSettings {
		ed.yearSettings[yearSetting.Year] = &config.YearSetting{
			Year:             yearSetting.Year,
			MaximumInflation: yearSetting.MaximumInflation,
		}
	}

	return ed, nil
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

	maxGasLimitPerBlock, err := strconv.ParseUint(economics.FeeSettings.MaxGasLimitPerBlock, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidMaxGasLimitPerBlock
	}

	maxGasLimitPerMetaBlock, err := strconv.ParseUint(economics.FeeSettings.MaxGasLimitPerMetaBlock, conversionBase, bitConversionSize)
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

	genesisTotalSupply, ok := big.NewInt(0).SetString(economics.GlobalSettings.GenesisTotalSupply, conversionBase)
	if !ok {
		return nil, process.ErrInvalidGenesisTotalSupply
	}

	return &EconomicsData{
		minGasPrice:             minGasPrice,
		minGasLimit:             minGasLimit,
		maxGasLimitPerBlock:     maxGasLimitPerBlock,
		maxGasLimitPerMetaBlock: maxGasLimitPerMetaBlock,
		gasPerDataByte:          gasPerDataByte,
		dataLimitForBaseCalc:    dataLimitForBaseCalc,
		genesisTotalSupply:      genesisTotalSupply,
	}, nil
}

func checkValues(economics *config.EconomicsConfig) error {
	if isPercentageInvalid(economics.RewardsSettings.LeaderPercentage) ||
		isPercentageInvalid(economics.RewardsSettings.DeveloperPercentage) ||
		isPercentageInvalid(economics.RewardsSettings.ProtocolSustainabilityPercentage) ||
		isPercentageInvalid(economics.GlobalSettings.MinimumInflation) {
		return process.ErrInvalidRewardsPercentages
	}

	for _, yearSetting := range economics.GlobalSettings.YearSettings {
		if isPercentageInvalid(yearSetting.MaximumInflation) {
			return process.ErrInvalidInflationPercentages
		}
	}

	if len(economics.RewardsSettings.ProtocolSustainabilityAddress) == 0 {
		return process.ErrNilProtocolSustainabilityAddress
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
func (ed *EconomicsData) MaxInflationRate(year uint32) float64 {
	ed.mutYearSettings.RLock()
	yearSetting, ok := ed.yearSettings[year]
	ed.mutYearSettings.RUnlock()

	if !ok {
		return ed.minInflation
	}

	return yearSetting.MaximumInflation
}

// GenesisTotalSupply will return the genesis total supply
func (ed *EconomicsData) GenesisTotalSupply() *big.Int {
	return ed.genesisTotalSupply
}

// MinGasPrice will return min gas price
func (ed *EconomicsData) MinGasPrice() uint64 {
	return ed.minGasPrice
}

// MinGasLimit will return min gas limit
func (ed *EconomicsData) MinGasLimit() uint64 {
	return ed.minGasLimit
}

// GasPerDataByte will return the gas required for a data byte
func (ed *EconomicsData) GasPerDataByte() uint64 {
	return ed.gasPerDataByte
}

// ComputeMoveBalanceFee computes the provided transaction's fee
func (ed *EconomicsData) ComputeMoveBalanceFee(tx process.TransactionWithFeeHandler) *big.Int {
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
	if tx.GetGasLimit() < requiredGasLimit {
		return process.ErrInsufficientGasLimitInTx
	}

	if tx.GetGasLimit() >= ed.maxGasLimitPerBlock {
		return process.ErrHigherGasLimitRequiredInTx
	}

	// The following is required to mitigate a "big value" attack
	if len(tx.GetValue().Bytes()) > len(ed.genesisTotalSupply.Bytes()) {
		return process.ErrTxValueOutOfBounds
	}

	if tx.GetValue().Cmp(ed.genesisTotalSupply) > 0 {
		return process.ErrTxValueTooBig
	}

	return nil
}

// MaxGasLimitPerBlock will return maximum gas limit allowed per block
func (ed *EconomicsData) MaxGasLimitPerBlock(shardID uint32) uint64 {
	if shardID == core.MetachainShardId {
		return ed.maxGasLimitPerMetaBlock
	}
	return ed.maxGasLimitPerBlock
}

// DeveloperPercentage will return the developer percentage value
func (ed *EconomicsData) DeveloperPercentage() float64 {
	return ed.developerPercentage
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (ed *EconomicsData) ProtocolSustainabilityPercentage() float64 {
	return ed.protocolSustainabilityPercentage
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (ed *EconomicsData) ProtocolSustainabilityAddress() string {
	return ed.protocolSustainabilityAddress
}

// ComputeGasLimit returns the gas limit need by the provided transaction in order to be executed
func (ed *EconomicsData) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	gasLimit := ed.minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * ed.gasPerDataByte

	return gasLimit
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *EconomicsData) IsInterfaceNil() bool {
	return ed == nil
}
