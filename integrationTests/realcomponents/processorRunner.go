package realcomponents

import (
	"crypto/rand"
	"io"
	"math/big"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/common/ordering"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	dbLookupFactory "github.com/multiversx/mx-chain-go/dblookupext/factory"
	"github.com/multiversx/mx-chain-go/factory"
	factoryBootstrap "github.com/multiversx/mx-chain-go/factory/bootstrap"
	factoryCore "github.com/multiversx/mx-chain-go/factory/core"
	factoryCrypto "github.com/multiversx/mx-chain-go/factory/crypto"
	factoryData "github.com/multiversx/mx-chain-go/factory/data"
	factoryNetwork "github.com/multiversx/mx-chain-go/factory/network"
	factoryProcessing "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/factory/runType"
	factoryState "github.com/multiversx/mx-chain-go/factory/state"
	factoryStatus "github.com/multiversx/mx-chain-go/factory/status"
	factoryStatusCore "github.com/multiversx/mx-chain-go/factory/statusCore"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
	nodesCoord "github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/cache"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/update/trigger"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

// ProcessorRunner is a test emulation to the nodeRunner component
type ProcessorRunner struct {
	closers              []io.Closer
	Config               config.Configs
	CoreComponents       factory.CoreComponentsHolder
	CryptoComponents     factory.CryptoComponentsHandler
	StatusCoreComponents factory.StatusCoreComponentsHolder
	NetworkComponents    factory.NetworkComponentsHolder
	BootstrapComponents  factory.BootstrapComponentsHolder
	DataComponents       factory.DataComponentsHolder
	StateComponents      factory.StateComponentsHolder
	NodesCoordinator     nodesCoord.NodesCoordinator
	StatusComponents     factory.StatusComponentsHolder
	ProcessComponents    factory.ProcessComponentsHolder
	RunTypeComponents    factory.RunTypeComponentsHolder
}

// NewProcessorRunner returns a new instance of ProcessorRunner
func NewProcessorRunner(tb testing.TB, config config.Configs) *ProcessorRunner {
	pr := &ProcessorRunner{
		Config:  config,
		closers: make([]io.Closer, 0),
	}

	pr.createComponents(tb)

	return pr
}

func (pr *ProcessorRunner) createComponents(tb testing.TB) {
	pr.createCoreComponents(tb)
	pr.createCryptoComponents(tb)
	pr.createRunTypeComponents(tb)
	pr.createStatusCoreComponents(tb)
	pr.createNetworkComponents(tb)
	pr.createBootstrapComponents(tb)
	pr.createDataComponents(tb)
	pr.createStateComponents(tb)
	pr.createStatusComponents(tb)
	pr.createProcessComponents(tb)
}

func (pr *ProcessorRunner) createRunTypeComponents(tb testing.TB) {
	initialAccounts, err := testscommon.ReadInitialAccounts(pr.Config.ConfigurationPathsHolder.Genesis)

	rtFactory, err := runType.NewRunTypeComponentsFactory(runType.ArgsRunTypeComponents{
		CoreComponents:   pr.CoreComponents,
		CryptoComponents: pr.CryptoComponents,
		Configs:          pr.Config,
		InitialAccounts:  initialAccounts,
	})
	require.Nil(tb, err)

	rtComp, err := runType.NewManagedRunTypeComponents(rtFactory)
	require.Nil(tb, err)

	err = rtComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, rtComp.CheckSubcomponents())

	pr.closers = append(pr.closers, rtComp)
	pr.RunTypeComponents = rtComp
}

func (pr *ProcessorRunner) createCoreComponents(tb testing.TB) {
	argsCore := factoryCore.CoreComponentsFactoryArgs{
		Config:                   *pr.Config.GeneralConfig,
		ConfigPathsHolder:        *pr.Config.ConfigurationPathsHolder,
		EpochConfig:              *pr.Config.EpochConfig,
		RoundConfig:              *pr.Config.RoundConfig,
		RatingsConfig:            *pr.Config.RatingsConfig,
		EconomicsConfig:          *pr.Config.EconomicsConfig,
		ImportDbConfig:           *pr.Config.ImportDbConfig,
		NodesFilename:            pr.Config.ConfigurationPathsHolder.Nodes,
		WorkingDirectory:         pr.Config.FlagsConfig.WorkingDir,
		ChanStopNodeProcess:      make(chan endProcess.ArgEndProcess),
		GenesisNodesSetupFactory: sharding.NewGenesisNodesSetupFactory(),
		RatingsDataFactory:       rating.NewRatingsDataFactory(),
	}
	coreFactory, err := factoryCore.NewCoreComponentsFactory(argsCore)
	require.Nil(tb, err)

	coreComp, err := factoryCore.NewManagedCoreComponents(coreFactory)
	require.Nil(tb, err)

	err = coreComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, coreComp.CheckSubcomponents())

	pr.closers = append(pr.closers, coreComp)
	pr.CoreComponents = coreComp
}

