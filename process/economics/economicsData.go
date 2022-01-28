package economics

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
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
	gasLimitSettings                 []config.GasLimitSetting
	gasLimitSettingEpoch             uint32
	maxGasLimitPerBlock              uint64
	maxGasLimitPerMiniBlock          uint64
	maxGasLimitPerMetaBlock          uint64
	maxGasLimitPerMetaMiniBlock      uint64
	maxGasLimitPerTx                 uint64
	mutGasLimitSettings              sync.RWMutex
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

// NewEconomicsData will create an object with information about economics parameters
func NewEconomicsData(args ArgsNewEconomicsData) (*economicsData, error) {
	if check.IfNil(args.BuiltInFunctionsCostHandler) {
		return nil, process.ErrNilBuiltInFunctionsCostHandler
	}

	err := checkValues(args.Economics)
	if err != nil {
		return nil, err
	}

	convertedData, err := convertValues(args.Economics)
	if err != nil {
		return nil, err
	}

	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}

	rewardsConfigs := make([]config.EpochRewardSettings, len(args.Economics.RewardsSettings.RewardsConfigByEpoch))
	_ = copy(rewardsConfigs, args.Economics.RewardsSettings.RewardsConfigByEpoch)

	sort.Slice(rewardsConfigs, func(i, j int) bool {
		return rewardsConfigs[i].EpochEnable < rewardsConfigs[j].EpochEnable
	})

	gasLimitSettings := make([]config.GasLimitSetting, len(args.Economics.FeeSettings.GasLimitSettings))
	_ = copy(gasLimitSettings, args.Economics.FeeSettings.GasLimitSettings)

	sort.Slice(gasLimitSettings, func(i, j int) bool {
		return gasLimitSettings[i].EnableEpoch < gasLimitSettings[j].EnableEpoch
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
		gasLimitSettings:                 gasLimitSettings,
		minGasPrice:                      convertedData.minGasPrice,
		gasPerDataByte:                   convertedData.gasPerDataByte,
		minInflation:                     args.Economics.GlobalSettings.MinimumInflation,
		genesisTotalSupply:               convertedData.genesisTotalSupply,
		penalizedTooMuchGasEnableEpoch:   args.PenalizedTooMuchGasEnableEpoch,
		gasPriceModifierEnableEpoch:      args.GasPriceModifierEnableEpoch,
		gasPriceModifier:                 args.Economics.FeeSettings.GasPriceModifier,
		statusHandler:                    statusHandler.NewNilStatusHandler(),
		builtInFunctionsCostHandler:      args.BuiltInFunctionsCostHandler,
	}
	log.Debug("economicsData: enable epoch for penalized too much gas", "epoch", ed.penalizedTooMuchGasEnableEpoch)
	log.Debug("economicsData: enable epoch for gas price modifier", "epoch", ed.gasPriceModifierEnableEpoch)

	ed.yearSettings = make(map[uint32]*config.YearSetting)
	for _, yearSetting := range args.Economics.GlobalSettings.YearSettings {
		ed.yearSettings[yearSetting.Year] = &config.YearSetting{
			Year:             yearSetting.Year,
			MaximumInflation: yearSetting.MaximumInflation,
		}
	}

	err = ed.setGasLimitSetting(gasLimitSettings[0])
	if err != nil {
		return nil, err
	}

	args.EpochNotifier.RegisterNotifyHandler(ed)

	return ed, nil
}

func (ed *economicsData) setGasLimitSetting(gasLimitSetting config.GasLimitSetting) error {
	var err error
	conversionBase := 10
	bitConversionSize := 64

	ed.gasLimitSettingEpoch = gasLimitSetting.EnableEpoch

	ed.maxGasLimitPerBlock, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerBlock, conversionBase, bitConversionSize)
	if err != nil {
		return process.ErrInvalidMaxGasLimitPerBlock
	}

	ed.maxGasLimitPerMiniBlock, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerMiniBlock, conversionBase, bitConversionSize)
	if err != nil {
		return process.ErrInvalidMaxGasLimitPerMiniBlock
	}

	ed.maxGasLimitPerMetaBlock, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerMetaBlock, conversionBase, bitConversionSize)
	if err != nil {
		return process.ErrInvalidMaxGasLimitPerMetaBlock
	}

	ed.maxGasLimitPerMetaMiniBlock, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerMetaMiniBlock, conversionBase, bitConversionSize)
	if err != nil {
		return process.ErrInvalidMaxGasLimitPerMetaMiniBlock
	}

	ed.maxGasLimitPerTx, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerTx, conversionBase, bitConversionSize)
	if err != nil {
		return process.ErrInvalidMaxGasLimitPerTx
	}

	ed.minGasLimit, err = strconv.ParseUint(gasLimitSetting.MinGasLimit, conversionBase, bitConversionSize)
	if err != nil {
		return process.ErrInvalidMinimumGasLimitForTx
	}

	return nil
}

