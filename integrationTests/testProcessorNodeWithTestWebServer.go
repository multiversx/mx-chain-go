package integrationTests

import (
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	dataTransaction "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
	nodeFacade "github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/external/blockAPI"
	"github.com/multiversx/mx-chain-go/node/external/transactionAPI"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	"github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/smartContract/builtInFunctions"
	"github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/process/txsimulator"
	txSimData "github.com/multiversx/mx-chain-go/process/txsimulator/data"
	"github.com/multiversx/mx-chain-go/process/txstatus"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
)

// TestProcessorNodeWithTestWebServer represents a TestProcessorNode with a test web server
type TestProcessorNodeWithTestWebServer struct {
	*TestProcessorNode
	facade Facade
	mutWs  sync.Mutex
	ws     *gin.Engine
}

// NewTestProcessorNodeWithTestWebServer returns a new TestProcessorNodeWithTestWebServer instance with a libp2p messenger
func NewTestProcessorNodeWithTestWebServer(
	maxShards uint32,
	nodeShardId uint32,
	txSignPrivKeyShardId uint32,
) *TestProcessorNodeWithTestWebServer {

	tpn := NewTestProcessorNode(ArgTestProcessorNode{
		MaxShards:            maxShards,
		NodeShardId:          nodeShardId,
		TxSignPrivKeyShardId: txSignPrivKeyShardId,
	})

	argFacade := createFacadeArg(tpn)
	facade, err := nodeFacade.NewNodeFacade(argFacade)
	log.LogIfError(err)

	ws := createGinServer(facade, argFacade.ApiRoutesConfig)

	return &TestProcessorNodeWithTestWebServer{
		TestProcessorNode: tpn,
		facade:            facade,
		ws:                ws,
	}
}

// DoRequest preforms a test request on the web server, returning the response ready to be parsed
func (node *TestProcessorNodeWithTestWebServer) DoRequest(request *http.Request) *httptest.ResponseRecorder {
	// this is a critical section, serialize each request
	node.mutWs.Lock()
	defer node.mutWs.Unlock()

	resp := httptest.NewRecorder()
	node.ws.ServeHTTP(resp, request)

	return resp
}

func createFacadeArg(tpn *TestProcessorNode) nodeFacade.ArgNodeFacade {
	apiResolver, txSimulator := createFacadeComponents(tpn)

	return nodeFacade.ArgNodeFacade{
		Node:                   tpn.Node,
		ApiResolver:            apiResolver,
		TxSimulatorProcessor:   txSimulator,
		RestAPIServerDebugMode: false,
		WsAntifloodConfig: config.WebServerAntifloodConfig{
			SimultaneousRequests:               1000,
			SameSourceRequests:                 1000,
			SameSourceResetIntervalInSec:       1,
			TrieOperationsDeadlineMilliseconds: 1,
			EndpointsThrottlers:                []config.EndpointsThrottlersConfig{},
		},
		FacadeConfig:    config.FacadeConfig{},
		ApiRoutesConfig: createTestApiConfig(),
		AccountsState:   tpn.AccntState,
		PeerState:       tpn.PeerState,
		Blockchain:      tpn.BlockChain,
	}
}

func createTestApiConfig() config.ApiRoutesConfig {
	routes := map[string][]string{
		"node":        {"/status", "/metrics", "/heartbeatstatus", "/statistics", "/p2pstatus", "/debug", "/peerinfo", "/bootstrapstatus", "/connected-peers-ratings", "/managed-keys/count", "/managed-keys", "/managed-keys/eligible", "/managed-keys/waiting"},
		"address":     {"/:address", "/:address/balance", "/:address/username", "/:address/code-hash", "/:address/key/:key", "/:address/esdt", "/:address/esdt/:tokenIdentifier"},
		"hardfork":    {"/trigger"},
		"network":     {"/status", "/total-staked", "/economics", "/config"},
		"log":         {"/log"},
		"validator":   {"/statistics"},
		"vm-values":   {"/hex", "/string", "/int", "/query"},
		"transaction": {"/send", "/simulate", "/send-multiple", "/cost", "/:txhash", "/pool"},
		"block":       {"/by-nonce/:nonce", "/by-hash/:hash", "/by-round/:round"},
	}

	routesConfig := config.ApiRoutesConfig{
		APIPackages: make(map[string]config.APIPackageConfig),
	}

	for name, endpoints := range routes {
		packageConfig := config.APIPackageConfig{}
		for _, routeName := range endpoints {
			route := config.RouteConfig{
				Name: routeName,
				Open: true,
			}
			packageConfig.Routes = append(packageConfig.Routes, route)
		}

		routesConfig.APIPackages[name] = packageConfig
	}

	return routesConfig
}