func (pr *ProcessorRunner) createCryptoComponents(tb testing.TB) {
	argsCrypto := factoryCrypto.CryptoComponentsFactoryArgs{
		ValidatorKeyPemFileName:              pr.Config.ConfigurationPathsHolder.ValidatorKey,
		AllValidatorKeysPemFileName:          pr.Config.ConfigurationPathsHolder.AllValidatorKeys,
		SkIndex:                              0,
		Config:                               *pr.Config.GeneralConfig,
		EnableEpochs:                         pr.Config.EpochConfig.EnableEpochs,
		PrefsConfig:                          *pr.Config.PreferencesConfig,
		CoreComponentsHolder:                 pr.CoreComponents,
		KeyLoader:                            core.NewKeyLoader(),
		ActivateBLSPubKeyMessageVerification: false,
		IsInImportMode:                       false,
		ImportModeNoSigCheck:                 false,
		P2pKeyPemFileName:                    "",
	}

	cryptoFactory, err := factoryCrypto.NewCryptoComponentsFactory(argsCrypto)
	require.Nil(tb, err)

	cryptoComp, err := factoryCrypto.NewManagedCryptoComponents(cryptoFactory)
	require.Nil(tb, err)

	err = cryptoComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, cryptoComp.CheckSubcomponents())

	pr.closers = append(pr.closers, cryptoComp)
	pr.CryptoComponents = cryptoComp
}

func (pr *ProcessorRunner) createStatusCoreComponents(tb testing.TB) {
	argsStatusCore := factoryStatusCore.StatusCoreComponentsFactoryArgs{
		Config:          *pr.Config.GeneralConfig,
		EpochConfig:     *pr.Config.EpochConfig,
		RoundConfig:     *pr.Config.RoundConfig,
		RatingsConfig:   *pr.Config.RatingsConfig,
		EconomicsConfig: *pr.Config.EconomicsConfig,
		CoreComp:        pr.CoreComponents,
	}

	statusCoreFactory, err := factoryStatusCore.NewStatusCoreComponentsFactory(argsStatusCore)
	require.Nil(tb, err)

	statusCoreComp, err := factoryStatusCore.NewManagedStatusCoreComponents(statusCoreFactory)
	require.Nil(tb, err)

	err = statusCoreComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, statusCoreComp.CheckSubcomponents())

	pr.closers = append(pr.closers, statusCoreComp)
	pr.StatusCoreComponents = statusCoreComp
}

func (pr *ProcessorRunner) createNetworkComponents(tb testing.TB) {
	argsNetwork := factoryNetwork.NetworkComponentsFactoryArgs{
		MainP2pConfig:         *pr.Config.MainP2pConfig,
		FullArchiveP2pConfig:  *pr.Config.FullArchiveP2pConfig,
		MainConfig:            *pr.Config.GeneralConfig,
		RatingsConfig:         *pr.Config.RatingsConfig,
		StatusHandler:         pr.StatusCoreComponents.AppStatusHandler(),
		Marshalizer:           pr.CoreComponents.InternalMarshalizer(),
		Syncer:                pr.CoreComponents.SyncTimer(),
		PreferredPeersSlices:  make([]string, 0),
		BootstrapWaitTime:     1,
		NodeOperationMode:     common.NormalOperation,
		ConnectionWatcherType: "",
		CryptoComponents:      pr.CryptoComponents,
	}

	networkFactory, err := factoryNetwork.NewNetworkComponentsFactory(argsNetwork)
	require.Nil(tb, err)

	networkComp, err := factoryNetwork.NewManagedNetworkComponents(networkFactory)
	require.Nil(tb, err)

	err = networkComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, networkComp.CheckSubcomponents())

	pr.closers = append(pr.closers, networkComp)
	pr.NetworkComponents = networkComp
}

