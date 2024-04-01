package components

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos/sposFactory"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	dataRetrieverFactory "github.com/multiversx/mx-chain-go/dataRetriever/factory"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/factory"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
)

// ArgsTestOnlyProcessingNode represents the DTO struct for the NewTestOnlyProcessingNode constructor function
type ArgsTestOnlyProcessingNode struct {
	Configs      config.Configs
	APIInterface APIConfigurator

	ChanStopNodeProcess    chan endProcess.ArgEndProcess
	SyncedBroadcastNetwork SyncedBroadcastNetworkHandler

	InitialRound           int64
	InitialNonce           uint64
	GasScheduleFilename    string
	NumShards              uint32
	ShardIDStr             string
	BypassTxSignatureCheck bool
	MinNodesPerShard       uint32
	MinNodesMeta           uint32
	RoundDurationInMillis  uint64
}

type testOnlyProcessingNode struct {
	closeHandler              *closeHandler
	CoreComponentsHolder      factory.CoreComponentsHandler
	StatusCoreComponents      factory.StatusCoreComponentsHandler
	StateComponentsHolder     factory.StateComponentsHandler
	StatusComponentsHolder    factory.StatusComponentsHandler
	CryptoComponentsHolder    factory.CryptoComponentsHandler
	NetworkComponentsHolder   factory.NetworkComponentsHandler
	BootstrapComponentsHolder factory.BootstrapComponentsHandler
	ProcessComponentsHolder   factory.ProcessComponentsHandler
	DataComponentsHolder      factory.DataComponentsHandler

	NodesCoordinator      nodesCoordinator.NodesCoordinator
	ChainHandler          chainData.ChainHandler
	ArgumentsParser       process.ArgumentsParser
	TransactionFeeHandler process.TransactionFeeHandler
	StoreService          dataRetriever.StorageService
	DataPool              dataRetriever.PoolsHolder
	broadcastMessenger    consensus.BroadcastMessenger

	httpServer    shared.UpgradeableHttpServerHandler
	facadeHandler shared.FacadeHandler
}

