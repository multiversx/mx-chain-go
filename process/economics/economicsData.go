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

var _ process.RewardsHandler = (*Data)(nil)
var _ process.FeeHandler = (*Data)(nil)

var epsilon float64 = 0.00000001
var log = logger.GetOrCreate("process/economics")

// Data will store information about economics
type Data struct {
	leaderPercentage                 float64
	protocolSustainabilityPercentage float64
	protocolSustainabilityAddress    string
	maxGasLimitPerBlock              uint64
	maxGasLimitPerMetaBlock          uint64
	gasPerDataByte                   uint64
	dataLimitForBaseCalc             uint64
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

// ArgsNewEconomicsData defines the arguments needed for new economics data
type ArgsNewEconomicsData struct {
	Economics                      *config.EconomicsConfig
	PenalizedTooMuchGasEnableEpoch uint32
	EpochNotifier                  process.EpochNotifier
	GasPriceModifierEnableEpoch    uint32
}

// NewEconomicsData will create and object with information about economics parameters
func NewEconomicsData(args ArgsNewEconomicsData) (*Data, error) {
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

	ed := &Data{
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

func convertValues(economics *config.EconomicsConfig) (*Data, error) {
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

	return &Data{
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
func (ed *Data) LeaderPercentage() float64 {
	return ed.leaderPercentage
}

// MinInflationRate will return the minimum inflation rate
func (ed *Data) MinInflationRate() float64 {
	return ed.minInflation
}

// MaxInflationRate will return the maximum inflation rate
func (ed *Data) MaxInflationRate(year uint32) float64 {
	ed.mutYearSettings.RLock()
	yearSetting, ok := ed.yearSettings[year]
	ed.mutYearSettings.RUnlock()

	if !ok {
		return ed.minInflation
	}

	return yearSetting.MaximumInflation
}

// GenesisTotalSupply will return the genesis total supply
func (ed *Data) GenesisTotalSupply() *big.Int {
	return ed.genesisTotalSupply
}

// MinGasPrice will return min gas price
func (ed *Data) MinGasPrice() uint64 {
	return ed.minGasPrice
}

// GasPriceModifier will return the gas price modifier
func (ed *Data) GasPriceModifier() float64 {
	return ed.gasPriceModifier
}

// MinGasLimit will return min gas limit
func (ed *Data) MinGasLimit() uint64 {
	return ed.minGasLimit
}

// GasPerDataByte will return the gas required for a data byte
func (ed *Data) GasPerDataByte() uint64 {
	return ed.gasPerDataByte
}

// ComputeMoveBalanceFee computes the provided transaction's fee
func (ed *Data) ComputeMoveBalanceFee(tx process.TransactionWithFeeHandler) *big.Int {
	return core.SafeMul(tx.GetGasPrice(), ed.ComputeGasLimit(tx))
}

// ComputeTxFee computes the provided transaction's fee using enable from epoch approach
func (ed *Data) ComputeTxFee(tx process.TransactionWithFeeHandler) *big.Int {
	if ed.flagPenalizedTooMuchGas.IsSet() {
		return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	return ed.ComputeMoveBalanceFee(tx)
}

// CheckValidityTxValues checks if the provided transaction is economically correct
func (ed *Data) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
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
func (ed *Data) MaxGasLimitPerBlock(shardID uint32) uint64 {
	if shardID == core.MetachainShardId {
		return ed.maxGasLimitPerMetaBlock
	}
	return ed.maxGasLimitPerBlock
}

// DeveloperPercentage will return the developer percentage value
func (ed *Data) DeveloperPercentage() float64 {
	return ed.developerPercentage
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (ed *Data) ProtocolSustainabilityPercentage() float64 {
	return ed.protocolSustainabilityPercentage
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (ed *Data) ProtocolSustainabilityAddress() string {
	return ed.protocolSustainabilityAddress
}

// ComputeGasLimit returns the gas limit need by the provided transaction in order to be executed
func (ed *Data) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	gasLimit := ed.minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * ed.gasPerDataByte

	return gasLimit
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (ed *Data) EpochConfirmed(epoch uint32) {
	ed.flagPenalizedTooMuchGas.Toggle(epoch >= ed.penalizedTooMuchGasEnableEpoch)
	log.Debug("economics: penalized too much gas", "enabled", ed.flagPenalizedTooMuchGas.IsSet())

	ed.flagGasPriceModifier.Toggle(epoch >= ed.gasPriceModifierEnableEpoch)
	log.Debug("economics: gas price modifier", "enabled", ed.flagPenalizedTooMuchGas.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *Data) IsInterfaceNil() bool {
	return ed == nil
}
