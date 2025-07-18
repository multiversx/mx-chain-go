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
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/heartbeat"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorErrors "github.com/multiversx/mx-chain-go/node/chainSimulator/errors"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
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
	BypassTxSignatureCheck     bool
	TempDir                    string
	PathToInitialConfig        string
	NumOfShards                uint32
	MinNodesPerShard           uint32
	MetaChainMinNodes          uint32
	Hysteresis                 float32
	NumNodesWaitingListShard   uint32
	NumNodesWaitingListMeta    uint32
	GenesisTimestamp           int64
	InitialRound               int64
	InitialEpoch               uint32
	InitialNonce               uint64
	RoundDurationInMillis      uint64
	RoundsPerEpoch             core.OptionalUint64
	ApiInterface               components.APIConfigurator
	AlterConfigsFunction       func(cfg *config.Configs)
	VmQueryDelayAfterStartInMs uint64
}

// ArgsBaseChainSimulator holds the arguments needed to create a new instance of simulator
type ArgsBaseChainSimulator struct {
	ArgsChainSimulator
	ConsensusGroupSize          uint32
	MetaChainConsensusGroupSize uint32
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
	return NewBaseChainSimulator(ArgsBaseChainSimulator{
		ArgsChainSimulator:          args,
		ConsensusGroupSize:          args.MinNodesPerShard,
		MetaChainConsensusGroupSize: args.MetaChainMinNodes,
	})
}

// NewBaseChainSimulator will create a new instance of simulator
func NewBaseChainSimulator(args ArgsBaseChainSimulator) (*simulator, error) {
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

func (s *simulator) createChainHandlers(args ArgsBaseChainSimulator) error {
	outputConfigs, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:                 args.NumOfShards,
		OriginalConfigsPath:         args.PathToInitialConfig,
		GenesisTimeStamp:            computeStartTimeBaseOnInitialRound(args.ArgsChainSimulator),
		RoundDurationInMillis:       args.RoundDurationInMillis,
		TempDir:                     args.TempDir,
		MinNodesPerShard:            args.MinNodesPerShard,
		ConsensusGroupSize:          args.ConsensusGroupSize,
		MetaChainMinNodes:           args.MetaChainMinNodes,
		MetaChainConsensusGroupSize: args.MetaChainConsensusGroupSize,
		Hysteresis:                  args.Hysteresis,
		RoundsPerEpoch:              args.RoundsPerEpoch,
		InitialEpoch:                args.InitialEpoch,
		AlterConfigsFunction:        args.AlterConfigsFunction,
		NumNodesWaitingListShard:    args.NumNodesWaitingListShard,
		NumNodesWaitingListMeta:     args.NumNodesWaitingListMeta,
	})
	if err != nil {
		return err
	}

	monitor := heartbeat.NewHeartbeatMonitor()

	for idx := -1; idx < int(args.NumOfShards); idx++ {
		shardIDStr := fmt.Sprintf("%d", idx)
		if idx == -1 {
			shardIDStr = "metachain"
		}

		node, errCreate := s.createTestNode(*outputConfigs, args, shardIDStr, monitor)
		if errCreate != nil {
			return errCreate
		}

		chainHandler, errCreate := process.NewBlocksCreator(node, monitor)
		if errCreate != nil {
			return errCreate
		}

		shardID := node.GetShardCoordinator().SelfId()
		s.nodes[shardID] = node
		s.handlers = append(s.handlers, chainHandler)

		var epochStartBlockHeader data.HeaderHandler

		if node.GetShardCoordinator().SelfId() == core.MetachainShardId {
			currentRootHash, errRootHash := node.GetProcessComponents().ValidatorsStatistics().RootHash()
			if errRootHash != nil {
				return errRootHash
			}

			allValidatorsInfo, errGet := node.GetProcessComponents().ValidatorsStatistics().GetValidatorInfoForRootHash(currentRootHash)
			if errGet != nil {
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

			epochStartBlockHeader = &block.MetaBlock{
				Nonce:     args.InitialNonce,
				Epoch:     args.InitialEpoch,
				Round:     uint64(args.InitialRound),
				TimeStamp: uint64(node.GetCoreComponents().RoundHandler().TimeStamp().Unix()),
			}
		} else {
			epochStartBlockHeader = &block.HeaderV2{
				Header: &block.Header{
					Nonce:     args.InitialNonce,
					Epoch:     args.InitialEpoch,
					Round:     uint64(args.InitialRound),
					TimeStamp: uint64(node.GetCoreComponents().RoundHandler().TimeStamp().Unix()),
				},
			}
		}

		err = node.GetProcessComponents().BlockchainHook().SetEpochStartHeader(epochStartBlockHeader)
		if err != nil {
			return err
		}
	}

	s.initialWalletKeys = outputConfigs.InitialWallets
	s.validatorsPrivateKeys = outputConfigs.ValidatorsPrivateKeys

	s.addProofs()
	s.setBasePeerIds()

	log.Info("running the chain simulator with the following parameters",
		"number of shards (including meta)", args.NumOfShards+1,
		"round per epoch", outputConfigs.Configs.GeneralConfig.EpochStartConfig.RoundsPerEpoch,
		"round duration", time.Millisecond*time.Duration(args.RoundDurationInMillis),
		"genesis timestamp", args.GenesisTimestamp,
		"original config path", args.PathToInitialConfig,
		"temporary path", args.TempDir)

	return nil
}

