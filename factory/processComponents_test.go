package factory_test

import (
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// TODO: write unit tests

func getProcessArgs() factory.ProcessComponentsFactoryArgs {
	coreArgs := getCoreArgs()
	return factory.ProcessComponentsFactoryArgs{
		CoreFactoryArgs:           &coreArgs,
		AccountsParser:            &mock.AccountsParserStub{},
		SmartContractParser:       &mock.SmartContractParserStub{},
		EconomicsData:             &economics.EconomicsData{},
		NodesConfig:               &sharding.NodesSetup{},
		GasSchedule:               make(map[string]map[string]uint64),
		Rounder:                   &mock.RounderMock{},
		ShardCoordinator:          mock.NewMultiShardsCoordinatorMock(2),
		NodesCoordinator:          &mock.NodesCoordinatorMock{},
		Data:                      &mock.DataComponentsMock{},
		CoreData:                  &mock.CoreComponentsMock{},
		Crypto:                    &mock.CryptoComponentsMock{},
		State:                     nil,
		Network:                   nil,
		CoreServiceContainer:      nil,
		RequestedItemsHandler:     nil,
		WhiteListHandler:          nil,
		WhiteListerVerifiedTxs:    nil,
		EpochStartNotifier:        nil,
		EpochStart:                nil,
		Rater:                     nil,
		RatingsData:               nil,
		SizeCheckDelta:            0,
		StateCheckpointModulus:    0,
		MaxComputableRounds:       0,
		NumConcurrentResolverJobs: 0,
		MinSizeInBytes:            0,
		MaxSizeInBytes:            0,
		MaxRating:                 0,
		ValidatorPubkeyConverter:  nil,
		SystemSCConfig:            nil,
		Version:                   "",
	}
}
