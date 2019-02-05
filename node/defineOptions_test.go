package node

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/mock"
	"github.com/stretchr/testify/assert"
)

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

func TestWithContext_NilContextShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithContext(nil)
	err := opt(node)

	assert.Equal(t, context.Background(), node.ctx)
	assert.Equal(t, ErrNilContext, err)
}

func TestWithContext_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	ctx, _ := context.WithCancel(context.Background())

	opt := WithContext(ctx)
	err := opt(node)

	assert.Equal(t, ctx, node.ctx)
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

	accounts := &mock.AccountsAdapterStub{}

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
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{})

	opt := WithBlockChain(blkc)
	err := opt(node)

	assert.True(t, node.blkc == blkc)
	assert.Nil(t, err)
}

func TestWithPrivateKey_NilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithPrivateKey(nil)
	err := opt(node)

	assert.Nil(t, node.privateKey)
	assert.Equal(t, ErrNilPrivateKey, err)
}

func TestWithPrivateKey_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	sk := &mock.PrivateKeyStub{}

	opt := WithPrivateKey(sk)
	err := opt(node)

	assert.True(t, node.privateKey == sk)
	assert.Nil(t, err)
}

func TestWithSingleSignKeyGenerator_NilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithSingleSignKeyGenerator(nil)
	err := opt(node)

	assert.Nil(t, node.singleSignKeyGen)
	assert.Equal(t, ErrNilSingleSignKeyGen, err)
}

func TestWithSingleSignKeyGenerator_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	keyGen := &mock.SingleSignKeyGenMock{}

	opt := WithSingleSignKeyGenerator(keyGen)
	err := opt(node)

	assert.True(t, node.singleSignKeyGen == keyGen)
	assert.Nil(t, err)
}

func TestWithInitialNodesPubKeys(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	pubKeys := []string{"pk1", "pk2", "pk3"}

	opt := WithInitialNodesPubKeys(pubKeys)
	err := opt(node)

	assert.Equal(t, pubKeys, node.initialNodesPubkeys)
	assert.Nil(t, err)
}

func TestWithPublicKey(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	pubKeys := []string{"pk1", "pk2", "pk3"}

	opt := WithInitialNodesPubKeys(pubKeys)
	err := opt(node)

	assert.Equal(t, pubKeys, node.initialNodesPubkeys)
	assert.Nil(t, err)
}

func TestWithPublicKey_NilPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithPublicKey(nil)
	err := opt(node)

	assert.Nil(t, node.publicKey)
	assert.Equal(t, ErrNilPublicKey, err)
}

func TestWithPublicKey_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	pk := &mock.SingleSignPublicKeyMock{}

	opt := WithPublicKey(pk)
	err := opt(node)

	assert.True(t, node.publicKey == pk)
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

	assert.Nil(t, node.syncer)
	assert.Equal(t, ErrNilSyncTimer, err)
}

func TestWithSyncer_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	sync := &mock.SyncStub{}

	opt := WithSyncer(sync)
	err := opt(node)

	assert.True(t, node.syncer == sync)
	assert.Nil(t, err)
}

func TestWithBlockProcessor_NilProcessorShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithBlockProcessor(nil)
	err := opt(node)

	assert.Nil(t, node.syncer)
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

func TestWithElasticSubrounds_TrueShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithElasticSubrounds(true)
	err := opt(node)

	assert.True(t, node.elasticSubrounds)
	assert.Nil(t, err)
}

func TestWithElasticSubrounds_FalseShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithElasticSubrounds(false)
	err := opt(node)

	assert.False(t, node.elasticSubrounds)
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

	dataPool := &mock.TransientDataPoolMock{}

	opt := WithDataPool(dataPool)
	err := opt(node)

	assert.True(t, node.dataPool == dataPool)
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

func TestWithMultisig_NilMultisigShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithMultisig(nil)
	err := opt(node)

	assert.Nil(t, node.multisig)
	assert.Equal(t, ErrNilMultiSig, err)
}

func TestWithMultisig_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	multisigner := &mock.MultisignMock{}

	opt := WithMultisig(multisigner)
	err := opt(node)

	assert.True(t, node.multisig == multisigner)
	assert.Nil(t, err)
}

func TestWithForkDetector_shouldWork(t *testing.T) {
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
