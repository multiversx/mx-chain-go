package parsers

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state/dataTrieValue"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/assert"
)

func TestNewDataTrieLeafParser(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		tlp, err := NewDataTrieLeafParser([]byte("address"), nil, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
		assert.True(t, check.IfNil(tlp))
		assert.Equal(t, errors.ErrNilMarshalizer, err)
	})

	t.Run("nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		tlp, err := NewDataTrieLeafParser([]byte("address"), &marshallerMock.MarshalizerMock{}, nil)
		assert.True(t, check.IfNil(tlp))
		assert.Equal(t, errors.ErrNilEnableEpochsHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tlp, err := NewDataTrieLeafParser([]byte("address"), &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
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
		tlp, _ := NewDataTrieLeafParser(address, &marshallerMock.MarshalizerMock{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

		keyVal, err := tlp.ParseLeaf(key, append(val, suffix...), core.NotSpecified)
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
			IsFlagEnabledInCurrentEpochCalled: func(flag core.EnableEpochFlag) bool {
				return flag == core.AutoBalanceDataTriesFlag
			},
		}
		tlp, _ := NewDataTrieLeafParser(address, &marshallerMock.MarshalizerMock{}, enableEpochsHandler)

		keyVal, err := tlp.ParseLeaf(key, append(val, suffix...), core.NotSpecified)
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
			IsFlagEnabledInCurrentEpochCalled: func(flag core.EnableEpochFlag) bool {
				return flag == core.AutoBalanceDataTriesFlag
			},
		}
		tlp, _ := NewDataTrieLeafParser(address, marshaller, enableEpochsHandler)

		keyVal, err := tlp.ParseLeaf(hasher.Compute(string(key)), serializedLeafData, core.AutoBalanceEnabled)
		assert.Nil(t, err)
		assert.Equal(t, key, keyVal.Key())
		assert.Equal(t, val, keyVal.Value())
	})

	t.Run("unmarshall bytes with appended data should not return empty data", func(t *testing.T) {
		t.Parallel()

		marshaller := &marshal.GogoProtoMarshalizer{}

		keyBytes := []byte("eth")
		valBytes := []byte("0xA2AA67319062488CAFfc7E52802a3308cAF78a54")
		addrBytes, err := hex.DecodeString("b080fe7e47edd5f32b619a7a439a0174ebda49ac27a5b112dd685470ae008001")
		assert.Nil(t, err)

		valWithAppendedData := append(valBytes, keyBytes...)
		valWithAppendedData = append(valWithAppendedData, addrBytes...)

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInCurrentEpochCalled: func(flag core.EnableEpochFlag) bool {
				return flag == core.AutoBalanceDataTriesFlag
			},
		}
		tlp, _ := NewDataTrieLeafParser(addrBytes, marshaller, enableEpochsHandler)

		keyVal, err := tlp.ParseLeaf(keyBytes, valWithAppendedData, core.NotSpecified)
		assert.Nil(t, err)
		assert.Equal(t, keyBytes, keyVal.Key())
		assert.Equal(t, valBytes, keyVal.Value())
	})
}
