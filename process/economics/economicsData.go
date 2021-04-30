package economics

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

var _ process.EconomicsDataHandler = (*economicsData)(nil)
var _ process.RewardsHandler = (*economicsData)(nil)
var _ process.FeeHandler = (*economicsData)(nil)

var epsilon = 0.00000001
var log = logger.GetOrCreate("process/economics")

// economicsData will store information about economics
type economicsData struct {
	rewardsSettings                  []config.EpochRewardSettings
	rewardsSettingEpoch              uint32
	leaderPercentage                 float64
	protocolSustainabilityPercentage float64
	protocolSustainabilityAddress    string
	developerPercentage              float64
	topUpGradientPoint               *big.Int
	topUpFactor                      float64
	mutRewardsSettings               sync.RWMutex
	maxGasLimitPerBlock              uint64
	maxGasLimitPerMetaBlock          uint64
	gasPerDataByte                   uint64
	minGasPrice                      uint64
	gasPriceModifier                 float64
	minGasLimit                      uint64
	genesisTotalSupply               *big.Int
	minInflation                     float64
	yearSettings                     map[uint32]*config.YearSetting
	mutYearSettings                  sync.RWMutex
	flagPenalizedTooMuchGas          atomic.Flag
	flagGasPriceModifier             atomic.Flag
	penalizedTooMuchGasEnableEpoch   uint32
	gasPriceModifierEnableEpoch      uint32
	statusHandler                    core.AppStatusHandler
	builtInFunctionsCostHandler      BuiltInFunctionsCostHandler
}

// ArgsNewEconomicsData defines the arguments needed for new economics economicsData
type ArgsNewEconomicsData struct {
	BuiltInFunctionsCostHandler    BuiltInFunctionsCostHandler
	Economics                      *config.EconomicsConfig
	EpochNotifier                  process.EpochNotifier
	PenalizedTooMuchGasEnableEpoch uint32
	GasPriceModifierEnableEpoch    uint32
}

