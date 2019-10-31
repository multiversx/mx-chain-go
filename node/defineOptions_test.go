package node

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestWithMessenger_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithMessenger(nil)
	err := opt(node)

	assert.Nil(t, node.messenger)
	assert.Equal(t, ErrNilMessenger, err)
}

func TestWithMessenger_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	messenger := &mock.MessengerStub{}

	opt := WithMessenger(messenger)
	err := opt(node)

	assert.True(t, node.messenger == messenger)
	assert.Nil(t, err)
}

func TestWithMarshalizer_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithMarshalizer(nil)
	err := opt(node)

	assert.Nil(t, node.marshalizer)
	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestWithMarshalizer_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	marshalizer := &mock.MarshalizerMock{}

	opt := WithMarshalizer(marshalizer)
	err := opt(node)

	assert.True(t, node.marshalizer == marshalizer)
	assert.Nil(t, err)
}

func TestWithHasher_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithHasher(nil)
	err := opt(node)

	assert.Nil(t, node.hasher)
	assert.Equal(t, ErrNilHasher, err)
}

func TestWithHasher_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	hasher := &mock.HasherMock{}

	opt := WithHasher(hasher)
	err := opt(node)

	assert.True(t, node.hasher == hasher)
	assert.Nil(t, err)
}

func TestWithAccountsAdapter_NilAccountsShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithAccountsAdapter(nil)
	err := opt(node)

	assert.Nil(t, node.accounts)
	assert.Equal(t, ErrNilAccountsAdapter, err)
}

func TestWithAccountsAdapter_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	accounts := &mock.AccountsStub{}

	opt := WithAccountsAdapter(accounts)
	err := opt(node)

	assert.True(t, node.accounts == accounts)
	assert.Nil(t, err)
}

func TestWithAddressConverter_NilConverterShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithAddressConverter(nil)
	err := opt(node)

	assert.Nil(t, node.addrConverter)
	assert.Equal(t, ErrNilAddressConverter, err)
}

func TestWithAddressConverter_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	converter := &mock.AddressConverterStub{}

	opt := WithAddressConverter(converter)
	err := opt(node)

	assert.True(t, node.addrConverter == converter)
	assert.Nil(t, err)
}

func TestWithBlockChain_NilBlockchainrShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithBlockChain(nil)
	err := opt(node)

	assert.Nil(t, node.blkc)
	assert.Equal(t, ErrNilBlockchain, err)
}

func TestWithBlockChain_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
	)

	opt := WithBlockChain(blkc)
	err := opt(node)

	assert.True(t, node.blkc == blkc)
	assert.Nil(t, err)
}

func TestWithDataStore_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithDataStore(nil)
	err := opt(node)

	assert.Nil(t, node.txSignPrivKey)
	assert.Equal(t, ErrNilStore, err)
}

func TestWithDataStore_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	store := &mock.ChainStorerMock{}

	opt := WithDataStore(store)
	err := opt(node)

	assert.True(t, node.store == store)
	assert.Nil(t, err)
}

func TestWithPrivateKey_NilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithTxSignPrivKey(nil)
	err := opt(node)

	assert.Nil(t, node.txSignPrivKey)
	assert.Equal(t, ErrNilPrivateKey, err)
}

func TestWithPrivateKey_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	sk := &mock.PrivateKeyStub{}

	opt := WithTxSignPrivKey(sk)
	err := opt(node)

	assert.True(t, node.txSignPrivKey == sk)
	assert.Nil(t, err)
}

func TestWithPrivateKey_NilBlsPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithPrivKey(nil)
	err := opt(node)

	assert.Nil(t, node.privKey)
	assert.Equal(t, ErrNilPrivateKey, err)
}

func TestWithBlsPrivateKey_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	sk := &mock.PrivateKeyStub{}

	opt := WithPrivKey(sk)
	err := opt(node)

	assert.True(t, node.privKey == sk)
	assert.Nil(t, err)
}

