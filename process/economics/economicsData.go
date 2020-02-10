package economics

import (
	"math"
	"math/big"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

// EconomicsData will store information about economics
type EconomicsData struct {
	rewardsValue         *big.Int
	communityPercentage  float64
	leaderPercentage     float64
	burnPercentage       float64
	maxGasLimitPerBlock  uint64
	gasPerDataByte       uint64
	dataLimitForBaseCalc uint64
	minGasPrice          uint64
	minGasLimit          uint64
	communityAddress     string
	burnAddress          string
	stakeValue           *big.Int
	unBondPeriod         uint64
	ratingsData          *RatingsData
	minStep              *big.Int
	totalSupply          *big.Int
	numNodes             uint32
	auctionEnableNonce   uint64
	stakeEnableNonce     uint64
}

const float64EqualityThreshold = 1e-9

// NewEconomicsData will create and object with information about economics parameters
func NewEconomicsData(economics *config.ConfigEconomics) (*EconomicsData, error) {
	//TODO check what happens if addresses are wrong
	data, err := convertValues(economics)
	if err != nil {
		return nil, err
	}

	notGreaterThanZero := data.rewardsValue.Cmp(big.NewInt(0))
	if notGreaterThanZero < 0 {
		return nil, process.ErrInvalidRewardsValue
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
		rewardsValue:         data.rewardsValue,
		communityPercentage:  economics.RewardsSettings.CommunityPercentage,
		leaderPercentage:     economics.RewardsSettings.LeaderPercentage,
		burnPercentage:       economics.RewardsSettings.BurnPercentage,
		maxGasLimitPerBlock:  data.maxGasLimitPerBlock,
		minGasPrice:          data.minGasPrice,
		minGasLimit:          data.minGasLimit,
		communityAddress:     economics.EconomicsAddresses.CommunityAddress,
		burnAddress:          economics.EconomicsAddresses.BurnAddress,
		stakeValue:           data.stakeValue,
		unBondPeriod:         data.unBondPeriod,
		gasPerDataByte:       data.gasPerDataByte,
		dataLimitForBaseCalc: data.dataLimitForBaseCalc,
		ratingsData:          rd,
		totalSupply:          data.totalSupply,
		minStep:              data.minStep,
		numNodes:             data.numNodes,
		auctionEnableNonce:   data.auctionEnableNonce,
		stakeEnableNonce:     data.stakeEnableNonce,
	}, nil
}

func convertValues(economics *config.ConfigEconomics) (*EconomicsData, error) {
	conversionBase := 10
	bitConversionSize := 64

	rewardsValue := new(big.Int)
	rewardsValue, ok := rewardsValue.SetString(economics.RewardsSettings.RewardsValue, conversionBase)
	if !ok {
		return nil, process.ErrInvalidRewardsValue
	}

	minGasPrice, err := strconv.ParseUint(economics.FeeSettings.MinGasPrice, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidMinimumGasPrice
	}

	minGasLimit, err := strconv.ParseUint(economics.FeeSettings.MinGasLimit, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidMinimumGasLimitForTx
	}

	stakeValue := new(big.Int)
	stakeValue, ok = stakeValue.SetString(economics.ValidatorSettings.StakeValue, conversionBase)
	if !ok {
		return nil, process.ErrInvalidRewardsValue
	}

	unBondPeriod, err := strconv.ParseUint(economics.ValidatorSettings.UnBondPeriod, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidUnBondPeriod
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

	totalSupply := new(big.Int)
	totalSupply, ok = totalSupply.SetString(economics.ValidatorSettings.TotalSupply, conversionBase)
	if !ok {
		return nil, process.ErrInvalidTotalSupply
	}

	minStepValue := new(big.Int)
	minStepValue, ok = minStepValue.SetString(economics.ValidatorSettings.MinStepValue, conversionBase)
	if !ok {
		return nil, process.ErrInvalidMinStepValue
	}

	auctionEnableNonce, err := strconv.ParseUint(economics.ValidatorSettings.AuctionEnableNonce, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidUnBondPeriod
	}

	stakeEnableNonce, err := strconv.ParseUint(economics.ValidatorSettings.StakeEnableNonce, conversionBase, bitConversionSize)
	if err != nil {
		return nil, process.ErrInvalidUnBondPeriod
	}

	return &EconomicsData{
		rewardsValue:         rewardsValue,
		minGasPrice:          minGasPrice,
		minGasLimit:          minGasLimit,
		stakeValue:           stakeValue,
		unBondPeriod:         unBondPeriod,
		maxGasLimitPerBlock:  maxGasLimitPerBlock,
		gasPerDataByte:       gasPerDataByte,
		dataLimitForBaseCalc: dataLimitForBaseCalc,
		totalSupply:          totalSupply,
		minStep:              minStepValue,
		numNodes:             economics.ValidatorSettings.NumNodes,
		auctionEnableNonce:   auctionEnableNonce,
		stakeEnableNonce:     stakeEnableNonce,
	}, nil
}

func checkValues(economics *config.ConfigEconomics) error {
	if isPercentageInvalid(economics.RewardsSettings.BurnPercentage) ||
		isPercentageInvalid(economics.RewardsSettings.CommunityPercentage) ||
		isPercentageInvalid(economics.RewardsSettings.LeaderPercentage) {
		return process.ErrInvalidRewardsPercentages
	}

	sumPercentage := economics.RewardsSettings.BurnPercentage
	sumPercentage += economics.RewardsSettings.CommunityPercentage
	sumPercentage += economics.RewardsSettings.LeaderPercentage
	isEqualsToOne := math.Abs(sumPercentage-1.0) <= float64EqualityThreshold
	if !isEqualsToOne {
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

// RewardsValue will return rewards value
func (ed *EconomicsData) RewardsValue() *big.Int {
	return ed.rewardsValue
}

// CommunityPercentage will return community reward percentage
func (ed *EconomicsData) CommunityPercentage() float64 {
	return ed.communityPercentage
}

// LeaderPercentage will return leader reward percentage
func (ed *EconomicsData) LeaderPercentage() float64 {
	return ed.leaderPercentage
}

// BurnPercentage will return burn percentage
func (ed *EconomicsData) BurnPercentage() float64 {
	return ed.burnPercentage
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

// ComputeGasLimit returns the gas limit need by the provided transaction in order to be executed
func (ed *EconomicsData) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	gasLimit := ed.minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * ed.gasPerDataByte
	//TODO reevaluate the formula or delete
	/* if dataLen < ed.dataLimitForBaseCalc || core.IsEmptyAddress(tx.GetRecvAddress()) {
		return gasLimit
	}

	overDataLimit := dataLen - ed.dataLimitForBaseCalc
	gasLimit += overDataLimit * overDataLimit * ed.gasPerDataByte */

	return gasLimit
}

// CommunityAddress will return community address
func (ed *EconomicsData) CommunityAddress() string {
	return ed.communityAddress
}

// BurnAddress will return burn address
func (ed *EconomicsData) BurnAddress() string {
	return ed.burnAddress
}

// StakeValue will return the minimum stake value
func (ed *EconomicsData) StakeValue() *big.Int {
	return ed.stakeValue
}

// UnBondPeriod will return the unbond period
func (ed *EconomicsData) UnBondPeriod() uint64 {
	return ed.unBondPeriod
}

// MinStepValue returns the step value which is considered in the node price determination
func (ed *EconomicsData) MinStepValue() *big.Int {
	return ed.minStep
}

// TotalSupply returns the total supply of the protocol
func (ed *EconomicsData) TotalSupply() *big.Int {
	return ed.totalSupply
}

// NumNodes returns the total node number for current setting
func (ed *EconomicsData) NumNodes() uint32 {
	return ed.numNodes
}

// AuctionEnableNonce returns whether full auction process is enabled
func (ed *EconomicsData) AuctionEnableNonce() uint64 {
	return ed.auctionEnableNonce
}

// StakeEnableNonce returns whether full auction process is enabled
func (ed *EconomicsData) StakeEnableNonce() uint64 {
	return ed.stakeEnableNonce
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *EconomicsData) IsInterfaceNil() bool {
	return ed == nil
}

// RatingsData will return the ratingsDataObject
func (ed *EconomicsData) RatingsData() *RatingsData {
	return ed.ratingsData
}
