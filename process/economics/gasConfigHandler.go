package economics

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/statusHandler"
)

const epsilon = 0.00000001

type gasConfig struct {
	gasLimitSettingEpoch        uint32
	maxGasLimitPerBlock         uint64
	maxGasLimitPerMiniBlock     uint64
	maxGasLimitPerMetaBlock     uint64
	maxGasLimitPerMetaMiniBlock uint64
	maxGasLimitPerTx            uint64
	minGasLimit                 uint64
	extraGasLimitGuardedTx      uint64
}

type gasConfigHandler struct {
	statusHandler          core.AppStatusHandler
	gasLimitSettings       []*gasConfig
	minGasPrice            uint64
	gasPerDataByte         uint64
	genesisTotalSupply     *big.Int
	maxGasPriceSetGuardian uint64
	mut                    sync.RWMutex
}

// newGasConfigHandler returns a new instance of gasConfigHandler
func newGasConfigHandler(economics *config.EconomicsConfig) (*gasConfigHandler, error) {
	gasConfigSlice, err := checkAndParseFeeSettings(economics.FeeSettings)
	if err != nil {
		return nil, err
	}

	sort.Slice(gasConfigSlice, func(i, j int) bool {
		return gasConfigSlice[i].gasLimitSettingEpoch < gasConfigSlice[j].gasLimitSettingEpoch
	})

	minGasPrice, gasPerDataByte, genesisTotalSupply, maxGasPriceSetGuardian, err := convertGenericValues(economics)
	if err != nil {
		return nil, err
	}

	return &gasConfigHandler{
		statusHandler:          statusHandler.NewNilStatusHandler(),
		gasLimitSettings:       gasConfigSlice,
		minGasPrice:            minGasPrice,
		gasPerDataByte:         gasPerDataByte,
		genesisTotalSupply:     genesisTotalSupply,
		maxGasPriceSetGuardian: maxGasPriceSetGuardian,
	}, nil
}

// setStatusHandler sets the provided status handler if not nil
func (handler *gasConfigHandler) setStatusHandler(statusHandler core.AppStatusHandler) error {
	if check.IfNil(statusHandler) {
		return core.ErrNilAppStatusHandler
	}

	handler.mut.Lock()
	handler.statusHandler = statusHandler
	handler.mut.Unlock()

	return nil
}

// getMinGasLimit returns min gas limit in a specific epoch
func (handler *gasConfigHandler) getMinGasLimit(epoch uint32) uint64 {
	gc := handler.getGasConfigForEpoch(epoch)
	return gc.minGasLimit
}

// getExtraGasLimitGuardedTx returns extra gas limit for guarded tx in a specific epoch
func (handler *gasConfigHandler) getExtraGasLimitGuardedTx(epoch uint32) uint64 {
	gc := handler.getGasConfigForEpoch(epoch)
	return gc.extraGasLimitGuardedTx
}

// getMaxGasLimitPerMetaBlock returns max gas limit per meta block in a specific epoch
func (handler *gasConfigHandler) getMaxGasLimitPerMetaBlock(epoch uint32) uint64 {
	gc := handler.getGasConfigForEpoch(epoch)
	return gc.maxGasLimitPerMetaBlock
}

// getMaxGasLimitPerBlock returns max gas limit per block in a specific epoch
func (handler *gasConfigHandler) getMaxGasLimitPerBlock(epoch uint32) uint64 {
	gc := handler.getGasConfigForEpoch(epoch)
	return gc.maxGasLimitPerBlock
}

// getMaxGasLimitPerMetaMiniBlock returns max gas limit per meta mini block in a specific epoch
func (handler *gasConfigHandler) getMaxGasLimitPerMetaMiniBlock(epoch uint32) uint64 {
	gc := handler.getGasConfigForEpoch(epoch)
	return gc.maxGasLimitPerMetaMiniBlock
}

// getMaxGasLimitPerMiniBlock returns max gas limit per mini block in a specific epoch
func (handler *gasConfigHandler) getMaxGasLimitPerMiniBlock(epoch uint32) uint64 {
	gc := handler.getGasConfigForEpoch(epoch)
	return gc.maxGasLimitPerMiniBlock
}