func createFacadeComponents(tpn *TestProcessorNode) (nodeFacade.ApiResolver, nodeFacade.TransactionSimulatorProcessor) {
	gasMap := wasmConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasMap, 1)
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasMap)
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:               gasScheduleNotifier,
		MapDNSAddresses:           make(map[string]struct{}),
		MapDNSV2Addresses:         make(map[string]struct{}),
		Marshalizer:               TestMarshalizer,
		Accounts:                  tpn.AccntState,
		ShardCoordinator:          tpn.ShardCoordinator,
		EpochNotifier:             tpn.EpochNotifier,
		EnableEpochsHandler:       tpn.EnableEpochsHandler,
		MaxNumNodesInTransferRole: 100,
		GuardedAccountHandler:     tpn.GuardedAccountHandler,
	}
	argsBuiltIn.AutomaticCrawlerAddresses = GenerateOneAddressPerShard(argsBuiltIn.ShardCoordinator)
	builtInFuncs, err := builtInFunctions.CreateBuiltInFunctionsFactory(argsBuiltIn)
	log.LogIfError(err)
	esdtTransferParser, _ := parsers.NewESDTTransferParser(TestMarshalizer)
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     TestAddressPubkeyConverter,
		ShardCoordinator:    tpn.ShardCoordinator,
		BuiltInFunctions:    builtInFuncs.BuiltInFunctionContainer(),
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtTransferParser,
		EnableEpochsHandler: tpn.EnableEpochsHandler,
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	log.LogIfError(err)

	txCostHandler, err := transaction.NewTransactionCostEstimator(
		txTypeHandler,
		tpn.EconomicsData,
		&mock.TransactionSimulatorStub{
			ProcessTxCalled: func(tx *dataTransaction.Transaction) (*txSimData.SimulationResults, error) {
				return &txSimData.SimulationResults{}, nil
			},
		},
		tpn.AccntState,
		tpn.ShardCoordinator,
		tpn.EnableEpochsHandler,
	)
	log.LogIfError(err)

	accountsWrapper := &trieIterators.AccountsWrapper{
		Mutex:           &sync.Mutex{},
		AccountsAdapter: tpn.AccntState,
	}

	args := trieIterators.ArgTrieIteratorProcessor{
		ShardID:            tpn.ShardCoordinator.SelfId(),
		Accounts:           accountsWrapper,
		QueryService:       tpn.SCQueryService,
		PublicKeyConverter: TestAddressPubkeyConverter,
	}
	totalStakedValueHandler, err := factory.CreateTotalStakedValueHandler(args)
	log.LogIfError(err)

	directStakedListHandler, err := factory.CreateDirectStakedListHandler(args)
	log.LogIfError(err)

	delegatedListHandler, err := factory.CreateDelegatedListHandler(args)
	log.LogIfError(err)

	logsFacade := &testscommon.LogsFacadeStub{}
	receiptsRepository := &testscommon.ReceiptsRepositoryStub{}

	argsDataFieldParser := &datafield.ArgsOperationDataFieldParser{
		AddressLength: TestAddressPubkeyConverter.Len(),
		Marshalizer:   TestMarshalizer,
	}
	dataFieldParser, err := datafield.NewOperationDataFieldParser(argsDataFieldParser)
	log.LogIfError(err)

	argsApiTransactionProc := &transactionAPI.ArgAPITransactionProcessor{
		Marshalizer:              TestMarshalizer,
		AddressPubKeyConverter:   TestAddressPubkeyConverter,
		ShardCoordinator:         tpn.ShardCoordinator,
		HistoryRepository:        tpn.HistoryRepository,
		StorageService:           tpn.Storage,
		DataPool:                 tpn.DataPool,
		Uint64ByteSliceConverter: TestUint64Converter,
		FeeComputer:              &testscommon.FeeComputerStub{},
		TxTypeHandler:            txTypeHandler,
		LogsFacade:               logsFacade,
		DataFieldParser:          dataFieldParser,
	}
	apiTransactionHandler, err := transactionAPI.NewAPITransactionProcessor(argsApiTransactionProc)
	log.LogIfError(err)

	statusCom, err := txstatus.NewStatusComputer(tpn.ShardCoordinator.SelfId(), TestUint64Converter, tpn.Storage)
	log.LogIfError(err)

	argsBlockAPI := &blockAPI.ArgAPIBlockProcessor{
		SelfShardID:                  tpn.ShardCoordinator.SelfId(),
		Store:                        tpn.Storage,
		Marshalizer:                  TestMarshalizer,
		Uint64ByteSliceConverter:     TestUint64Converter,
		HistoryRepo:                  tpn.HistoryRepository,
		APITransactionHandler:        apiTransactionHandler,
		StatusComputer:               statusCom,
		Hasher:                       TestHasher,
		AddressPubkeyConverter:       TestAddressPubkeyConverter,
		LogsFacade:                   logsFacade,
		ReceiptsRepository:           receiptsRepository,
		AlteredAccountsProvider:      &testscommon.AlteredAccountsProviderStub{},
		AccountsRepository:           &state.AccountsRepositoryStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	blockAPIHandler, err := blockAPI.CreateAPIBlockProcessor(argsBlockAPI)
	log.LogIfError(err)

	apiInternalBlockProcessor, err := blockAPI.CreateAPIInternalBlockProcessor(argsBlockAPI)
	log.LogIfError(err)

	argsApiResolver := external.ArgNodeApiResolver{
		SCQueryService:           tpn.SCQueryService,
		StatusMetricsHandler:     &testscommon.StatusMetricsStub{},
		TxCostHandler:            txCostHandler,
		TotalStakedValueHandler:  totalStakedValueHandler,
		DirectStakedListHandler:  directStakedListHandler,
		DelegatedListHandler:     delegatedListHandler,
		APITransactionHandler:    apiTransactionHandler,
		APIBlockHandler:          blockAPIHandler,
		APIInternalBlockHandler:  apiInternalBlockProcessor,
		GenesisNodesSetupHandler: &mock.NodesSetupStub{},
		ValidatorPubKeyConverter: &testscommon.PubkeyConverterMock{},
		AccountsParser:           &genesisMocks.AccountsParserStub{},
		GasScheduleNotifier:      &testscommon.GasScheduleNotifierMock{},
		ManagedPeersMonitor:      &testscommon.ManagedPeersMonitorStub{},
	}

	apiResolver, err := external.NewNodeApiResolver(argsApiResolver)
	log.LogIfError(err)

	argSimulator := txsimulator.ArgsTxSimulator{
		TransactionProcessor:      tpn.TxProcessor,
		IntermediateProcContainer: tpn.InterimProcContainer,
		AddressPubKeyConverter:    TestAddressPubkeyConverter,
		ShardCoordinator:          tpn.ShardCoordinator,
		Marshalizer:               TestMarshalizer,
		Hasher:                    TestHasher,
		VMOutputCacher:            &testscommon.CacherMock{},
	}

	txSimulator, err := txsimulator.NewTransactionSimulator(argSimulator)
	log.LogIfError(err)

	return apiResolver, txSimulator
}