func TestWithSingleSignKeyGenerator_NilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithKeyGen(nil)
	err := opt(node)

	assert.Nil(t, node.keyGen)
	assert.Equal(t, ErrNilSingleSignKeyGen, err)
}

func TestWithSingleSignKeyGenerator_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	keyGen := &mock.KeyGenMock{}

	opt := WithKeyGen(keyGen)
	err := opt(node)

	assert.True(t, node.keyGen == keyGen)
	assert.Nil(t, err)
}

func TestWithInitialNodesPubKeys(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	pubKeys := make(map[uint32][]string, 1)
	pubKeys[0] = []string{"pk1", "pk2", "pk3"}

	opt := WithInitialNodesPubKeys(pubKeys)
	err := opt(node)

	assert.Equal(t, pubKeys, node.initialNodesPubkeys)
	assert.Nil(t, err)
}

func TestWithPublicKey(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	pubKeys := make(map[uint32][]string, 1)
	pubKeys[0] = []string{"pk1", "pk2", "pk3"}

	opt := WithInitialNodesPubKeys(pubKeys)
	err := opt(node)

	assert.Equal(t, pubKeys, node.initialNodesPubkeys)
	assert.Nil(t, err)
}

func TestWithPublicKey_NilPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithTxSignPubKey(nil)
	err := opt(node)

	assert.Nil(t, node.txSignPubKey)
	assert.Equal(t, ErrNilPublicKey, err)
}

func TestWithPublicKey_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	pk := &mock.PublicKeyMock{}

	opt := WithTxSignPubKey(pk)
	err := opt(node)

	assert.True(t, node.txSignPubKey == pk)
	assert.Nil(t, err)
}

func TestWithRoundDuration_ZeroDurationShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithRoundDuration(0)
	err := opt(node)

	assert.Equal(t, uint64(0), node.roundDuration)
	assert.Equal(t, ErrZeroRoundDurationNotSupported, err)
}

func TestWithRoundDuration_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	duration := uint64(5664)

	opt := WithRoundDuration(duration)
	err := opt(node)

	assert.True(t, node.roundDuration == duration)
	assert.Nil(t, err)
}

func TestWithConsensusGroupSize_NegativeGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithConsensusGroupSize(-1)
	err := opt(node)

	assert.Equal(t, 0, node.consensusGroupSize)
	assert.Equal(t, ErrNegativeOrZeroConsensusGroupSize, err)
}

func TestWithConsensusGroupSize_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	groupSize := 567

	opt := WithConsensusGroupSize(groupSize)
	err := opt(node)

	assert.True(t, node.consensusGroupSize == groupSize)
	assert.Nil(t, err)
}

func TestWithSyncer_NilSyncerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithSyncer(nil)
	err := opt(node)

	assert.Nil(t, node.syncTimer)
	assert.Equal(t, ErrNilSyncTimer, err)
}

func TestWithSyncer_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	sync := &mock.SyncStub{}

	opt := WithSyncer(sync)
	err := opt(node)

	assert.True(t, node.syncTimer == sync)
	assert.Nil(t, err)
}

func TestWithRounder_NilRounderShouldErr(t *testing.T) {
	t.Parallel()
	node, _ := NewNode()
	opt := WithRounder(nil)
	err := opt(node)
	assert.Nil(t, node.rounder)
	assert.Equal(t, ErrNilRounder, err)
}

func TestWithRounder_ShouldWork(t *testing.T) {
	t.Parallel()
	node, _ := NewNode()
	rnd := &mock.RounderMock{}
	opt := WithRounder(rnd)
	err := opt(node)
	assert.True(t, node.rounder == rnd)
	assert.Nil(t, err)
}

func TestWithBlockProcessor_NilProcessorShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithBlockProcessor(nil)
	err := opt(node)

	assert.Nil(t, node.syncTimer)
	assert.Equal(t, ErrNilBlockProcessor, err)
}

func TestWithBlockProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	bp := &mock.BlockProcessorStub{}

	opt := WithBlockProcessor(bp)
	err := opt(node)

	assert.True(t, node.blockProcessor == bp)
	assert.Nil(t, err)
}

func TestWithGenesisTime(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	aTime := time.Time{}.Add(time.Duration(uint64(78)))

	opt := WithGenesisTime(aTime)
	err := opt(node)

	assert.Equal(t, node.genesisTime, aTime)
	assert.Nil(t, err)
}

func TestWithDataPool_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithDataPool(nil)
	err := opt(node)

	assert.Nil(t, node.dataPool)
	assert.Equal(t, ErrNilDataPool, err)
}

func TestWithDataPool_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	dataPool := &mock.PoolsHolderStub{}

	opt := WithDataPool(dataPool)
	err := opt(node)

	assert.True(t, node.dataPool == dataPool)
	assert.Nil(t, err)
}

func TestWithMetaDataPool_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithMetaDataPool(nil)
	err := opt(node)

	assert.Nil(t, node.dataPool)
	assert.Equal(t, ErrNilDataPool, err)
}

func TestWithMetaDataPool_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	dataPool := &mock.MetaPoolsHolderStub{}

	opt := WithMetaDataPool(dataPool)
	err := opt(node)

	assert.True(t, node.metaDataPool == dataPool)
	assert.Nil(t, err)
}

func TestWithShardCoordinator_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithShardCoordinator(nil)
	err := opt(node)

	assert.Nil(t, node.shardCoordinator)
	assert.Equal(t, ErrNilShardCoordinator, err)
}

func TestWithShardCoordinator_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	shardCoordinator := mock.NewOneShardCoordinatorMock()

	opt := WithShardCoordinator(shardCoordinator)
	err := opt(node)

	assert.True(t, node.shardCoordinator == shardCoordinator)
	assert.Nil(t, err)
}

func TestWithNodesCoordinator_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithNodesCoordinator(nil)
	err := opt(node)

	assert.Nil(t, node.nodesCoordinator)
	assert.Equal(t, ErrNilNodesCoordinator, err)
}

func TestWithNodesCoordinator_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	nodesCoordinator := &mock.NodesCoordinatorMock{}

	opt := WithNodesCoordinator(nodesCoordinator)
	err := opt(node)

	assert.True(t, node.nodesCoordinator == nodesCoordinator)
	assert.Nil(t, err)
}

func TestWithUint64ByteSliceConverter_NilConverterShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithUint64ByteSliceConverter(nil)
	err := opt(node)

	assert.Nil(t, node.uint64ByteSliceConverter)
	assert.Equal(t, ErrNilUint64ByteSliceConverter, err)
}

func TestWithUint64ByteSliceConverter_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	converter := mock.NewNonceHashConverterMock()

	opt := WithUint64ByteSliceConverter(converter)
	err := opt(node)

	assert.True(t, node.uint64ByteSliceConverter == converter)
	assert.Nil(t, err)
}

func TestWithInitialNodesBalances_NilBalancesShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithInitialNodesBalances(nil)
	err := opt(node)

	assert.Nil(t, node.initialNodesBalances)
	assert.Equal(t, ErrNilBalances, err)
}

func TestWithInitialNodesBalances_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	balances := map[string]*big.Int{
		"pk1": big.NewInt(45),
		"pk2": big.NewInt(56),
	}

	opt := WithInitialNodesBalances(balances)
	err := opt(node)

	assert.Equal(t, node.initialNodesBalances, balances)
	assert.Nil(t, err)
}

func TestWithSinglesig_NilBlsSinglesigShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithSingleSigner(nil)
	err := opt(node)

	assert.Nil(t, node.singleSigner)
	assert.Equal(t, ErrNilSingleSig, err)
}

func TestWithSinglesig_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	singlesigner := &mock.SinglesignMock{}

	opt := WithSingleSigner(singlesigner)
	err := opt(node)

	assert.True(t, node.singleSigner == singlesigner)
	assert.Nil(t, err)
}

