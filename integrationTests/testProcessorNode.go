package integrationTests

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	metafactoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/metachain"
	factoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	metaProcess "github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/pkg/errors"
)

// TestHasher represents a Sha256 hasher
var TestHasher = sha256.Sha256{}

// TestMarshalizer represents a JSON marshalizer
var TestMarshalizer = &marshal.JsonMarshalizer{}

// TestAddressConverter represents a plain address converter
var TestAddressConverter, _ = addressConverters.NewPlainAddressConverter(32, "0x")

// TestMultiSig represents a mock multisig
var TestMultiSig = mock.NewMultiSigner(1)

// TestUint64Converter represents an uint64 to byte slice converter
var TestUint64Converter = uint64ByteSlice.NewBigEndianConverter()

// TestProcessorNode represents a container type of class used in integration tests
// with all its fields exported
type TestProcessorNode struct {
	ShardCoordinator sharding.Coordinator
	Messenger        p2p.Messenger

	SingleSigner  crypto.SingleSigner
	SkTxSign      crypto.PrivateKey
	PkTxSign      crypto.PublicKey
	PkTxSignBytes []byte
	KeygenTxSign  crypto.KeyGenerator
	TxSignAddress state.AddressContainer

	ShardDataPool dataRetriever.PoolsHolder
	MetaDataPool  dataRetriever.MetaPoolsHolder
	Storage       dataRetriever.StorageService
	AccntState    state.AccountsAdapter
	BlockChain    data.ChainHandler
	GenesisBlocks map[uint32]data.HeaderHandler

	InterceptorsContainer process.InterceptorsContainer
	ResolversContainer    dataRetriever.ResolversContainer
	ResolverFinder        dataRetriever.ResolversFinder
	RequestHandler        process.RequestHandler

	InterimProcContainer   process.IntermediateProcessorContainer
	TxProcessor            process.TransactionProcessor
	TxCoordinator          process.TransactionCoordinator
	ScrForwarder           process.IntermediateTransactionHandler
	VmProcessor            vmcommon.VMExecutionHandler
	VmDataGetter           vmcommon.VMExecutionHandler
	BlockchainHook         vmcommon.BlockchainHook
	ArgsParser             process.ArgumentsParser
	ScProcessor            process.SmartContractProcessor
	PreProcessorsContainer process.PreProcessorsContainer

	ForkDetector       process.ForkDetector
	BlockTracker       process.BlocksTracker
	BlockProcessor     process.BlockProcessor
	BroadcastMessenger consensus.BroadcastMessenger

	//Node is used to call the functionality already implemented in it
	Node         *node.Node
	ScDataGetter external.ScDataGetter

	CounterHdrRecv int32
	CounterMbRecv  int32
	CounterTxRecv  int32
	CounterMetaRcv int32
}

// NewTestProcessorNode returns a new TestProcessorNode instance
func NewTestProcessorNode(maxShards uint32, nodeShardId uint32, txSignPrivKeyShardId uint32, initialNodeAddr string) *TestProcessorNode {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator: shardCoordinator,
		Messenger:        messenger,
	}

	tpn.initCrypto(txSignPrivKeyShardId)
	tpn.initDataPools()
	tpn.initStorage()
	tpn.AccntState, _, _ = CreateAccountsDB(tpn.ShardCoordinator)
	tpn.initChainHandler()
	tpn.GenesisBlocks = CreateGenesisBlocks(tpn.ShardCoordinator)
	tpn.initInterceptors()
	tpn.initResolvers()
	tpn.initInnerProcessors()
	tpn.initBlockProcessor()
	tpn.BroadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		TestMarshalizer,
		tpn.Messenger,
		tpn.ShardCoordinator,
		tpn.SkTxSign,
		tpn.SingleSigner,
	)
	tpn.setGenesisBlock()
	tpn.initNode()
	tpn.ScDataGetter, _ = smartContract.NewSCDataGetter(tpn.VmDataGetter)
	tpn.addHandlersForCounters()

	return tpn
}