func (pr *ProcessorRunner) createBootstrapComponents(tb testing.TB) {
	argsBootstrap := factoryBootstrap.BootstrapComponentsFactoryArgs{
		Config:               *pr.Config.GeneralConfig,
		RoundConfig:          *pr.Config.RoundConfig,
		PrefConfig:           *pr.Config.PreferencesConfig,
		ImportDbConfig:       *pr.Config.ImportDbConfig,
		FlagsConfig:          *pr.Config.FlagsConfig,
		WorkingDir:           pr.Config.FlagsConfig.WorkingDir,
		CoreComponents:       pr.CoreComponents,
		CryptoComponents:     pr.CryptoComponents,
		NetworkComponents:    pr.NetworkComponents,
		StatusCoreComponents: pr.StatusCoreComponents,
		RunTypeComponents:    pr.RunTypeComponents,
	}

	bootstrapFactory, err := factoryBootstrap.NewBootstrapComponentsFactory(argsBootstrap)
	require.Nil(tb, err)

	bootstrapComp, err := factoryBootstrap.NewManagedBootstrapComponents(bootstrapFactory)
	require.Nil(tb, err)

	err = bootstrapComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, bootstrapComp.CheckSubcomponents())

	pr.closers = append(pr.closers, bootstrapComp)
	pr.BootstrapComponents = bootstrapComp
}

func (pr *ProcessorRunner) createDataComponents(tb testing.TB) {
	argsData := factoryData.DataComponentsFactoryArgs{
		Config:                          *pr.Config.GeneralConfig,
		PrefsConfig:                     pr.Config.PreferencesConfig.Preferences,
		ShardCoordinator:                pr.BootstrapComponents.ShardCoordinator(),
		Core:                            pr.CoreComponents,
		StatusCore:                      pr.StatusCoreComponents,
		Crypto:                          pr.CryptoComponents,
		CurrentEpoch:                    0,
		CreateTrieEpochRootHashStorer:   false,
		NodeProcessingMode:              common.Normal,
		FlagsConfigs:                    config.ContextFlagsConfig{},
		AdditionalStorageServiceCreator: pr.RunTypeComponents.AdditionalStorageServiceCreator(),
	}

	dataFactory, err := factoryData.NewDataComponentsFactory(argsData)
	require.Nil(tb, err)

	dataComp, err := factoryData.NewManagedDataComponents(dataFactory)
	require.Nil(tb, err)

	err = dataComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, dataComp.CheckSubcomponents())

	pr.closers = append(pr.closers, dataComp)
	pr.DataComponents = dataComp
}

func (pr *ProcessorRunner) createStateComponents(tb testing.TB) {
	argsState := factoryState.StateComponentsFactoryArgs{
		Config:                   *pr.Config.GeneralConfig,
		Core:                     pr.CoreComponents,
		StatusCore:               pr.StatusCoreComponents,
		StorageService:           pr.DataComponents.StorageService(),
		ProcessingMode:           common.Normal,
		ShouldSerializeSnapshots: false,
		ChainHandler:             pr.DataComponents.Blockchain(),
		AccountsCreator:          pr.RunTypeComponents.AccountsCreator(),
	}

	stateFactory, err := factoryState.NewStateComponentsFactory(argsState)
	require.Nil(tb, err)

	stateComp, err := factoryState.NewManagedStateComponents(stateFactory)
	require.Nil(tb, err)

	err = stateComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, stateComp.CheckSubcomponents())

	pr.closers = append(pr.closers, stateComp)
	pr.StateComponents = stateComp
}

