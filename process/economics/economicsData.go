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

var _ process.EconomicsDataHandler = (*economicsData)(nil)
var _ process.RewardsHandler = (*economicsData)(nil)
var _ process.FeeHandler = (*economicsData)(nil)

var epsilon = 0.00000001
var log = logger.GetOrCreate("process/economics")

// economicsData will store information about economics
type economicsData struct {
	leaderPercentage                 float64
	protocolSustainabilityPercentage float64
	protocolSustainabilityAddress    string
	maxGasLimitPerBlock              uint64
	maxGasLimitPerMetaBlock          uint64
	gasPerDataByte                   uint64
	minGasPrice                      uint64
	gasPriceModifier                 float64
	minGasLimit                      uint64
	developerPercentage              float64
	genesisTotalSupply               *big.Int
	minInflation                     float64
	yearSettings                     map[uint32]*config.YearSetting
	mutYearSettings                  sync.RWMutex
	flagPenalizedTooMuchGas          atomic.Flag
	flagGasPriceModifier             atomic.Flag
	penalizedTooMuchGasEnableEpoch   uint32
	gasPriceModifierEnableEpoch      uint32
}

// ArgsNewEconomicsData defines the arguments needed for new economics economicsData
type ArgsNewEconomicsData struct {
	Economics                      *config.EconomicsConfig
	PenalizedTooMuchGasEnableEpoch uint32
	EpochNotifier                  process.EpochNotifier
	GasPriceModifierEnableEpoch    uint32
}

