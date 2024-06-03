package chainSimulator

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorErrors "github.com/multiversx/mx-chain-go/node/chainSimulator/errors"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	processing "github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/rating"
	mxChainSharding "github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/sharding"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const delaySendTxs = time.Millisecond

var log = logger.GetOrCreate("chainSimulator")

type transactionWithResult struct {
	hexHash string
	tx      *transaction.Transaction
	result  *transaction.ApiTransactionResult
}

// ArgsChainSimulator holds the arguments needed to create a new instance of simulator
type ArgsChainSimulator struct {
	BypassTxSignatureCheck         bool
	TempDir                        string
	PathToInitialConfig            string
	NumOfShards                    uint32
	MinNodesPerShard               uint32
	ConsensusGroupSize             uint32
	MetaChainMinNodes              uint32
	MetaChainConsensusGroupSize    uint32
	NumNodesWaitingListShard       uint32
	NumNodesWaitingListMeta        uint32
	GenesisTimestamp               int64
	InitialRound                   int64
	InitialEpoch                   uint32
	InitialNonce                   uint64
	RoundDurationInMillis          uint64
	RoundsPerEpoch                 core.OptionalUint64
	ApiInterface                   components.APIConfigurator
	AlterConfigsFunction           func(cfg *config.Configs)
	CreateGenesisNodesSetup        func(nodesFilePath string, addressPubkeyConverter core.PubkeyConverter, validatorPubkeyConverter core.PubkeyConverter, genesisMaxNumShards uint32) (mxChainSharding.GenesisNodesSetupHandler, error)
	CreateRatingsData              func(arg rating.RatingsDataArg) (processing.RatingsInfoHandler, error)
	CreateIncomingHeaderSubscriber func(config *config.NotifierConfig, dataPool dataRetriever.PoolsHolder, mainChainNotarizationStartRound uint64, runTypeComponents factory.RunTypeComponentsHolder) (processing.IncomingHeaderSubscriber, error)
	CreateRunTypeComponents        func(args runType.ArgsRunTypeComponents) (factory.RunTypeComponentsHolder, error)
	NodeFactory                    node.NodeFactory
	ChainProcessorFactory          ChainHandlerFactory
}

type simulator struct {
	chanStopNodeProcess    chan endProcess.ArgEndProcess
	syncedBroadcastNetwork components.SyncedBroadcastNetworkHandler
	handlers               []ChainHandler
	initialWalletKeys      *dtos.InitialWalletKeys
	initialStakedKeys      map[string]*dtos.BLSKey
	validatorsPrivateKeys  []crypto.PrivateKey
	nodes                  map[uint32]process.NodeHandler
	numOfShards            uint32
	mutex                  sync.RWMutex
}