func (pr *ProcessorRunner) createStatusComponents(tb testing.TB) {
	nodesShufflerOut, err := factoryBootstrap.CreateNodesShuffleOut(
		pr.CoreComponents.GenesisNodesSetup(),
		pr.Config.GeneralConfig.EpochStartConfig,
		pr.CoreComponents.ChanStopNodeProcess(),
	)
	require.Nil(tb, err)

	bootstrapStorer, err := pr.DataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit)
	require.Nil(tb, err)

	pr.NodesCoordinator, err = factoryBootstrap.CreateNodesCoordinator(
		nodesShufflerOut,
		pr.CoreComponents.GenesisNodesSetup(),
		pr.Config.PreferencesConfig.Preferences,
		pr.CoreComponents.EpochStartNotifierWithConfirm(),
		pr.CryptoComponents.PublicKey(),
		pr.CoreComponents.InternalMarshalizer(),
		pr.CoreComponents.Hasher(),
		pr.CoreComponents.Rater(),
		bootstrapStorer,
		pr.CoreComponents.NodesShuffler(),
		pr.BootstrapComponents.ShardCoordinator().SelfId(),
		pr.BootstrapComponents.EpochBootstrapParams(),
		pr.BootstrapComponents.EpochBootstrapParams().Epoch(),
		pr.CoreComponents.ChanStopNodeProcess(),
		pr.CoreComponents.NodeTypeProvider(),
		pr.CoreComponents.EnableEpochsHandler(),
		pr.DataComponents.Datapool().CurrentEpochValidatorInfo(),
		pr.BootstrapComponents.NodesCoordinatorRegistryFactory(),
		pr.RunTypeComponents.NodesCoordinatorWithRaterCreator(),
	)
	require.Nil(tb, err)

	argsStatus := factoryStatus.StatusComponentsFactoryArgs{
		Config:               *pr.Config.GeneralConfig,
		ExternalConfig:       *pr.Config.ExternalConfig,
		EconomicsConfig:      *pr.Config.EconomicsConfig,
		ShardCoordinator:     pr.BootstrapComponents.ShardCoordinator(),
		NodesCoordinator:     pr.NodesCoordinator,
		EpochStartNotifier:   pr.CoreComponents.EpochStartNotifierWithConfirm(),
		CoreComponents:       pr.CoreComponents,
		StatusCoreComponents: pr.StatusCoreComponents,
		NetworkComponents:    pr.NetworkComponents,
		StateComponents:      pr.StateComponents,
		CryptoComponents:     pr.CryptoComponents,
		IsInImportMode:       false,
	}

	statusFactory, err := factoryStatus.NewStatusComponentsFactory(argsStatus)
	require.Nil(tb, err)

	statusComp, err := factoryStatus.NewManagedStatusComponents(statusFactory)
	require.Nil(tb, err)

	err = statusComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, statusComp.CheckSubcomponents())

	pr.closers = append(pr.closers, statusComp)
	pr.StatusComponents = statusComp
}

