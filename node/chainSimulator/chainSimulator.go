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

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	processSov "github.com/multiversx/mx-chain-go/process"
	mxChainSharding "github.com/multiversx/mx-chain-go/sharding"
)

const delaySendTxs = time.Millisecond

var log = logger.GetOrCreate("chainSimulator")

type transactionWithResult struct {
	hexHash string
	tx      *transaction.Transaction
	result  *transaction.ApiTransactionResult
}

// ArgsChainSimulator holds the arguments needed to create a new instance of Simulator
type ArgsChainSimulator struct {
	BypassTxSignatureCheck      bool
	TempDir                     string
	PathToInitialConfig         string
	NumOfShards                 uint32
	MinNodesPerShard            uint32
	MetaChainMinNodes           uint32
	NumNodesWaitingListShard    uint32
	NumNodesWaitingListMeta     uint32
	GenesisTimestamp            int64
	InitialRound                int64
	InitialEpoch                uint32
	InitialNonce                uint64
	RoundDurationInMillis       uint64
	RoundsPerEpoch              core.OptionalUint64
	ApiInterface                components.APIConfigurator
	AlterConfigsFunction        func(cfg *config.Configs)
	CreateIncomingHeaderHandler func(config *config.NotifierConfig, dataPool dataRetriever.PoolsHolder, mainChainNotarizationStartRound uint64, runTypeComponents factory.RunTypeComponentsHolder) (processSov.IncomingHeaderSubscriber, error)
	GetRunTypeComponents        func(coreComponents factory.CoreComponentsHolder, cryptoComponents factory.CryptoComponentsHolder) (factory.RunTypeComponentsHolder, error)
}

