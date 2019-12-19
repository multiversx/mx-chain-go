package node

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sync/atomic"
	"time"

	nodeCmdFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat/storage"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/dataValidators"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/process/sync/storageBootstrap"
	procTx "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

// SendTransactionsPipe is the pipe used for sending new transactions
const SendTransactionsPipe = "send transactions pipe"

// HeartbeatTopic is the topic used for heartbeat signaling
const HeartbeatTopic = "heartbeat"

var log = logger.GetOrCreate("node")

// Option represents a functional configuration parameter that can operate
//  over the None struct.
type Option func(*Node) error

// Node is a structure that passes the configuration parameters and initializes
//  required services as requested
type Node struct {
	marshalizer              marshal.Marshalizer
	sizeCheckDelta           uint32
	ctx                      context.Context
	hasher                   hashing.Hasher
	feeHandler               process.FeeHandler
	initialNodesPubkeys      map[uint32][]string
	initialNodesBalances     map[string]*big.Int
	roundDuration            uint64
	consensusGroupSize       int
	messenger                P2PMessenger
	syncTimer                ntp.SyncTimer
	rounder                  consensus.Rounder
	blockProcessor           process.BlockProcessor
	genesisTime              time.Time
	epochStartTrigger        epochStart.TriggerHandler
	accounts                 state.AccountsAdapter
	addrConverter            state.AddressConverter
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	interceptorsContainer    process.InterceptorsContainer
	resolversFinder          dataRetriever.ResolversFinder
	heartbeatMonitor         *heartbeat.Monitor
	heartbeatSender          *heartbeat.Sender
	appStatusHandler         core.AppStatusHandler
	validatorStatistics      process.ValidatorStatisticsProcessor

	txSignPrivKey     crypto.PrivateKey
	txSignPubKey      crypto.PublicKey
	pubKey            crypto.PublicKey
	privKey           crypto.PrivateKey
	keyGen            crypto.KeyGenerator
	keyGenForAccounts crypto.KeyGenerator
	singleSigner      crypto.SingleSigner
	txSingleSigner    crypto.SingleSigner
	multiSigner       crypto.MultiSigner
	forkDetector      process.ForkDetector

	blkc             data.ChainHandler
	dataPool         dataRetriever.PoolsHolder
	metaDataPool     dataRetriever.MetaPoolsHolder
	store            dataRetriever.StorageService
	shardCoordinator sharding.Coordinator
	nodesCoordinator sharding.NodesCoordinator

	consensusTopic string
	consensusType  string

	isRunning                bool
	txStorageSize            uint32
	currentSendingGoRoutines int32
	bootstrapRoundIndex      uint64

	indexer               indexer.Indexer
	blackListHandler      process.BlackListHandler
	bootStorer            process.BootStorer
	requestedItemsHandler dataRetriever.RequestedItemsHandler
	headerSigVerifier     spos.RandSeedVerifier

	chainID []byte
}

// ApplyOptions can set up different configurable options of a Node instance
func (n *Node) ApplyOptions(opts ...Option) error {
	if n.IsRunning() {
		return errors.New("cannot apply options while node is running")
	}
	for _, opt := range opts {
		err := opt(n)
		if err != nil {
			return errors.New("error applying option: " + err.Error())
		}
	}
	return nil
}

// NewNode creates a new Node instance
func NewNode(opts ...Option) (*Node, error) {
	node := &Node{
		ctx:                      context.Background(),
		currentSendingGoRoutines: 0,
		appStatusHandler:         statusHandler.NewNilStatusHandler(),
	}
	for _, opt := range opts {
		err := opt(node)
		if err != nil {
			return nil, errors.New("error applying option: " + err.Error())
		}
	}

	return node, nil
}

// GetAppStatusHandler will return the current status handler
func (n *Node) GetAppStatusHandler() core.AppStatusHandler {
	return n.appStatusHandler
}

// IsRunning will return the current state of the node
func (n *Node) IsRunning() bool {
	return n.isRunning
}