// getMaxGasLimitPerBlockForSafeCrossShard returns maximum gas limit per block for safe cross shard in a specific epoch
func (handler *gasConfigHandler) getMaxGasLimitPerBlockForSafeCrossShard(epoch uint32) uint64 {
	gc := handler.getGasConfigForEpoch(epoch)
	return core.MinUint64(gc.maxGasLimitPerBlock, gc.maxGasLimitPerMetaBlock)
}

// getMaxGasLimitPerMiniBlockForSafeCrossShard returns maximum gas limit per mini block for safe cross shard in a specific epoch
func (handler *gasConfigHandler) getMaxGasLimitPerMiniBlockForSafeCrossShard(epoch uint32) uint64 {
	gc := handler.getGasConfigForEpoch(epoch)
	return core.MinUint64(gc.maxGasLimitPerMiniBlock, gc.maxGasLimitPerMetaMiniBlock)
}

// getMaxGasLimitPerTx returns max gas limit per tx in a specific epoch
func (handler *gasConfigHandler) getMaxGasLimitPerTx(epoch uint32) uint64 {
	gc := handler.getGasConfigForEpoch(epoch)
	return gc.maxGasLimitPerTx
}

func (handler *gasConfigHandler) updateGasConfigMetrics(epoch uint32) {
	gc := handler.getGasConfigForEpoch(epoch)

	log.Debug("economics: gasConfigHandler",
		"epoch", gc.gasLimitSettingEpoch,
		"maxGasLimitPerBlock", gc.maxGasLimitPerBlock,
		"maxGasLimitPerMiniBlock", gc.maxGasLimitPerMiniBlock,
		"maxGasLimitPerMetaBlock", gc.maxGasLimitPerMetaBlock,
		"maxGasLimitPerMetaMiniBlock", gc.maxGasLimitPerMetaMiniBlock,
		"maxGasLimitPerTx", gc.maxGasLimitPerTx,
		"minGasLimit", gc.minGasLimit,
	)

	handler.mut.RLock()
	handler.statusHandler.SetUInt64Value(common.MetricMaxGasPerTransaction, gc.maxGasLimitPerTx)
	handler.mut.RUnlock()
}

func (handler *gasConfigHandler) getGasConfigForEpoch(epoch uint32) *gasConfig {
	gasConfigSetting := handler.gasLimitSettings[0]
	for i := 1; i < len(handler.gasLimitSettings); i++ {
		if epoch >= handler.gasLimitSettings[i].gasLimitSettingEpoch {
			gasConfigSetting = handler.gasLimitSettings[i]
		}
	}

	return gasConfigSetting
}

func checkAndParseFeeSettings(feeSettings config.FeeSettings) ([]*gasConfig, error) {
	if feeSettings.GasPriceModifier > 1.0 || feeSettings.GasPriceModifier < epsilon {
		return nil, process.ErrInvalidGasModifier
	}

	if len(feeSettings.GasLimitSettings) == 0 {
		return nil, process.ErrEmptyGasLimitSettings
	}

	gasConfigSlice := make([]*gasConfig, 0, len(feeSettings.GasLimitSettings))
	for _, gasLimitSetting := range feeSettings.GasLimitSettings {
		gc, err := checkAndParseGasLimitSettings(gasLimitSetting)
		if err != nil {
			return nil, err
		}

		gasConfigSlice = append(gasConfigSlice, gc)
	}

	return gasConfigSlice, nil
}

