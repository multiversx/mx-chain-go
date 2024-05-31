package economics

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/statusHandler"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.EconomicsDataHandler = (*economicsData)(nil)
var _ process.RewardsHandler = (*economicsData)(nil)
var _ process.FeeHandler = (*economicsData)(nil)

var log = logger.GetOrCreate("process/economics")

// economicsData will store information about economics
type economicsData struct {
	*gasConfigHandler
	*rewardsConfigHandler
	gasPriceModifier    float64
	minInflation        float64
	yearSettings        map[uint32]*config.YearSetting
	mutYearSettings     sync.RWMutex
	statusHandler       core.AppStatusHandler
	enableEpochsHandler common.EnableEpochsHandler
	txVersionHandler    process.TxVersionCheckerHandler
	mut                 sync.RWMutex
}

// ArgsNewEconomicsData defines the arguments needed for new economics economicsData
type ArgsNewEconomicsData struct {
	TxVersionChecker    process.TxVersionCheckerHandler
	Economics           *config.EconomicsConfig
	EpochNotifier       process.EpochNotifier
	EnableEpochsHandler common.EnableEpochsHandler
}

// NewEconomicsData will create an object with information about economics parameters
func NewEconomicsData(args ArgsNewEconomicsData) (*economicsData, error) {
	if check.IfNil(args.TxVersionChecker) {
		return nil, process.ErrNilTransactionVersionChecker
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.GasPriceModifierFlag,
		common.PenalizedTooMuchGasFlag,
	})
	if err != nil {
		return nil, err
	}

	err = checkEconomicsConfig(args.Economics)
	if err != nil {
		return nil, err
	}

	ed := &economicsData{
		minInflation:        args.Economics.GlobalSettings.MinimumInflation,
		gasPriceModifier:    args.Economics.FeeSettings.GasPriceModifier,
		statusHandler:       statusHandler.NewNilStatusHandler(),
		enableEpochsHandler: args.EnableEpochsHandler,
		txVersionHandler:    args.TxVersionChecker,
	}

	ed.yearSettings = make(map[uint32]*config.YearSetting)
	for _, yearSetting := range args.Economics.GlobalSettings.YearSettings {
		ed.yearSettings[yearSetting.Year] = &config.YearSetting{
			Year:             yearSetting.Year,
			MaximumInflation: yearSetting.MaximumInflation,
		}
	}

	ed.gasConfigHandler, err = newGasConfigHandler(args.Economics)
	if err != nil {
		return nil, err
	}

	ed.rewardsConfigHandler, err = newRewardsConfigHandler(args.Economics.RewardsSettings)
	if err != nil {
		return nil, err
	}

	args.EpochNotifier.RegisterNotifyHandler(ed)

	return ed, nil
}

func checkEconomicsConfig(economics *config.EconomicsConfig) error {
	if isPercentageInvalid(economics.GlobalSettings.MinimumInflation) {
		return process.ErrInvalidInflationPercentages
	}

	if len(economics.RewardsSettings.RewardsConfigByEpoch) == 0 {
		return process.ErrEmptyEpochRewardsConfig
	}

	if len(economics.GlobalSettings.YearSettings) == 0 {
		return process.ErrEmptyYearSettings
	}
	for _, yearSetting := range economics.GlobalSettings.YearSettings {
		if isPercentageInvalid(yearSetting.MaximumInflation) {
			return process.ErrInvalidInflationPercentages
		}
	}

	return nil
}

// SetStatusHandler will set the provided status handler if not nil
func (ed *economicsData) SetStatusHandler(statusHandler core.AppStatusHandler) error {
	if check.IfNil(statusHandler) {
		return core.ErrNilAppStatusHandler
	}
	ed.mut.Lock()
	ed.statusHandler = statusHandler
	ed.mut.Unlock()

	err := ed.gasConfigHandler.setStatusHandler(statusHandler)
	if err != nil {
		return err
	}
	return ed.rewardsConfigHandler.setStatusHandler(statusHandler)
}