// Start will create a new messenger and and set up the Node state as running
func (n *Node) Start() error {
	err := n.P2PBootstrap()
	if err == nil {
		n.isRunning = true
	}
	return err
}

// Stop closes the messenger and undos everything done in Start
func (n *Node) Stop() error {
	if !n.IsRunning() {
		return nil
	}
	err := n.messenger.Close()
	if err != nil {
		return err
	}

	return nil
}

// P2PBootstrap will try to connect to many peers as possible
func (n *Node) P2PBootstrap() error {
	if n.messenger == nil {
		return ErrNilMessenger
	}

	return n.messenger.Bootstrap()
}

// CreateShardedStores instantiate sharded cachers for Transactions and Headers
func (n *Node) CreateShardedStores() error {
	if n.shardCoordinator == nil {
		return ErrNilShardCoordinator
	}

	if n.dataPool == nil {
		return ErrNilDataPool
	}

	transactionsDataStore := n.dataPool.Transactions()
	headersDataStore := n.dataPool.Headers()

	if transactionsDataStore == nil {
		return errors.New("nil transaction sharded data store")
	}

	if headersDataStore == nil {
		return errors.New("nil header sharded data store")
	}

	shards := n.shardCoordinator.NumberOfShards()
	currentShardId := n.shardCoordinator.SelfId()

	transactionsDataStore.CreateShardStore(process.ShardCacherIdentifier(currentShardId, currentShardId))
	for i := uint32(0); i < shards; i++ {
		if i == n.shardCoordinator.SelfId() {
			continue
		}
		transactionsDataStore.CreateShardStore(process.ShardCacherIdentifier(i, currentShardId))
		transactionsDataStore.CreateShardStore(process.ShardCacherIdentifier(currentShardId, i))
	}

	return nil
}

// StartConsensus will start the consensus service for the current node
func (n *Node) StartConsensus() error {
	isGenesisBlockNotInitialized := n.blkc.GetGenesisHeaderHash() == nil ||
		n.blkc.GetGenesisHeader() == nil
	if isGenesisBlockNotInitialized {
		return ErrGenesisBlockNotInitialized
	}

	chronologyHandler, err := n.createChronologyHandler(n.rounder, n.appStatusHandler)
	if err != nil {
		return err
	}

	bootstrapper, err := n.createBootstrapper(n.rounder)
	if err != nil {
		return err
	}

	err = bootstrapper.SetStatusHandler(n.GetAppStatusHandler())
	if err != nil {
		log.Debug("cannot set app status handler for shard bootstrapper")
	}

	bootstrapper.StartSync()

	consensusState, err := n.createConsensusState()
	if err != nil {
		return err
	}

	consensusService, err := sposFactory.GetConsensusCoreFactory(n.consensusType)
	if err != nil {
		return err
	}

	broadcastMessenger, err := sposFactory.GetBroadcastMessenger(
		n.marshalizer,
		n.messenger,
		n.shardCoordinator,
		n.privKey,
		n.singleSigner)

	if err != nil {
		return err
	}

	netInputMarshalizer := n.marshalizer
	if n.sizeCheckDelta > 0 {
		netInputMarshalizer = marshal.NewSizeCheckUnmarshalizer(n.marshalizer, n.sizeCheckDelta)
	}
	worker, err := spos.NewWorker(
		consensusService,
		n.blkc,
		n.blockProcessor,
		bootstrapper,
		broadcastMessenger,
		consensusState,
		n.forkDetector,
		n.keyGen,
		netInputMarshalizer,
		n.rounder,
		n.shardCoordinator,
		n.singleSigner,
		n.syncTimer,
		n.headerSigVerifier,
		n.chainID,
	)
	if err != nil {
		return err
	}

	err = n.createConsensusTopic(worker, n.shardCoordinator)
	if err != nil {
		return err
	}

	consensusDataContainer, err := spos.NewConsensusCore(
		n.blkc,
		n.blockProcessor,
		bootstrapper,
		broadcastMessenger,
		chronologyHandler,
		n.hasher,
		n.marshalizer,
		n.privKey,
		n.singleSigner,
		n.multiSigner,
		n.rounder,
		n.shardCoordinator,
		n.nodesCoordinator,
		n.syncTimer,
	)
	if err != nil {
		return err
	}

	fct, err := sposFactory.GetSubroundsFactory(
		consensusDataContainer,
		consensusState,
		worker,
		n.consensusType,
		n.appStatusHandler,
		n.indexer,
		n.chainID,
	)
	if err != nil {
		return err
	}

	err = fct.GenerateSubrounds()
	if err != nil {
		return err
	}

	go chronologyHandler.StartRounds()

	return nil
}

