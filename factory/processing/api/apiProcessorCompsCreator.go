package api

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	metachainEpochStart "github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	factoryDisabled "github.com/multiversx/mx-chain-go/factory/disabled"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/sharding"
)

type apiProcessorCompsCreator struct {
}

type APIProcessComps struct {
	StakingDataProviderAPI peer.StakingDataProviderAPI
	AuctionListSelector    epochStart.AuctionListSelector
}

type ArgsCreateAPIProcessComps struct {
	ArgsStakingDataProvider metachainEpochStart.StakingDataProviderArgs
	ShardCoordinator        sharding.Coordinator
	EpochNotifier           process.EpochNotifier
	SoftAuctionConfig       config.SoftAuctionConfig
	EnableEpochs            config.EnableEpochs
	Denomination            int
}

func NewAPIProcessorCompsCreator() *apiProcessorCompsCreator {
	return &apiProcessorCompsCreator{}
}

func (c *apiProcessorCompsCreator) CreateAPIComps(args ArgsCreateAPIProcessComps) (*APIProcessComps, error) {
	if args.ShardCoordinator.SelfId() == core.MetachainShardId {
		return createAPIComps(args)
	}

	return &APIProcessComps{
		StakingDataProviderAPI: factoryDisabled.NewDisabledStakingDataProvider(),
		AuctionListSelector:    factoryDisabled.NewDisabledAuctionListSelector(),
	}, nil
}

func createAPIComps(args ArgsCreateAPIProcessComps) (*APIProcessComps, error) {
	stakingDataProviderAPI, err := metachainEpochStart.NewStakingDataProvider(args.ArgsStakingDataProvider)
	if err != nil {
		return nil, err
	}

	maxNodesChangeConfigProviderAPI, err := notifier.NewNodesConfigProviderAPI(args.EpochNotifier, args.EnableEpochs)
	if err != nil {
		return nil, err
	}

	extendedShardCoordinator, castOk := args.ShardCoordinator.(metachainEpochStart.ShardCoordinatorHandler)
	if !castOk {
		return nil, fmt.Errorf("%w when trying to cast shard coordinator to extended shard coordinator", process.ErrWrongTypeAssertion)
	}

	argsAuctionListSelectorAPI := metachainEpochStart.AuctionListSelectorArgs{
		ShardCoordinator:             extendedShardCoordinator,
		StakingDataProvider:          stakingDataProviderAPI,
		MaxNodesChangeConfigProvider: maxNodesChangeConfigProviderAPI,
		SoftAuctionConfig:            args.SoftAuctionConfig,
		Denomination:                 args.Denomination,
		AuctionListDisplayHandler:    factoryDisabled.NewDisabledAuctionListDisplayer(),
	}

	auctionListSelectorAPI, err := metachainEpochStart.NewAuctionListSelector(argsAuctionListSelectorAPI)
	if err != nil {
		return nil, err
	}

	return &APIProcessComps{
		StakingDataProviderAPI: stakingDataProviderAPI,
		AuctionListSelector:    auctionListSelectorAPI,
	}, nil
}

func (c *apiProcessorCompsCreator) IsInterfaceNil() bool {
	return c == nil
}
