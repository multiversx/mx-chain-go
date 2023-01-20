package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockValidatorInfoBuff() []byte {
	vi := &state.ValidatorInfo{
		PublicKey: []byte("provided pk"),
		ShardId:   123,
		List:      string(common.EligibleList),
		Index:     10,
		Rating:    10,
	}

	marshalizerMock := marshallerMock.MarshalizerMock{}
	buff, _ := marshalizerMock.Marshal(vi)

	return buff
}

func TestNewInterceptedValidatorInfoDataFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil core components should error", func(t *testing.T) {
		t.Parallel()

		_, cryptoComponents := createMockComponentHolders()
		args := createMockArgument(nil, cryptoComponents)

		ividf, err := NewInterceptedValidatorInfoDataFactory(*args)
		assert.Equal(t, process.ErrNilCoreComponentsHolder, err)
		assert.True(t, check.IfNil(ividf))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, cryptoComponents := createMockComponentHolders()
		coreComponents.IntMarsh = nil
		args := createMockArgument(coreComponents, cryptoComponents)

		ividf, err := NewInterceptedValidatorInfoDataFactory(*args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(ividf))
	})
	t.Run("nil hasher should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, cryptoComponents := createMockComponentHolders()
		coreComponents.Hash = nil
		args := createMockArgument(coreComponents, cryptoComponents)

		ividf, err := NewInterceptedValidatorInfoDataFactory(*args)
		assert.Equal(t, process.ErrNilHasher, err)
		assert.True(t, check.IfNil(ividf))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ividf, err := NewInterceptedValidatorInfoDataFactory(*createMockArgument(createMockComponentHolders()))
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ividf))
	})
}

func TestInterceptedValidatorInfoDataFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("nil buff should error", func(t *testing.T) {
		t.Parallel()
		ividf, _ := NewInterceptedValidatorInfoDataFactory(*createMockArgument(createMockComponentHolders()))
		require.False(t, check.IfNil(ividf))

		ivi, err := ividf.Create(nil)
		assert.NotNil(t, err)
		assert.True(t, check.IfNil(ivi))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		ividf, _ := NewInterceptedValidatorInfoDataFactory(*createMockArgument(createMockComponentHolders()))
		require.False(t, check.IfNil(ividf))

		ivi, err := ividf.Create(createMockValidatorInfoBuff())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ivi))
	})
}