// GetBalance gets the balance for a specific address
func (n *Node) GetBalance(addressHex string) (*big.Int, error) {
	if n.addrConverter == nil || n.addrConverter.IsInterfaceNil() || n.accounts == nil || n.accounts.IsInterfaceNil() {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}

	address, err := n.addrConverter.CreateAddressFromHex(addressHex)
	if err != nil {
		return nil, errors.New("invalid address, could not decode from hex: " + err.Error())
	}
	accWrp, err := n.accounts.GetExistingAccount(address)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param: " + err.Error())
	}

	if accWrp == nil || accWrp.IsInterfaceNil() {
		return big.NewInt(0), nil
	}

	account, ok := accWrp.(*state.Account)
	if !ok {
		return big.NewInt(0), nil
	}

	return account.Balance, nil
}

// createChronologyHandler method creates a chronology object
func (n *Node) createChronologyHandler(rounder consensus.Rounder, appStatusHandler core.AppStatusHandler) (consensus.ChronologyHandler, error) {
	chr, err := chronology.NewChronology(
		n.genesisTime,
		rounder,
		n.syncTimer,
	)

	if err != nil {
		return nil, err
	}

	err = chr.SetAppStatusHandler(appStatusHandler)
	if err != nil {
		return nil, err
	}

	return chr, nil
}

//TODO move this func in structs.go
func (n *Node) createBootstrapper(rounder consensus.Rounder) (process.Bootstrapper, error) {
	if n.shardCoordinator.SelfId() < n.shardCoordinator.NumberOfShards() {
		return n.createShardBootstrapper(rounder)
	}

	if n.shardCoordinator.SelfId() == sharding.MetachainShardId {
		return n.createMetaChainBootstrapper(rounder)
	}

	return nil, sharding.ErrShardIdOutOfRange
}

func (n *Node) createShardBootstrapper(rounder consensus.Rounder) (process.Bootstrapper, error) {
	storageBootstrapArguments := storageBootstrap.ArgsStorageBootstrapper{
		ResolversFinder:     n.resolversFinder,
		BootStorer:          n.bootStorer,
		ForkDetector:        n.forkDetector,
		BlockProcessor:      n.blockProcessor,
		ChainHandler:        n.blkc,
		Marshalizer:         n.marshalizer,
		Store:               n.store,
		Uint64Converter:     n.uint64ByteSliceConverter,
		BootstrapRoundIndex: n.bootstrapRoundIndex,
		ShardCoordinator:    n.shardCoordinator,
	}

	shardStorageBootstrapper, err := storageBootstrap.NewShardStorageBootstrapper(storageBootstrapArguments)
	if err != nil {
		return nil, err
	}

	bootstrap, err := sync.NewShardBootstrap(
		n.dataPool,
		n.store,
		n.blkc,
		rounder,
		n.blockProcessor,
		n.rounder.TimeDuration(),
		n.hasher,
		n.marshalizer,
		n.forkDetector,
		n.resolversFinder,
		n.shardCoordinator,
		n.accounts,
		n.blackListHandler,
		n.messenger,
		n.bootStorer,
		shardStorageBootstrapper,
		n.requestedItemsHandler,
	)
	if err != nil {
		return nil, err
	}

	return bootstrap, nil
}