// NewTestOnlyProcessingNode creates a new instance of a node that is able to only process transactions
func NewTestOnlyProcessingNode(args ArgsTestOnlyProcessingNode) (*testOnlyProcessingNode, error) {
	instance := &testOnlyProcessingNode{
		ArgumentsParser: smartContract.NewArgumentParser(),
		StoreService:    CreateStore(args.NumShards),
		closeHandler:    NewCloseHandler(),
	}

	var err error
	instance.TransactionFeeHandler = postprocess.NewFeeAccumulator()

	instance.CoreComponentsHolder, err = CreateCoreComponents(ArgsCoreComponentsHolder{
		Config:              *args.Configs.GeneralConfig,
		EnableEpochsConfig:  args.Configs.EpochConfig.EnableEpochs,
		RoundsConfig:        *args.Configs.RoundConfig,
		EconomicsConfig:     *args.Configs.EconomicsConfig,
		ChanStopNodeProcess: args.ChanStopNodeProcess,
		NumShards:           args.NumShards,
		WorkingDir:          args.Configs.FlagsConfig.WorkingDir,
		GasScheduleFilename: args.GasScheduleFilename,
		NodesSetupPath:      args.Configs.ConfigurationPathsHolder.Nodes,
		InitialRound:        args.InitialRound,
		MinNodesPerShard:    args.MinNodesPerShard,
		MinNodesMeta:        args.MinNodesMeta,
		RoundDurationInMs:   args.RoundDurationInMillis,
		RatingConfig:        *args.Configs.RatingsConfig,
	})
	if err != nil {
		return nil, err
	}

	instance.StatusCoreComponents, err = CreateStatusCoreComponents(args.Configs, instance.CoreComponentsHolder)
	if err != nil {
		return nil, err
	}

	instance.CryptoComponentsHolder, err = CreateCryptoComponents(ArgsCryptoComponentsHolder{
		Config:                      *args.Configs.GeneralConfig,
		EnableEpochsConfig:          args.Configs.EpochConfig.EnableEpochs,
		Preferences:                 *args.Configs.PreferencesConfig,
		CoreComponentsHolder:        instance.CoreComponentsHolder,
		BypassTxSignatureCheck:      args.BypassTxSignatureCheck,
		AllValidatorKeysPemFileName: args.Configs.ConfigurationPathsHolder.AllValidatorKeys,
	})
	if err != nil {
		return nil, err
	}

	instance.NetworkComponentsHolder, err = CreateNetworkComponents(args.SyncedBroadcastNetwork)
	if err != nil {
		return nil, err
	}

	instance.BootstrapComponentsHolder, err = CreateBootstrapComponents(ArgsBootstrapComponentsHolder{
		CoreComponents:       instance.CoreComponentsHolder,
		CryptoComponents:     instance.CryptoComponentsHolder,
		NetworkComponents:    instance.NetworkComponentsHolder,
		StatusCoreComponents: instance.StatusCoreComponents,
		WorkingDir:           args.Configs.FlagsConfig.WorkingDir,
		FlagsConfig:          *args.Configs.FlagsConfig,
		ImportDBConfig:       *args.Configs.ImportDbConfig,
		PrefsConfig:          *args.Configs.PreferencesConfig,
		Config:               *args.Configs.GeneralConfig,
		ShardIDStr:           args.ShardIDStr,
	})
	if err != nil {
		return nil, err
	}

	selfShardID := instance.GetShardCoordinator().SelfId()
	instance.StatusComponentsHolder, err = CreateStatusComponents(
		selfShardID,
		instance.StatusCoreComponents.AppStatusHandler(),
		args.Configs.GeneralConfig.GeneralSettings.StatusPollingIntervalSec,
	)
	if err != nil {
		return nil, err
	}

	err = instance.createBlockChain(selfShardID)
	if err != nil {
		return nil, err
	}

	instance.StateComponentsHolder, err = CreateStateComponents(ArgsStateComponents{
		Config:         *args.Configs.GeneralConfig,
		CoreComponents: instance.CoreComponentsHolder,
		StatusCore:     instance.StatusCoreComponents,
		StoreService:   instance.StoreService,
		ChainHandler:   instance.ChainHandler,
	})
	if err != nil {
		return nil, err
	}

	err = instance.createDataPool(args)
	if err != nil {
		return nil, err
	}
	err = instance.createNodesCoordinator(args.Configs.PreferencesConfig.Preferences, *args.Configs.GeneralConfig)
	if err != nil {
		return nil, err
	}

	instance.DataComponentsHolder, err = CreateDataComponents(ArgsDataComponentsHolder{
		Chain:              instance.ChainHandler,
		StorageService:     instance.StoreService,
		DataPool:           instance.DataPool,
		InternalMarshaller: instance.CoreComponentsHolder.InternalMarshalizer(),
	})
	if err != nil {
		return nil, err
	}

	instance.ProcessComponentsHolder, err = CreateProcessComponents(ArgsProcessComponentsHolder{
		CoreComponents:           instance.CoreComponentsHolder,
		CryptoComponents:         instance.CryptoComponentsHolder,
		NetworkComponents:        instance.NetworkComponentsHolder,
		BootstrapComponents:      instance.BootstrapComponentsHolder,
		StateComponents:          instance.StateComponentsHolder,
		StatusComponents:         instance.StatusComponentsHolder,
		StatusCoreComponents:     instance.StatusCoreComponents,
		FlagsConfig:              *args.Configs.FlagsConfig,
		ImportDBConfig:           *args.Configs.ImportDbConfig,
		PrefsConfig:              *args.Configs.PreferencesConfig,
		Config:                   *args.Configs.GeneralConfig,
		EconomicsConfig:          *args.Configs.EconomicsConfig,
		SystemSCConfig:           *args.Configs.SystemSCConfig,
		EpochConfig:              *args.Configs.EpochConfig,
		RoundConfig:              *args.Configs.RoundConfig,
		ConfigurationPathsHolder: *args.Configs.ConfigurationPathsHolder,
		NodesCoordinator:         instance.NodesCoordinator,
		DataComponents:           instance.DataComponentsHolder,
		GenesisNonce:             args.InitialNonce,
		GenesisRound:             uint64(args.InitialRound),
	})
	if err != nil {
		return nil, err
	}

	err = instance.StatusComponentsHolder.SetForkDetector(instance.ProcessComponentsHolder.ForkDetector())
	if err != nil {
		return nil, err
	}

	err = instance.StatusComponentsHolder.StartPolling()
	if err != nil {
		return nil, err
	}

	err = instance.createBroadcastMessenger()
	if err != nil {
		return nil, err
	}

	err = instance.createFacade(args.Configs, args.APIInterface)
	if err != nil {
		return nil, err
	}

	err = instance.createHttpServer(args.Configs)
	if err != nil {
		return nil, err
	}

	instance.collectClosableComponents(args.APIInterface)

	return instance, nil
}