type Simulator struct {
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

// NewChainSimulator will create a new instance of Simulator
func NewChainSimulator(args ArgsChainSimulator) (*Simulator, error) {
	syncedBroadcastNetwork := components.NewSyncedBroadcastNetwork()

	if args.CreateIncomingHeaderHandler == nil {
		args.CreateIncomingHeaderHandler = func(_ *config.NotifierConfig, _ dataRetriever.PoolsHolder, _ uint64, _ factory.RunTypeComponentsHolder) (processSov.IncomingHeaderSubscriber, error) {
			return nil, nil
		}
	}
	if args.GetRunTypeComponents == nil {
		args.GetRunTypeComponents = func(coreComponents factory.CoreComponentsHolder, cryptoComponents factory.CryptoComponentsHolder) (factory.RunTypeComponentsHolder, error) {
			return createRunTypeComponents(coreComponents, cryptoComponents)
		}
	}

	instance := &Simulator{
		syncedBroadcastNetwork: syncedBroadcastNetwork,
		nodes:                  make(map[uint32]process.NodeHandler),
		handlers:               make([]ChainHandler, 0, args.NumOfShards),
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

func (s *Simulator) createChainHandlers(args ArgsChainSimulator) error {
	outputConfigs, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:              args.NumOfShards,
		OriginalConfigsPath:      args.PathToInitialConfig,
		GenesisTimeStamp:         computeStartTimeBaseOnInitialRound(args),
		RoundDurationInMillis:    args.RoundDurationInMillis,
		TempDir:                  args.TempDir,
		MinNodesPerShard:         args.MinNodesPerShard,
		MetaChainMinNodes:        args.MetaChainMinNodes,
		RoundsPerEpoch:           args.RoundsPerEpoch,
		InitialEpoch:             args.InitialEpoch,
		AlterConfigsFunction:     args.AlterConfigsFunction,
		NumNodesWaitingListShard: args.NumNodesWaitingListShard,
		NumNodesWaitingListMeta:  args.NumNodesWaitingListMeta,
	})
	if err != nil {
		return err
	}

	for idx := 0; idx < int(args.NumOfShards); idx++ {
		shardIDStr := fmt.Sprintf("%d", idx)

		node, errCreate := s.createTestNode(*outputConfigs, args, shardIDStr)
		if errCreate != nil {
			return errCreate
		}

		chainHandler, errCreate := process.NewBlocksCreator(node)
		if errCreate != nil {
			return errCreate
		}

		shardID := node.GetShardCoordinator().SelfId()
		s.nodes[shardID] = node
		s.handlers = append(s.handlers, chainHandler)
	}

	s.initialWalletKeys = outputConfigs.InitialWallets
	s.validatorsPrivateKeys = outputConfigs.ValidatorsPrivateKeys

	log.Info("running the chain Simulator with the following parameters",
		"number of shards (including meta)", args.NumOfShards+1,
		"round per epoch", outputConfigs.Configs.GeneralConfig.EpochStartConfig.RoundsPerEpoch,
		"round duration", time.Millisecond*time.Duration(args.RoundDurationInMillis),
		"genesis timestamp", args.GenesisTimestamp,
		"original config path", args.PathToInitialConfig,
		"temporary path", args.TempDir)

	return nil
}

func createRunTypeComponents(coreComponents factory.CoreComponentsHolder, _ factory.CryptoComponentsHolder) (factory.RunTypeComponentsHolder, error) {
	runTypeComponentsFactory, _ := runType.NewRunTypeComponentsFactory(coreComponents)
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

func computeStartTimeBaseOnInitialRound(args ArgsChainSimulator) int64 {
	return args.GenesisTimestamp + int64(args.RoundDurationInMillis/1000)*args.InitialRound
}

func (s *Simulator) createTestNode(
	outputConfigs configs.ArgsConfigsSimulator, args ArgsChainSimulator, shardIDStr string,
) (process.NodeHandler, error) {
	argsTestOnlyProcessorNode := components.ArgsTestOnlyProcessingNode{
		Configs:                     outputConfigs.Configs,
		ChanStopNodeProcess:         s.chanStopNodeProcess,
		SyncedBroadcastNetwork:      s.syncedBroadcastNetwork,
		NumShards:                   s.numOfShards,
		GasScheduleFilename:         outputConfigs.GasScheduleFilename,
		ShardIDStr:                  shardIDStr,
		APIInterface:                args.ApiInterface,
		BypassTxSignatureCheck:      args.BypassTxSignatureCheck,
		InitialRound:                args.InitialRound,
		InitialNonce:                args.InitialNonce,
		MinNodesPerShard:            args.MinNodesPerShard,
		MinNodesMeta:                args.MetaChainMinNodes,
		RoundDurationInMillis:       args.RoundDurationInMillis,
		CreateIncomingHeaderHandler: args.CreateIncomingHeaderHandler,
		GetRunTypeComponents:        args.GetRunTypeComponents,
	}

	return components.NewTestOnlyProcessingNode(argsTestOnlyProcessorNode)
}

// GenerateBlocks will generate the provided number of blocks
func (s *Simulator) GenerateBlocks(numOfBlocks int) error {
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
func (s *Simulator) GenerateBlocksUntilEpochIsReached(targetEpoch int32) error {
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

func (s *Simulator) isTargetEpochReached(targetEpoch int32) (bool, error) {
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

func (s *Simulator) incrementRoundOnAllValidators() {
	for _, node := range s.handlers {
		node.IncrementRound()
	}
}

func (s *Simulator) allNodesCreateBlocks() error {
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
func (s *Simulator) GetNodeHandler(shardID uint32) process.NodeHandler {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.nodes[shardID]
}

// GetRestAPIInterfaces will return a map with the rest api interfaces for every node
func (s *Simulator) GetRestAPIInterfaces() map[uint32]string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	resMap := make(map[uint32]string)
	for shardID, node := range s.nodes {
		resMap[shardID] = node.GetFacadeHandler().RestApiInterface()
	}

	return resMap
}

// GetInitialWalletKeys will return the initial wallet keys
func (s *Simulator) GetInitialWalletKeys() *dtos.InitialWalletKeys {
	return s.initialWalletKeys
}

// AddValidatorKeys will add the provided validators private keys in the keys handler on all nodes
func (s *Simulator) AddValidatorKeys(validatorsPrivateKeys [][]byte) error {
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
// if the target shard ID value does not correspond to a node handled by the chain Simulator, the address will be generated in a random shard ID
func (s *Simulator) GenerateAndMintWalletAddress(targetShardID uint32, value *big.Int) (dtos.WalletAddress, error) {
	addressConverter := s.nodes[core.SovereignChainShardId].GetCoreComponents().AddressPubKeyConverter()
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

func (s *Simulator) setValidatorKeysForNode(node process.NodeHandler, validatorsPrivateKeys [][]byte) error {
	for idx, privateKey := range validatorsPrivateKeys {

		err := node.GetCryptoComponents().ManagedPeersHolder().AddManagedPeer(privateKey)
		if err != nil {
			return fmt.Errorf("cannot add private key for shard=%d, index=%d, error=%s", node.GetShardCoordinator().SelfId(), idx, err.Error())
		}
	}

	return nil
}

// GetValidatorPrivateKeys will return the initial validators private keys
func (s *Simulator) GetValidatorPrivateKeys() []crypto.PrivateKey {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.validatorsPrivateKeys
}

// SetKeyValueForAddress will set the provided state for a given address
func (s *Simulator) SetKeyValueForAddress(address string, keyValueMap map[string]string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	addressConverter := s.nodes[core.SovereignChainShardId].GetCoreComponents().AddressPubKeyConverter()
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

func (s *Simulator) setKeyValueSystemAccount(keyValueMap map[string]string) error {
	for shard, node := range s.nodes {
		err := node.SetKeyValueForAddress(core.SystemAccountAddress, keyValueMap)
		if err != nil {
			return fmt.Errorf("%w for shard %d", err, shard)
		}
	}

	return nil
}

// SetStateMultiple will set state for multiple addresses
func (s *Simulator) SetStateMultiple(stateSlice []*dtos.AddressState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	addressConverter := s.nodes[core.SovereignChainShardId].GetCoreComponents().AddressPubKeyConverter()
	for _, state := range stateSlice {
		addressBytes, err := addressConverter.Decode(state.Address)
		if err != nil {
			return err
		}

		if bytes.Equal(addressBytes, core.SystemAccountAddress) {
			err = s.setStateSystemAccount(state)
		} else {
			err = s.nodes[core.SovereignChainShardId].SetStateForAddress(addressBytes, state)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// SendTxAndGenerateBlockTilTxIsExecuted will send the provided transaction and generate block until the transaction is executed
func (s *Simulator) SendTxAndGenerateBlockTilTxIsExecuted(txToSend *transaction.Transaction, maxNumOfBlocksToGenerateWhenExecutingTx int) (*transaction.ApiTransactionResult, error) {
	result, err := s.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txToSend}, maxNumOfBlocksToGenerateWhenExecutingTx)
	if err != nil {
		return nil, err
	}

	return result[0], nil
}

// SendTxsAndGenerateBlocksTilAreExecuted will send the provided transactions and generate block until all transactions are executed
func (s *Simulator) SendTxsAndGenerateBlocksTilAreExecuted(txsToSend []*transaction.Transaction, maxNumOfBlocksToGenerateWhenExecutingTx int) ([]*transaction.ApiTransactionResult, error) {
	if len(txsToSend) == 0 {
		return nil, errEmptySliceOfTxs
	}
	if maxNumOfBlocksToGenerateWhenExecutingTx == 0 {
		return nil, errInvalidMaxNumOfBlocks
	}

	transactionStatus := make([]*transactionWithResult, 0, len(txsToSend))
	for idx, tx := range txsToSend {
		if tx == nil {
			return nil, fmt.Errorf("%w on position %d", errNilTransaction, idx)
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

func (s *Simulator) computeTransactionsStatus(txsWithResult []*transactionWithResult) bool {
	allAreExecuted := true
	for _, resultTx := range txsWithResult {
		if resultTx.result != nil {
			continue
		}

		result, errGet := s.GetNodeHandler(core.SovereignChainShardId).GetFacadeHandler().GetTransaction(resultTx.hexHash, true)
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

func (s *Simulator) sendTx(tx *transaction.Transaction) (string, error) {
	node := s.GetNodeHandler(core.SovereignChainShardId)
	err := node.GetFacadeHandler().ValidateTransaction(tx)
	if err != nil {
		return "", err
	}

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

func (s *Simulator) setStateSystemAccount(state *dtos.AddressState) error {
	for shard, node := range s.nodes {
		err := node.SetStateForAddress(core.SystemAccountAddress, state)
		if err != nil {
			return fmt.Errorf("%w for shard %d", err, shard)
		}
	}

	return nil
}

// GetAccount will fetch the account of the provided address
func (s *Simulator) GetAccount(address dtos.WalletAddress) (api.AccountResponse, error) {
	account, _, err := s.GetNodeHandler(core.SovereignChainShardId).GetFacadeHandler().GetAccount(address.Bech32, api.AccountQueryOptions{})
	return account, err
}

// Close will stop and close the Simulator
func (s *Simulator) Close() {
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
		log.Error("error closing chain Simulator", "error", components.AggregateErrors(errorStrings, components.ErrClose))
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *Simulator) IsInterfaceNil() bool {
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