// LeaderPercentage returns leader reward percentage
func (ed *economicsData) LeaderPercentage() float64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.LeaderPercentageInEpoch(currentEpoch)
}

// LeaderPercentageInEpoch returns leader reward percentage in a specific epoch
func (ed *economicsData) LeaderPercentageInEpoch(epoch uint32) float64 {
	return ed.getLeaderPercentage(epoch)
}

// MinInflationRate returns the minimum inflation rate
func (ed *economicsData) MinInflationRate() float64 {
	return ed.minInflation
}

// MaxInflationRate returns the maximum inflation rate
func (ed *economicsData) MaxInflationRate(year uint32) float64 {
	ed.mutYearSettings.RLock()
	yearSetting, ok := ed.yearSettings[year]
	ed.mutYearSettings.RUnlock()

	if !ok {
		return ed.minInflation
	}

	return yearSetting.MaximumInflation
}

// GenesisTotalSupply returns the genesis total supply
func (ed *economicsData) GenesisTotalSupply() *big.Int {
	return ed.genesisTotalSupply
}

// MinGasPrice returns min gas price
func (ed *economicsData) MinGasPrice() uint64 {
	return ed.minGasPrice
}

// MinGasPriceForProcessing returns the minimum allowed gas price for processing
func (ed *economicsData) MinGasPriceForProcessing() uint64 {
	priceModifier := ed.GasPriceModifier()

	return uint64(float64(ed.minGasPrice) * priceModifier)
}

// GasPriceModifier returns the gas price modifier
func (ed *economicsData) GasPriceModifier() float64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.GasPriceModifierInEpoch(currentEpoch)
}

// GasPriceModifierInEpoch returns the gas price modifier in a specific epoch
func (ed *economicsData) GasPriceModifierInEpoch(epoch uint32) float64 {
	if !ed.enableEpochsHandler.IsFlagEnabledInEpoch(common.GasPriceModifierFlag, epoch) {
		return 1.0
	}
	return ed.gasPriceModifier
}

// MinGasLimit returns min gas limit
func (ed *economicsData) MinGasLimit() uint64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.MinGasLimitInEpoch(currentEpoch)
}

// MinGasLimitInEpoch returns min gas limit in a specific epoch
func (ed *economicsData) MinGasLimitInEpoch(epoch uint32) uint64 {
	return ed.getMinGasLimit(epoch)
}

// ExtraGasLimitGuardedTx returns the extra gas limit required by the guarded transactions
func (ed *economicsData) ExtraGasLimitGuardedTx() uint64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ExtraGasLimitGuardedTxInEpoch(currentEpoch)
}

// ExtraGasLimitGuardedTxInEpoch returns the extra gas limit required by the guarded transactions in a specific epoch
func (ed *economicsData) ExtraGasLimitGuardedTxInEpoch(epoch uint32) uint64 {
	return ed.getExtraGasLimitGuardedTx(epoch)
}

// MaxGasPriceSetGuardian returns the maximum gas price for set guardian transactions
func (ed *economicsData) MaxGasPriceSetGuardian() uint64 {
	return ed.maxGasPriceSetGuardian
}

// GasPerDataByte returns the gas required for a economicsData byte
func (ed *economicsData) GasPerDataByte() uint64 {
	return ed.gasPerDataByte
}

// ComputeMoveBalanceFee computes the provided transaction's fee
func (ed *economicsData) ComputeMoveBalanceFee(tx data.TransactionWithFeeHandler) *big.Int {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ComputeMoveBalanceFeeInEpoch(tx, currentEpoch)
}

// ComputeMoveBalanceFeeInEpoch computes the provided transaction's fee in a specific epoch
func (ed *economicsData) ComputeMoveBalanceFeeInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int {
	if isSmartContractResult(tx) {
		return big.NewInt(0)
	}

	return core.SafeMul(ed.GasPriceForMove(tx), ed.ComputeGasLimitInEpoch(tx, epoch))
}

// ComputeFeeForProcessing will compute the fee using the gas price modifier, the gas to use and the actual gas price
func (ed *economicsData) ComputeFeeForProcessing(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ComputeFeeForProcessingInEpoch(tx, gasToUse, currentEpoch)
}

