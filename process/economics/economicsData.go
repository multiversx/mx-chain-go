package economics

import (
	"github.com/ElrondNetwork/elrond-go/config"
)

// EconomicsData will store information about economics
type EconomicsData struct {
	rewardsValue        uint64
	communityPercentage float64
	leaderPercentage    float64
	burnPercentage      float64

	minGasPrice      uint64
	minGasLimitForTx uint64
	minTxFee         uint64

	communityAddress string
	burnAddress      string
}

// NewEconomicsData will create and object with information about economics parameters
func NewEconomicsData(economics *config.ConfigEconomics) *EconomicsData {
	return &EconomicsData{
		rewardsValue:        economics.RewardsSettings.RewardsValue,
		communityPercentage: economics.RewardsSettings.CommunityPercentage,
		leaderPercentage:    economics.RewardsSettings.LeaderPercentage,
		burnPercentage:      economics.RewardsSettings.BurnPercentage,
		minGasPrice:         economics.FeeSettings.MinGasPrice,
		minGasLimitForTx:    economics.FeeSettings.MinGasLimitForTx,
		minTxFee:            economics.FeeSettings.MinTxFee,
		communityAddress:    economics.EconomicsAddresses.CommunityAddress,
		burnAddress:         economics.EconomicsAddresses.BurnAddress,
	}
}

// RewardsValue will return rewards value
func (ed *EconomicsData) RewardsValue() uint64 {
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

// MinGasPrice will return minimum gas price
func (ed *EconomicsData) MinGasPrice() uint64 {
	return ed.minGasPrice
}

// MinGasLimitForTx will return minimum gas limit
func (ed *EconomicsData) MinGasLimitForTx() uint64 {
	return ed.minGasLimitForTx
}

// MinTxFee will return minimum transaction fee
func (ed *EconomicsData) MinTxFee() uint64 {
	return ed.minTxFee
}

// CommunityAddress will return community address
func (ed *EconomicsData) CommunityAddress() string {
	return ed.communityAddress
}

// BurnAddress will return burn address
func (ed *EconomicsData) BurnAddress() string {
	return ed.burnAddress
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *EconomicsData) IsInterfaceNil() bool {
	if ed == nil {
		return true
	}
	return false
}