func (tpn *TestProcessorNode) initCrypto(txSignPrivKeyShardId uint32) {
	suite := kyber.NewBlakeSHA256Ed25519()
	tpn.SingleSigner = &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	for {
		pkBytes, _ := pk.ToByteArray()
		addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(pkBytes)
		if tpn.ShardCoordinator.ComputeId(addr) == txSignPrivKeyShardId {
			break
		}
		sk, pk = keyGen.GeneratePair()
	}

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s in shard %d\n", hex.EncodeToString(pkBuff), txSignPrivKeyShardId)

	tpn.SkTxSign = sk
	tpn.PkTxSign = pk
	tpn.PkTxSignBytes, _ = pk.ToByteArray()
	tpn.KeygenTxSign = keyGen
	tpn.TxSignAddress, _ = TestAddressConverter.CreateAddressFromPublicKeyBytes(tpn.PkTxSignBytes)
}

func (tpn *TestProcessorNode) initDataPools() {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.MetaDataPool = CreateTestMetaDataPool()
	} else {
		tpn.ShardDataPool = CreateTestShardDataPool()
	}
}

func (tpn *TestProcessorNode) initStorage() {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.Storage = CreateMetaStore(tpn.ShardCoordinator)
	} else {
		tpn.Storage = CreateShardStore(tpn.ShardCoordinator.NumberOfShards())
	}
}

func (tpn *TestProcessorNode) initChainHandler() {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.BlockChain = CreateMetaChain()
	} else {
		tpn.BlockChain = CreateShardChain()
	}
}

func (tpn *TestProcessorNode) initInterceptors() {
	var err error
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		interceptorContainerFactory, _ := metaProcess.NewInterceptorsContainerFactory(
			tpn.ShardCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			TestHasher,
			TestMultiSig,
			tpn.MetaDataPool,
			&mock.ChronologyValidatorMock{},
		)

		tpn.InterceptorsContainer, err = interceptorContainerFactory.Create()
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		interceptorContainerFactory, _ := shard.NewInterceptorsContainerFactory(
			tpn.ShardCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			TestHasher,
			tpn.KeygenTxSign,
			tpn.SingleSigner,
			TestMultiSig,
			tpn.ShardDataPool,
			TestAddressConverter,
			&mock.ChronologyValidatorMock{},
		)

		tpn.InterceptorsContainer, err = interceptorContainerFactory.Create()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (tpn *TestProcessorNode) initResolvers() {
	dataPacker, _ := partitioning.NewSizeDataPacker(TestMarshalizer)

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		resolversContainerFactory, _ := metafactoryDataRetriever.NewResolversContainerFactory(
			tpn.ShardCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			tpn.MetaDataPool,
			TestUint64Converter,
		)

		tpn.ResolversContainer, _ = resolversContainerFactory.Create()
		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewMetaResolverRequestHandler(
			tpn.ResolverFinder,
			factory.ShardHeadersForMetachainTopic,
		)
	} else {
		resolversContainerFactory, _ := factoryDataRetriever.NewResolversContainerFactory(
			tpn.ShardCoordinator,
			tpn.Messenger,
			tpn.Storage,
			TestMarshalizer,
			tpn.ShardDataPool,
			TestUint64Converter,
			dataPacker,
		)

		tpn.ResolversContainer, _ = resolversContainerFactory.Create()
		tpn.ResolverFinder, _ = containers.NewResolversFinder(tpn.ResolversContainer, tpn.ShardCoordinator)
		tpn.RequestHandler, _ = requestHandlers.NewShardResolverRequestHandler(
			tpn.ResolverFinder,
			factory.TransactionTopic,
			factory.UnsignedTransactionTopic,
			factory.MiniBlocksTopic,
			factory.MetachainBlocksTopic,
			100,
		)
	}
}

