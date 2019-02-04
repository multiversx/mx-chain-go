package transaction

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//------- NewTxResolver

func TestNewTxResolver_NilResolverShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		nil,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilResolver, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilTxPoolShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		&mock.ResolverStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilTxDataPool, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilTxStorageShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		&mock.ResolverStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilTxStorage, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		&mock.ResolverStub{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		nil,
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(i func(rd process.RequestData) ([]byte, error)) {
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

func TestTxResolver_ResolveTxRequestWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) ([]byte, error)) {
	}

	txRes, _ := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	buff, err := txRes.resolveTxRequest(process.RequestData{Type: process.NonceType, Value: []byte("aaa")})

	assert.Nil(t, buff)
	assert.Equal(t, process.ErrResolveNotHashType, err)
}

func TestTxResolver_ResolveTxRequestNilValueShouldRetNil(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) ([]byte, error)) {
	}

	txRes, _ := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	buff, err := txRes.resolveTxRequest(process.RequestData{Type: process.HashType, Value: nil})

	assert.Nil(t, buff)
	assert.Equal(t, process.ErrNilValue, err)
}

func TestTxResolver_ResolveTxRequestFoundInTxPoolShouldRetVal(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) ([]byte, error)) {
	}

	marshalizer := &mock.MarshalizerMock{}
	buffToExpect, err := marshalizer.Marshal("value")
	assert.Nil(t, err)

	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal([]byte("aaa"), key) {
			return "value", true
		}

		return nil, false
	}

	txRes, _ := NewTxResolver(
		res,
		txPool,
		&mock.StorerStub{},
		marshalizer,
	)

	buff, err := txRes.ResolveTxRequest(process.RequestData{Type: process.HashType, Value: []byte("aaa")})

	//we are 100% sure that the txPool resolved the request because storage is a stub
	//and any call to its method whould have panic-ed (&mock.StorerStub{} has uninitialized ...Called fields)
	assert.Nil(t, err)
	assert.Equal(t, buffToExpect, buff)
}

func TestTxResolver_ResolveTxRequestFoundInTxPoolMarshalizerFailShouldRetNilAndErr(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) ([]byte, error)) {
	}

	marshalizer := &mock.MarshalizerStub{}
	marshalizer.MarshalCalled = func(obj interface{}) (i []byte, e error) {
		return nil, errors.New("MarshalizerMock generic error")
	}

	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal([]byte("aaa"), key) {
			return "value", true
		}

		return nil, false
	}

	txRes, _ := NewTxResolver(
		res,
		txPool,
		&mock.StorerStub{},
		marshalizer,
	)

	buff, err := txRes.ResolveTxRequest(process.RequestData{Type: process.HashType, Value: []byte("aaa")})
	//Same as above test, we are sure that the marshalizer from txPool request failed as the code would have panic-ed
	//otherwise
	assert.Nil(t, buff)
	assert.Equal(t, "MarshalizerMock generic error", err.Error())
}

func TestTxResolver_ResolveTxRequestFoundInTxStorageShouldRetValAndError(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) ([]byte, error)) {
	}

	marshalizer := &mock.MarshalizerMock{}

	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		//not found in txPool
		return nil, false
	}

	expectedBuff := []byte("bbb")

	txStorage := &mock.StorerStub{}
	txStorage.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal([]byte("aaa"), key) {
			return expectedBuff, nil
		}

		return nil, nil
	}

	txRes, _ := NewTxResolver(
		res,
		txPool,
		txStorage,
		marshalizer,
	)

	buff, _ := txRes.ResolveTxRequest(process.RequestData{Type: process.HashType, Value: []byte("aaa")})

	assert.Equal(t, expectedBuff, buff)

}

func TestTxResolver_ResolveTxRequestFoundInTxStorageCheckRetError(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) ([]byte, error)) {
	}

	marshalizer := &mock.MarshalizerMock{}

	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		//not found in txPool
		return nil, false
	}

	expectedBuff := []byte("bbb")

	txStorage := &mock.StorerStub{}
	txStorage.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal([]byte("aaa"), key) {
			return expectedBuff, errors.New("just checking output error")
		}

		return nil, nil
	}

	txRes, _ := NewTxResolver(
		res,
		txPool,
		txStorage,
		marshalizer,
	)

	_, err := txRes.ResolveTxRequest(process.RequestData{Type: process.HashType, Value: []byte("aaa")})
	assert.Equal(t, "just checking output error", err.Error())

}

//------- RequestTransactionFromHash

func TestTxResolver_RequestTransactionFromHashShouldWork(t *testing.T) {
	t.Parallel()

	res := &mock.ResolverStub{}
	res.SetResolverHandlerCalled = func(h func(rd process.RequestData) ([]byte, error)) {
	}

	requested := process.RequestData{}

	res.RequestDataCalled = func(rd process.RequestData) error {
		requested = rd
		return nil
	}

	buffRequested := []byte("aaaa")

	txRes, _ := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, txRes.RequestTransactionFromHash(buffRequested))
	assert.Equal(t, process.RequestData{
		Type:  process.HashType,
		Value: buffRequested,
	}, requested)

}