// ComputeFeeForProcessingInEpoch will compute the fee using the gas price modifier, the gas to use and the actual gas price in a specific epoch
func (ed *economicsData) ComputeFeeForProcessingInEpoch(tx data.TransactionWithFeeHandler, gasToUse uint64, epoch uint32) *big.Int {
	gasPrice := ed.GasPriceForProcessingInEpoch(tx, epoch)
	return core.SafeMul(gasPrice, gasToUse)
}

// GasPriceForProcessing computes the price for the gas in addition to balance movement and data
func (ed *economicsData) GasPriceForProcessing(tx data.TransactionWithFeeHandler) uint64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.GasPriceForProcessingInEpoch(tx, currentEpoch)
}

// GasPriceForProcessingInEpoch computes the price for the gas in addition to balance movement and data in a specific epoch
func (ed *economicsData) GasPriceForProcessingInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) uint64 {
	return uint64(float64(tx.GetGasPrice()) * ed.GasPriceModifierInEpoch(epoch))
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
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ComputeTxFeeInEpoch(tx, currentEpoch)
}

// ComputeTxFeeInEpoch computes the provided transaction's fee in a specific epoch
func (ed *economicsData) ComputeTxFeeInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) *big.Int {
	if ed.enableEpochsHandler.IsFlagEnabledInEpoch(common.GasPriceModifierFlag, epoch) {
		if isSmartContractResult(tx) {
			return ed.ComputeFeeForProcessingInEpoch(tx, tx.GetGasLimit(), epoch)
		}

		gasLimitForMoveBalance, difference := ed.SplitTxGasInCategoriesInEpoch(tx, epoch)
		moveBalanceFee := core.SafeMul(ed.GasPriceForMove(tx), gasLimitForMoveBalance)
		if tx.GetGasLimit() <= gasLimitForMoveBalance {
			return moveBalanceFee
		}

		extraFee := ed.ComputeFeeForProcessingInEpoch(tx, difference, epoch)
		moveBalanceFee.Add(moveBalanceFee, extraFee)
		return moveBalanceFee
	}

	if ed.enableEpochsHandler.IsFlagEnabledInEpoch(common.PenalizedTooMuchGasFlag, epoch) {
		return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	return ed.ComputeMoveBalanceFeeInEpoch(tx, epoch)
}

// SplitTxGasInCategories returns the gas split per categories
func (ed *economicsData) SplitTxGasInCategories(tx data.TransactionWithFeeHandler) (gasLimitMove, gasLimitProcess uint64) {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.SplitTxGasInCategoriesInEpoch(tx, currentEpoch)
}

// SplitTxGasInCategoriesInEpoch returns the gas split per categories in a specific epoch
func (ed *economicsData) SplitTxGasInCategoriesInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) (gasLimitMove, gasLimitProcess uint64) {
	var err error
	gasLimitMove = ed.ComputeGasLimitInEpoch(tx, epoch)
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
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.CheckValidityTxValuesInEpoch(tx, currentEpoch)
}

// CheckValidityTxValuesInEpoch checks if the provided transaction is economically correct in a specific epoch
func (ed *economicsData) CheckValidityTxValuesInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) error {
	if ed.minGasPrice > tx.GetGasPrice() {
		return process.ErrInsufficientGasPriceInTx
	}

	if !isSmartContractResult(tx) {
		requiredGasLimit := ed.ComputeGasLimitInEpoch(tx, epoch)
		if tx.GetGasLimit() < requiredGasLimit {
			return process.ErrInsufficientGasLimitInTx
		}
	}

	// The following check should be kept as it is in order to avoid backwards compatibility issues
	if tx.GetGasLimit() >= ed.getMaxGasLimitPerBlock(epoch) {
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

// MaxGasLimitPerBlock returns maximum gas limit allowed per block
func (ed *economicsData) MaxGasLimitPerBlock(shardID uint32) uint64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.MaxGasLimitPerBlockInEpoch(shardID, currentEpoch)
}

// MaxGasLimitPerBlockInEpoch returns maximum gas limit allowed per block in a specific epoch
func (ed *economicsData) MaxGasLimitPerBlockInEpoch(shardID uint32, epoch uint32) uint64 {
	if shardID == core.MetachainShardId {
		return ed.getMaxGasLimitPerMetaBlock(epoch)
	}
	return ed.getMaxGasLimitPerBlock(epoch)
}

// MaxGasLimitPerMiniBlock returns maximum gas limit allowed per mini block
func (ed *economicsData) MaxGasLimitPerMiniBlock(shardID uint32) uint64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.MaxGasLimitPerMiniBlockInEpoch(shardID, currentEpoch)
}