func convertValues(economics *config.EconomicsConfig) (*economicsData, error) {
	conversionBase := 10
	bitConversionSize := 64

	minGasPrice, err := strconv.ParseUint(economics.FeeSettings.MinGasPrice, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidMinimumGasPrice
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
		minGasPrice:        minGasPrice,
		gasPerDataByte:     gasPerDataByte,
		genesisTotalSupply: genesisTotalSupply,
	}, nil
}

func checkValues(economics *config.EconomicsConfig) error {
	if isPercentageInvalid(economics.GlobalSettings.MinimumInflation) {
		return process.ErrInvalidRewardsPercentages
	}

	if len(economics.RewardsSettings.RewardsConfigByEpoch) == 0 {
		return process.ErrEmptyEpochRewardsConfig
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

	if len(economics.GlobalSettings.YearSettings) == 0 {
		return process.ErrEmptyYearSettings
	}
	for _, yearSetting := range economics.GlobalSettings.YearSettings {
		if isPercentageInvalid(yearSetting.MaximumInflation) {
			return process.ErrInvalidInflationPercentages
		}
	}

	if economics.FeeSettings.GasPriceModifier > 1.0 || economics.FeeSettings.GasPriceModifier < epsilon {
		return process.ErrInvalidGasModifier
	}

	if len(economics.FeeSettings.GasLimitSettings) == 0 {
		return process.ErrEmptyGasLimitSettings
	}
	conversionBase := 10
	bitConversionSize := 64
	for _, gasLimitSetting := range economics.FeeSettings.GasLimitSettings {
		minGasLimit, err := strconv.ParseUint(gasLimitSetting.MinGasLimit, conversionBase, bitConversionSize)
		if err != nil {
			return process.ErrInvalidMinimumGasLimitForTx
		}

		maxGasLimitPerBlock, err := strconv.ParseUint(gasLimitSetting.MaxGasLimitPerBlock, conversionBase, bitConversionSize)
		if err != nil {
			return fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerBlock, gasLimitSetting.EnableEpoch)
		}

		maxGasLimitPerMiniBlock, err := strconv.ParseUint(gasLimitSetting.MaxGasLimitPerMiniBlock, conversionBase, bitConversionSize)
		if err != nil {
			return fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerMiniBlock, gasLimitSetting.EnableEpoch)
		}

		maxGasLimitPerMetaBlock, err := strconv.ParseUint(gasLimitSetting.MaxGasLimitPerMetaBlock, conversionBase, bitConversionSize)
		if err != nil {
			return fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerMetaBlock, gasLimitSetting.EnableEpoch)
		}

		maxGasLimitPerMetaMiniBlock, err := strconv.ParseUint(gasLimitSetting.MaxGasLimitPerMetaMiniBlock, conversionBase, bitConversionSize)
		if err != nil {
			return fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerMetaMiniBlock, gasLimitSetting.EnableEpoch)
		}

		maxGasLimitPerTx, err := strconv.ParseUint(gasLimitSetting.MaxGasLimitPerTx, conversionBase, bitConversionSize)
		if err != nil {
			return fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerTx, gasLimitSetting.EnableEpoch)
		}

		if maxGasLimitPerBlock < minGasLimit {
			return fmt.Errorf("%w: maxGasLimitPerBlock = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerBlock, maxGasLimitPerBlock, minGasLimit, gasLimitSetting.EnableEpoch)
		}
		if maxGasLimitPerMiniBlock < minGasLimit {
			return fmt.Errorf("%w: maxGasLimitPerMiniBlock = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerMiniBlock, maxGasLimitPerMiniBlock, minGasLimit, gasLimitSetting.EnableEpoch)
		}
		if maxGasLimitPerMetaBlock < minGasLimit {
			return fmt.Errorf("%w: maxGasLimitPerMetaBlock = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerMetaBlock, maxGasLimitPerMetaBlock, minGasLimit, gasLimitSetting.EnableEpoch)
		}
		if maxGasLimitPerMetaMiniBlock < minGasLimit {
			return fmt.Errorf("%w: maxGasLimitPerMetaMiniBlock = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerMetaMiniBlock, maxGasLimitPerMetaMiniBlock, minGasLimit, gasLimitSetting.EnableEpoch)
		}
		if maxGasLimitPerTx < minGasLimit {
			return fmt.Errorf("%w: maxGasLimitPerTx = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerTx, maxGasLimitPerTx, minGasLimit, gasLimitSetting.EnableEpoch)
		}
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
func (ed *economicsData) ComputeMoveBalanceFee(tx data.TransactionWithFeeHandler) *big.Int {
	if isSmartContractResult(tx) {
		return big.NewInt(0)
	}

	return core.SafeMul(ed.GasPriceForMove(tx), ed.ComputeGasLimit(tx))
}

// ComputeFeeForProcessing will compute the fee using the gas price modifier, the gas to use and the actual gas price
func (ed *economicsData) ComputeFeeForProcessing(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	gasPrice := ed.GasPriceForProcessing(tx)
	return core.SafeMul(gasPrice, gasToUse)
}

// GasPriceForProcessing computes the price for the gas in addition to balance movement and data
func (ed *economicsData) GasPriceForProcessing(tx data.TransactionWithFeeHandler) uint64 {
	return uint64(float64(tx.GetGasPrice()) * ed.GasPriceModifier())
}

// GasPriceForMove returns the gas price for transferring funds
func (ed *economicsData) GasPriceForMove(tx data.TransactionWithFeeHandler) uint64 {
	return tx.GetGasPrice()
}

func isSmartContractResult(tx data.TransactionWithFeeHandler) bool {
	_, isSCR := tx.(*smartContractResult.SmartContractResult)
	return isSCR
}

// ComputeTxFee computes the provided transaction's fee using enable from epoch approach
func (ed *economicsData) ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int {
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
func (ed *economicsData) SplitTxGasInCategories(tx data.TransactionWithFeeHandler) (gasLimitMove, gasLimitProcess uint64) {
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
func (ed *economicsData) CheckValidityTxValues(tx data.TransactionWithFeeHandler) error {
	if ed.minGasPrice > tx.GetGasPrice() {
		return process.ErrInsufficientGasPriceInTx
	}

	if !isSmartContractResult(tx) {
		requiredGasLimit := ed.ComputeGasLimit(tx)
		if tx.GetGasLimit() < requiredGasLimit {
			return process.ErrInsufficientGasLimitInTx
		}
	}

	//The following check should be kept as it is in order to avoid backwards compatibility issues
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
	ed.mutGasLimitSettings.RLock()
	defer ed.mutGasLimitSettings.RUnlock()

	if shardID == core.MetachainShardId {
		return ed.maxGasLimitPerMetaBlock
	}
	return ed.maxGasLimitPerBlock
}

// MaxGasLimitPerMiniBlock will return maximum gas limit allowed per mini block
func (ed *economicsData) MaxGasLimitPerMiniBlock(shardID uint32) uint64 {
	ed.mutGasLimitSettings.RLock()
	defer ed.mutGasLimitSettings.RUnlock()

	if shardID == core.MetachainShardId {
		return ed.maxGasLimitPerMetaMiniBlock
	}
	return ed.maxGasLimitPerMiniBlock
}

// MaxGasLimitPerBlockForSafeCrossShard will return maximum gas limit per block for safe cross shard
func (ed *economicsData) MaxGasLimitPerBlockForSafeCrossShard() uint64 {
	ed.mutGasLimitSettings.RLock()
	defer ed.mutGasLimitSettings.RUnlock()

	return core.MinUint64(ed.maxGasLimitPerBlock, ed.maxGasLimitPerMetaBlock)
}

// MaxGasLimitPerMiniBlockForSafeCrossShard will return maximum gas limit per mini block for safe cross shard
func (ed *economicsData) MaxGasLimitPerMiniBlockForSafeCrossShard() uint64 {
	ed.mutGasLimitSettings.RLock()
	defer ed.mutGasLimitSettings.RUnlock()

	return core.MinUint64(ed.maxGasLimitPerMiniBlock, ed.maxGasLimitPerMetaMiniBlock)
}

// MaxGasLimitPerTx will return maximum gas limit per tx
func (ed *economicsData) MaxGasLimitPerTx() uint64 {
	ed.mutGasLimitSettings.RLock()
	defer ed.mutGasLimitSettings.RUnlock()

	return ed.maxGasLimitPerTx
}

// DeveloperPercentage will return the developer percentage value
func (ed *economicsData) DeveloperPercentage() float64 {
	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

	return ed.developerPercentage
}

// ProtocolSustainabilityPercentage will return the protocol sustainability percentage value
func (ed *economicsData) ProtocolSustainabilityPercentage() float64 {
	ed.mutRewardsSettings.RLock()
	defer ed.mutRewardsSettings.RUnlock()

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
func (ed *economicsData) ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64 {
	gasLimit := ed.minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * ed.gasPerDataByte

	return gasLimit
}

// ComputeGasUsedAndFeeBasedOnRefundValue will compute gas used value and transaction fee using refund value from a SCR
func (ed *economicsData) ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if refundValue.Cmp(big.NewInt(0)) == 0 {
		if ed.builtInFunctionsCostHandler.IsBuiltInFuncCall(tx) {
			builtInCost := ed.builtInFunctionsCostHandler.ComputeBuiltInCost(tx)
			computedGasLimit := ed.ComputeGasLimit(tx)

			gasLimitWithBuiltInCost := builtInCost + computedGasLimit
			txFee := ed.ComputeTxFeeBasedOnGasUsed(tx, gasLimitWithBuiltInCost)

			gasLimitWithoutMoveBalance := tx.GetGasLimit() - computedGasLimit
			// transaction will consume all the gas if sender provided too much gas
			if isTooMuchGasProvided(gasLimitWithoutMoveBalance, gasLimitWithoutMoveBalance-builtInCost) {
				return tx.GetGasLimit(), ed.ComputeTxFee(tx)
			}

			return gasLimitWithBuiltInCost, txFee
		}

		txFee := ed.ComputeTxFee(tx)
		return tx.GetGasLimit(), txFee
	}

	txFee := ed.ComputeTxFee(tx)
	flagCorrectTxFee := !ed.flagPenalizedTooMuchGas.IsSet() && !ed.flagGasPriceModifier.IsSet()
	if flagCorrectTxFee {
		txFee = core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	txFee = big.NewInt(0).Sub(txFee, refundValue)

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
func (ed *economicsData) ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
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
func (ed *economicsData) EpochConfirmed(epoch uint32, _ uint64) {
	ed.flagPenalizedTooMuchGas.SetValue(epoch >= ed.penalizedTooMuchGasEnableEpoch)
	log.Debug("economics: penalized too much gas", "enabled", ed.flagPenalizedTooMuchGas.IsSet())

	ed.flagGasPriceModifier.SetValue(epoch >= ed.gasPriceModifierEnableEpoch)
	log.Debug("economics: gas price modifier", "enabled", ed.flagGasPriceModifier.IsSet())
	ed.statusHandler.SetStringValue(common.MetricGasPriceModifier, fmt.Sprintf("%g", ed.GasPriceModifier()))
	ed.setRewardsEpochConfig(epoch)
	ed.setGasLimitConfig(epoch)
}

func (ed *economicsData) setRewardsEpochConfig(currentEpoch uint32) {
	ed.mutRewardsSettings.Lock()
	defer ed.mutRewardsSettings.Unlock()

	rewardSetting := ed.rewardsSettings[0]
	for i, setting := range ed.rewardsSettings {
		// as we go from epoch k to epoch k+1 we set the config for epoch k before computing the economics/rewards
		if currentEpoch > setting.EpochEnable {
			rewardSetting = ed.rewardsSettings[i]
		}
	}

	if ed.rewardsSettingEpoch != rewardSetting.EpochEnable {
		ed.rewardsSettingEpoch = rewardSetting.EpochEnable
		ed.leaderPercentage = rewardSetting.LeaderPercentage
		ed.protocolSustainabilityPercentage = rewardSetting.ProtocolSustainabilityPercentage
		ed.protocolSustainabilityAddress = rewardSetting.ProtocolSustainabilityAddress
		ed.developerPercentage = rewardSetting.DeveloperPercentage
		ed.topUpFactor = rewardSetting.TopUpFactor
		// config was checked before for validity
		ed.topUpGradientPoint, _ = big.NewInt(0).SetString(rewardSetting.TopUpGradientPoint, 10)

		// TODO: add all metrics
		ed.statusHandler.SetStringValue(common.MetricLeaderPercentage, fmt.Sprintf("%f", rewardSetting.LeaderPercentage))
		ed.statusHandler.SetStringValue(common.MetricRewardsTopUpGradientPoint, rewardSetting.TopUpGradientPoint)
		ed.statusHandler.SetStringValue(common.MetricTopUpFactor, fmt.Sprintf("%f", rewardSetting.TopUpFactor))
	}

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

func (ed *economicsData) setGasLimitConfig(currentEpoch uint32) {
	ed.mutGasLimitSettings.Lock()
	defer ed.mutGasLimitSettings.Unlock()

	gasLimitSetting := ed.gasLimitSettings[0]
	for i := 1; i < len(ed.gasLimitSettings); i++ {
		if currentEpoch >= ed.gasLimitSettings[i].EnableEpoch {
			gasLimitSetting = ed.gasLimitSettings[i]
		}
	}

	if ed.gasLimitSettingEpoch != gasLimitSetting.EnableEpoch {
		err := ed.setGasLimitSetting(gasLimitSetting)
		if err != nil {
			log.Error("setGasLimitConfig", "error", err.Error())
		}
	}

	log.Debug("economics: GasLimitConfig",
		"epoch", ed.gasLimitSettingEpoch,
		"maxGasLimitPerBlock", ed.maxGasLimitPerBlock,
		"maxGasLimitPerMiniBlock", ed.maxGasLimitPerMiniBlock,
		"maxGasLimitPerMetaBlock", ed.maxGasLimitPerMetaBlock,
		"maxGasLimitPerMetaMiniBlock", ed.maxGasLimitPerMetaMiniBlock,
		"maxGasLimitPerTx", ed.maxGasLimitPerTx,
		"minGasLimit", ed.minGasLimit,
	)

	ed.statusHandler.SetUInt64Value(common.MetricMaxGasPerTransaction, ed.maxGasLimitPerTx)
}

// ComputeGasLimitBasedOnBalance will compute gas limit for the given transaction based on the balance
func (ed *economicsData) ComputeGasLimitBasedOnBalance(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error) {
	balanceWithoutTransferValue := big.NewInt(0).Sub(balance, tx.GetValue())
	if balanceWithoutTransferValue.Cmp(big.NewInt(0)) < 1 {
		return 0, process.ErrInsufficientFunds
	}

	moveBalanceFee := ed.ComputeMoveBalanceFee(tx)
	if moveBalanceFee.Cmp(balanceWithoutTransferValue) > 0 {
		return 0, process.ErrInsufficientFunds
	}

	if !ed.flagGasPriceModifier.IsSet() {
		gasPriceBig := big.NewInt(0).SetUint64(tx.GetGasPrice())
		gasLimitBig := big.NewInt(0).Div(balanceWithoutTransferValue, gasPriceBig)

		return gasLimitBig.Uint64(), nil
	}

	remainedBalanceAfterMoveBalanceFee := big.NewInt(0).Sub(balanceWithoutTransferValue, moveBalanceFee)
	gasPriceBigForProcessing := ed.GasPriceForProcessing(tx)
	gasPriceBigForProcessingBig := big.NewInt(0).SetUint64(gasPriceBigForProcessing)
	gasLimitFromRemainedBalanceBig := big.NewInt(0).Div(remainedBalanceAfterMoveBalanceFee, gasPriceBigForProcessingBig)

	gasLimitMoveBalance := ed.ComputeGasLimit(tx)
	totalGasLimit := gasLimitMoveBalance + gasLimitFromRemainedBalanceBig.Uint64()

	return totalGasLimit, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *economicsData) IsInterfaceNil() bool {
	return ed == nil
}
