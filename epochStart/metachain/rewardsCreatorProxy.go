package metachain

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochStartRewardsCreator defines the functionality for the metachain to create rewards at end of epoch
type EpochStartRewardsCreator interface {
	CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error)
	VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error
	GetProtocolSustainabilityRewards() *big.Int
	GetLocalTxCache() epochStart.TransactionCacher
	CreateMarshalizedData(body *block.Body) map[string][][]byte
	GetRewardsTxs(body *block.Body) map[string]data.TransactionHandler
	SaveTxBlockToStorage(metaBlock *block.MetaBlock, body *block.Body)
	DeleteTxsFromStorage(metaBlock *block.MetaBlock, body *block.Body)
	RemoveBlockDataFromPools(metaBlock *block.MetaBlock, body *block.Body)
	IsInterfaceNil() bool
}

type configuredRewardsCreator string

const (
	rCreatorV1 configuredRewardsCreator = "rewardsCreatorV1"
	rCreatorV2                          = "rewardsCreatorV2"
)

// RewardsCreatorProxyArgs holds the proxy arguments
type RewardsCreatorProxyArgs struct {
	BaseRewardsCreatorArgs
	StakingDataProvider   epochStart.StakingDataProvider
	EconomicsDataProvider epochStart.EpochEconomicsDataProvider
	TopUpRewardFactor     float64
	TopUpGradientPoint    *big.Int
	EpochEnableV2         uint32
}

type rewardsCreatorProxy struct {
	rc            EpochStartRewardsCreator
	epochEnableV2 uint32
	configuredRC  configuredRewardsCreator
	args          *RewardsCreatorProxyArgs
	mutRc         sync.Mutex
}

func NewRewardsCreatorProxy(args RewardsCreatorProxyArgs) (*rewardsCreatorProxy, error) {
	argsV1 := ArgsNewRewardsCreator{
		BaseRewardsCreatorArgs: args.BaseRewardsCreatorArgs,
	}

	rcV1, err := NewEpochStartRewardsCreator(argsV1)
	if err != nil {
		return nil, err
	}

	log.Info("rewardsCreatorProxy", "configured", rCreatorV1)

	return &rewardsCreatorProxy{
		rc:            rcV1,
		configuredRC:  rCreatorV1,
		epochEnableV2: args.EpochEnableV2,
		args:          &args,
	}, nil
}

// CreateRewardsMiniBlocks proxies the CreateRewardsMiniBlocks method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	err := rcp.changeRewardCreatorIfNeeded(metaBlock.Epoch)
	if err != nil {
		return nil, err
	}
	return rcp.rc.CreateRewardsMiniBlocks(metaBlock, validatorsInfo)
}

// VerifyRewardsMiniBlocks proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error {
	err := rcp.changeRewardCreatorIfNeeded(metaBlock.Epoch)
	if err != nil {
		return err
	}
	return rcp.rc.VerifyRewardsMiniBlocks(metaBlock, validatorsInfo)
}

// GetProtocolSustainabilityRewards proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) GetProtocolSustainabilityRewards() *big.Int {
	return rcp.rc.GetProtocolSustainabilityRewards()
}

// GetLocalTxCache proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) GetLocalTxCache() epochStart.TransactionCacher {
	return rcp.rc.GetLocalTxCache()
}

// CreateMarshalizedData proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) CreateMarshalizedData(body *block.Body) map[string][][]byte {
	return rcp.rc.CreateMarshalizedData(body)
}

// GetRewardsTxs proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) GetRewardsTxs(body *block.Body) map[string]data.TransactionHandler {
	return rcp.rc.GetRewardsTxs(body)
}

// SaveTxBlockToStorage proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) SaveTxBlockToStorage(metaBlock *block.MetaBlock, body *block.Body) {
	rcp.rc.SaveTxBlockToStorage(metaBlock, body)
}

// // RemoveBlockDataFromPools proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) DeleteTxsFromStorage(metaBlock *block.MetaBlock, body *block.Body) {
	rcp.rc.DeleteTxsFromStorage(metaBlock, body)
}

// RemoveBlockDataFromPools proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) RemoveBlockDataFromPools(metaBlock *block.MetaBlock, body *block.Body) {
	rcp.rc.RemoveBlockDataFromPools(metaBlock, body)
}

// IsInterfaceNil returns nil if the underlying object is nil
func (rcp *rewardsCreatorProxy) IsInterfaceNil() bool {
	return rcp == nil
}

func (rcp *rewardsCreatorProxy) changeRewardCreatorIfNeeded(epoch uint32) error {
	rcp.mutRc.Lock()
	defer rcp.mutRc.Unlock()

	switch rcp.configuredRC {
	case rCreatorV1:
		if epoch > rcp.epochEnableV2 {
			return rcp.switchToRewardsCreatorV2()
		}
	case rCreatorV2:
		if epoch <= rcp.epochEnableV2 {
			return rcp.switchToRewardsCreatorV1()
		}
	default:
		return epochStart.ErrInvalidRewardsCreatorProxyConfig
	}

	return nil
}

// to be called under locked mutex
func (rcp *rewardsCreatorProxy) switchToRewardsCreatorV1() error {
	argsV1 := ArgsNewRewardsCreator{
		BaseRewardsCreatorArgs: rcp.args.BaseRewardsCreatorArgs,
	}

	rcV1, err := NewEpochStartRewardsCreator(argsV1)
	if err != nil {
		return err
	}

	rcp.rc = rcV1
	rcp.configuredRC = rCreatorV1

	log.Info("rewardsCreatorProxy.switchToRewardsCreatorV1")

	return nil
}

// to be called under locked mutex
func (rcp *rewardsCreatorProxy) switchToRewardsCreatorV2() error {
	argsV2 := RewardsCreatorArgsV2{
		BaseRewardsCreatorArgs: rcp.args.BaseRewardsCreatorArgs,
		StakingDataProvider:    rcp.args.StakingDataProvider,
		EconomicsDataProvider:  rcp.args.EconomicsDataProvider,
		TopUpRewardFactor:      rcp.args.TopUpRewardFactor,
		TopUpGradientPoint:     rcp.args.TopUpGradientPoint,
	}

	rcV2, err := NewEpochStartRewardsCreatorV2(argsV2)
	if err != nil {
		return err
	}
	rcp.rc = rcV2

	log.Info("rewardsCreatorProxy.switchToRewardsCreatorV2")

	return nil
}