// NewChainSimulator will create a new instance of simulator
func NewChainSimulator(args ArgsChainSimulator) (*simulator, error) {
	setSimulatorRunTypeArguments(&args)

	instance := &simulator{
		syncedBroadcastNetwork: components.NewSyncedBroadcastNetwork(),
		nodes:                  make(map[uint32]process.NodeHandler),
		handlers:               make([]ChainHandler, 0, args.NumOfShards+1),
		numOfShards:            args.NumOfShards,
		chanStopNodeProcess:    make(chan endProcess.ArgEndProcess),
		mutex:                  sync.RWMutex{},
		initialStakedKeys:      make(map[string]*dtos.BLSKey),
	}

	err := instance.createChainHandlers(args)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func setSimulatorRunTypeArguments(args *ArgsChainSimulator) {
	if args.CreateGenesisNodesSetup == nil {
		args.CreateGenesisNodesSetup = func(nodesFilePath string, addressPubkeyConverter core.PubkeyConverter, validatorPubkeyConverter core.PubkeyConverter, genesisMaxNumShards uint32) (mxChainSharding.GenesisNodesSetupHandler, error) {
			return mxChainSharding.NewNodesSetup(nodesFilePath, addressPubkeyConverter, validatorPubkeyConverter, genesisMaxNumShards)
		}
	}
	if args.CreateRatingsData == nil {
		args.CreateRatingsData = func(arg rating.RatingsDataArg) (processing.RatingsInfoHandler, error) {
			return rating.NewRatingsData(arg)
		}
	}
	if args.CreateIncomingHeaderSubscriber == nil {
		args.CreateIncomingHeaderSubscriber = func(_ *config.NotifierConfig, _ dataRetriever.PoolsHolder, _ uint64, _ factory.RunTypeComponentsHolder) (processing.IncomingHeaderSubscriber, error) {
			return &sovereign.IncomingHeaderSubscriberStub{}, nil
		}
	}
	if args.CreateRunTypeComponents == nil {
		args.CreateRunTypeComponents = func(args runType.ArgsRunTypeComponents) (factory.RunTypeComponentsHolder, error) {
			return createRunTypeComponents(args)
		}
	}
	if args.NodeFactory == nil {
		args.NodeFactory = node.NewNodeFactory()
	}
	if args.ChainProcessorFactory == nil {
		args.ChainProcessorFactory = NewChainHandlerFactory()
	}
}

func createRunTypeComponents(args runType.ArgsRunTypeComponents) (factory.RunTypeComponentsHolder, error) {
	runTypeComponentsFactory, err := runType.NewRunTypeComponentsFactory(args)
	if err != nil {
		return nil, err
	}
	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(runTypeComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedRunTypeComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedRunTypeComponents, nil
}

func (s *simulator) createChainHandlers(args ArgsChainSimulator) error {
	outputConfigs, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:                 args.NumOfShards,
		OriginalConfigsPath:         args.PathToInitialConfig,
		GenesisTimeStamp:            computeStartTimeBaseOnInitialRound(args),
		RoundDurationInMillis:       args.RoundDurationInMillis,
		TempDir:                     args.TempDir,
		MinNodesPerShard:            args.MinNodesPerShard,
		ConsensusGroupSize:          args.ConsensusGroupSize,
		MetaChainMinNodes:           args.MetaChainMinNodes,
		MetaChainConsensusGroupSize: args.MetaChainConsensusGroupSize,
		RoundsPerEpoch:              args.RoundsPerEpoch,
		InitialEpoch:                args.InitialEpoch,
		AlterConfigsFunction:        args.AlterConfigsFunction,
		NumNodesWaitingListShard:    args.NumNodesWaitingListShard,
		NumNodesWaitingListMeta:     args.NumNodesWaitingListMeta,
	})
	if err != nil {
		return err
	}

	for idx := -1; idx < int(args.NumOfShards); idx++ {
		shardIDStr := fmt.Sprintf("%d", idx)
		if idx == -1 {
			if args.MetaChainMinNodes == 0 {
				continue
			}
			shardIDStr = "metachain"
		}

		node, errCreate := s.createTestNode(*outputConfigs, args, shardIDStr)
		if errCreate != nil {
			return errCreate
		}

		chainHandler, errCreate := args.ChainProcessorFactory.CreateChainHandler(node)
		if errCreate != nil {
			return errCreate
		}

		shardID := node.GetShardCoordinator().SelfId()
		s.nodes[shardID] = node
		s.handlers = append(s.handlers, chainHandler)

		if node.GetShardCoordinator().SelfId() == core.MetachainShardId {
			currentRootHash, errRootHash := node.GetProcessComponents().ValidatorsStatistics().RootHash()
			if errRootHash != nil {
				return errRootHash
			}

			allValidatorsInfo, errGet := node.GetProcessComponents().ValidatorsStatistics().GetValidatorInfoForRootHash(currentRootHash)
			if errRootHash != nil {
				return errGet
			}

			err = node.GetProcessComponents().EpochSystemSCProcessor().ProcessSystemSmartContract(
				allValidatorsInfo,
				node.GetDataComponents().Blockchain().GetGenesisHeader(),
			)
			if err != nil {
				return err
			}

			_, err = node.GetStateComponents().AccountsAdapter().Commit()
			if err != nil {
				return err
			}
		}
	}

	s.initialWalletKeys = outputConfigs.InitialWallets
	s.validatorsPrivateKeys = outputConfigs.ValidatorsPrivateKeys

	log.Info("running the chain simulator with the following parameters",
		"number of shards (including meta)", args.NumOfShards+1,
		"round per epoch", outputConfigs.Configs.GeneralConfig.EpochStartConfig.RoundsPerEpoch,
		"round duration", time.Millisecond*time.Duration(args.RoundDurationInMillis),
		"genesis timestamp", args.GenesisTimestamp,
		"original config path", args.PathToInitialConfig,
		"temporary path", args.TempDir)

	return nil
}