func (n *Node) createMetaChainBootstrapper(rounder consensus.Rounder) (process.Bootstrapper, error) {
	storageBootstrapArguments := storageBootstrap.ArgsStorageBootstrapper{
		ResolversFinder:     n.resolversFinder,
		BootStorer:          n.bootStorer,
		ForkDetector:        n.forkDetector,
		BlockProcessor:      n.blockProcessor,
		ChainHandler:        n.blkc,
		Marshalizer:         n.marshalizer,
		Store:               n.store,
		Uint64Converter:     n.uint64ByteSliceConverter,
		BootstrapRoundIndex: n.bootstrapRoundIndex,
		ShardCoordinator:    n.shardCoordinator,
	}

	metaStorageBootstrapper, err := storageBootstrap.NewMetaStorageBootstrapper(storageBootstrapArguments)
	if err != nil {
		return nil, err
	}

	bootstrap, err := sync.NewMetaBootstrap(
		n.metaDataPool,
		n.store,
		n.blkc,
		rounder,
		n.blockProcessor,
		n.rounder.TimeDuration(),
		n.hasher,
		n.marshalizer,
		n.forkDetector,
		n.resolversFinder,
		n.shardCoordinator,
		n.accounts,
		n.blackListHandler,
		n.messenger,
		n.bootStorer,
		metaStorageBootstrapper,
		n.requestedItemsHandler,
	)

	if err != nil {
		return nil, err
	}

	return bootstrap, nil
}

// createConsensusState method creates a consensusState object
func (n *Node) createConsensusState() (*spos.ConsensusState, error) {
	selfId, err := n.pubKey.ToByteArray()

	if err != nil {
		return nil, err
	}

	roundConsensus := spos.NewRoundConsensus(
		n.initialNodesPubkeys[n.shardCoordinator.SelfId()],
		n.consensusGroupSize,
		string(selfId))

	roundConsensus.ResetRoundState()

	roundThreshold := spos.NewRoundThreshold()

	roundStatus := spos.NewRoundStatus()
	roundStatus.ResetRoundStatus()

	consensusState := spos.NewConsensusState(
		roundConsensus,
		roundThreshold,
		roundStatus)

	return consensusState, nil
}

// createConsensusTopic creates a consensus topic for node
func (n *Node) createConsensusTopic(messageProcessor p2p.MessageProcessor, shardCoordinator sharding.Coordinator) error {
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return ErrNilShardCoordinator
	}
	if messageProcessor == nil || messageProcessor.IsInterfaceNil() {
		return ErrNilMessenger
	}

	n.consensusTopic = core.ConsensusTopic + shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())
	if n.messenger.HasTopicValidator(n.consensusTopic) {
		return ErrValidatorAlreadySet
	}

	if !n.messenger.HasTopic(n.consensusTopic) {
		err := n.messenger.CreateTopic(n.consensusTopic, true)
		if err != nil {
			return err
		}
	}

	return n.messenger.RegisterMessageProcessor(n.consensusTopic, messageProcessor)
}

// SendTransaction will send a new transaction on the topic channel
func (n *Node) SendTransaction(
	nonce uint64,
	senderHex string,
	receiverHex string,
	value string,
	gasPrice uint64,
	gasLimit uint64,
	transactionData string,
	signature []byte) (string, error) {

	if n.shardCoordinator == nil || n.shardCoordinator.IsInterfaceNil() {
		return "", ErrNilShardCoordinator
	}

	sender, err := n.addrConverter.CreateAddressFromHex(senderHex)
	if err != nil {
		return "", err
	}

	receiver, err := n.addrConverter.CreateAddressFromHex(receiverHex)
	if err != nil {
		return "", err
	}

	senderShardId := n.shardCoordinator.ComputeId(sender)

	valAsBigInt, ok := big.NewInt(0).SetString(value, 10)
	if !ok {
		return "", ErrInvalidValue
	}

	tx := transaction.Transaction{
		Nonce:     nonce,
		Value:     valAsBigInt,
		RcvAddr:   receiver.Bytes(),
		SndAddr:   sender.Bytes(),
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      transactionData,
		Signature: signature,
	}

	err = n.validateTx(&tx)
	if err != nil {
		return "", err
	}

	txBuff, err := n.marshalizer.Marshal(&tx)
	if err != nil {
		return "", err
	}

	txHexHash := hex.EncodeToString(n.hasher.Compute(string(txBuff)))

	marshalizedTx, err := n.marshalizer.Marshal([][]byte{txBuff})
	if err != nil {
		return "", errors.New("could not marshal transaction")
	}

	//the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.TransactionTopic + n.shardCoordinator.CommunicationIdentifier(senderShardId)

	n.messenger.BroadcastOnChannel(
		SendTransactionsPipe,
		identifier,
		marshalizedTx,
	)

	return txHexHash, nil
}

