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
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
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
	topUpGradientPoint               *big.Int
	topUpFactor                      float64
}

// ArgsNewEconomicsData defines the arguments needed for new economics economicsData
type ArgsNewEconomicsData struct {
	Economics                      *config.EconomicsConfig
	EpochNotifier                  process.EpochNotifier
	PenalizedTooMuchGasEnableEpoch uint32
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

	topUpGradientPoint, ok := big.NewInt(0).SetString(args.Economics.RewardsSettings.TopUpGradientPoint, 10)
	if !ok {
		return nil, process.ErrInvalidRewardsTopUpGradientPoint
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

// MinGasPriceForProcessing returns the minimum allowed gas price for processing
func (ed *economicsData) MinGasPriceForProcessing() uint64 {
	priceModifier := ed.GasPriceModifier()

	return uint64(float64(ed.minGasPrice) * priceModifier)
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
	if isSmartContractResult(tx) {
		return big.NewInt(0)
	}

	return core.SafeMul(ed.GasPriceForMove(tx), ed.ComputeGasLimit(tx))
}

// ComputeFeeForProcessing will compute the fee using the gas price modifier, the gas to use and the actual gas price
func (ed *economicsData) ComputeFeeForProcessing(tx process.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	gasPrice := ed.GasPriceForProcessing(tx)
	return core.SafeMul(gasPrice, gasToUse)
}

// GasPriceForProcessing computes the price for the gas in addition to balance movement and data
func (ed *economicsData) GasPriceForProcessing(tx process.TransactionWithFeeHandler) uint64 {
	return uint64(float64(tx.GetGasPrice()) * ed.GasPriceModifier())
}

// GasPriceForMove returns the gas price for transferring funds
func (ed *economicsData) GasPriceForMove(tx process.TransactionWithFeeHandler) uint64 {
	return tx.GetGasPrice()
}

func isSmartContractResult(tx process.TransactionWithFeeHandler) bool {
	_, isSCR := tx.(*smartContractResult.SmartContractResult)
	return isSCR
}

// ComputeTxFee computes the provided transaction's fee using enable from epoch approach
func (ed *economicsData) ComputeTxFee(tx process.TransactionWithFeeHandler) *big.Int {
	if ed.flagGasPriceModifier.IsSet() {
		if isSmartContractResult(tx) {
			return ed.ComputeFeeForProcessing(tx, tx.GetGasLimit())
		}

		gasLimitForMoveBalance, difference := ed.SplitTxGasInCategories(tx)
		moveBalanceFee := core.SafeMul(ed.GasPriceForMove(tx), gasLimitForMoveBalance)
		if tx.GetGasLimit() <= gasLimitForMoveBalance {
			return moveBalanceFee
		}

		extraFee := ed.ComputeFeeForProcessing(tx, difference)
		moveBalanceFee.Add(moveBalanceFee, extraFee)
		return moveBalanceFee
	}

	if ed.flagPenalizedTooMuchGas.IsSet() {
		return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	return ed.ComputeMoveBalanceFee(tx)
}

// SplitTxGasInCategories returns the gas split per categories
func (ed *economicsData) SplitTxGasInCategories(tx process.TransactionWithFeeHandler) (gasLimitMove, gasLimitProcess uint64) {
	var err error
	gasLimitMove = ed.ComputeGasLimit(tx)
	gasLimitProcess, err = core.SafeSubUint64(tx.GetGasLimit(), gasLimitMove)
	if err != nil {
		log.Warn("SplitTxGasInCategories - insufficient gas for move",
			"providedGas", tx.GetGasLimit(),
			"computedMinimumRequired", gasLimitMove,
			"dataLen", len(tx.GetData()),
		)
	}

	return
}

// CheckValidityTxValues checks if the provided transaction is economically correct
func (ed *economicsData) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	if ed.minGasPrice > tx.GetGasPrice() {
		return process.ErrInsufficientGasPriceInTx
	}

	if !isSmartContractResult(tx) {
		requiredGasLimit := ed.ComputeGasLimit(tx)
		if tx.GetGasLimit() < requiredGasLimit {
			return process.ErrInsufficientGasLimitInTx
		}
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

// RewardsTopUpGradientPoint returns the rewards top-up gradient point
func (ed *economicsData) RewardsTopUpGradientPoint() *big.Int {
	return big.NewInt(0).Set(ed.topUpGradientPoint)
}

// RewardsTopUpFactor returns the rewards top-up factor
func (ed *economicsData) RewardsTopUpFactor() float64 {
	return ed.topUpFactor
}

// ComputeGasLimit returns the gas limit need by the provided transaction in order to be executed
func (ed *economicsData) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	gasLimit := ed.minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * ed.gasPerDataByte

	return gasLimit
}

// ComputeGasUsedAndFeeBasedOnRefundValue will compute gas used value and transaction fee using refund value from a SCR
func (ed *economicsData) ComputeGasUsedAndFeeBasedOnRefundValue(tx process.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if refundValue.Cmp(big.NewInt(0)) == 0 {
		txFee := ed.ComputeTxFee(tx)
		return tx.GetGasLimit(), txFee
	}

	gasRefund := ed.computeGasRefund(refundValue, tx.GetGasPrice())
	gasUsed, err := core.SafeSubUint64(tx.GetGasLimit(), gasRefund)
	if err != nil {
		log.Warn("ComputeGasUsedAndFeeBasedOnRefundValue", "SafeSubUint64", err.Error())
	}

	txFee := big.NewInt(0).Sub(ed.ComputeTxFee(tx), refundValue)

	return gasUsed, txFee
}

// ComputeTxFeeBasedOnGasUsed will compute transaction fee
func (ed *economicsData) ComputeTxFeeBasedOnGasUsed(tx process.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	moveBalanceGasLimit := ed.ComputeGasLimit(tx)
	moveBalanceFee := ed.ComputeMoveBalanceFee(tx)
	if gasUsed <= moveBalanceGasLimit {
		return moveBalanceFee
	}

	computeFeeForProcessing := ed.ComputeFeeForProcessing(tx, gasUsed-moveBalanceGasLimit)
	txFee := big.NewInt(0).Add(moveBalanceFee, computeFeeForProcessing)

	return txFee
}

func (ed *economicsData) computeGasRefund(refundValue *big.Int, gasPrice uint64) uint64 {
	gasPriceBig := big.NewInt(0).SetUint64(gasPrice)
	gasRefund := big.NewInt(0).Div(refundValue, gasPriceBig).Uint64()
	gasRefund = uint64(ed.GasPriceModifier() * float64(gasRefund))

	return gasRefund
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
