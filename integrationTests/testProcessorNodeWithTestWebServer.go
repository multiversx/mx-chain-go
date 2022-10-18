package integrationTests

import (
	"net/http"
	"net/http/httptest"
	"sync"

	arwenConfig "github.com/ElrondNetwork/wasm-vm/config"
	dataTransaction "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/config"
	nodeFacade "github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/external/blockAPI"
	"github.com/ElrondNetwork/elrond-go/node/external/transactionAPI"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators/factory"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
	txSimData "github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genesisMocks"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	datafield "github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
		"node":        {"/status", "/metrics", "/heartbeatstatus", "/statistics", "/p2pstatus", "/debug", "/peerinfo"},
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
	gasMap := arwenConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasMap, 1)
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasMap)
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:               gasScheduleNotifier,
		MapDNSAddresses:           make(map[string]struct{}),
		Marshalizer:               TestMarshalizer,
		Accounts:                  tpn.AccntState,
		ShardCoordinator:          tpn.ShardCoordinator,
		EpochNotifier:             tpn.EpochNotifier,
		EnableEpochsHandler:       tpn.EnableEpochsHandler,
		MaxNumNodesInTransferRole: 100,
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
		AddressLength:    TestAddressPubkeyConverter.Len(),
		Marshalizer:      TestMarshalizer,
		ShardCoordinator: tpn.ShardCoordinator,
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
		ChainID:                  string(tpn.ChainID),
	}
	apiTransactionHandler, err := transactionAPI.NewAPITransactionProcessor(argsApiTransactionProc)
	log.LogIfError(err)

	statusCom, err := txstatus.NewStatusComputer(tpn.ShardCoordinator.SelfId(), TestUint64Converter, tpn.Storage)
	log.LogIfError(err)

	argsBlockAPI := &blockAPI.ArgAPIBlockProcessor{
		SelfShardID:              tpn.ShardCoordinator.SelfId(),
		Store:                    tpn.Storage,
		Marshalizer:              TestMarshalizer,
		Uint64ByteSliceConverter: TestUint64Converter,
		HistoryRepo:              tpn.HistoryRepository,
		APITransactionHandler:    apiTransactionHandler,
		StatusComputer:           statusCom,
		Hasher:                   TestHasher,
		AddressPubkeyConverter:   TestAddressPubkeyConverter,
		LogsFacade:               logsFacade,
		ReceiptsRepository:       receiptsRepository,
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