// SendBulkTransactions sends the provided transactions as a bulk, optimizing transfer between nodes
func (n *Node) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	transactionsByShards := make(map[uint32][][]byte, 0)

	if txs == nil || len(txs) == 0 {
		return 0, ErrNoTxToProcess
	}

	for _, tx := range txs {
		senderBytes, err := n.addrConverter.CreateAddressFromPublicKeyBytes(tx.SndAddr)
		if err != nil {
			continue
		}

		senderShardId := n.shardCoordinator.ComputeId(senderBytes)
		marshalizedTx, err := n.marshalizer.Marshal(tx)
		if err != nil {
			continue
		}

		err = n.validateTx(tx)
		if err != nil {
			continue
		}

		transactionsByShards[senderShardId] = append(transactionsByShards[senderShardId], marshalizedTx)
	}

	numOfSentTxs := uint64(0)
	for shardId, txs := range transactionsByShards {
		err := n.sendBulkTransactionsFromShard(txs, shardId)
		if err != nil {
			log.Debug("sendBulkTransactionsFromShard", "error", err.Error())
		} else {
			numOfSentTxs += uint64(len(txs))
		}
	}

	return numOfSentTxs, nil
}

func (n *Node) validateTx(tx *transaction.Transaction) error {
	txValidator, err := dataValidators.NewTxValidator(n.accounts, n.shardCoordinator, nodeCmdFactory.MaxTxNonceDeltaAllowed)
	if err != nil {
		return nil
	}

	marshalizedTx, err := n.marshalizer.Marshal(tx)
	if err != nil {
		return err
	}

	intTx, err := procTx.NewInterceptedTransaction(
		marshalizedTx,
		n.marshalizer,
		n.hasher,
		n.keyGenForAccounts,
		n.txSingleSigner,
		n.addrConverter,
		n.shardCoordinator,
		n.feeHandler,
	)
	if err != nil {
		return err
	}

	err = intTx.CheckValidity()
	if err != nil {
		return err
	}

	return txValidator.CheckTxValidity(intTx)
}

func (n *Node) sendBulkTransactionsFromShard(transactions [][]byte, senderShardId uint32) error {
	dataPacker, err := partitioning.NewSimpleDataPacker(n.marshalizer)
	if err != nil {
		return err
	}

	//the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.TransactionTopic + n.shardCoordinator.CommunicationIdentifier(senderShardId)

	packets, err := dataPacker.PackDataInChunks(transactions, core.MaxBulkTransactionSize)
	if err != nil {
		return err
	}

	atomic.AddInt32(&n.currentSendingGoRoutines, int32(len(packets)))
	for _, buff := range packets {
		go func(bufferToSend []byte) {
			err = n.messenger.BroadcastOnChannelBlocking(
				SendTransactionsPipe,
				identifier,
				bufferToSend,
			)
			if err != nil {
				log.Debug("BroadcastOnChannelBlocking", "error", err.Error())
			}

			atomic.AddInt32(&n.currentSendingGoRoutines, -1)
		}(buff)
	}

	return nil
}