func createGinServer(facade Facade, apiConfig config.ApiRoutesConfig) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())

	groupsMap := createGroups(facade)
	for groupName, groupHandler := range groupsMap {
		ginGroup := ws.Group(groupName)
		groupHandler.RegisterRoutes(ginGroup, apiConfig)
	}

	return ws
}

func createGroups(facade Facade) map[string]shared.GroupHandler {
	groupsMap := make(map[string]shared.GroupHandler)
	addressGroup, err := groups.NewAddressGroup(facade)
	if err == nil {
		groupsMap["address"] = addressGroup
	}

	blockGroup, err := groups.NewBlockGroup(facade)
	if err == nil {
		groupsMap["block"] = blockGroup
	}

	hardforkGroup, err := groups.NewHardforkGroup(facade)
	if err == nil {
		groupsMap["hardfork"] = hardforkGroup
	}

	networkGroup, err := groups.NewNetworkGroup(facade)
	if err == nil {
		groupsMap["network"] = networkGroup
	}

	nodeGroup, err := groups.NewNodeGroup(facade)
	if err == nil {
		groupsMap["node"] = nodeGroup
	}

	proofGroup, err := groups.NewProofGroup(facade)
	if err == nil {
		groupsMap["proof"] = proofGroup
	}

	transactionGroup, err := groups.NewTransactionGroup(facade)
	if err == nil {
		groupsMap["transaction"] = transactionGroup
	}

	validatorGroup, err := groups.NewValidatorGroup(facade)
	if err == nil {
		groupsMap["validator"] = validatorGroup
	}

	vmValuesGroup, err := groups.NewVmValuesGroup(facade)
	if err == nil {
		groupsMap["vm-values"] = vmValuesGroup
	}

	return groupsMap
}