// MaxGasLimitPerMiniBlockInEpoch returns maximum gas limit allowed per mini block in a specific epoch
func (ed *economicsData) MaxGasLimitPerMiniBlockInEpoch(shardID uint32, epoch uint32) uint64 {
	if shardID == core.MetachainShardId {
		return ed.getMaxGasLimitPerMetaMiniBlock(epoch)
	}
	return ed.getMaxGasLimitPerMiniBlock(epoch)
}

// MaxGasLimitPerBlockForSafeCrossShard returns maximum gas limit per block for safe cross shard
func (ed *economicsData) MaxGasLimitPerBlockForSafeCrossShard() uint64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.MaxGasLimitPerBlockForSafeCrossShardInEpoch(currentEpoch)
}

// MaxGasLimitPerBlockForSafeCrossShardInEpoch returns maximum gas limit per block for safe cross shard in a specific epoch
func (ed *economicsData) MaxGasLimitPerBlockForSafeCrossShardInEpoch(epoch uint32) uint64 {
	return ed.getMaxGasLimitPerBlockForSafeCrossShard(epoch)
}

// MaxGasLimitPerMiniBlockForSafeCrossShard returns maximum gas limit per mini block for safe cross shard
func (ed *economicsData) MaxGasLimitPerMiniBlockForSafeCrossShard() uint64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.MaxGasLimitPerMiniBlockForSafeCrossShardInEpoch(currentEpoch)
}

// MaxGasLimitPerMiniBlockForSafeCrossShardInEpoch returns maximum gas limit per mini block for safe cross shard in a specific epoch
func (ed *economicsData) MaxGasLimitPerMiniBlockForSafeCrossShardInEpoch(epoch uint32) uint64 {
	return ed.getMaxGasLimitPerMiniBlockForSafeCrossShard(epoch)
}

// MaxGasLimitPerTx returns maximum gas limit per tx
func (ed *economicsData) MaxGasLimitPerTx() uint64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.MaxGasLimitPerTxInEpoch(currentEpoch)
}

// MaxGasLimitPerTxInEpoch returns maximum gas limit per tx in a specific epoch
func (ed *economicsData) MaxGasLimitPerTxInEpoch(epoch uint32) uint64 {
	return ed.getMaxGasLimitPerTx(epoch)
}

// DeveloperPercentage returns the developer percentage value
func (ed *economicsData) DeveloperPercentage() float64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.DeveloperPercentageInEpoch(currentEpoch)
}

// DeveloperPercentageInEpoch returns the developer percentage value in a specific epoch
func (ed *economicsData) DeveloperPercentageInEpoch(epoch uint32) float64 {
	return ed.getDeveloperPercentage(epoch)
}

// ProtocolSustainabilityPercentage returns the protocol sustainability percentage value
func (ed *economicsData) ProtocolSustainabilityPercentage() float64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ProtocolSustainabilityPercentageInEpoch(currentEpoch)
}

// ProtocolSustainabilityPercentageInEpoch returns the protocol sustainability percentage value in a specific epoch
func (ed *economicsData) ProtocolSustainabilityPercentageInEpoch(epoch uint32) float64 {
	return ed.getProtocolSustainabilityPercentage(epoch)
}

// ProtocolSustainabilityAddress returns the protocol sustainability address
func (ed *economicsData) ProtocolSustainabilityAddress() string {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ProtocolSustainabilityAddressInEpoch(currentEpoch)
}