func checkAndParseGasLimitSettings(gasLimitSetting config.GasLimitSetting) (*gasConfig, error) {
	conversionBase := 10
	bitConversionSize := 64

	gc := &gasConfig{}
	var err error

	gc.gasLimitSettingEpoch = gasLimitSetting.EnableEpoch
	gc.minGasLimit, err = strconv.ParseUint(gasLimitSetting.MinGasLimit, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidMinimumGasLimitForTx
	}

	gc.maxGasLimitPerBlock, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerBlock, conversionBase, bitConversionSize)
	if err != nil {
		return nil, fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerBlock, gasLimitSetting.EnableEpoch)
	}

	gc.maxGasLimitPerMiniBlock, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerMiniBlock, conversionBase, bitConversionSize)
	if err != nil {
		return nil, fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerMiniBlock, gasLimitSetting.EnableEpoch)
	}

	gc.maxGasLimitPerMetaBlock, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerMetaBlock, conversionBase, bitConversionSize)
	if err != nil {
		return nil, fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerMetaBlock, gasLimitSetting.EnableEpoch)
	}

	gc.maxGasLimitPerMetaMiniBlock, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerMetaMiniBlock, conversionBase, bitConversionSize)
	if err != nil {
		return nil, fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerMetaMiniBlock, gasLimitSetting.EnableEpoch)
	}

	gc.maxGasLimitPerTx, err = strconv.ParseUint(gasLimitSetting.MaxGasLimitPerTx, conversionBase, bitConversionSize)
	if err != nil {
		return nil, fmt.Errorf("%w for epoch %d", process.ErrInvalidMaxGasLimitPerTx, gasLimitSetting.EnableEpoch)
	}

	gc.extraGasLimitGuardedTx, err = strconv.ParseUint(gasLimitSetting.ExtraGasLimitGuardedTx, conversionBase, bitConversionSize)
	if err != nil {
		return nil, fmt.Errorf("%w for epoch %d", process.ErrInvalidExtraGasLimitGuardedTx, gasLimitSetting.EnableEpoch)
	}

	if gc.maxGasLimitPerBlock < gc.minGasLimit {
		return nil, fmt.Errorf("%w: maxGasLimitPerBlock = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerBlock, gc.maxGasLimitPerBlock, gc.minGasLimit, gasLimitSetting.EnableEpoch)
	}
	if gc.maxGasLimitPerMiniBlock < gc.minGasLimit {
		return nil, fmt.Errorf("%w: maxGasLimitPerMiniBlock = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerMiniBlock, gc.maxGasLimitPerMiniBlock, gc.minGasLimit, gasLimitSetting.EnableEpoch)
	}
	if gc.maxGasLimitPerMetaBlock < gc.minGasLimit {
		return nil, fmt.Errorf("%w: maxGasLimitPerMetaBlock = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerMetaBlock, gc.maxGasLimitPerMetaBlock, gc.minGasLimit, gasLimitSetting.EnableEpoch)
	}
	if gc.maxGasLimitPerMetaMiniBlock < gc.minGasLimit {
		return nil, fmt.Errorf("%w: maxGasLimitPerMetaMiniBlock = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerMetaMiniBlock, gc.maxGasLimitPerMetaMiniBlock, gc.minGasLimit, gasLimitSetting.EnableEpoch)
	}
	if gc.maxGasLimitPerTx < gc.minGasLimit {
		return nil, fmt.Errorf("%w: maxGasLimitPerTx = %d minGasLimit = %d in epoch %d", process.ErrInvalidMaxGasLimitPerTx, gc.maxGasLimitPerTx, gc.minGasLimit, gasLimitSetting.EnableEpoch)
	}

	return gc, nil
}

func convertGenericValues(economics *config.EconomicsConfig) (uint64, uint64, *big.Int, uint64, error) {
	conversionBase := 10
	bitConversionSize := 64

	minGasPrice, err := strconv.ParseUint(economics.FeeSettings.MinGasPrice, conversionBase, bitConversionSize)
	if err != nil {
		return 0, 0, nil, 0, process.ErrInvalidMinimumGasPrice
	}

	gasPerDataByte, err := strconv.ParseUint(economics.FeeSettings.GasPerDataByte, conversionBase, bitConversionSize)
	if err != nil {
		return 0, 0, nil, 0, process.ErrInvalidGasPerDataByte
	}

	genesisTotalSupply, ok := big.NewInt(0).SetString(economics.GlobalSettings.GenesisTotalSupply, conversionBase)
	if !ok {
		return 0, 0, nil, 0, process.ErrInvalidGenesisTotalSupply
	}

	maxGasPriceSetGuardian, err := strconv.ParseUint(economics.FeeSettings.MaxGasPriceSetGuardian, conversionBase, bitConversionSize)
	if err != nil {
		return 0, 0, nil, 0, process.ErrInvalidMaxGasPriceSetGuardian
	}

	return minGasPrice, gasPerDataByte, genesisTotalSupply, maxGasPriceSetGuardian, nil
}
