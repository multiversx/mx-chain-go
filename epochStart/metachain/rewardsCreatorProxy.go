package metachain

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
)

type configuredRewardsCreator string

var _ process.RewardsCreator = (*rewardsCreatorProxy)(nil)

const (
	rCreatorV1 configuredRewardsCreator = "rewardsCreatorV1"
	rCreatorV2 configuredRewardsCreator = "rewardsCreatorV2"
)

// RewardsCreatorProxyArgs holds the proxy arguments
type RewardsCreatorProxyArgs struct {
	BaseRewardsCreatorArgs
	StakingDataProvider   epochStart.StakingDataProvider
	EconomicsDataProvider epochStart.EpochEconomicsDataProvider
	RewardsHandler        process.RewardsHandler
	EpochEnableV2         uint32
}

type rewardsCreatorProxy struct {
	rc            epochStart.RewardsCreator
	epochEnableV2 uint32
	configuredRC  configuredRewardsCreator
	args          *RewardsCreatorProxyArgs
	mutRc         sync.Mutex
}

// NewRewardsCreatorProxy creates a rewards creator proxy instance
func NewRewardsCreatorProxy(args RewardsCreatorProxyArgs) (*rewardsCreatorProxy, error) {
	var err error

	rcProxy := &rewardsCreatorProxy{
		epochEnableV2: args.EpochEnableV2,
		configuredRC:  rCreatorV1,
		args:          &args,
	}
	log.Debug("rewardsProxy: enable epoch for switch jail waiting", "epoch", args.BaseRewardsCreatorArgs.RewardsFix1EpochEnable)
	log.Debug("rewardsProxy: enable epoch for staking v2", "epoch", args.EpochEnableV2)

	rcProxy.rc, err = rcProxy.createRewardsCreatorV1()
	if err != nil {
		return nil, err
	}

	// test creation of v2 rewards creator
	_, err = rcProxy.createRewardsCreatorV2()
	if err != nil {
		return nil, err
	}

	log.Info("rewardsCreatorProxy", "configured", rCreatorV1)

	return rcProxy, nil
}

// CreateRewardsMiniBlocks proxies the CreateRewardsMiniBlocks method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) CreateRewardsMiniBlocks(
	metaBlock data.MetaHeaderHandler,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) (block.MiniBlockSlice, error) {
	err := rcp.changeRewardCreatorIfNeeded(metaBlock.GetEpoch())
	if err != nil {
		return nil, err
	}
	return rcp.rc.CreateRewardsMiniBlocks(metaBlock, validatorsInfo, computedEconomics)
}

// VerifyRewardsMiniBlocks proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) VerifyRewardsMiniBlocks(
	metaBlock data.MetaHeaderHandler,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) error {
	err := rcp.changeRewardCreatorIfNeeded(metaBlock.GetEpoch())
	if err != nil {
		return err
	}
	return rcp.rc.VerifyRewardsMiniBlocks(metaBlock, validatorsInfo, computedEconomics)
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
func (rcp *rewardsCreatorProxy) SaveTxBlockToStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	rcp.rc.SaveTxBlockToStorage(metaBlock, body)
}

// DeleteTxsFromStorage proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) DeleteTxsFromStorage(metaBlock data.MetaHeaderHandler, body *block.Body) {
	rcp.rc.DeleteTxsFromStorage(metaBlock, body)
}

// RemoveBlockDataFromPools proxies the same method of the configured rewardsCreator instance
func (rcp *rewardsCreatorProxy) RemoveBlockDataFromPools(metaBlock data.MetaHeaderHandler, body *block.Body) {
	rcp.rc.RemoveBlockDataFromPools(metaBlock, body)
}

// IsInterfaceNil returns nil if the underlying object is nil
func (rcp *rewardsCreatorProxy) IsInterfaceNil() bool {
	return rcp == nil
}

func (rcp *rewardsCreatorProxy) changeRewardCreatorIfNeeded(epoch uint32) error {
	rcp.mutRc.Lock()
	defer rcp.mutRc.Unlock()

	if epoch > rcp.epochEnableV2 {
		if rcp.configuredRC != rCreatorV2 {
			return rcp.switchToRewardsCreatorV2()
		}

		return nil
	}

	if rcp.configuredRC != rCreatorV1 {
		return rcp.switchToRewardsCreatorV1()
	}

	return nil
}

// to be called under locked mutex
func (rcp *rewardsCreatorProxy) switchToRewardsCreatorV1() error {
	rcV1, err := rcp.createRewardsCreatorV1()
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
	rcV2, err := rcp.createRewardsCreatorV2()
	if err != nil {
		return err
	}

	rcp.rc = rcV2
	rcp.configuredRC = rCreatorV2

	log.Info("rewardsCreatorProxy.switchToRewardsCreatorV2")

	return nil
}

func (rcp *rewardsCreatorProxy) createRewardsCreatorV2() (*rewardsCreatorV2, error) {
	argsV2 := RewardsCreatorArgsV2{
		BaseRewardsCreatorArgs: rcp.args.BaseRewardsCreatorArgs,
		StakingDataProvider:    rcp.args.StakingDataProvider,
		EconomicsDataProvider:  rcp.args.EconomicsDataProvider,
		RewardsHandler:         rcp.args.RewardsHandler,
	}

	return NewRewardsCreatorV2(argsV2)
}

func (rcp *rewardsCreatorProxy) createRewardsCreatorV1() (*rewardsCreator, error) {
	argsV1 := ArgsNewRewardsCreator{
		BaseRewardsCreatorArgs: rcp.args.BaseRewardsCreatorArgs,
	}

	return NewRewardsCreator(argsV1)
}