// ProtocolSustainabilityAddressInEpoch returns the protocol sustainability address in a specific epoch
func (ed *economicsData) ProtocolSustainabilityAddressInEpoch(epoch uint32) string {
	return ed.getProtocolSustainabilityAddress(epoch)
}

// RewardsTopUpGradientPoint returns the rewards top-up gradient point
func (ed *economicsData) RewardsTopUpGradientPoint() *big.Int {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.RewardsTopUpGradientPointInEpoch(currentEpoch)
}

// RewardsTopUpGradientPointInEpoch returns the rewards top-up gradient point in a specific epoch
func (ed *economicsData) RewardsTopUpGradientPointInEpoch(epoch uint32) *big.Int {
	return big.NewInt(0).Set(ed.getTopUpGradientPoint(epoch))
}

// RewardsTopUpFactor returns the rewards top-up factor
func (ed *economicsData) RewardsTopUpFactor() float64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.RewardsTopUpFactorInEpoch(currentEpoch)
}

// RewardsTopUpFactorInEpoch returns the rewards top-up factor in a specific epoch
func (ed *economicsData) RewardsTopUpFactorInEpoch(epoch uint32) float64 {
	return ed.getTopUpFactor(epoch)
}

// ComputeGasLimit returns the gas limit need by the provided transaction in order to be executed
func (ed *economicsData) ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64 {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ComputeGasLimitInEpoch(tx, currentEpoch)
}

// ComputeGasLimitInEpoch returns the gas limit need by the provided transaction in order to be executed in a specific epoch
func (ed *economicsData) ComputeGasLimitInEpoch(tx data.TransactionWithFeeHandler, epoch uint32) uint64 {
	gasLimit := ed.getMinGasLimit(epoch)

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * ed.gasPerDataByte
	txInstance, ok := tx.(*transaction.Transaction)
	if ok && ed.txVersionHandler.IsGuardedTransaction(txInstance) {
		gasLimit += ed.getExtraGasLimitGuardedTx(epoch)
	}

	return gasLimit
}

// ComputeGasUsedAndFeeBasedOnRefundValue will compute gas used value and transaction fee using refund value from a SCR
func (ed *economicsData) ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ComputeGasUsedAndFeeBasedOnRefundValueInEpoch(tx, refundValue, currentEpoch)
}

// ComputeGasUsedAndFeeBasedOnRefundValueInEpoch will compute gas used value and transaction fee using refund value from a SCR in a specific epoch
func (ed *economicsData) ComputeGasUsedAndFeeBasedOnRefundValueInEpoch(tx data.TransactionWithFeeHandler, refundValue *big.Int, epoch uint32) (uint64, *big.Int) {
	if refundValue.Cmp(big.NewInt(0)) == 0 {
		txFee := ed.ComputeTxFeeInEpoch(tx, epoch)

		return tx.GetGasLimit(), txFee
	}

	txFee := ed.ComputeTxFeeInEpoch(tx, epoch)

	isPenalizedTooMuchGasFlagEnabled := ed.enableEpochsHandler.IsFlagEnabledInEpoch(common.PenalizedTooMuchGasFlag, epoch)
	isGasPriceModifierFlagEnabled := ed.enableEpochsHandler.IsFlagEnabledInEpoch(common.GasPriceModifierFlag, epoch)
	flagCorrectTxFee := !isPenalizedTooMuchGasFlagEnabled && !isGasPriceModifierFlagEnabled
	if flagCorrectTxFee {
		txFee = core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
	}

	txFee = big.NewInt(0).Sub(txFee, refundValue)

	moveBalanceGasUnits := ed.ComputeGasLimitInEpoch(tx, epoch)
	moveBalanceFee := ed.ComputeMoveBalanceFeeInEpoch(tx, epoch)

	scOpFee := big.NewInt(0).Sub(txFee, moveBalanceFee)
	gasPriceForProcessing := big.NewInt(0).SetUint64(ed.GasPriceForProcessingInEpoch(tx, epoch))
	scOpGasUnits := big.NewInt(0).Div(scOpFee, gasPriceForProcessing)

	gasUsed := moveBalanceGasUnits + scOpGasUnits.Uint64()

	return gasUsed, txFee
}

