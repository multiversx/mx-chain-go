package interceptors_test

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

//------- NewHeaderInterceptorBase

func TestNewHeaderInterceptorBase_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	hi, err := interceptors.NewHeaderInterceptorBase(
		nil,
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptorBase_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	hi, err := interceptors.NewHeaderInterceptorBase(
		&mock.MarshalizerMock{},
		nil,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilHeadersStorage, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptorBase_NilMultiSignerShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	hi, err := interceptors.NewHeaderInterceptorBase(
		&mock.MarshalizerMock{},
		storer,
		nil,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Nil(t, hi)
	assert.Equal(t, process.ErrNilMultiSigVerifier, err)
}

func TestNewHeaderInterceptorBase_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	hi, err := interceptors.NewHeaderInterceptorBase(
		&mock.MarshalizerMock{},
		storer,
		mock.NewMultiSigner(),
		nil,
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptorBase_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	hi, err := interceptors.NewHeaderInterceptorBase(
		&mock.MarshalizerMock{},
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		nil,
		mock.NewNodesCoordinatorMock(),
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptorBase_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	hi, err := interceptors.NewHeaderInterceptorBase(
		&mock.MarshalizerMock{},
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		nil,
	)

	assert.Equal(t, process.ErrNilNodesCoordinator, err)
	assert.Nil(t, hi)
}

func TestNewHeaderInterceptorBase_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	hib, err := interceptors.NewHeaderInterceptorBase(
		&mock.MarshalizerMock{},
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, hib)
}

//------- ProcessReceivedMessage

func TestHeaderInterceptorBase_ParseReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	hib, _ := interceptors.NewHeaderInterceptorBase(
		&mock.MarshalizerMock{},
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	hdr, err := hib.ParseReceivedMessage(nil)

	assert.Nil(t, hdr)
	assert.Equal(t, process.ErrNilMessage, err)
}

func TestHeaderInterceptorBase_ParseReceivedMessageNilDataToProcessShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	hib, _ := interceptors.NewHeaderInterceptorBase(
		&mock.MarshalizerMock{},
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	msg := &mock.P2PMessageMock{}
	hdr, err := hib.ParseReceivedMessage(msg)

	assert.Nil(t, hdr)
	assert.Equal(t, process.ErrNilDataToProcess, err)
}

func TestHeaderInterceptorBase_ParseReceivedMessageMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	storer := &mock.StorerStub{}
	hib, _ := interceptors.NewHeaderInterceptorBase(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errMarshalizer
			},
		},
		storer,
		mock.NewMultiSigner(),
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		mock.NewNodesCoordinatorMock(),
	)

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}
	hdr, err := hib.ParseReceivedMessage(msg)

	assert.Nil(t, hdr)
	assert.Equal(t, errMarshalizer, err)
}

func TestHeaderInterceptorBase_ParseReceivedMessageSanityCheckFailedShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	multisigner := mock.NewMultiSigner()

	nodesCoordinator := mock.NewNodesCoordinatorMock()
	hib, _ := interceptors.NewHeaderInterceptorBase(
		marshalizer,
		storer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		nodesCoordinator,
	)

	hdr := block.NewInterceptedHeader(multisigner, nodesCoordinator, marshalizer, hasher)
	buff, _ := marshalizer.Marshal(hdr)
	msg := &mock.P2PMessageMock{
		DataField: buff,
	}
	hdr, err := hib.ParseReceivedMessage(msg)

	assert.Nil(t, hdr)
	assert.Equal(t, process.ErrNilPubKeysBitmap, err)
}

func createNodesCoordinator() sharding.NodesCoordinator {
	validators := make(map[uint32][]sharding.Validator, 0)

	//shard0
	shardValidators := make([]sharding.Validator, 0)
	for i := 0; i < 16; i++ {
		pubKeyStr := fmt.Sprintf("pk_shard0_%d", i)
		addrStr := fmt.Sprintf("addr_shard0_%d", i)
		v, _ := sharding.NewValidator(big.NewInt(0), 1, []byte(pubKeyStr), []byte(addrStr))
		shardValidators = append(shardValidators, v)
	}

	//metachain
	pubKeyBytes := []byte("pk_meta")
	addrBytes := []byte("addr_meta")
	v, _ := sharding.NewValidator(big.NewInt(0), 1, pubKeyBytes, addrBytes)

	validators[0] = shardValidators
	validators[sharding.MetachainShardId] = []sharding.Validator{v}

	nodesCoordinator := mock.NewNodesCoordinatorMock()
	nodesCoordinator.SetNodesPerShards(validators)

	return nodesCoordinator
}

func TestHeaderInterceptorBase_ParseReceivedMessageValsOkShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	testedNonce := uint64(67)
	multisigner := mock.NewMultiSigner()
	storer := &mock.StorerStub{}
	storer.HasCalled = func(key []byte) error {
		return errors.New("key not found")
	}

	nodesCoordinator := createNodesCoordinator()

	hib, _ := interceptors.NewHeaderInterceptorBase(
		marshalizer,
		storer,
		multisigner,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock(),
		nodesCoordinator,
	)

	hdr := block.NewInterceptedHeader(multisigner, nodesCoordinator, marshalizer, mock.HasherMock{})
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

	hdrIntercepted, err := hib.ParseReceivedMessage(msg)
	if hdrIntercepted != nil {
		//hdrIntercepted will have a "real" hash computed
		hdrIntercepted.SetHash(hdr.Hash())
	}

	assert.Equal(t, hdr, hdrIntercepted)
	assert.Nil(t, err)
}
