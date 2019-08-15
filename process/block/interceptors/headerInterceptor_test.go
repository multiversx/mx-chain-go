package interceptors_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

var durTimeout = time.Second

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
	headerValidator := &mock.HeaderValidatorStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		nil,
		headers,
		headersNonces,
		headerValidator,
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
	headerValidator := &mock.HeaderValidatorStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		nil,
		headersNonces,
		headerValidator,
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
	headerValidator := &mock.HeaderValidatorStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		nil,
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHeaderHandlerValidatorShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		nil,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilHeaderHandlerValidator, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		headerValidator,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		headerValidator,
		mock.NewMultiSigner(),
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		nil,
		&mock.ChronologyValidatorStub{},
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_NilChronologyValidatorShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		nil,
	)

	assert.Equal(t, process.ErrNilChronologyValidator, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	hi, err := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, hi)
}

//------- ParseReceivedMessage

func TestHeaderInterceptor_ParseReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}

	headerValidator := &mock.HeaderValidatorStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	hdr, err := hi.ParseReceivedMessage(nil)

	assert.Nil(t, hdr)
	assert.Equal(t, process.ErrNilMessage, err)
}

func TestHeaderInterceptor_ParseReceivedMessageNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	msg := &mock.P2PMessageMock{}
	hdr, err := hi.ParseReceivedMessage(msg)

	assert.Nil(t, hdr)
	assert.Equal(t, process.ErrNilDataToProcess, err)
}

func TestHeaderInterceptor_ParseReceivedMessageMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errMarshalizer
			},
		},
		headers,
		headersNonces,
		headerValidator,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ChronologyValidatorStub{},
	)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}
	hdr, err := hi.ParseReceivedMessage(msg)

	assert.Nil(t, hdr)
	assert.Equal(t, errMarshalizer, err)
}

func TestHeaderInterceptor_ParseReceivedMessageSanityCheckFailedShouldErr(t *testing.T) {
	t.Parallel()

	headerValidator := &mock.HeaderValidatorStub{}
	marshalizer := &mock.MarshalizerMock{}
	multisigner := mock.NewMultiSigner()
	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	chronologyValidator := &mock.ChronologyValidatorStub{
		ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
			return nil
		},
	}

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		headerValidator,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		chronologyValidator,
	)

	hdr := block.NewInterceptedHeader(multisigner, chronologyValidator)
	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}
	hdr, err := hi.ParseReceivedMessage(msg)

	assert.Nil(t, hdr)
	assert.Equal(t, process.ErrNilPubKeysBitmap, err)
}

func TestHeaderInterceptor_ParseReceivedMessageValsOkShouldWork(t *testing.T) {
	t.Parallel()

	testedNonce := uint64(67)

	headerValidator := &mock.HeaderValidatorStub{
		IsHeaderValidForProcessingCalled: func(headerHandler data.HeaderHandler) bool {
			return true
		},
	}

	marshalizer := &mock.MarshalizerMock{}
	multisigner := mock.NewMultiSigner()
	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	chronologyValidator := &mock.ChronologyValidatorStub{
		ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
			return nil
		},
	}

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		headerValidator,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		chronologyValidator,
	)

	hdr := block.NewInterceptedHeader(multisigner, chronologyValidator)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
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

	hdrIntercepted, err := hi.ParseReceivedMessage(msg)
	if hdrIntercepted != nil {
		//hdrIntercepted will have a "real" hash computed
		hdrIntercepted.SetHash(hdr.Hash())
	}

	assert.Equal(t, hdr, hdrIntercepted)
	assert.Nil(t, err)
}

//------- ProcessReceivedMessage

func TestHeaderInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	headers := &mock.CacherStub{}
	headersNonces := &mock.Uint64SyncMapCacherStub{}
	headerValidator := &mock.HeaderValidatorStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		&mock.MarshalizerMock{},
		headers,
		headersNonces,
		headerValidator,
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

	headerValidator := &mock.HeaderValidatorStub{
		IsHeaderValidForProcessingCalled: func(headerHandler data.HeaderHandler) bool {
			return true
		},
	}

	nodesCoordinator := mock.NewNodesCoordinatorMock()
	nodes := generateValidatorsMap(3, 3, 1)
	nodesCoordinator.SetNodesPerShards(nodes)

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		headerValidator,
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

func TestHeaderInterceptor_ProcessReceivedMessageTestHdrNonces(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	chanDone := make(chan struct{}, 1)
	testedNonce := uint64(67)
	headers := &mock.CacherStub{}
	multisigner := mock.NewMultiSigner()
	chronologyValidator := &mock.ChronologyValidatorStub{
		ValidateReceivedBlockCalled: func(shardID uint32, epoch uint32, nonce uint64, round uint64) error {
			return nil
		},
	}

	headerValidator := &mock.HeaderValidatorStub{
		IsHeaderValidForProcessingCalled: func(headerHandler data.HeaderHandler) bool {
			return true
		},
	}

	hdrsNonces := &mock.Uint64SyncMapCacherStub{}

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		hdrsNonces,
		headerValidator,
		multisigner,
		mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(2),
		chronologyValidator,
	)

	hdr := block.NewInterceptedHeader(multisigner, chronologyValidator)
	hdr.Nonce = testedNonce
	hdr.ShardId = 0
	hdr.PrevHash = make([]byte, 0)
	hdr.PubKeysBitmap = make([]byte, 0)
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
		return false, false
	}

	hdrsNonces.MergeCalled = func(nonce uint64, src dataRetriever.ShardIdHashMap) {
		if testedNonce == nonce {
			chanDone <- struct{}{}
		}
	}

	assert.Nil(t, hi.ProcessReceivedMessage(msg))
	select {
	case <-chanDone:
	case <-time.After(durTimeout):
		assert.Fail(t, "timeout while waiting for block to be inserted in the pool")
	}
}

func TestHeaderInterceptor_ProcessReceivedMessageIsNotValidShouldNotAdd(t *testing.T) {
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

	headerValidator := &mock.HeaderValidatorStub{
		IsHeaderValidForProcessingCalled: func(headerHandler data.HeaderHandler) bool {
			return false
		},
	}

	nodesCoordinator := mock.NewNodesCoordinatorMock()
	nodes := generateValidatorsMap(3, 3, 1)
	nodesCoordinator.SetNodesPerShards(nodes)

	hi, _ := interceptors.NewHeaderInterceptor(
		marshalizer,
		headers,
		headersNonces,
		headerValidator,
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

	headerValidator := &mock.HeaderValidatorStub{
		IsHeaderValidForProcessingCalled: func(headerHandler data.HeaderHandler) bool {
			return true
		},
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
		headerValidator,
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