func (node *testOnlyProcessingNode) createBlockChain(selfShardID uint32) error {
	var err error
	if selfShardID == core.MetachainShardId {
		node.ChainHandler, err = blockchain.NewMetaChain(node.StatusCoreComponents.AppStatusHandler())
	} else {
		node.ChainHandler, err = blockchain.NewBlockChain(node.StatusCoreComponents.AppStatusHandler())
	}

	return err
}

func (node *testOnlyProcessingNode) createDataPool(args ArgsTestOnlyProcessingNode) error {
	var err error

	argsDataPool := dataRetrieverFactory.ArgsDataPool{
		Config:           args.Configs.GeneralConfig,
		EconomicsData:    node.CoreComponentsHolder.EconomicsData(),
		ShardCoordinator: node.BootstrapComponentsHolder.ShardCoordinator(),
		Marshalizer:      node.CoreComponentsHolder.InternalMarshalizer(),
		PathManager:      node.CoreComponentsHolder.PathHandler(),
	}

	node.DataPool, err = dataRetrieverFactory.NewDataPoolFromConfig(argsDataPool)

	return err
}

func (node *testOnlyProcessingNode) createNodesCoordinator(pref config.PreferencesConfig, generalConfig config.Config) error {
	nodesShufflerOut, err := bootstrapComp.CreateNodesShuffleOut(
		node.CoreComponentsHolder.GenesisNodesSetup(),
		generalConfig.EpochStartConfig,
		node.CoreComponentsHolder.ChanStopNodeProcess(),
	)
	if err != nil {
		return err
	}

	bootstrapStorer, err := node.StoreService.GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return err
	}

	node.NodesCoordinator, err = bootstrapComp.CreateNodesCoordinator(
		nodesShufflerOut,
		node.CoreComponentsHolder.GenesisNodesSetup(),
		pref,
		node.CoreComponentsHolder.EpochStartNotifierWithConfirm(),
		node.CryptoComponentsHolder.PublicKey(),
		node.CoreComponentsHolder.InternalMarshalizer(),
		node.CoreComponentsHolder.Hasher(),
		node.CoreComponentsHolder.Rater(),
		bootstrapStorer,
		node.CoreComponentsHolder.NodesShuffler(),
		node.BootstrapComponentsHolder.ShardCoordinator().SelfId(),
		node.BootstrapComponentsHolder.EpochBootstrapParams(),
		node.BootstrapComponentsHolder.EpochBootstrapParams().Epoch(),
		node.CoreComponentsHolder.ChanStopNodeProcess(),
		node.CoreComponentsHolder.NodeTypeProvider(),
		node.CoreComponentsHolder.EnableEpochsHandler(),
		node.DataPool.CurrentEpochValidatorInfo(),
		node.BootstrapComponentsHolder.NodesCoordinatorRegistryFactory(),
		nodesCoordinator.NewIndexHashedNodesCoordinatorWithRaterFactory(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (node *testOnlyProcessingNode) createBroadcastMessenger() error {
	broadcastMessenger, err := sposFactory.GetBroadcastMessenger(
		node.CoreComponentsHolder.InternalMarshalizer(),
		node.CoreComponentsHolder.Hasher(),
		node.NetworkComponentsHolder.NetworkMessenger(),
		node.ProcessComponentsHolder.ShardCoordinator(),
		node.CryptoComponentsHolder.PeerSignatureHandler(),
		node.DataComponentsHolder.Datapool().Headers(),
		node.ProcessComponentsHolder.InterceptorsContainer(),
		node.CoreComponentsHolder.AlarmScheduler(),
		node.CryptoComponentsHolder.KeysHandler(),
	)
	if err != nil {
		return err
	}

	node.broadcastMessenger, err = NewInstantBroadcastMessenger(broadcastMessenger, node.BootstrapComponentsHolder.ShardCoordinator())
	return err
}

// GetProcessComponents will return the process components
func (node *testOnlyProcessingNode) GetProcessComponents() factory.ProcessComponentsHolder {
	return node.ProcessComponentsHolder
}

// GetChainHandler will return the chain handler
func (node *testOnlyProcessingNode) GetChainHandler() chainData.ChainHandler {
	return node.ChainHandler
}

// GetBroadcastMessenger will return the broadcast messenger
func (node *testOnlyProcessingNode) GetBroadcastMessenger() consensus.BroadcastMessenger {
	return node.broadcastMessenger
}

// GetShardCoordinator will return the shard coordinator
func (node *testOnlyProcessingNode) GetShardCoordinator() sharding.Coordinator {
	return node.BootstrapComponentsHolder.ShardCoordinator()
}

// GetCryptoComponents will return the crypto components
func (node *testOnlyProcessingNode) GetCryptoComponents() factory.CryptoComponentsHolder {
	return node.CryptoComponentsHolder
}

// GetCoreComponents will return the core components
func (node *testOnlyProcessingNode) GetCoreComponents() factory.CoreComponentsHolder {
	return node.CoreComponentsHolder
}

// GetStateComponents will return the state components
func (node *testOnlyProcessingNode) GetStateComponents() factory.StateComponentsHolder {
	return node.StateComponentsHolder
}

// GetFacadeHandler will return the facade handler
func (node *testOnlyProcessingNode) GetFacadeHandler() shared.FacadeHandler {
	return node.facadeHandler
}

// GetStatusCoreComponents will return the status core components
func (node *testOnlyProcessingNode) GetStatusCoreComponents() factory.StatusCoreComponentsHolder {
	return node.StatusCoreComponents
}

func (node *testOnlyProcessingNode) collectClosableComponents(apiInterface APIConfigurator) {
	node.closeHandler.AddComponent(node.ProcessComponentsHolder)
	node.closeHandler.AddComponent(node.DataComponentsHolder)
	node.closeHandler.AddComponent(node.StateComponentsHolder)
	node.closeHandler.AddComponent(node.StatusComponentsHolder)
	node.closeHandler.AddComponent(node.BootstrapComponentsHolder)
	node.closeHandler.AddComponent(node.NetworkComponentsHolder)
	node.closeHandler.AddComponent(node.StatusCoreComponents)
	node.closeHandler.AddComponent(node.CoreComponentsHolder)
	node.closeHandler.AddComponent(node.facadeHandler)

	// TODO remove this after http server fix
	shardID := node.GetShardCoordinator().SelfId()
	if facade.DefaultRestPortOff != apiInterface.RestApiInterface(shardID) {
		node.closeHandler.AddComponent(node.httpServer)
	}
}

// SetKeyValueForAddress will set the provided state for the given address
func (node *testOnlyProcessingNode) SetKeyValueForAddress(address []byte, keyValueMap map[string]string) error {
	userAccount, err := node.getUserAccount(address)
	if err != nil {
		return err
	}

	err = setKeyValueMap(userAccount, keyValueMap)
	if err != nil {
		return err
	}

	accountsAdapter := node.StateComponentsHolder.AccountsAdapter()
	err = accountsAdapter.SaveAccount(userAccount)
	if err != nil {
		return err
	}

	_, err = accountsAdapter.Commit()

	return err
}

func setKeyValueMap(userAccount state.UserAccountHandler, keyValueMap map[string]string) error {
	for keyHex, valueHex := range keyValueMap {
		keyDecoded, err := hex.DecodeString(keyHex)
		if err != nil {
			return fmt.Errorf("cannot decode key, error: %w", err)
		}
		valueDecoded, err := hex.DecodeString(valueHex)
		if err != nil {
			return fmt.Errorf("cannot decode value, error: %w", err)
		}

		err = userAccount.SaveKeyValue(keyDecoded, valueDecoded)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetStateForAddress will set the state for the give address
func (node *testOnlyProcessingNode) SetStateForAddress(address []byte, addressState *dtos.AddressState) error {
	userAccount, err := node.getUserAccount(address)
	if err != nil {
		return err
	}

	err = setNonceAndBalanceForAccount(userAccount, addressState.Nonce, addressState.Balance)
	if err != nil {
		return err
	}

	err = setKeyValueMap(userAccount, addressState.Keys)
	if err != nil {
		return err
	}

	err = node.setScDataIfNeeded(address, userAccount, addressState)
	if err != nil {
		return err
	}

	rootHash, err := base64.StdEncoding.DecodeString(addressState.RootHash)
	if err != nil {
		return err
	}
	if len(rootHash) != 0 {
		userAccount.SetRootHash(rootHash)
	}

	accountsAdapter := node.StateComponentsHolder.AccountsAdapter()
	err = accountsAdapter.SaveAccount(userAccount)
	if err != nil {
		return err
	}

	_, err = accountsAdapter.Commit()
	return err
}

func setNonceAndBalanceForAccount(userAccount state.UserAccountHandler, nonce *uint64, balance string) error {
	if nonce != nil {
		// set nonce to zero
		userAccount.IncreaseNonce(-userAccount.GetNonce())
		// set nonce with the provided value
		userAccount.IncreaseNonce(*nonce)
	}

	if balance == "" {
		return nil
	}

	providedBalance, ok := big.NewInt(0).SetString(balance, 10)
	if !ok {
		return errors.New("cannot convert string balance to *big.Int")
	}

	// set balance to zero
	userBalance := userAccount.GetBalance()
	err := userAccount.AddToBalance(userBalance.Neg(userBalance))
	if err != nil {
		return err
	}
	// set provided balance
	return userAccount.AddToBalance(providedBalance)
}

func (node *testOnlyProcessingNode) setScDataIfNeeded(address []byte, userAccount state.UserAccountHandler, addressState *dtos.AddressState) error {
	if !core.IsSmartContractAddress(address) {
		return nil
	}

	if addressState.Code != "" {
		decodedCode, err := hex.DecodeString(addressState.Code)
		if err != nil {
			return err
		}
		userAccount.SetCode(decodedCode)
	}

	if addressState.CodeHash != "" {
		codeHash, errD := base64.StdEncoding.DecodeString(addressState.CodeHash)
		if errD != nil {
			return errD
		}
		userAccount.SetCodeHash(codeHash)
	}

	if addressState.CodeMetadata != "" {
		decodedCodeMetadata, errD := base64.StdEncoding.DecodeString(addressState.CodeMetadata)
		if errD != nil {
			return errD
		}
		userAccount.SetCodeMetadata(decodedCodeMetadata)
	}

	if addressState.Owner != "" {
		ownerAddress, errD := node.CoreComponentsHolder.AddressPubKeyConverter().Decode(addressState.Owner)
		if errD != nil {
			return errD
		}
		userAccount.SetOwnerAddress(ownerAddress)
	}

	if addressState.DeveloperRewards != "" {
		developerRewards, ok := big.NewInt(0).SetString(addressState.DeveloperRewards, 10)
		if !ok {
			return errors.New("cannot convert string developer rewards to *big.Int")
		}
		userAccount.AddToDeveloperReward(developerRewards)
	}

	return nil
}

func (node *testOnlyProcessingNode) getUserAccount(address []byte) (state.UserAccountHandler, error) {
	accountsAdapter := node.StateComponentsHolder.AccountsAdapter()
	account, err := accountsAdapter.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(state.UserAccountHandler)
	if !ok {
		return nil, errors.New("cannot cast AccountHandler to UserAccountHandler")
	}

	return userAccount, nil
}

// Close will call the Close methods on all inner components
func (node *testOnlyProcessingNode) Close() error {
	return node.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (node *testOnlyProcessingNode) IsInterfaceNil() bool {
	return node == nil
}