// ComputeTxFeeBasedOnGasUsed will compute transaction fee
func (ed *economicsData) ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	currenEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ComputeTxFeeBasedOnGasUsedInEpoch(tx, gasUsed, currenEpoch)
}

// ComputeTxFeeBasedOnGasUsedInEpoch will compute transaction fee in a specific epoch
func (ed *economicsData) ComputeTxFeeBasedOnGasUsedInEpoch(tx data.TransactionWithFeeHandler, gasUsed uint64, epoch uint32) *big.Int {
	moveBalanceGasLimit := ed.ComputeGasLimitInEpoch(tx, epoch)
	moveBalanceFee := ed.ComputeMoveBalanceFeeInEpoch(tx, epoch)
	if gasUsed <= moveBalanceGasLimit {
		return moveBalanceFee
	}

	computeFeeForProcessing := ed.ComputeFeeForProcessingInEpoch(tx, gasUsed-moveBalanceGasLimit, epoch)
	txFee := big.NewInt(0).Add(moveBalanceFee, computeFeeForProcessing)

	return txFee
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (ed *economicsData) EpochConfirmed(epoch uint32, _ uint64) {
	ed.mut.RLock()
	ed.statusHandler.SetStringValue(common.MetricGasPriceModifier, fmt.Sprintf("%g", ed.GasPriceModifierInEpoch(epoch)))
	ed.mut.RUnlock()

	ed.updateRewardsConfigMetrics(epoch)
	ed.updateGasConfigMetrics(epoch)
}

// ComputeGasLimitBasedOnBalance will compute gas limit for the given transaction based on the balance
func (ed *economicsData) ComputeGasLimitBasedOnBalance(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error) {
	currentEpoch := ed.enableEpochsHandler.GetCurrentEpoch()
	return ed.ComputeGasLimitBasedOnBalanceInEpoch(tx, balance, currentEpoch)
}

// ComputeGasLimitBasedOnBalanceInEpoch will compute gas limit for the given transaction based on the balance in a specific epoch
func (ed *economicsData) ComputeGasLimitBasedOnBalanceInEpoch(tx data.TransactionWithFeeHandler, balance *big.Int, epoch uint32) (uint64, error) {
	balanceWithoutTransferValue := big.NewInt(0).Sub(balance, tx.GetValue())
	if balanceWithoutTransferValue.Cmp(big.NewInt(0)) < 1 {
		return 0, process.ErrInsufficientFunds
	}

	moveBalanceFee := ed.ComputeMoveBalanceFeeInEpoch(tx, epoch)
	if moveBalanceFee.Cmp(balanceWithoutTransferValue) > 0 {
		return 0, process.ErrInsufficientFunds
	}

	if !ed.enableEpochsHandler.IsFlagEnabledInEpoch(common.GasPriceModifierFlag, epoch) {
		gasPriceBig := big.NewInt(0).SetUint64(tx.GetGasPrice())
		gasLimitBig := big.NewInt(0).Div(balanceWithoutTransferValue, gasPriceBig)

		return gasLimitBig.Uint64(), nil
	}

	remainedBalanceAfterMoveBalanceFee := big.NewInt(0).Sub(balanceWithoutTransferValue, moveBalanceFee)
	gasPriceBigForProcessing := ed.GasPriceForProcessingInEpoch(tx, epoch)
	gasPriceBigForProcessingBig := big.NewInt(0).SetUint64(gasPriceBigForProcessing)
	gasLimitFromRemainedBalanceBig := big.NewInt(0).Div(remainedBalanceAfterMoveBalanceFee, gasPriceBigForProcessingBig)

	gasLimitMoveBalance := ed.ComputeGasLimitInEpoch(tx, epoch)
	totalGasLimit := gasLimitMoveBalance + gasLimitFromRemainedBalanceBig.Uint64()

	return totalGasLimit, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *economicsData) IsInterfaceNil() bool {
	return ed == nil
}