// NewEconomicsData will create and object with information about economics parameters
func NewEconomicsData(args ArgsNewEconomicsData) (*economicsData, error) {
	if check.IfNil(args.BuiltInFunctionsCostHandler) {
		return nil, process.ErrNilBuiltInFunctionsCostHandler
	}

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

	rewardsConfigs := make([]config.EpochRewardSettings, len(args.Economics.RewardsSettings.RewardsConfigByEpoch))
	_ = copy(rewardsConfigs, args.Economics.RewardsSettings.RewardsConfigByEpoch)

	sort.Slice(rewardsConfigs, func(i, j int) bool {
		return rewardsConfigs[i].EpochEnable < rewardsConfigs[j].EpochEnable
	})

	// validity checked in checkValues above
	topUpGradientPoint, _ := big.NewInt(0).SetString(rewardsConfigs[0].TopUpGradientPoint, 10)

	ed := &economicsData{
		rewardsSettings:                  rewardsConfigs,
		rewardsSettingEpoch:              rewardsConfigs[0].EpochEnable,
		leaderPercentage:                 rewardsConfigs[0].LeaderPercentage,
		protocolSustainabilityPercentage: rewardsConfigs[0].ProtocolSustainabilityPercentage,
		protocolSustainabilityAddress:    rewardsConfigs[0].ProtocolSustainabilityAddress,
		developerPercentage:              rewardsConfigs[0].DeveloperPercentage,
		topUpFactor:                      rewardsConfigs[0].TopUpFactor,
		topUpGradientPoint:               topUpGradientPoint,
		maxGasLimitPerBlock:              convertedData.maxGasLimitPerBlock,
		maxGasLimitPerMetaBlock:          convertedData.maxGasLimitPerMetaBlock,
		minGasPrice:                      convertedData.minGasPrice,
		minGasLimit:                      convertedData.minGasLimit,
		gasPerDataByte:                   convertedData.gasPerDataByte,
		minInflation:                     args.Economics.GlobalSettings.MinimumInflation,
		genesisTotalSupply:               convertedData.genesisTotalSupply,
		penalizedTooMuchGasEnableEpoch:   args.PenalizedTooMuchGasEnableEpoch,
		gasPriceModifierEnableEpoch:      args.GasPriceModifierEnableEpoch,
		gasPriceModifier:                 args.Economics.FeeSettings.GasPriceModifier,
		statusHandler:                    statusHandler.NewNilStatusHandler(),
		builtInFunctionsCostHandler:      args.BuiltInFunctionsCostHandler,
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
	if isPercentageInvalid(economics.GlobalSettings.MinimumInflation) {
		return process.ErrInvalidRewardsPercentages
	}

	for _, rewardsConfig := range economics.RewardsSettings.RewardsConfigByEpoch {
		if isPercentageInvalid(rewardsConfig.LeaderPercentage) ||
			isPercentageInvalid(rewardsConfig.DeveloperPercentage) ||
			isPercentageInvalid(rewardsConfig.ProtocolSustainabilityPercentage) ||
			isPercentageInvalid(rewardsConfig.TopUpFactor) {
			return process.ErrInvalidRewardsPercentages
		}

		if len(rewardsConfig.ProtocolSustainabilityAddress) == 0 {
			return process.ErrNilProtocolSustainabilityAddress
		}

		_, ok := big.NewInt(0).SetString(rewardsConfig.TopUpGradientPoint, 10)
		if !ok {
			return process.ErrInvalidRewardsTopUpGradientPoint
		}
	}

	for _, yearSetting := range economics.GlobalSettings.YearSettings {
		if isPercentageInvalid(yearSetting.MaximumInflation) {
			return process.ErrInvalidInflationPercentages
		}
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

// SetStatusHandler will set the provided status handler if not nil
func (ed *economicsData) SetStatusHandler(statusHandler core.AppStatusHandler) error {
	if check.IfNil(statusHandler) {
		return core.ErrNilAppStatusHandler
	}

	ed.statusHandler = statusHandler

	return nil
}

// LeaderPercentage will return leader reward percentage
func (ed *economicsData) LeaderPercentage() float64 {
	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

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
	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

	if shardID == core.MetachainShardId {
		return ed.maxGasLimitPerMetaBlock
	}
	return ed.maxGasLimitPerBlock
}

// DeveloperPercentage will return the developer percentage value
func (ed *economicsData) DeveloperPercentage() float64 {
	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

	return ed.developerPercentage
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (ed *economicsData) ProtocolSustainabilityPercentage() float64 {
	return ed.protocolSustainabilityPercentage
}

// ProtocolSustainabilityAddress will return the protocol sustainability address
func (ed *economicsData) ProtocolSustainabilityAddress() string {
	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

	return ed.protocolSustainabilityAddress
}

// RewardsTopUpGradientPoint returns the rewards top-up gradient point
func (ed *economicsData) RewardsTopUpGradientPoint() *big.Int {
	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

	return big.NewInt(0).Set(ed.topUpGradientPoint)
}

// RewardsTopUpFactor returns the rewards top-up factor
func (ed *economicsData) RewardsTopUpFactor() float64 {
	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

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
		if ed.builtInFunctionsCostHandler.IsBuiltInFuncCall(tx) {
			cost := ed.builtInFunctionsCostHandler.ComputeBuiltInCost(tx)
			gasLimit := ed.ComputeGasLimit(tx)

			gasLimitWithBuiltInCost := cost + gasLimit
			txFee := ed.ComputeTxFeeBasedOnGasUsed(tx, gasLimitWithBuiltInCost)

			// transaction will consume all the gas if sender provided too much gas
			if isTooMuchGasProvided(tx.GetGasLimit(), tx.GetGasLimit()-gasLimitWithBuiltInCost) {
				return tx.GetGasLimit(), ed.ComputeTxFee(tx)
			}

			return gasLimitWithBuiltInCost, txFee
		}

		txFee := ed.ComputeTxFee(tx)
		return tx.GetGasLimit(), txFee
	}

	txFee := big.NewInt(0).Sub(ed.ComputeTxFee(tx), refundValue)

	moveBalanceGasUnits := ed.ComputeGasLimit(tx)
	moveBalanceFee := ed.ComputeMoveBalanceFee(tx)

	scOpFee := big.NewInt(0).Sub(txFee, moveBalanceFee)
	gasPriceForProcessing := big.NewInt(0).SetUint64(ed.GasPriceForProcessing(tx))
	scOpGasUnits := big.NewInt(0).Div(scOpFee, gasPriceForProcessing)

	gasUsed := moveBalanceGasUnits + scOpGasUnits.Uint64()

	return gasUsed, txFee
}

func isTooMuchGasProvided(gasProvided uint64, gasRemained uint64) bool {
	if gasProvided <= gasRemained {
		return false
	}

	gasUsed := gasProvided - gasRemained
	return gasProvided > gasUsed*process.MaxGasFeeHigherFactorAccepted
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

// EpochConfirmed is called whenever a new epoch is confirmed
func (ed *economicsData) EpochConfirmed(epoch uint32) {
	ed.flagPenalizedTooMuchGas.Toggle(epoch >= ed.penalizedTooMuchGasEnableEpoch)
	log.Debug("economics: penalized too much gas", "enabled", ed.flagPenalizedTooMuchGas.IsSet())

	ed.flagGasPriceModifier.Toggle(epoch >= ed.gasPriceModifierEnableEpoch)
	log.Debug("economics: gas price modifier", "enabled", ed.flagGasPriceModifier.IsSet())
	ed.statusHandler.SetStringValue(core.MetricGasPriceModifier, fmt.Sprintf("%g", ed.GasPriceModifier()))
	ed.setRewardsEpochConfig(epoch)
}

func (ed *economicsData) setRewardsEpochConfig(currentEpoch uint32) {
	rewardSetting := ed.rewardsSettings[0]
	for i, setting := range ed.rewardsSettings {
		// as we go from epoch k to epoch k+1 we set the config for epoch k before computing the economics/rewards
		if currentEpoch > setting.EpochEnable {
			rewardSetting = ed.rewardsSettings[i]
		}
	}

	ed.mutRewardsSettings.Lock()
	defer ed.mutRewardsSettings.Unlock()

	if ed.rewardsSettingEpoch == rewardSetting.EpochEnable {
		log.Debug("economics: RewardsConfig",
			"epoch", ed.rewardsSettingEpoch,
			"leaderPercentage", ed.leaderPercentage,
			"protocolSustainabilityPercentage", ed.protocolSustainabilityPercentage,
			"protocolSustainabilityAddress", ed.protocolSustainabilityAddress,
			"developerPercentage", ed.developerPercentage,
			"topUpFactor", ed.topUpFactor,
			"topUpGradientPoint", ed.topUpGradientPoint,
		)
		return
	}

	ed.rewardsSettingEpoch = rewardSetting.EpochEnable
	ed.leaderPercentage = rewardSetting.LeaderPercentage
	ed.protocolSustainabilityPercentage = rewardSetting.ProtocolSustainabilityPercentage
	ed.protocolSustainabilityAddress = rewardSetting.ProtocolSustainabilityAddress
	ed.developerPercentage = rewardSetting.DeveloperPercentage
	ed.topUpFactor = rewardSetting.TopUpFactor
	// config was checked before for validity
	ed.topUpGradientPoint, _ = big.NewInt(0).SetString(rewardSetting.TopUpGradientPoint, 10)

	// TODO: add all metrics
	ed.statusHandler.SetStringValue(core.MetricLeaderPercentage, fmt.Sprintf("%f", rewardSetting.LeaderPercentage))
	ed.statusHandler.SetStringValue(core.MetricRewardsTopUpGradientPoint, rewardSetting.TopUpGradientPoint)
	ed.statusHandler.SetStringValue(core.MetricTopUpFactor, fmt.Sprintf("%f", rewardSetting.TopUpFactor))

	log.Debug("economics: RewardsConfig",
		"epoch", ed.rewardsSettingEpoch,
		"leaderPercentage", ed.leaderPercentage,
		"protocolSustainabilityPercentage", ed.protocolSustainabilityPercentage,
		"protocolSustainabilityAddress", ed.protocolSustainabilityAddress,
		"developerPercentage", ed.developerPercentage,
		"topUpFactor", ed.topUpFactor,
		"topUpGradientPoint", ed.topUpGradientPoint,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *economicsData) IsInterfaceNil() bool {
	return ed == nil
}