// CreateTransaction will return a transaction from all the required fields
func (n *Node) CreateTransaction(
	nonce uint64,
	value string,
	receiverHex string,
	senderHex string,
	gasPrice uint64,
	gasLimit uint64,
	data string,
	signatureHex string,
	challenge string,
) (*transaction.Transaction, error) {

	if n.addrConverter == nil || n.addrConverter.IsInterfaceNil() {
		return nil, ErrNilAddressConverter
	}

	if n.accounts == nil || n.accounts.IsInterfaceNil() {
		return nil, ErrNilAccountsAdapter
	}

	receiverAddress, err := n.addrConverter.CreateAddressFromHex(receiverHex)
	if err != nil {
		return nil, errors.New("could not create receiver address from provided param")
	}

	senderAddress, err := n.addrConverter.CreateAddressFromHex(senderHex)
	if err != nil {
		return nil, errors.New("could not create sender address from provided param")
	}

	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return nil, errors.New("could not fetch signature bytes")
	}

	challengeBytes, err := hex.DecodeString(challenge)
	if err != nil {
		return nil, errors.New("could not fetch challenge bytes")
	}

	valAsBigInt, ok := big.NewInt(0).SetString(value, 10)
	if !ok {
		return nil, ErrInvalidValue
	}

	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     valAsBigInt,
		RcvAddr:   receiverAddress.Bytes(),
		SndAddr:   senderAddress.Bytes(),
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      data,
		Signature: signatureBytes,
		Challenge: challengeBytes,
	}, nil
}

//GetTransaction gets the transaction
func (n *Node) GetTransaction(_ string) (*transaction.Transaction, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// GetCurrentPublicKey will return the current node's public key
func (n *Node) GetCurrentPublicKey() string {
	if n.txSignPubKey != nil {
		pkey, _ := n.txSignPubKey.ToByteArray()
		return fmt.Sprintf("%x", pkey)
	}
	return ""
}

// GetAccount will return acount details for a given address
func (n *Node) GetAccount(address string) (*state.Account, error) {
	if n.addrConverter == nil || n.addrConverter.IsInterfaceNil() {
		return nil, ErrNilAddressConverter
	}
	if n.accounts == nil || n.accounts.IsInterfaceNil() {
		return nil, ErrNilAccountsAdapter
	}

	addr, err := n.addrConverter.CreateAddressFromHex(address)
	if err != nil {
		return nil, err
	}

	accWrp, err := n.accounts.GetExistingAccount(addr)
	if err != nil {
		if err == state.ErrAccNotFound {
			return &state.Account{
				Balance:  big.NewInt(0),
				Nonce:    0,
				RootHash: nil,
				CodeHash: nil,
			}, nil
		}
		return nil, errors.New("could not fetch sender address from provided param: " + err.Error())
	}

	account, ok := accWrp.(*state.Account)
	if !ok {
		return nil, errors.New("account is not of type with balance and nonce")
	}

	return account, nil
}

// StartHeartbeat starts the node's heartbeat processing/signaling module
func (n *Node) StartHeartbeat(hbConfig config.HeartbeatConfig, versionNumber string, nodeDisplayName string) error {
	if !hbConfig.Enabled {
		return nil
	}

	err := n.checkConfigParams(hbConfig)
	if err != nil {
		return err
	}

	if n.messenger.HasTopicValidator(HeartbeatTopic) {
		return ErrValidatorAlreadySet
	}

	if !n.messenger.HasTopic(HeartbeatTopic) {
		err := n.messenger.CreateTopic(HeartbeatTopic, true)
		if err != nil {
			return err
		}
	}

	n.heartbeatSender, err = heartbeat.NewSender(
		n.messenger,
		n.singleSigner,
		n.privKey,
		n.marshalizer,
		HeartbeatTopic,
		n.shardCoordinator,
		versionNumber,
		nodeDisplayName,
	)
	if err != nil {
		return err
	}

	heartbeatStorageUnit := n.store.GetStorer(dataRetriever.HeartbeatUnit)

	heartBeatMsgProcessor, err := heartbeat.NewMessageProcessor(
		n.singleSigner,
		n.keyGen,
		n.marshalizer)
	if err != nil {
		return err
	}

	heartbeatStorer, err := storage.NewHeartbeatDbStorer(heartbeatStorageUnit, n.marshalizer)
	timer := &heartbeat.RealTimer{}
	netInputMarshalizer := n.marshalizer
	if n.sizeCheckDelta > 0 {
		netInputMarshalizer = marshal.NewSizeCheckUnmarshalizer(n.marshalizer, n.sizeCheckDelta)
	}
	n.heartbeatMonitor, err = heartbeat.NewMonitor(
		netInputMarshalizer,
		time.Second*time.Duration(hbConfig.DurationInSecToConsiderUnresponsive),
		n.initialNodesPubkeys,
		n.genesisTime,
		heartBeatMsgProcessor,
		heartbeatStorer,
		timer,
	)
	if err != nil {
		return err
	}

	err = n.heartbeatMonitor.SetAppStatusHandler(n.appStatusHandler)
	if err != nil {
		return err
	}

	err = n.messenger.RegisterMessageProcessor(HeartbeatTopic, n.heartbeatMonitor)
	if err != nil {
		return err
	}

	go n.startSendingHeartbeats(hbConfig)

	return nil
}

func (n *Node) checkConfigParams(config config.HeartbeatConfig) error {
	if config.DurationInSecToConsiderUnresponsive < 1 {
		return ErrNegativeDurationInSecToConsiderUnresponsive
	}
	if config.MaxTimeToWaitBetweenBroadcastsInSec < 1 {
		return ErrNegativeMaxTimeToWaitBetweenBroadcastsInSec
	}
	if config.MinTimeToWaitBetweenBroadcastsInSec < 1 {
		return ErrNegativeMinTimeToWaitBetweenBroadcastsInSec
	}
	if config.MaxTimeToWaitBetweenBroadcastsInSec <= config.MinTimeToWaitBetweenBroadcastsInSec {
		return ErrWrongValues
	}
	if config.DurationInSecToConsiderUnresponsive <= config.MaxTimeToWaitBetweenBroadcastsInSec {
		return ErrWrongValues
	}

	return nil
}

func (n *Node) startSendingHeartbeats(config config.HeartbeatConfig) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	for {
		diffSeconds := config.MaxTimeToWaitBetweenBroadcastsInSec - config.MinTimeToWaitBetweenBroadcastsInSec
		diffNanos := int64(diffSeconds) * time.Second.Nanoseconds()
		randomNanos := r.Int63n(diffNanos)
		timeToWait := time.Second*time.Duration(config.MinTimeToWaitBetweenBroadcastsInSec) + time.Duration(randomNanos)

		time.Sleep(timeToWait)

		err := n.heartbeatSender.SendHeartbeat()
		if err != nil {
			log.Debug("SendHeartbeat", "error", err.Error())
		}
	}
}