func (pr *ProcessorRunner) createProcessComponents(tb testing.TB) {
	whiteListCache, err := storageunit.NewCache(storageFactory.GetCacherFromConfig(pr.Config.GeneralConfig.WhiteListPool))
	require.Nil(tb, err)

	whiteListRequest, err := interceptors.NewWhiteListDataVerifier(whiteListCache)
	require.Nil(tb, err)

	whiteListCacheVerified, err := storageunit.NewCache(storageFactory.GetCacherFromConfig(pr.Config.GeneralConfig.WhiteListerVerifiedTxs))
	require.Nil(tb, err)

	whiteListerVerifiedTxs, err := interceptors.NewWhiteListDataVerifier(whiteListCacheVerified)
	require.Nil(tb, err)

	smartContractParser, err := parsing.NewSmartContractsParser(
		pr.Config.ConfigurationPathsHolder.SmartContracts,
		pr.CoreComponents.AddressPubKeyConverter(),
		pr.CryptoComponents.TxSignKeyGen(),
	)
	require.Nil(tb, err)

	argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig:  pr.Config.EpochConfig.GasSchedule,
		ConfigDir:          pr.Config.ConfigurationPathsHolder.GasScheduleDirectoryName,
		EpochNotifier:      pr.CoreComponents.EpochNotifier(),
		WasmVMChangeLocker: pr.CoreComponents.WasmVMChangeLocker(),
	}
	gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasScheduleNotifier)
	require.Nil(tb, err)

	historyRepoFactoryArgs := &dbLookupFactory.ArgsHistoryRepositoryFactory{
		SelfShardID:              pr.BootstrapComponents.ShardCoordinator().SelfId(),
		Config:                   pr.Config.GeneralConfig.DbLookupExtensions,
		Hasher:                   pr.CoreComponents.Hasher(),
		Marshalizer:              pr.CoreComponents.InternalMarshalizer(),
		Store:                    pr.DataComponents.StorageService(),
		Uint64ByteSliceConverter: pr.CoreComponents.Uint64ByteSliceConverter(),
	}
	historyRepositoryFactory, err := dbLookupFactory.NewHistoryRepositoryFactory(historyRepoFactoryArgs)
	require.Nil(tb, err)

	historyRepository, err := historyRepositoryFactory.Create()
	require.Nil(tb, err)

	requestedItemsHandler := cache.NewTimeCache(
		time.Duration(uint64(time.Millisecond) * pr.CoreComponents.GenesisNodesSetup().GetRoundDuration()))

	importStartHandler, err := trigger.NewImportStartHandler(filepath.Join(pr.Config.FlagsConfig.DbDir, common.DefaultDBPath), pr.Config.FlagsConfig.Version)
	require.Nil(tb, err)

	txExecutionOrderHandler := ordering.NewOrderedCollection()

	argsProcess := factoryProcessing.ProcessComponentsFactoryArgs{
		Config:         *pr.Config.GeneralConfig,
		EpochConfig:    *pr.Config.EpochConfig,
		RoundConfig:    *pr.Config.RoundConfig,
		PrefConfigs:    *pr.Config.PreferencesConfig,
		ImportDBConfig: *pr.Config.ImportDbConfig,
		FlagsConfig: config.ContextFlagsConfig{
			Version:    "test",
			WorkingDir: pr.Config.FlagsConfig.WorkingDir,
		},
		SmartContractParser:     smartContractParser,
		GasSchedule:             gasScheduleNotifier,
		NodesCoordinator:        pr.NodesCoordinator,
		RequestedItemsHandler:   requestedItemsHandler,
		WhiteListHandler:        whiteListRequest,
		WhiteListerVerifiedTxs:  whiteListerVerifiedTxs,
		MaxRating:               pr.Config.RatingsConfig.General.MaxRating,
		SystemSCConfig:          pr.Config.SystemSCConfig,
		ImportStartHandler:      importStartHandler,
		HistoryRepo:             historyRepository,
		Data:                    pr.DataComponents,
		CoreData:                pr.CoreComponents,
		Crypto:                  pr.CryptoComponents,
		State:                   pr.StateComponents,
		Network:                 pr.NetworkComponents,
		BootstrapComponents:     pr.BootstrapComponents,
		StatusComponents:        pr.StatusComponents,
		StatusCoreComponents:    pr.StatusCoreComponents,
		TxExecutionOrderHandler: txExecutionOrderHandler,
		RunTypeComponents:       pr.RunTypeComponents,
	}

	processFactory, err := factoryProcessing.NewProcessComponentsFactory(argsProcess)
	require.Nil(tb, err)

	processComp, err := factoryProcessing.NewManagedProcessComponents(processFactory)
	require.Nil(tb, err)

	err = processComp.Create()
	require.Nil(tb, err)
	require.Nil(tb, processComp.CheckSubcomponents())

	pr.closers = append(pr.closers, processComp)
	pr.ProcessComponents = processComp
}

// Close will close all inner components
func (pr *ProcessorRunner) Close(tb testing.TB) {
	for i := len(pr.closers) - 1; i >= 0; i-- {
		err := pr.closers[i].Close()
		require.Nil(tb, err)
	}
}

// GenerateAddress will generate an address for the given shardID
func (pr *ProcessorRunner) GenerateAddress(shardID uint32) []byte {
	address := make([]byte, 32)

	for {
		_, _ = rand.Read(address)
		if pr.BootstrapComponents.ShardCoordinator().ComputeId(address) == shardID {
			return address
		}
	}
}

// AddBalanceToAccount will add the provided balance to the account
func (pr *ProcessorRunner) AddBalanceToAccount(tb testing.TB, address []byte, balanceToAdd *big.Int) {
	userAccount := pr.GetUserAccount(tb, address)

	err := userAccount.AddToBalance(balanceToAdd)
	require.Nil(tb, err)

	err = pr.StateComponents.AccountsAdapter().SaveAccount(userAccount)
	require.Nil(tb, err)
}

// GetUserAccount will return the user account for the provided address
func (pr *ProcessorRunner) GetUserAccount(tb testing.TB, address []byte) state.UserAccountHandler {
	acc, err := pr.StateComponents.AccountsAdapter().LoadAccount(address)
	require.Nil(tb, err)

	userAccount, ok := acc.(state.UserAccountHandler)
	require.True(tb, ok)

	return userAccount
}