func (tpn *TestProcessorNode) initInnerProcessors() {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		return
	}

	interimProcFactory, _ := shard.NewIntermediateProcessorsContainerFactory(
		tpn.ShardCoordinator,
		TestMarshalizer,
		TestHasher,
		TestAddressConverter,
		tpn.Storage,
	)
	tpn.InterimProcContainer, _ = interimProcFactory.Create()
	tpn.ScrForwarder, _ = tpn.InterimProcContainer.Get(dataBlock.SmartContractResultBlock)

	tpn.VmProcessor, tpn.BlockchainHook = CreateIeleVMAndBlockchainHook(tpn.AccntState)
	tpn.VmDataGetter, _ = CreateIeleVMAndBlockchainHook(tpn.AccntState)

	tpn.ArgsParser, _ = smartContract.NewAtArgumentParser()
	tpn.ScProcessor, _ = smartContract.NewSmartContractProcessor(
		tpn.VmProcessor,
		tpn.ArgsParser,
		TestHasher,
		TestMarshalizer,
		tpn.AccntState,
		tpn.BlockchainHook.(*hooks.VMAccountsDB),
		TestAddressConverter,
		tpn.ShardCoordinator,
		tpn.ScrForwarder,
	)

	tpn.TxProcessor, _ = transaction.NewTxProcessor(
		tpn.AccntState,
		TestHasher,
		TestAddressConverter,
		TestMarshalizer,
		tpn.ShardCoordinator,
		tpn.ScProcessor,
	)

	fact, _ := shard.NewPreProcessorsContainerFactory(
		tpn.ShardCoordinator,
		tpn.Storage,
		TestMarshalizer,
		TestHasher,
		tpn.ShardDataPool,
		TestAddressConverter,
		tpn.AccntState,
		tpn.RequestHandler,
		tpn.TxProcessor,
		tpn.ScProcessor,
		tpn.ScProcessor.(process.SmartContractResultProcessor),
	)
	tpn.PreProcessorsContainer, _ = fact.Create()

	tpn.TxCoordinator, _ = coordinator.NewTransactionCoordinator(
		tpn.ShardCoordinator,
		tpn.AccntState,
		tpn.ShardDataPool,
		tpn.RequestHandler,
		tpn.PreProcessorsContainer,
		tpn.InterimProcContainer,
	)
}

func (tpn *TestProcessorNode) initBlockProcessor() {
	var err error

	tpn.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
	}

	tpn.BlockTracker = &mock.BlocksTrackerMock{
		AddBlockCalled: func(headerHandler data.HeaderHandler) {
		},
		RemoveNotarisedBlocksCalled: func(headerHandler data.HeaderHandler) error {
			return nil
		},
		UnnotarisedBlocksCalled: func() []data.HeaderHandler {
			return make([]data.HeaderHandler, 0)
		},
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.BlockProcessor, err = block.NewMetaProcessor(
			&mock.ServiceContainerMock{},
			tpn.AccntState,
			tpn.MetaDataPool,
			tpn.ForkDetector,
			tpn.ShardCoordinator,
			TestHasher,
			TestMarshalizer,
			tpn.Storage,
			tpn.GenesisBlocks,
			tpn.RequestHandler,
			TestUint64Converter,
		)
	} else {
		tpn.BlockProcessor, err = block.NewShardProcessor(
			nil,
			tpn.ShardDataPool,
			tpn.Storage,
			TestHasher,
			TestMarshalizer,
			tpn.AccntState,
			tpn.ShardCoordinator,
			tpn.ForkDetector,
			tpn.BlockTracker,
			tpn.GenesisBlocks,
			tpn.RequestHandler,
			tpn.TxCoordinator,
			TestUint64Converter,
		)
	}

	if err != nil {
		fmt.Printf("Error creating blockprocessor: %s\n", err.Error())
	}
}

func (tpn *TestProcessorNode) setGenesisBlock() {
	genesisBlock := tpn.GenesisBlocks[tpn.ShardCoordinator.SelfId()]
	_ = tpn.BlockChain.SetGenesisHeader(genesisBlock)
	hash, _ := core.CalculateHash(TestMarshalizer, TestHasher, genesisBlock)
	tpn.BlockChain.SetGenesisHeaderHash(hash)
}