func computeStartTimeBaseOnInitialRound(args ArgsChainSimulator) int64 {
	return args.GenesisTimestamp + int64(args.RoundDurationInMillis/1000)*args.InitialRound
}

func (s *simulator) createTestNode(
	outputConfigs configs.ArgsConfigsSimulator, args ArgsChainSimulator, shardIDStr string,
) (process.NodeHandler, error) {
	argsTestOnlyProcessorNode := components.ArgsTestOnlyProcessingNode{
		Configs:                        outputConfigs.Configs,
		ChanStopNodeProcess:            s.chanStopNodeProcess,
		SyncedBroadcastNetwork:         s.syncedBroadcastNetwork,
		NumShards:                      s.numOfShards,
		GasScheduleFilename:            outputConfigs.GasScheduleFilename,
		ShardIDStr:                     shardIDStr,
		APIInterface:                   args.ApiInterface,
		BypassTxSignatureCheck:         args.BypassTxSignatureCheck,
		InitialRound:                   args.InitialRound,
		InitialNonce:                   args.InitialNonce,
		MinNodesPerShard:               args.MinNodesPerShard,
		ConsensusGroupSize:             args.ConsensusGroupSize,
		MinNodesMeta:                   args.MetaChainMinNodes,
		MetaChainConsensusGroupSize:    args.MetaChainConsensusGroupSize,
		RoundDurationInMillis:          args.RoundDurationInMillis,
		CreateGenesisNodesSetup:        args.CreateGenesisNodesSetup,
		CreateRatingsData:              args.CreateRatingsData,
		CreateIncomingHeaderSubscriber: args.CreateIncomingHeaderSubscriber,
		CreateRunTypeComponents:        args.CreateRunTypeComponents,
		NodeFactory:                    args.NodeFactory,
	}

	return components.NewTestOnlyProcessingNode(argsTestOnlyProcessorNode)
}