// SetESDTForAccount will set the provided ESDT balance to the account
func (pr *ProcessorRunner) SetESDTForAccount(
	tb testing.TB,
	address []byte,
	tokenIdentifier string,
	esdtNonce uint64,
	esdtValue *big.Int,
) {
	userAccount := pr.GetUserAccount(tb, address)

	esdtData := &esdt.ESDigitalToken{
		Value:      esdtValue,
		Properties: []byte{},
	}

	esdtDataBytes, err := pr.CoreComponents.InternalMarshalizer().Marshal(esdtData)
	require.Nil(tb, err)

	key := append([]byte(core.ProtectedKeyPrefix), []byte(core.ESDTKeyIdentifier)...)
	key = append(key, tokenIdentifier...)
	if esdtNonce > 0 {
		key = append(key, big.NewInt(0).SetUint64(esdtNonce).Bytes()...)
	}

	err = userAccount.SaveKeyValue(key, esdtDataBytes)
	require.Nil(tb, err)

	err = pr.StateComponents.AccountsAdapter().SaveAccount(userAccount)
	require.Nil(tb, err)

	pr.saveNewTokenOnSystemAccount(tb, key, esdtData)

	_, err = pr.StateComponents.AccountsAdapter().Commit()
	require.Nil(tb, err)
}

func (pr *ProcessorRunner) saveNewTokenOnSystemAccount(tb testing.TB, tokenKey []byte, esdtData *esdt.ESDigitalToken) {
	esdtDataOnSystemAcc := esdtData
	esdtDataOnSystemAcc.Properties = nil
	esdtDataOnSystemAcc.Reserved = []byte{1}
	esdtDataOnSystemAcc.Value.Set(esdtData.Value)

	esdtDataBytes, err := pr.CoreComponents.InternalMarshalizer().Marshal(esdtData)
	require.Nil(tb, err)

	sysAccount, err := pr.StateComponents.AccountsAdapter().LoadAccount(core.SystemAccountAddress)
	require.Nil(tb, err)

	sysUserAccount, ok := sysAccount.(state.UserAccountHandler)
	require.True(tb, ok)

	err = sysUserAccount.SaveKeyValue(tokenKey, esdtDataBytes)
	require.Nil(tb, err)

	err = pr.StateComponents.AccountsAdapter().SaveAccount(sysAccount)
	require.Nil(tb, err)
}

// ExecuteTransactionAsScheduled will execute the provided transaction as scheduled
func (pr *ProcessorRunner) ExecuteTransactionAsScheduled(tb testing.TB, tx *transaction.Transaction) error {
	hash, err := core.CalculateHash(pr.CoreComponents.InternalMarshalizer(), pr.CoreComponents.Hasher(), tx)
	require.Nil(tb, err)
	pr.ProcessComponents.ScheduledTxsExecutionHandler().AddScheduledTx(hash, tx)

	return pr.ProcessComponents.ScheduledTxsExecutionHandler().Execute(hash)
}

// CreateDeploySCTx will return the transaction and the hash for the deployment smart-contract transaction
func (pr *ProcessorRunner) CreateDeploySCTx(
	tb testing.TB,
	owner []byte,
	contractPath string,
	gasLimit uint64,
	initialHexParameters []string,
) (*transaction.Transaction, []byte) {
	scCode := wasm.GetSCCode(contractPath)
	ownerAccount := pr.GetUserAccount(tb, owner)

	txDataComponents := append([]string{wasm.CreateDeployTxData(scCode)}, initialHexParameters...)

	tx := &transaction.Transaction{
		Nonce:    ownerAccount.GetNonce(),
		Value:    big.NewInt(0),
		RcvAddr:  vm.CreateEmptyAddress(),
		SndAddr:  owner,
		GasPrice: pr.CoreComponents.EconomicsData().MinGasPrice(),
		GasLimit: gasLimit,
		Data:     []byte(strings.Join(txDataComponents, "@")),
	}

	hash, err := core.CalculateHash(pr.CoreComponents.InternalMarshalizer(), pr.CoreComponents.Hasher(), tx)
	require.Nil(tb, err)

	return tx, hash
}
