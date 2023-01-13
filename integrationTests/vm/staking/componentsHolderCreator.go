package staking

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/nodetype"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/integrationTests"
	integrationMocks "github.com/multiversx/mx-chain-go/integrationTests/mock"
	mockFactory "github.com/multiversx/mx-chain-go/node/mock/factory"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	stateFactory "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/stakingcommon"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/trie"
)

func createComponentHolders(numOfShards uint32) (
	factory.CoreComponentsHolder,
	factory.DataComponentsHolder,
	factory.BootstrapComponentsHolder,
	factory.StatusComponentsHolder,
	factory.StateComponentsHandler,
) {
	coreComponents := createCoreComponents()
	statusComponents := createStatusComponents()
	stateComponents := createStateComponents(coreComponents)
	dataComponents := createDataComponents(coreComponents, numOfShards)
	bootstrapComponents := createBootstrapComponents(coreComponents.InternalMarshalizer(), numOfShards)

	return coreComponents, dataComponents, bootstrapComponents, statusComponents, stateComponents
}

func createCoreComponents() factory.CoreComponentsHolder {
	return &integrationMocks.CoreComponentsStub{
		InternalMarshalizerField:           &marshal.GogoProtoMarshalizer{},
		HasherField:                        sha256.NewSha256(),
		Uint64ByteSliceConverterField:      uint64ByteSlice.NewBigEndianConverter(),
		StatusHandlerField:                 statusHandler.NewStatusMetrics(),
		RoundHandlerField:                  &mock.RoundHandlerMock{RoundTimeDuration: time.Second},
		EpochStartNotifierWithConfirmField: notifier.NewEpochStartSubscriptionHandler(),
		EpochNotifierField:                 forking.NewGenericEpochNotifier(),
		RaterField:                         &testscommon.RaterMock{Chance: 5},
		AddressPubKeyConverterField:        testscommon.NewPubkeyConverterMock(addressLength),
		EconomicsDataField:                 stakingcommon.CreateEconomicsData(),
		ChanStopNodeProcessField:           endProcess.GetDummyEndProcessChannel(),
		NodeTypeProviderField:              nodetype.NewNodeTypeProvider(core.NodeTypeValidator),
		ProcessStatusHandlerInternal:       statusHandler.NewProcessStatusHandler(),
	}
}

func createDataComponents(coreComponents factory.CoreComponentsHolder, numOfShards uint32) factory.DataComponentsHolder {
	genesisBlock := createGenesisMetaBlock()
	genesisBlockHash, _ := coreComponents.InternalMarshalizer().Marshal(genesisBlock)
	genesisBlockHash = coreComponents.Hasher().Compute(string(genesisBlockHash))

	blockChain, _ := blockchain.NewMetaChain(coreComponents.StatusHandler())
	_ = blockChain.SetGenesisHeader(createGenesisMetaBlock())
	blockChain.SetGenesisHeaderHash(genesisBlockHash)

	chainStorer := dataRetriever.NewChainStorer()
	chainStorer.AddStorer(dataRetriever.BootstrapUnit, integrationTests.CreateMemUnit())
	chainStorer.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, integrationTests.CreateMemUnit())
	chainStorer.AddStorer(dataRetriever.MetaBlockUnit, integrationTests.CreateMemUnit())
	chainStorer.AddStorer(dataRetriever.MiniBlockUnit, integrationTests.CreateMemUnit())
	chainStorer.AddStorer(dataRetriever.BlockHeaderUnit, integrationTests.CreateMemUnit())
	for i := uint32(0); i < numOfShards; i++ {
		unit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		chainStorer.AddStorer(unit, integrationTests.CreateMemUnit())
	}

	return &mockFactory.DataComponentsMock{
		Store:         chainStorer,
		DataPool:      dataRetrieverMock.NewPoolsHolderMock(),
		BlockChain:    blockChain,
		EconomicsData: coreComponents.EconomicsData(),
	}
}

func createBootstrapComponents(
	marshaller marshal.Marshalizer,
	numOfShards uint32,
) factory.BootstrapComponentsHolder {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(numOfShards, core.MetachainShardId)
	ncr, _ := nodesCoordinator.NewNodesCoordinatorRegistryFactory(
		marshaller,
		stakingV4EnableEpoch,
	)

	return &mainFactoryMocks.BootstrapComponentsStub{
		ShCoordinator:        shardCoordinator,
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32) data.HeaderHandler {
				return &block.MetaBlock{Epoch: epoch}
			},
		},
		NodesCoordinatorRegistryFactoryField: ncr,
	}
}

func createStatusComponents() factory.StatusComponentsHolder {
	return &integrationMocks.StatusComponentsStub{
		Outport:          &testscommon.OutportStub{},
		AppStatusHandler: &statusHandlerMock.AppStatusHandlerStub{},
	}
}

func createStateComponents(coreComponents factory.CoreComponentsHolder) factory.StateComponentsHandler {
	trieFactoryManager, _ := trie.NewTrieStorageManagerWithoutPruning(integrationTests.CreateMemUnit())
	userAccountsDB := createAccountsDB(coreComponents, stateFactory.NewAccountCreator(), trieFactoryManager)
	peerAccountsDB := createAccountsDB(coreComponents, stateFactory.NewPeerAccountCreator(), trieFactoryManager)

	return &testscommon.StateComponentsMock{
		PeersAcc: peerAccountsDB,
		Accounts: userAccountsDB,
	}
}

func createAccountsDB(
	coreComponents factory.CoreComponentsHolder,
	accountFactory state.AccountFactory,
	trieStorageManager common.StorageManager,
) *state.AccountsDB {
	tr, _ := trie.NewTrie(trieStorageManager, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), 5)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(10, testscommon.NewMemDbMock(), coreComponents.InternalMarshalizer())
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)

	argsAccountsDb := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                coreComponents.Hasher(),
		Marshaller:            coreComponents.InternalMarshalizer(),
		AccountFactory:        accountFactory,
		StoragePruningManager: spm,
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  coreComponents.ProcessStatusHandler(),
	}
	adb, _ := state.NewAccountsDB(argsAccountsDb)
	return adb
}