func (tpn *TestProcessorNode) initNode() {
	var err error

	tpn.Node, err = node.NewNode(
		node.WithMessenger(tpn.Messenger),
		node.WithMarshalizer(TestMarshalizer),
		node.WithHasher(TestHasher),
		node.WithHasher(TestHasher),
		node.WithAddressConverter(TestAddressConverter),
		node.WithAccountsAdapter(tpn.AccntState),
		node.WithKeyGen(tpn.KeygenTxSign),
		node.WithShardCoordinator(tpn.ShardCoordinator),
		node.WithBlockChain(tpn.BlockChain),
		node.WithUint64ByteSliceConverter(TestUint64Converter),
		node.WithMultiSigner(TestMultiSig),
		node.WithSingleSigner(tpn.SingleSigner),
		node.WithTxSignPrivKey(tpn.SkTxSign),
		node.WithTxSignPubKey(tpn.PkTxSign),
		node.WithInterceptorsContainer(tpn.InterceptorsContainer),
		node.WithResolversFinder(tpn.ResolverFinder),
		node.WithBlockProcessor(tpn.BlockProcessor),
		node.WithDataStore(tpn.Storage),
		node.WithSyncer(&mock.SyncTimerMock{}),
	)
	if err != nil {
		fmt.Printf("Error creating node: %s\n", err.Error())
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		err = tpn.Node.ApplyOptions(
			node.WithMetaDataPool(tpn.MetaDataPool),
		)
	} else {
		err = tpn.Node.ApplyOptions(
			node.WithDataPool(tpn.ShardDataPool),
		)
	}

	if err != nil {
		fmt.Printf("Error creating node: %s\n", err.Error())
	}
}

// SendTransaction can send a transaction (it does the dispatching)
func (tpn *TestProcessorNode) SendTransaction(tx *dataTransaction.Transaction) (string, error) {
	txHash, err := tpn.Node.SendTransaction(
		tx.Nonce,
		hex.EncodeToString(tx.SndAddr),
		hex.EncodeToString(tx.RcvAddr),
		tx.Value,
		tx.GasPrice,
		tx.GasLimit,
		tx.Data,
		tx.Signature,
	)
	return txHash, err
}

func (tpn *TestProcessorNode) addHandlersForCounters() {
	hdrHandlers := func(key []byte) {
		atomic.AddInt32(&tpn.CounterHdrRecv, 1)
	}
	metaHandlers := func(key []byte) {
		atomic.AddInt32(&tpn.CounterMetaRcv, 1)
	}

	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		tpn.MetaDataPool.ShardHeaders().RegisterHandler(hdrHandlers)
		tpn.MetaDataPool.MetaChainBlocks().RegisterHandler(metaHandlers)
	} else {
		txHandler := func(key []byte) {
			atomic.AddInt32(&tpn.CounterTxRecv, 1)
		}
		mbHandlers := func(key []byte) {
			atomic.AddInt32(&tpn.CounterMbRecv, 1)
		}

		tpn.ShardDataPool.UnsignedTransactions().RegisterHandler(txHandler)
		tpn.ShardDataPool.Transactions().RegisterHandler(txHandler)
		tpn.ShardDataPool.Headers().RegisterHandler(hdrHandlers)
		tpn.ShardDataPool.MetaBlocks().RegisterHandler(metaHandlers)
		tpn.ShardDataPool.MiniBlocks().RegisterHandler(mbHandlers)
	}

}

// LoadTxSignSkBytes alters the already generated sk/pk pair
func (tpn *TestProcessorNode) LoadTxSignSkBytes(skBytes []byte) {
	newSk, _ := tpn.KeygenTxSign.PrivateKeyFromByteArray(skBytes)
	newPk := newSk.GeneratePublic()

	tpn.SkTxSign = newSk
	tpn.PkTxSign = newPk
	tpn.PkTxSignBytes, _ = newPk.ToByteArray()
	tpn.TxSignAddress, _ = TestAddressConverter.CreateAddressFromPublicKeyBytes(tpn.PkTxSignBytes)
}