// GenerateBlocks will generate the provided number of blocks
func (s *simulator) GenerateBlocks(numOfBlocks int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for idx := 0; idx < numOfBlocks; idx++ {
		s.incrementRoundOnAllValidators()
		err := s.allNodesCreateBlocks()
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateBlocksUntilEpochIsReached will generate blocks until the epoch is reached
func (s *simulator) GenerateBlocksUntilEpochIsReached(targetEpoch int32) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	maxNumberOfRounds := 10000
	for idx := 0; idx < maxNumberOfRounds; idx++ {
		s.incrementRoundOnAllValidators()
		err := s.allNodesCreateBlocks()
		if err != nil {
			return err
		}

		epochReachedOnAllNodes, err := s.isTargetEpochReached(targetEpoch)
		if err != nil {
			return err
		}

		if epochReachedOnAllNodes {
			return nil
		}
	}
	return fmt.Errorf("exceeded rounds to generate blocks")
}

// ForceResetValidatorStatisticsCache will force the reset of the cache used for the validators statistics endpoint
func (s *simulator) ForceResetValidatorStatisticsCache() error {
	metachainNode := s.GetNodeHandler(core.MetachainShardId)
	if check.IfNil(metachainNode) {
		return errNilMetachainNode
	}

	return metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
}

func (s *simulator) isTargetEpochReached(targetEpoch int32) (bool, error) {
	metachainNode := s.nodes[core.MetachainShardId]
	metachainEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()

	for shardID, n := range s.nodes {
		if shardID != core.MetachainShardId {
			if int32(n.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()) < int32(metachainEpoch-1) {
				return false, fmt.Errorf("shard %d is with at least 2 epochs behind metachain shard node epoch %d, metachain node epoch %d",
					shardID, n.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch(), metachainEpoch)
			}
		}

		if int32(n.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()) < targetEpoch {
			return false, nil
		}
	}

	return true, nil
}

func (s *simulator) incrementRoundOnAllValidators() {
	for _, node := range s.handlers {
		node.IncrementRound()
	}
}

func (s *simulator) allNodesCreateBlocks() error {
	for _, node := range s.handlers {
		// TODO MX-15150 remove this when we remove all goroutines
		time.Sleep(2 * time.Millisecond)

		err := node.CreateNewBlock()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetNodeHandler returns the node handler from the provided shardID
func (s *simulator) GetNodeHandler(shardID uint32) process.NodeHandler {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.nodes[shardID]
}

// GetRestAPIInterfaces will return a map with the rest api interfaces for every node
func (s *simulator) GetRestAPIInterfaces() map[uint32]string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	resMap := make(map[uint32]string)
	for shardID, node := range s.nodes {
		resMap[shardID] = node.GetFacadeHandler().RestApiInterface()
	}

	return resMap
}

// GetInitialWalletKeys will return the initial wallet keys
func (s *simulator) GetInitialWalletKeys() *dtos.InitialWalletKeys {
	return s.initialWalletKeys
}

// AddValidatorKeys will add the provided validators private keys in the keys handler on all nodes
func (s *simulator) AddValidatorKeys(validatorsPrivateKeys [][]byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, node := range s.nodes {
		err := s.setValidatorKeysForNode(node, validatorsPrivateKeys)
		if err != nil {
			return err
		}
	}

	return nil
}

// GenerateAndMintWalletAddress will generate an address in the provided shard and will mint that address with the provided value
// if the target shard ID value does not correspond to a node handled by the chain simulator, the address will be generated in a random shard ID
func (s *simulator) GenerateAndMintWalletAddress(targetShardID uint32, value *big.Int) (dtos.WalletAddress, error) {
	addressConverter := s.nodes[0].GetCoreComponents().AddressPubKeyConverter()
	nodeHandler := s.GetNodeHandler(targetShardID)
	var buff []byte
	if check.IfNil(nodeHandler) {
		buff = generateAddress(addressConverter.Len())
	} else {
		buff = generateAddressInShard(nodeHandler.GetShardCoordinator(), addressConverter.Len())
	}

	address, err := addressConverter.Encode(buff)
	if err != nil {
		return dtos.WalletAddress{}, err
	}

	err = s.SetStateMultiple([]*dtos.AddressState{
		{
			Address: address,
			Balance: value.String(),
		},
	})

	return dtos.WalletAddress{
		Bech32: address,
		Bytes:  buff,
	}, err
}

func generateAddressInShard(shardCoordinator mxChainSharding.Coordinator, len int) []byte {
	for {
		buff := generateAddress(len)
		shardID := shardCoordinator.ComputeId(buff)
		if shardID == shardCoordinator.SelfId() {
			return buff
		}
	}
}

func generateAddress(len int) []byte {
	buff := make([]byte, len)
	_, _ = rand.Read(buff)

	return buff
}

func (s *simulator) setValidatorKeysForNode(node process.NodeHandler, validatorsPrivateKeys [][]byte) error {
	for idx, privateKey := range validatorsPrivateKeys {

		err := node.GetCryptoComponents().ManagedPeersHolder().AddManagedPeer(privateKey)
		if err != nil {
			return fmt.Errorf("cannot add private key for shard=%d, index=%d, error=%s", node.GetShardCoordinator().SelfId(), idx, err.Error())
		}
	}

	return nil
}

// GetValidatorPrivateKeys will return the initial validators private keys
func (s *simulator) GetValidatorPrivateKeys() []crypto.PrivateKey {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.validatorsPrivateKeys
}

// SetKeyValueForAddress will set the provided state for a given address
func (s *simulator) SetKeyValueForAddress(address string, keyValueMap map[string]string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	addressConverter := s.nodes[0].GetCoreComponents().AddressPubKeyConverter()
	addressBytes, err := addressConverter.Decode(address)
	if err != nil {
		return err
	}

	if bytes.Equal(addressBytes, core.SystemAccountAddress) {
		return s.setKeyValueSystemAccount(keyValueMap)
	}

	shardID := sharding.ComputeShardID(addressBytes, s.numOfShards)
	testNode, ok := s.nodes[shardID]
	if !ok {
		return fmt.Errorf("cannot find a test node for the computed shard id, computed shard id: %d", shardID)
	}

	return testNode.SetKeyValueForAddress(addressBytes, keyValueMap)
}

func (s *simulator) setKeyValueSystemAccount(keyValueMap map[string]string) error {
	for shard, node := range s.nodes {
		err := node.SetKeyValueForAddress(core.SystemAccountAddress, keyValueMap)
		if err != nil {
			return fmt.Errorf("%w for shard %d", err, shard)
		}
	}

	return nil
}

// SetStateMultiple will set state for multiple addresses
func (s *simulator) SetStateMultiple(stateSlice []*dtos.AddressState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	addressConverter := s.nodes[0].GetCoreComponents().AddressPubKeyConverter()
	for _, stateValue := range stateSlice {
		addressBytes, err := addressConverter.Decode(stateValue.Address)
		if err != nil {
			return err
		}

		if bytes.Equal(addressBytes, core.SystemAccountAddress) {
			err = s.setStateSystemAccount(stateValue)
		} else {
			shardID := sharding.ComputeShardID(addressBytes, s.numOfShards)
			err = s.nodes[shardID].SetStateForAddress(addressBytes, stateValue)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveAccounts will try to remove all accounts data for the addresses provided
func (s *simulator) RemoveAccounts(addresses []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	addressConverter := s.nodes[0].GetCoreComponents().AddressPubKeyConverter()
	for _, address := range addresses {
		addressBytes, err := addressConverter.Decode(address)
		if err != nil {
			return err
		}

		if bytes.Equal(addressBytes, core.SystemAccountAddress) {
			err = s.removeAllSystemAccounts()
		} else {
			shardID := sharding.ComputeShardID(addressBytes, s.numOfShards)
			err = s.nodes[shardID].RemoveAccount(addressBytes)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// SendTxAndGenerateBlockTilTxIsExecuted will send the provided transaction and generate block until the transaction is executed
func (s *simulator) SendTxAndGenerateBlockTilTxIsExecuted(txToSend *transaction.Transaction, maxNumOfBlocksToGenerateWhenExecutingTx int) (*transaction.ApiTransactionResult, error) {
	result, err := s.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txToSend}, maxNumOfBlocksToGenerateWhenExecutingTx)
	if err != nil {
		return nil, err
	}

	return result[0], nil
}

// SendTxsAndGenerateBlocksTilAreExecuted will send the provided transactions and generate block until all transactions are executed
func (s *simulator) SendTxsAndGenerateBlocksTilAreExecuted(txsToSend []*transaction.Transaction, maxNumOfBlocksToGenerateWhenExecutingTx int) ([]*transaction.ApiTransactionResult, error) {
	if len(txsToSend) == 0 {
		return nil, chainSimulatorErrors.ErrEmptySliceOfTxs
	}
	if maxNumOfBlocksToGenerateWhenExecutingTx == 0 {
		return nil, chainSimulatorErrors.ErrInvalidMaxNumOfBlocks
	}

	transactionStatus := make([]*transactionWithResult, 0, len(txsToSend))
	for idx, tx := range txsToSend {
		if tx == nil {
			return nil, fmt.Errorf("%w on position %d", chainSimulatorErrors.ErrNilTransaction, idx)
		}

		txHashHex, err := s.sendTx(tx)
		if err != nil {
			return nil, err
		}

		transactionStatus = append(transactionStatus, &transactionWithResult{
			hexHash: txHashHex,
			tx:      tx,
		})
	}

	time.Sleep(delaySendTxs)

	for count := 0; count < maxNumOfBlocksToGenerateWhenExecutingTx; count++ {
		err := s.GenerateBlocks(1)
		if err != nil {
			return nil, err
		}

		txsAreExecuted := s.computeTransactionsStatus(transactionStatus)
		if txsAreExecuted {
			return getApiTransactionsFromResult(transactionStatus), nil
		}
	}

	return nil, errors.New("something went wrong. Transaction(s) is/are still in pending")
}

func (s *simulator) computeTransactionsStatus(txsWithResult []*transactionWithResult) bool {
	allAreExecuted := true
	for _, resultTx := range txsWithResult {
		if resultTx.result != nil {
			continue
		}

		sentTx := resultTx.tx
		destinationShardID := s.GetNodeHandler(0).GetShardCoordinator().ComputeId(sentTx.RcvAddr)
		result, errGet := s.GetNodeHandler(destinationShardID).GetFacadeHandler().GetTransaction(resultTx.hexHash, true)
		if errGet == nil && result.Status != transaction.TxStatusPending {
			log.Info("############## transaction was executed ##############", "txHash", resultTx.hexHash)
			resultTx.result = result
			continue
		}

		allAreExecuted = false
	}

	return allAreExecuted
}

func getApiTransactionsFromResult(txWithResult []*transactionWithResult) []*transaction.ApiTransactionResult {
	result := make([]*transaction.ApiTransactionResult, 0, len(txWithResult))
	for _, tx := range txWithResult {
		result = append(result, tx.result)
	}

	return result
}

func (s *simulator) sendTx(tx *transaction.Transaction) (string, error) {
	shardID := s.GetNodeHandler(0).GetShardCoordinator().ComputeId(tx.SndAddr)
	err := s.GetNodeHandler(shardID).GetFacadeHandler().ValidateTransaction(tx)
	if err != nil {
		return "", err
	}

	node := s.GetNodeHandler(shardID)
	txHash, err := core.CalculateHash(node.GetCoreComponents().InternalMarshalizer(), node.GetCoreComponents().Hasher(), tx)
	if err != nil {
		return "", err
	}

	txHashHex := hex.EncodeToString(txHash)
	_, err = node.GetFacadeHandler().SendBulkTransactions([]*transaction.Transaction{tx})
	if err != nil {
		return "", err
	}

	for {
		recoveredTx, _ := node.GetFacadeHandler().GetTransaction(txHashHex, false)
		if recoveredTx != nil {
			log.Info("############## send transaction ##############", "txHash", txHashHex)
			return txHashHex, nil
		}

		time.Sleep(delaySendTxs)
	}
}

func (s *simulator) setStateSystemAccount(state *dtos.AddressState) error {
	for shard, node := range s.nodes {
		err := node.SetStateForAddress(core.SystemAccountAddress, state)
		if err != nil {
			return fmt.Errorf("%w for shard %d", err, shard)
		}
	}

	return nil
}

func (s *simulator) removeAllSystemAccounts() error {
	for shard, node := range s.nodes {
		err := node.RemoveAccount(core.SystemAccountAddress)
		if err != nil {
			return fmt.Errorf("%w for shard %d", err, shard)
		}
	}

	return nil
}

// GetAccount will fetch the account of the provided address
func (s *simulator) GetAccount(address dtos.WalletAddress) (api.AccountResponse, error) {
	destinationShardID := s.GetNodeHandler(0).GetShardCoordinator().ComputeId(address.Bytes)

	account, _, err := s.GetNodeHandler(destinationShardID).GetFacadeHandler().GetAccount(address.Bech32, api.AccountQueryOptions{})
	return account, err
}

// Close will stop and close the simulator
func (s *simulator) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var errorStrings []string
	for _, n := range s.nodes {
		err := n.Close()
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	if len(errorStrings) != 0 {
		log.Error("error closing chain simulator", "error", components.AggregateErrors(errorStrings, components.ErrClose))
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simulator) IsInterfaceNil() bool {
	return s == nil
}

// GenerateBlsPrivateKeys will generate bls keys
func GenerateBlsPrivateKeys(numOfKeys int) ([][]byte, []string, error) {
	blockSigningGenerator := signing.NewKeyGenerator(mcl.NewSuiteBLS12())

	secretKeysBytes := make([][]byte, 0, numOfKeys)
	blsKeysHex := make([]string, 0, numOfKeys)
	for idx := 0; idx < numOfKeys; idx++ {
		secretKey, publicKey := blockSigningGenerator.GeneratePair()

		secretKeyBytes, err := secretKey.ToByteArray()
		if err != nil {
			return nil, nil, err
		}

		secretKeysBytes = append(secretKeysBytes, secretKeyBytes)

		publicKeyBytes, err := publicKey.ToByteArray()
		if err != nil {
			return nil, nil, err
		}

		blsKeysHex = append(blsKeysHex, hex.EncodeToString(publicKeyBytes))
	}

	return secretKeysBytes, blsKeysHex, nil
}