func (s *simulator) setBasePeerIds() {
	peerIds := make(map[uint32]core.PeerID, 0)
	for _, nodeHandler := range s.nodes {
		peerID := nodeHandler.GetNetworkComponents().NetworkMessenger().ID()
		peerIds[nodeHandler.GetShardCoordinator().SelfId()] = peerID
	}

	for _, nodeHandler := range s.nodes {
		nodeHandler.SetBasePeers(peerIds)
	}
}

func (s *simulator) addProofs() {
	proofs := make([]*block.HeaderProof, 0, len(s.nodes))

	for shardID, nodeHandler := range s.nodes {
		hash := nodeHandler.GetChainHandler().GetGenesisHeaderHash()
		proofs = append(proofs, &block.HeaderProof{
			HeaderShardId: shardID,
			HeaderHash:    hash,
		})
	}

	metachainProofsPool := s.GetNodeHandler(core.MetachainShardId).GetDataComponents().Datapool().Proofs()
	for _, proof := range proofs {
		_ = metachainProofsPool.AddProof(proof)

		if proof.HeaderShardId != core.MetachainShardId {
			_ = s.GetNodeHandler(proof.HeaderShardId).GetDataComponents().Datapool().Proofs().AddProof(proof)
		}
	}
}

func computeStartTimeBaseOnInitialRound(args ArgsChainSimulator) int64 {
	return args.GenesisTimestamp + int64(args.RoundDurationInMillis/1000)*args.InitialRound
}

func (s *simulator) createTestNode(
	outputConfigs configs.ArgsConfigsSimulator, args ArgsBaseChainSimulator, shardIDStr string, monitor factory.HeartbeatV2Monitor,
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
		ConsensusGroupSize:          args.ConsensusGroupSize,
		MinNodesMeta:                args.MetaChainMinNodes,
		MetaChainConsensusGroupSize: args.MetaChainConsensusGroupSize,
		RoundDurationInMillis:       args.RoundDurationInMillis,
		VmQueryDelayAfterStartInMs:  args.VmQueryDelayAfterStartInMs,
		Monitor:                     monitor,
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

// ForceChangeOfEpoch will force the change of current epoch
// This method will call the epoch change trigger and generate block till a new epoch is reached
func (s *simulator) ForceChangeOfEpoch() error {
	s.mutex.Lock()
	log.Info("force change of epoch")
	for shardID, node := range s.nodes {
		err := node.ForceChangeOfEpoch()
		if err != nil {
			s.mutex.Unlock()
			return fmt.Errorf("force change of epoch shardID-%d: error-%w", shardID, err)
		}
	}

	epoch := s.nodes[core.MetachainShardId].GetProcessComponents().EpochStartTrigger().Epoch()
	s.mutex.Unlock()

	err := s.GenerateBlocksUntilEpochIsReached(int32(epoch + 1))
	if err != nil {
		return err
	}

	s.incrementRoundOnAllValidators()

	return s.allNodesCreateBlocks()
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
	wallet := s.GenerateAddressInShard(targetShardID)

	err := s.SetStateMultiple([]*dtos.AddressState{
		{
			Address: wallet.Bech32,
			Balance: value.String(),
		},
	})

	return wallet, err
}

// GenerateAddressInShard will generate a wallet address based on the provided shard
func (s *simulator) GenerateAddressInShard(providedShardID uint32) dtos.WalletAddress {
	converter := s.nodes[core.MetachainShardId].GetCoreComponents().AddressPubKeyConverter()
	nodeHandler := s.GetNodeHandler(providedShardID)
	if check.IfNil(nodeHandler) {
		return generateWalletAddress(converter)
	}

	for {
		buff := generateAddress(converter.Len())
		if nodeHandler.GetShardCoordinator().ComputeId(buff) == providedShardID {
			return generateWalletAddressFromBuffer(converter, buff)
		}
	}
}

func generateWalletAddress(converter core.PubkeyConverter) dtos.WalletAddress {
	buff := generateAddress(converter.Len())
	return generateWalletAddressFromBuffer(converter, buff)
}

func generateWalletAddressFromBuffer(converter core.PubkeyConverter, buff []byte) dtos.WalletAddress {
	return dtos.WalletAddress{
		Bech32: converter.SilentEncode(buff, log),
		Bytes:  buff,
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

	addressConverter := s.nodes[core.MetachainShardId].GetCoreComponents().AddressPubKeyConverter()
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

	addressConverter := s.nodes[core.MetachainShardId].GetCoreComponents().AddressPubKeyConverter()
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

	addressConverter := s.nodes[core.MetachainShardId].GetCoreComponents().AddressPubKeyConverter()
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
	contractDeploySCAddress := make([]byte, s.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Len())
	for _, resultTx := range txsWithResult {
		if resultTx.result != nil {
			continue
		}

		sentTx := resultTx.tx
		destinationShardID := s.GetNodeHandler(0).GetShardCoordinator().ComputeId(sentTx.RcvAddr)
		if bytes.Equal(sentTx.RcvAddr, contractDeploySCAddress) {
			destinationShardID = s.GetNodeHandler(0).GetShardCoordinator().ComputeId(sentTx.SndAddr)
		}

		result, errGet := s.GetNodeHandler(destinationShardID).GetFacadeHandler().GetTransaction(resultTx.hexHash, true)
		if errGet == nil && result.Status != transaction.TxStatusPending {
			log.Trace("############## transaction was executed ##############", "txHash", resultTx.hexHash)
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
			log.Trace("############## send transaction ##############", "txHash", txHashHex)
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