// ProposeBlock proposes a new block
func (tpn *TestProcessorNode) ProposeBlock(round uint64) (data.BodyHandler, data.HeaderHandler, [][]byte) {
	haveTime := func() bool { return true }

	blockBody, err := tpn.BlockProcessor.CreateBlockBody(round, haveTime)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, nil
	}
	blockHeader, err := tpn.BlockProcessor.CreateBlockHeader(blockBody, round, haveTime)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, nil
	}

	blockHeader.SetRound(round)
	blockHeader.SetNonce(round)
	blockHeader.SetPubKeysBitmap(make([]byte, 0))
	sig, _ := TestMultiSig.AggregateSigs(nil)
	blockHeader.SetSignature(sig)
	currHdr := tpn.BlockChain.GetCurrentBlockHeader()
	if currHdr == nil {
		currHdr = tpn.BlockChain.GetGenesisHeader()
	}

	buff, _ := TestMarshalizer.Marshal(currHdr)
	blockHeader.SetPrevHash(TestHasher.Compute(string(buff)))
	blockHeader.SetPrevRandSeed(currHdr.GetRandSeed())
	blockHeader.SetRandSeed(sig)

	shardBlockBody, ok := blockBody.(dataBlock.Body)
	txHashes := make([][]byte, 0)
	if !ok {
		return blockBody, blockHeader, txHashes
	}

	for _, mb := range shardBlockBody {
		for _, hash := range mb.TxHashes {
			copiedHash := make([]byte, len(hash))
			copy(copiedHash, hash)
			txHashes = append(txHashes, copiedHash)
		}
	}

	return blockBody, blockHeader, txHashes
}

// BroadcastBlock broadcasts the block and body to the connected peers
func (tpn *TestProcessorNode) BroadcastBlock(body data.BodyHandler, header data.HeaderHandler) {
	_ = tpn.BroadcastMessenger.BroadcastBlock(body, header)
	_ = tpn.BroadcastMessenger.BroadcastHeader(header)
	miniBlocks, transactions, _ := tpn.BlockProcessor.MarshalizedDataToBroadcast(header, body)
	_ = tpn.BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
	_ = tpn.BroadcastMessenger.BroadcastTransactions(transactions)
}

// CommitBlock commits the block and body
func (tpn *TestProcessorNode) CommitBlock(body data.BodyHandler, header data.HeaderHandler) {
	_ = tpn.BlockProcessor.CommitBlock(tpn.BlockChain, header, body)
}

// GetShardHeader returns the first *dataBlock.Header stored in datapools having the nonce provided as parameter
func (tpn *TestProcessorNode) GetShardHeader(nonce uint64) (*dataBlock.Header, error) {
	invalidCachers := tpn.ShardDataPool == nil || tpn.ShardDataPool.Headers() == nil || tpn.ShardDataPool.HeadersNonces() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	syncMapHashNonce, ok := tpn.ShardDataPool.HeadersNonces().Get(nonce)
	if !ok {
		return nil, errors.New(fmt.Sprintf("no hash-nonce link in HeadersNonces for nonce %d", nonce))
	}

	headerHash, ok := syncMapHashNonce.Load(tpn.ShardCoordinator.SelfId())
	if !ok {
		return nil, errors.New(fmt.Sprintf("no hash-nonce hash in HeadersNonces for nonce %d", nonce))
	}

	headerObject, ok := tpn.ShardDataPool.Headers().Get(headerHash)
	if !ok {
		return nil, errors.New(fmt.Sprintf("no header found for hash %s", hex.EncodeToString(headerHash)))
	}

	header, ok := headerObject.(*dataBlock.Header)
	if !ok {
		return nil, errors.New(fmt.Sprintf("not a *dataBlock.Header stored in headers found for hash %s", hex.EncodeToString(headerHash)))
	}

	return header, nil
}

