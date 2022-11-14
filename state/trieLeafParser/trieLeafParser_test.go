package trieLeafParser

import (
	"github.com/ElrondNetwork/elrond-go/state/dataTrieValue"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/testscommon/enableEpochsHandlerMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieLeafParser(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		tlp, err := NewTrieLeafParser([]byte("address"), nil, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		assert.True(t, check.IfNil(tlp))
		assert.Equal(t, errors.ErrNilMarshalizer, err)
	})

	t.Run("nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		tlp, err := NewTrieLeafParser([]byte("address"), &marshallerMock.MarshalizerMock{}, nil)
		assert.True(t, check.IfNil(tlp))
		assert.Equal(t, errors.ErrNilEnableEpochsHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tlp, err := NewTrieLeafParser([]byte("address"), &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(tlp))
	})
}

func TestTrieLeafParser_ParseLeaf(t *testing.T) {
	t.Parallel()

	t.Run("auto balance disabled", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		val := []byte("val")
		address := []byte("address")
		suffix := append(key, address...)
		tlp, _ := NewTrieLeafParser(address, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

		keyVal, err := tlp.ParseLeaf(key, append(val, suffix...))
		assert.Nil(t, err)
		assert.Equal(t, key, keyVal.Key())
		assert.Equal(t, val, keyVal.Value())
	})

	t.Run("auto balance enabled - val with appended data", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		val := []byte("val")
		address := []byte("address")
		suffix := append(key, address...)
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tlp, _ := NewTrieLeafParser(address, &marshallerMock.MarshalizerMock{}, enableEpochsHandler)

		keyVal, err := tlp.ParseLeaf(key, append(val, suffix...))
		assert.Nil(t, err)
		assert.Equal(t, key, keyVal.Key())
		assert.Equal(t, val, keyVal.Value())
	})

	t.Run("auto balance enabled - val as struct", func(t *testing.T) {
		t.Parallel()

		marshaller := &marshallerMock.MarshalizerMock{}
		hasher := &hashingMocks.HasherMock{}
		key := []byte("key")
		val := []byte("val")
		address := []byte("address")
		leafData := dataTrieValue.TrieLeafData{
			Value:   val,
			Key:     key,
			Address: address,
		}
		serializedLeafData, _ := marshaller.Marshal(leafData)
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tlp, _ := NewTrieLeafParser(address, marshaller, enableEpochsHandler)

		keyVal, err := tlp.ParseLeaf(hasher.Compute(string(key)), serializedLeafData)
		assert.Nil(t, err)
		assert.Equal(t, key, keyVal.Key())
		assert.Equal(t, val, keyVal.Value())
	})
}