func TestWithMultisig_NilMultisigShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithMultiSigner(nil)
	err := opt(node)

	assert.Nil(t, node.multiSigner)
	assert.Equal(t, ErrNilMultiSig, err)
}

func TestWithMultisig_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	multisigner := &mock.MultisignMock{}

	opt := WithMultiSigner(multisigner)
	err := opt(node)

	assert.True(t, node.multiSigner == multisigner)
	assert.Nil(t, err)
}

func TestWithForkDetector_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	forkDetector := &mock.ForkDetectorMock{}
	opt := WithForkDetector(forkDetector)
	err := opt(node)

	assert.True(t, node.forkDetector == forkDetector)
	assert.Nil(t, err)
}

func TestWithForkDetector_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithForkDetector(nil)
	err := opt(node)

	assert.Nil(t, node.forkDetector)
	assert.Equal(t, ErrNilForkDetector, err)
}

func TestWithInterceptorsContainer_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	interceptorsContainer := &mock.InterceptorsContainerStub{}
	opt := WithInterceptorsContainer(interceptorsContainer)

	err := opt(node)

	assert.True(t, node.interceptorsContainer == interceptorsContainer)
	assert.Nil(t, err)
}

func TestWithInterceptorsContainer_NilContainerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithInterceptorsContainer(nil)
	err := opt(node)

	assert.Nil(t, node.interceptorsContainer)
	assert.Equal(t, ErrNilInterceptorsContainer, err)
}

func TestWithResolversFinder_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	resolversFinder := &mock.ResolversFinderStub{}
	opt := WithResolversFinder(resolversFinder)

	err := opt(node)

	assert.True(t, node.resolversFinder == resolversFinder)
	assert.Nil(t, err)
}

func TestWithResolversContainer_NilContainerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithResolversFinder(nil)
	err := opt(node)

	assert.Nil(t, node.resolversFinder)
	assert.Equal(t, ErrNilResolversFinder, err)
}

func TestWithConsensusBls_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	consensusType := "bls"
	opt := WithConsensusType(consensusType)
	err := opt(node)

	assert.Equal(t, consensusType, node.consensusType)
	assert.Nil(t, err)
}

func TestWithAppStatusHandler_NilAshShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithAppStatusHandler(nil)
	err := opt(node)

	assert.Equal(t, ErrNilStatusHandler, err)
}

func TestWithAppStatusHandler_OkAshShouldPass(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithAppStatusHandler(statusHandler.NewNilStatusHandler())
	err := opt(node)

	assert.IsType(t, &statusHandler.NilStatusHandler{}, node.appStatusHandler)
	assert.Nil(t, err)
}

func TestWithIndexer_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	indexer := &mock.IndexerMock{}
	opt := WithIndexer(indexer)
	err := opt(node)

	assert.Equal(t, indexer, node.indexer)
	assert.Nil(t, err)
}

func TestWithKeyGenForAccounts_NilKeygenShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithKeyGenForAccounts(nil)
	err := opt(node)

	assert.Equal(t, ErrNilKeyGenForBalances, err)
}

func TestWithKeyGenForAccounts_OkKeygenShouldPass(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	keyGen := &mock.KeyGenMock{}
	opt := WithKeyGenForAccounts(keyGen)
	err := opt(node)

	assert.True(t, node.keyGenForAccounts == keyGen)
	assert.Nil(t, err)
}

func TestWithTxFeeHandler_NilTxFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithTxFeeHandler(nil)
	err := opt(node)

	assert.Equal(t, ErrNilTxFeeHandler, err)
}

func TestWithTxFeeHandler_OkHandlerShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	txFeeHandler := &mock.FeeHandlerStub{}
	opt := WithTxFeeHandler(txFeeHandler)
	err := opt(node)

	assert.True(t, node.feeHandler == txFeeHandler)
	assert.Nil(t, err)
}