// GetHeartbeats returns the heartbeat status for each public key defined in genesis.json
func (n *Node) GetHeartbeats() []heartbeat.PubKeyHeartbeat {
	if n.heartbeatMonitor == nil {
		return nil
	}
	return n.heartbeatMonitor.GetHeartbeats()
}

// ValidatorStatisticsApi will return the statistics for all the validators from the initial nodes pub keys
func (n *Node) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	mapToReturn := make(map[string]*state.ValidatorApiResponse)
	for _, pubKeyInShards := range n.initialNodesPubkeys {
		for _, pubKey := range pubKeyInShards {
			acc, err := n.validatorStatistics.GetPeerAccount([]byte(pubKey))
			if err != nil {
				log.Debug("validator api: get peer account", "error", err.Error())
				continue
			}

			peerAcc, ok := acc.(*state.PeerAccount)
			if !ok {
				log.Debug("validator api: convert to peer account", "error", ErrCannotConvertToPeerAccount)
				continue
			}

			strKey := hex.EncodeToString([]byte(pubKey))
			mapToReturn[strKey] = &state.ValidatorApiResponse{
				NrLeaderSuccess:    peerAcc.LeaderSuccessRate.NrSuccess,
				NrLeaderFailure:    peerAcc.LeaderSuccessRate.NrFailure,
				NrValidatorSuccess: peerAcc.ValidatorSuccessRate.NrSuccess,
				NrValidatorFailure: peerAcc.ValidatorSuccessRate.NrFailure,
			}
		}
	}

	return mapToReturn, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *Node) IsInterfaceNil() bool {
	if n == nil {
		return true
	}
	return false
}
