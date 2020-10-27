package node

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

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

func TestWithGenesisTime(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	aTime := time.Time{}.Add(time.Duration(uint64(78)))

	opt := WithGenesisTime(aTime)
	err := opt(node)

	assert.Equal(t, node.genesisTime, aTime)
	assert.Nil(t, err)
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

func TestWithRequestedItemsHandler_NilRequestedItemsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithRequestedItemsHandler(nil)
	err := opt(node)

	assert.Equal(t, ErrNilRequestedItemsHandler, err)
}

func TestWithRequestedItemsHandler_OkRequestedItemsHandlerShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	requestedItemsHeanlder := &mock.TimeCacheStub{}
	opt := WithRequestedItemsHandler(requestedItemsHeanlder)
	err := opt(node)

	assert.True(t, node.requestedItemsHandler == requestedItemsHeanlder)
	assert.Nil(t, err)
}

func TestWithBootstrapRoundIndex(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()
	roundIndex := uint64(0)
	opt := WithBootstrapRoundIndex(roundIndex)

	err := opt(node)
	assert.Equal(t, roundIndex, node.bootstrapRoundIndex)
	assert.Nil(t, err)
}

func TestWithPeerDenialEvaluator_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithPeerDenialEvaluator(nil)
	err := opt(node)

	assert.True(t, errors.Is(err, ErrNilPeerDenialEvaluator))
}

func TestWithPeerDenialEvaluator_OkHandlerShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	blackListHandler := &mock.PeerDenialEvaluatorStub{}
	opt := WithPeerDenialEvaluator(blackListHandler)
	err := opt(node)

	assert.True(t, node.peerDenialEvaluator == blackListHandler)
	assert.Nil(t, err)
}

func TestWithNetworkShardingCollector_NilNetworkShardingCollectorShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithNetworkShardingCollector(nil)
	err := opt(node)

	assert.Equal(t, ErrNilNetworkShardingCollector, err)
}

func TestWithNetworkShardingCollector_OkHandlerShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	networkShardingCollector := &mock.NetworkShardingCollectorStub{}
	opt := WithNetworkShardingCollector(networkShardingCollector)
	err := opt(node)

	assert.True(t, node.networkShardingCollector == networkShardingCollector)
	assert.Nil(t, err)
}

func TestWithTxAccumulator_NilAccumulatorShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithTxAccumulator(nil)
	err := opt(node)

	assert.Equal(t, ErrNilTxAccumulator, err)
}

func TestWithHardforkTrigger_NilHardforkTriggerShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithHardforkTrigger(nil)
	err := opt(node)

	assert.Equal(t, ErrNilHardforkTrigger, err)
}

func TestWithHardforkTrigger_ShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	hardforkTrigger := &mock.HardforkTriggerStub{}
	opt := WithHardforkTrigger(hardforkTrigger)
	err := opt(node)

	assert.Nil(t, err)
	assert.True(t, node.hardforkTrigger == hardforkTrigger)
}

func TestWithSignatureSize(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()
	signatureSize := 48
	opt := WithSignatureSize(signatureSize)

	err := opt(node)
	assert.Equal(t, signatureSize, node.signatureSize)
	assert.Nil(t, err)
}

func TestWithPublicKeySize(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()
	publicKeySize := 96
	opt := WithPublicKeySize(publicKeySize)

	err := opt(node)
	assert.Equal(t, publicKeySize, node.publicKeySize)
	assert.Nil(t, err)
}

func TestWithNodeStopChannel_NilNodeStopChannelShouldErr(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	opt := WithNodeStopChannel(nil)
	err := opt(node)

	assert.Equal(t, ErrNilNodeStopChannel, err)
}

func TestWithNodeStopChannel_OkNodeStopChannelShouldWork(t *testing.T) {
	t.Parallel()

	node, _ := NewNode()

	ch := make(chan endProcess.ArgEndProcess, 1)
	opt := WithNodeStopChannel(ch)
	err := opt(node)

	assert.True(t, node.chanStopNodeProcess == ch)
	assert.Nil(t, err)
}

// TODO: add the missing options tests
