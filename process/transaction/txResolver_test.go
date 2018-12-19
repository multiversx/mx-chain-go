package transaction

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewTxResolver

func TestNewTxResolver_NilResolverShouldErr(t *testing.T) {
	t.Parallel()

	_, err := NewTxResolver(
		nil,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilResolver, err)
}

func TestNewTxResolver_NilTxPoolShouldErr(t *testing.T) {
	t.Parallel()

	_, err := NewTxResolver(
		&mock.ResolverStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilTxDataPool, err)
}

func TestNewTxResolver_NilTxStorageShouldErr(t *testing.T) {
	t.Parallel()

	_, err := NewTxResolver(
		&mock.ResolverStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilTxStorage, err)
}

func TestNewTxResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	_, err := NewTxResolver(
		&mock.ResolverStub{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		nil,
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewTxResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(i func(rd process.RequestData) []byte) {
		wasCalled = true
	}

	txRes, err := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, txRes)
	assert.True(t, wasCalled)
}

//------- resolveTxRequest

func TestTxResolver_ResolveTxRequestWrongTypeShouldRetNil(t *testing.T) {
	t.Parallel()

	var handler func(rd process.RequestData) []byte

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) []byte) {
		handler = h
	}

	_, err := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)
	assert.Nil(t, handler(process.RequestData{Type: process.NonceType, Value: []byte("aaa")}))
}

func TestTxResolver_ResolveTxRequestNilValueShouldRetNil(t *testing.T) {
	t.Parallel()

	var handler func(rd process.RequestData) []byte

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) []byte) {
		handler = h
	}

	txRes, err := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, txRes)
	assert.Nil(t, handler(process.RequestData{Type: process.HashType, Value: nil}))
}

func TestTxResolver_ResolveTxRequestFoundInTxPoolShouldRetVal(t *testing.T) {
	t.Parallel()

	var handler func(rd process.RequestData) []byte

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) []byte) {
		handler = h
	}

	marshalizer := &mock.MarshalizerMock{}
	buffToExpect, err := marshalizer.Marshal("value")
	assert.Nil(t, err)

	txPool := &mock.ShardedDataStub{}
	txPool.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
		if bytes.Equal([]byte("aaa"), key) {
			return map[uint32]interface{}{
				0: "value",
			}
		}

		return nil
	}

	txRes, err := NewTxResolver(
		res,
		txPool,
		&mock.StorerStub{},
		marshalizer,
	)
	assert.Nil(t, err)
	assert.NotNil(t, txRes)
	assert.Equal(t, buffToExpect, handler(process.RequestData{Type: process.HashType, Value: []byte("aaa")}))
}

func TestTxResolver_ResolveTxRequestFoundInTxPoolMarshalizerFailShouldRetNil(t *testing.T) {
	t.Parallel()

	var handler func(rd process.RequestData) []byte

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) []byte) {
		handler = h
	}

	marshalizer := &mock.MarshalizerMock{}
	marshalizer.Fail = true

	txPool := &mock.ShardedDataStub{}
	txPool.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
		if bytes.Equal([]byte("aaa"), key) {
			return map[uint32]interface{}{
				0: "value",
			}
		}

		return nil
	}

	txRes, err := NewTxResolver(
		res,
		txPool,
		&mock.StorerStub{},
		marshalizer,
	)
	assert.Nil(t, err)
	assert.NotNil(t, txRes)
	assert.Nil(t, handler(process.RequestData{Type: process.HashType, Value: []byte("aaa")}))
}

func TestTxResolver_ResolveTxRequestFoundInTxStorageShouldRetVal(t *testing.T) {
	t.Parallel()

	var handler func(rd process.RequestData) []byte

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) []byte) {
		handler = h
	}

	marshalizer := &mock.MarshalizerMock{}

	txPool := &mock.ShardedDataStub{}
	txPool.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
		//not found in txPool
		return nil
	}

	txStorage := &mock.StorerStub{}
	txStorage.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal([]byte("aaa"), key) {
			return []byte("bbb"), nil
		}

		return nil, nil
	}

	txRes, err := NewTxResolver(
		res,
		txPool,
		txStorage,
		marshalizer,
	)
	assert.Nil(t, err)
	assert.NotNil(t, txRes)
	assert.Equal(t, []byte("bbb"), handler(process.RequestData{Type: process.HashType, Value: []byte("aaa")}))
}

func TestTxResolver_ResolveTxRequestNotFoundShouldRetNil(t *testing.T) {
	t.Parallel()

	var handler func(rd process.RequestData) []byte

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) []byte) {
		handler = h
	}

	marshalizer := &mock.MarshalizerMock{}

	txPool := &mock.ShardedDataStub{}
	txPool.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
		//not found in txPool
		return nil
	}

	txStorage := &mock.StorerStub{}
	txStorage.GetCalled = func(key []byte) (i []byte, e error) {
		//not found in storage
		return nil, nil
	}

	txRes, err := NewTxResolver(
		res,
		txPool,
		txStorage,
		marshalizer,
	)
	assert.Nil(t, err)
	assert.NotNil(t, txRes)
	assert.Nil(t, handler(process.RequestData{Type: process.HashType, Value: []byte("aaa")}))
}

//------- RequestTransactionFromHash

func TestTxResolver_RequestTransactionFromHashShouldWork(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) []byte) {
	}

	requested := process.RequestData{}

	res.RequestDataCalled = func(rd process.RequestData) error {
		requested = rd
		return nil
	}

	marshalizer := &mock.MarshalizerMock{}

	txPool := &mock.ShardedDataStub{}
	txPool.SearchDataCalled = func(key []byte) (shardValuesPairs map[uint32]interface{}) {
		//not found in txPool
		return nil
	}

	txStorage := &mock.StorerStub{}
	txStorage.GetCalled = func(key []byte) (i []byte, e error) {
		//not found in storage
		return nil, nil
	}

	txRes, err := NewTxResolver(
		res,
		txPool,
		txStorage,
		marshalizer,
	)
	assert.Nil(t, err)
	assert.NotNil(t, txRes)

	assert.Nil(t, txRes.RequestTransactionFromHash([]byte("aaaa")))
	assert.Equal(t, process.HashType, requested.Type)
	assert.Equal(t, []byte("aaaa"), requested.Value)
}