// NewEconomicsData will create and object with information about economics parameters
func NewEconomicsData(args ArgsNewEconomicsData) (*economicsData, error) {
	convertedData, err := convertValues(args.Economics)
	if err != nil {
		return nil, err
	}

	err = checkValues(args.Economics)
	if err != nil {
		return nil, err
	}

	if convertedData.maxGasLimitPerBlock < convertedData.minGasLimit {
		return nil, process.ErrInvalidMaxGasLimitPerBlock
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	ed := &economicsData{
		leaderPercentage:                 args.Economics.RewardsSettings.LeaderPercentage,
		protocolSustainabilityPercentage: args.Economics.RewardsSettings.ProtocolSustainabilityPercentage,
		protocolSustainabilityAddress:    args.Economics.RewardsSettings.ProtocolSustainabilityAddress,
		maxGasLimitPerBlock:              convertedData.maxGasLimitPerBlock,
		maxGasLimitPerMetaBlock:          convertedData.maxGasLimitPerMetaBlock,
		minGasPrice:                      convertedData.minGasPrice,
		minGasLimit:                      convertedData.minGasLimit,
		gasPerDataByte:                   convertedData.gasPerDataByte,
		developerPercentage:              args.Economics.RewardsSettings.DeveloperPercentage,
		minInflation:                     args.Economics.GlobalSettings.MinimumInflation,
		genesisTotalSupply:               convertedData.genesisTotalSupply,
		penalizedTooMuchGasEnableEpoch:   args.PenalizedTooMuchGasEnableEpoch,
		gasPriceModifierEnableEpoch:      args.GasPriceModifierEnableEpoch,
		gasPriceModifier:                 args.Economics.FeeSettings.GasPriceModifier,
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

func convertValues(economics *config.EconomicsConfig) (*economicsData, error) {
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

	genesisTotalSupply, ok := big.NewInt(0).SetString(economics.GlobalSettings.GenesisTotalSupply, conversionBase)
	if !ok {
		return nil, process.ErrInvalidGenesisTotalSupply
	}

	return &economicsData{
		minGasPrice:             minGasPrice,
		minGasLimit:             minGasLimit,
		maxGasLimitPerBlock:     maxGasLimitPerBlock,
		maxGasLimitPerMetaBlock: maxGasLimitPerMetaBlock,
		gasPerDataByte:          gasPerDataByte,
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

	if economics.FeeSettings.GasPriceModifier > 1.0 || economics.FeeSettings.GasPriceModifier < epsilon {
		return process.ErrInvalidGasModifier
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
func (ed *economicsData) LeaderPercentage() float64 {
	return ed.leaderPercentage
}

// MinInflationRate will return the minimum inflation rate
func (ed *economicsData) MinInflationRate() float64 {
	return ed.minInflation
}

// MaxInflationRate will return the maximum inflation rate
func (ed *economicsData) MaxInflationRate(year uint32) float64 {
	ed.mutYearSettings.RLock()
	yearSetting, ok := ed.yearSettings[year]
	ed.mutYearSettings.RUnlock()

	if !ok {
		return ed.minInflation
	}

	return yearSetting.MaximumInflation
}

// GenesisTotalSupply will return the genesis total supply
func (ed *economicsData) GenesisTotalSupply() *big.Int {
	return ed.genesisTotalSupply
}

// MinGasPrice will return min gas price
func (ed *economicsData) MinGasPrice() uint64 {
	return ed.minGasPrice
}

// GasPriceModifier will return the gas price modifier
func (ed *economicsData) GasPriceModifier() float64 {
	if !ed.flagGasPriceModifier.IsSet() {
		return 1.0
	}
	return ed.gasPriceModifier
}

// MinGasLimit will return min gas limit
func (ed *economicsData) MinGasLimit() uint64 {
	return ed.minGasLimit
}

// GasPerDataByte will return the gas required for a economicsData byte
func (ed *economicsData) GasPerDataByte() uint64 {
	return ed.gasPerDataByte
}

// ComputeMoveBalanceFee computes the provided transaction's fee
func (ed *economicsData) ComputeMoveBalanceFee(tx process.TransactionWithFeeHandler) *big.Int {
	return core.SafeMul(tx.GetGasPrice(), ed.ComputeGasLimit(tx))
}

// ComputeFeeForProcessing will compute the fee using the gas price modifier, the gas to use and the actual gas price
func (ed *economicsData) ComputeFeeForProcessing(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	if !ed.flagGasPriceModifier.IsSet() {
		return core.SafeMul(tx.GetGasPrice(), gasToUse)
	}

	modifiedGasPrice := uint64(float64(tx.GetGasPrice()) * ed.gasPriceModifier)
	return core.SafeMul(modifiedGasPrice, gasToUse)
}

// ComputeTxFee computes the provided transaction's fee using enable from epoch approach
func (ed *economicsData) ComputeTxFee(tx process.TransactionWithFeeHandler) *big.Int {
	if ed.flagGasPriceModifier.IsSet() {
		gasLimitForMoveBalance := ed.ComputeGasLimit(tx)
		moveBalanceFee := core.SafeMul(tx.GetGasPrice(), gasLimitForMoveBalance)
		if tx.GetGasLimit() <= gasLimitForMoveBalance {
			return moveBalanceFee
		}

		difference := tx.GetGasLimit() - gasLimitForMoveBalance
		extraFee := ed.ComputeFeeForProcessing(tx, difference)
		moveBalanceFee.Add(moveBalanceFee, extraFee)
		return moveBalanceFee
	}

	if ed.flagPenalizedTooMuchGas.IsSet() {
		return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	return ed.ComputeMoveBalanceFee(tx)
}

// CheckValidityTxValues checks if the provided transaction is economically correct
func (ed *economicsData) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	if ed.minGasPrice > tx.GetGasPrice() {
		return process.ErrInsufficientGasPriceInTx
	}

	requiredGasLimit := ed.ComputeGasLimit(tx)
	if tx.GetGasLimit() < requiredGasLimit {
		return process.ErrInsufficientGasLimitInTx
	}

	if tx.GetGasLimit() >= ed.maxGasLimitPerBlock {
		return process.ErrMoreGasThanGasLimitPerBlock
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
func (ed *economicsData) MaxGasLimitPerBlock(shardID uint32) uint64 {
	if shardID == core.MetachainShardId {
		return ed.maxGasLimitPerMetaBlock
	}
	return ed.maxGasLimitPerBlock
}

// DeveloperPercentage will return the developer percentage value
func (ed *economicsData) DeveloperPercentage() float64 {
	return ed.developerPercentage
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (ed *economicsData) ProtocolSustainabilityPercentage() float64 {
	return ed.protocolSustainabilityPercentage
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (ed *economicsData) ProtocolSustainabilityAddress() string {
	return ed.protocolSustainabilityAddress
}

// ComputeGasLimit returns the gas limit need by the provided transaction in order to be executed
func (ed *economicsData) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	gasLimit := ed.minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * ed.gasPerDataByte

	return gasLimit
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (ed *economicsData) EpochConfirmed(epoch uint32) {
	ed.flagPenalizedTooMuchGas.Toggle(epoch >= ed.penalizedTooMuchGasEnableEpoch)
	log.Debug("economics: penalized too much gas", "enabled", ed.flagPenalizedTooMuchGas.IsSet())

	ed.flagGasPriceModifier.Toggle(epoch >= ed.gasPriceModifierEnableEpoch)
	log.Debug("economics: gas price modifier", "enabled", ed.flagGasPriceModifier.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *economicsData) IsInterfaceNil() bool {
	return ed == nil
}