// GetBlockBody returns the body for provided header parameter
func (tpn *TestProcessorNode) GetBlockBody(header *dataBlock.Header) (dataBlock.Body, error) {
	invalidCachers := tpn.ShardDataPool == nil || tpn.ShardDataPool.MiniBlocks() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	body := dataBlock.Body{}
	for _, miniBlockHeader := range header.MiniBlockHeaders {
		miniBlockHash := miniBlockHeader.Hash

		mbObject, ok := tpn.ShardDataPool.MiniBlocks().Get(miniBlockHash)
		if !ok {
			return nil, errors.New(fmt.Sprintf("no miniblock found for hash %s", hex.EncodeToString(miniBlockHash)))
		}

		mb, ok := mbObject.(*dataBlock.MiniBlock)
		if !ok {
			return nil, errors.New(fmt.Sprintf("not a *dataBlock.MiniBlock stored in miniblocks found for hash %s", hex.EncodeToString(miniBlockHash)))
		}

		body = append(body, mb)
	}

	return body, nil
}

// GetMetaHeader returns the first *dataBlock.MetaBlock stored in datapools having the nonce provided as parameter
func (tpn *TestProcessorNode) GetMetaHeader(nonce uint64) (*dataBlock.MetaBlock, error) {
	invalidCachers := tpn.MetaDataPool == nil || tpn.MetaDataPool.MetaChainBlocks() == nil || tpn.MetaDataPool.HeadersNonces() == nil
	if invalidCachers {
		return nil, errors.New("invalid data pool")
	}

	syncMapHashNonce, ok := tpn.MetaDataPool.HeadersNonces().Get(nonce)
	if !ok {
		return nil, errors.New(fmt.Sprintf("no hash-nonce link in HeadersNonces for nonce %d", nonce))
	}

	headerHash, ok := syncMapHashNonce.Load(tpn.ShardCoordinator.SelfId())
	if !ok {
		return nil, errors.New(fmt.Sprintf("no hash-nonce hash in HeadersNonces for nonce %d", nonce))
	}

	headerObject, ok := tpn.MetaDataPool.MetaChainBlocks().Get(headerHash)
	if !ok {
		return nil, errors.New(fmt.Sprintf("no header found for hash %s", hex.EncodeToString(headerHash)))
	}

	header, ok := headerObject.(*dataBlock.MetaBlock)
	if !ok {
		return nil, errors.New(fmt.Sprintf("not a *dataBlock.MetaBlock stored in headers found for hash %s", hex.EncodeToString(headerHash)))
	}

	return header, nil
}

// SetAccountNonce sets the account nonce with journal
func (tpn *TestProcessorNode) SetAccountNonce(nonce uint64) error {
	nodePubKeyBytes, _ := tpn.SkTxSign.GeneratePublic().ToByteArray()
	nodeAddress := CreateAddressFromAddrBytes(nodePubKeyBytes)
	nodeAccount, _ := tpn.AccntState.GetAccountWithJournal(nodeAddress)
	err := nodeAccount.(*state.Account).SetNonceWithJournal(nonce)
	if err != nil {
		return err
	}

	_, err = tpn.AccntState.Commit()
	if err != nil {
		return err
	}

	return nil
}

// SyncNode tries to process and commit a block already stored in data pool with provided nonce
func (tpn *TestProcessorNode) SyncNode(nonce uint64) error {
	if tpn.ShardCoordinator.SelfId() == sharding.MetachainShardId {
		return tpn.syncMetaNode(nonce)
	} else {
		return tpn.syncShardNode(nonce)
	}
}

func (tpn *TestProcessorNode) syncShardNode(nonce uint64) error {
	header, err := tpn.GetShardHeader(nonce)
	if err != nil {
		return err
	}

	body, err := tpn.GetBlockBody(header)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.ProcessBlock(
		tpn.BlockChain,
		header,
		body,
		func() time.Duration {
			return time.Second * 2
		},
	)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.CommitBlock(tpn.BlockChain, header, body)
	if err != nil {
		return err
	}

	return nil
}

func (tpn *TestProcessorNode) syncMetaNode(nonce uint64) error {
	header, err := tpn.GetMetaHeader(nonce)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.ProcessBlock(
		tpn.BlockChain,
		header,
		&dataBlock.MetaBlockBody{},
		func() time.Duration {
			return time.Second * 2
		},
	)
	if err != nil {
		return err
	}

	err = tpn.BlockProcessor.CommitBlock(tpn.BlockChain, header, &dataBlock.MetaBlockBody{})
	if err != nil {
		return err
	}

	return nil
}
