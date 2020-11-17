package economics

import (
	"math/big"
	"strconv"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.RewardsHandler = (*EconomicsData)(nil)
var _ process.FeeHandler = (*EconomicsData)(nil)

var log = logger.GetOrCreate("process/economics")

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
	flagPenalizedTooMuchGas          atomic.Flag
	penalizedTooMuchGasEnableEpoch   uint32
	topUpGradientPoint               *big.Int
	topUpFactor                      float64
}

// ArgsNewEconomicsData defines the arguments needed for new economics data
type ArgsNewEconomicsData struct {
	Economics                      *config.EconomicsConfig
	PenalizedTooMuchGasEnableEpoch uint32
	EpochNotifier                  process.EpochNotifier
}

// NewEconomicsData will create and object with information about economics parameters
func NewEconomicsData(args ArgsNewEconomicsData) (*EconomicsData, error) {
	data, err := convertValues(args.Economics)
	if err != nil {
		return nil, err
	}

	err = checkValues(args.Economics)
	if err != nil {
		return nil, err
	}

	if data.maxGasLimitPerBlock < data.minGasLimit {
		return nil, process.ErrInvalidMaxGasLimitPerBlock
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	topUpGradientPoint, ok := big.NewInt(0).SetString(args.Economics.RewardsSettings.TopUpGradientPoint, 10)
	if !ok {
		return nil, process.ErrInvalidRewardsTopUpGradientPoint
	}

	ed := &EconomicsData{
		leaderPercentage:                 args.Economics.RewardsSettings.LeaderPercentage,
		protocolSustainabilityPercentage: args.Economics.RewardsSettings.ProtocolSustainabilityPercentage,
		protocolSustainabilityAddress:    args.Economics.RewardsSettings.ProtocolSustainabilityAddress,
		maxGasLimitPerBlock:              data.maxGasLimitPerBlock,
		maxGasLimitPerMetaBlock:          data.maxGasLimitPerMetaBlock,
		minGasPrice:                      data.minGasPrice,
		minGasLimit:                      data.minGasLimit,
		gasPerDataByte:                   data.gasPerDataByte,
		dataLimitForBaseCalc:             data.dataLimitForBaseCalc,
		developerPercentage:              args.Economics.RewardsSettings.DeveloperPercentage,
		minInflation:                     args.Economics.GlobalSettings.MinimumInflation,
		genesisTotalSupply:               data.genesisTotalSupply,
		penalizedTooMuchGasEnableEpoch:   args.PenalizedTooMuchGasEnableEpoch,
		topUpGradientPoint:               topUpGradientPoint,
		topUpFactor:                      args.Economics.RewardsSettings.TopUpFactor,
	}

	ed.yearSettings = make(map[uint32]*config.YearSetting)
	for _, yearSetting := range args.Economics.GlobalSettings.YearSettings {
		ed.yearSettings[yearSetting.Year] = &config.YearSetting{
			Year:             yearSetting.Year,
			MaximumInflation: yearSetting.MaximumInflation,
		}
	}

	args.EpochNotifier.RegisterNotifyHandler(ed)

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
		isPercentageInvalid(economics.GlobalSettings.MinimumInflation) ||
		isPercentageInvalid(economics.RewardsSettings.TopUpFactor) {
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
	return core.SafeMul(tx.GetGasPrice(), ed.ComputeGasLimit(tx))
}

// ComputeTxFee computes the provided transaction's fee using enable from epoch approach
func (ed *EconomicsData) ComputeTxFee(tx process.TransactionWithFeeHandler) *big.Int {
	if ed.flagPenalizedTooMuchGas.IsSet() {
		return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	return ed.ComputeMoveBalanceFee(tx)
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

// RewardsTopUpGradientPoint returns the rewards top-up gradient point
func (ed *EconomicsData) RewardsTopUpGradientPoint() *big.Int {
	return big.NewInt(0).Set(ed.topUpGradientPoint)
}

// RewardsTopUpFactor returns the rewards top-up factor
func (ed *EconomicsData) RewardsTopUpFactor() float64 {
	return ed.topUpFactor
}

// ComputeGasLimit returns the gas limit need by the provided transaction in order to be executed
func (ed *EconomicsData) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	gasLimit := ed.minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * ed.gasPerDataByte

	return gasLimit
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (ed *EconomicsData) EpochConfirmed(epoch uint32) {
	ed.flagPenalizedTooMuchGas.Toggle(epoch >= ed.penalizedTooMuchGasEnableEpoch)
	log.Debug("EconomicsData: penalized too much gas", "enabled", ed.flagPenalizedTooMuchGas.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *EconomicsData) IsInterfaceNil() bool {
	return ed == nil
}
