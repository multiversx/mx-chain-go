package interceptors_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

var durTimeout = time.Duration(time.Second)

func generateValidatorsMap(shardSize, metachainSize, nbShards uint32) map[uint32][]sharding.Validator {
	nodes := make(map[uint32][]sharding.Validator)

	for shard := uint32(0); shard < nbShards; shard++ {
		shardNodes := make([]sharding.Validator, 0)
		for valIdx := uint32(0); valIdx < shardSize; valIdx++ {
			pk := fmt.Sprintf("pubKey_sh%d_node%d", shard, valIdx)
			addr := fmt.Sprintf("address_sh%d_node%d", shard, valIdx)
			v, _ := sharding.NewValidator(big.NewInt(0), 1, []byte(pk), []byte(addr))
			shardNodes = append(shardNodes, v)
		}
		nodes[shard] = shardNodes
	}

	metaNodes := make([]sharding.Validator, 0)
	for mValIdx := uint32(0); mValIdx < metachainSize; mValIdx++ {
		pk := fmt.Sprintf("pubKey_meta_node%d", mValIdx)
		addr := fmt.Sprintf("address_meta_node%d", mValIdx)
		v, _ := sharding.NewValidator(big.NewInt(0), 1, []byte(pk), []byte(addr))
		metaNodes = append(metaNodes, v)
	}
	nodes[sharding.MetachainShardId] = metaNodes

	return nodes
}

//------- NewHeaderInterceptor

func TestNewHeaderInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		nil,
		headers,
		headersNonces,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	headersNonces := &mock.Uint64SyncMapCacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		nil,
		headersNonces,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHeadersNoncesShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		nil,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	storer := &mock.StorerStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, hi)
}

//------- ProcessReceivedMessage

func TestHeaderInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	storer := &mock.StorerStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilMessage, hi.ProcessReceivedMessage(nil))
}

func TestHeaderInterceptor_ProcessReceivedMessageValsOkShouldWork(t *testing.T) {
	t.Parallel()

	chanDone := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	testedNonce := uint64(67)
	marshalizer := &mock.MarshalizerMock{}
	headers := &mock.CacherStub{}

	hasher := mock.HasherMock{}
	multisigner := mock.NewMultiSigner()
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headersNonces.MergeCalled = func(nonce uint64, src dataRetriever.ShardIdHashMap) {
		if nonce == testedNonce {
			wg.Done()
		}
	}

	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return errors.New("key not found")
	}

	nodesCoordinator := mock.NewNodesCoordinatorMock()
	nodes := generateValidatorsMap(3, 3, 1)
	nodesCoordinator.SetNodesPerShards(nodes)

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		nodesCoordinator,
	)

	hdr := block.NewInterceptedHeader(multisigner, nodesCoordinator, marshalizer, hasher)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = []byte{1}
	hdr.BlockBodyType = dataBlock.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.RootHash = make([]byte, 0)
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.MiniBlockHeaders = make([]dataBlock.MiniBlockHeader, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	headers.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			wg.Done()
		}
		return false, false
	}

	go func() {
		wg.Wait()
		chanDone <- struct{}{}
	}()

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
	case <-time.After(durTimeout):
		assert.Fail(t, "timeout while waiting for block to be inserted in the pool")
	}
}

func TestHeaderInterceptor_ProcessReceivedMessageIsInStorageShouldNotAdd(t *testing.T) {
	t.Parallel()

	chanDone := make(chan struct{}, 10)
	testedNonce := uint64(67)
	marshalizer := &mock.MarshalizerMock{}
	headers := &mock.CacherStub{}

	hasher := mock.HasherMock{}
	multisigner := mock.NewMultiSigner()
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headersNonces.MergeCalled = func(nonce uint64, src dataRetriever.ShardIdHashMap) {
		if nonce == testedNonce {
			chanDone <- struct{}{}
		}
	}

	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return nil
	}

	nodesCoordinator := mock.NewNodesCoordinatorMock()
	nodes := generateValidatorsMap(3, 3, 1)
	nodesCoordinator.SetNodesPerShards(nodes)

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		nodesCoordinator,
	)

	hdr := block.NewInterceptedHeader(multisigner, nodesCoordinator, marshalizer, hasher)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = []byte{1}
	hdr.BlockBodyType = dataBlock.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.MiniBlockHeaders = make([]dataBlock.MiniBlockHeader, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	headers.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			chanDone <- struct{}{}
		}
		return false, false
	}

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
		assert.Fail(t, "should have not add block in pool")
	case <-time.After(durTimeout):
	}
}

func TestHeaderInterceptor_ProcessReceivedMessageNotForCurrentShardShouldNotAdd(t *testing.T) {
	t.Parallel()

	chanDone := make(chan struct{}, 10)
	testedNonce := uint64(67)
	marshalizer := &mock.MarshalizerMock{}
	headers := &mock.CacherStub{}

	hasher := mock.HasherMock{}
	multisigner := mock.NewMultiSigner()
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headersNonces.MergeCalled = func(nonce uint64, src dataRetriever.ShardIdHashMap) {
		if nonce == testedNonce {
			chanDone <- struct{}{}
		}
	}

	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return errors.New("key not found")
	}
	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 2
	shardCoordinator.SetNoShards(5)

	nodesCoordinator := &mock.NodesCoordinatorMock{
		NbShards:           5,
		ShardConsensusSize: 1,
		MetaConsensusSize:  1,
		ShardId:            2,
	}

	nodes := generateValidatorsMap(3, 3, 5)
	nodesCoordinator.SetNodesPerShards(nodes)

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		storer,
		multisigner,
		mock.HasherMock{},
		shardCoordinator,
		nodesCoordinator,
	)

	hdr := block.NewInterceptedHeader(multisigner, nodesCoordinator, marshalizer, hasher)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = []byte{1}
	hdr.BlockBodyType = dataBlock.TxBlock
	hdr.Signature = make([]byte, 0)
	hdr.RootHash = make([]byte, 0)
	hdr.SetHash([]byte("aaa"))
	hdr.PrevRandSeed = make([]byte, 0)
	hdr.RandSeed = make([]byte, 0)
	hdr.MiniBlockHeaders = make([]dataBlock.MiniBlockHeader, 0)

	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	headers.HasOrAddCalled = func(key []byte, value interface{}) (ok, evicted bool) {
		aaaHash := mock.HasherMock{}.Compute(string(buff))
		if bytes.Equal(aaaHash, key) {
			chanDone <- struct{}{}
		}
		return false, false
	}

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
}
