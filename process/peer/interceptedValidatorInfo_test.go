package peer

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgInterceptedValidatorInfo() ArgInterceptedValidatorInfo {
	args := ArgInterceptedValidatorInfo{
		Marshalizer: marshallerMock.MarshalizerMock{},
		Hasher:      &hashingMocks.HasherMock{},
	}
	args.DataBuff, _ = args.Marshalizer.Marshal(createMockShardValidatorInfo())

	return args
}

func TestNewInterceptedValidatorInfo(t *testing.T) {
	t.Parallel()

	t.Run("nil data buff should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedValidatorInfo()
		args.DataBuff = nil

		ivi, err := NewInterceptedValidatorInfo(args)
		assert.Equal(t, process.ErrNilBuffer, err)
		assert.True(t, check.IfNil(ivi))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedValidatorInfo()
		args.Marshalizer = nil

		ivi, err := NewInterceptedValidatorInfo(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(ivi))
	})
	t.Run("nil hasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedValidatorInfo()
		args.Hasher = nil

		ivi, err := NewInterceptedValidatorInfo(args)
		assert.Equal(t, process.ErrNilHasher, err)
		assert.True(t, check.IfNil(ivi))
	})
	t.Run("unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		args := createMockArgInterceptedValidatorInfo()
		args.Marshalizer = &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}

		ivi, err := NewInterceptedValidatorInfo(args)
		assert.Equal(t, expectedErr, err)
		assert.True(t, check.IfNil(ivi))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ivi, err := NewInterceptedValidatorInfo(createMockArgInterceptedValidatorInfo())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(ivi))
	})
}

func TestInterceptedValidatorInfo_CheckValidity(t *testing.T) {
	t.Parallel()

	t.Run("publicKeyProperty too short", testInterceptedValidatorInfoPropertyLen(publicKeyProperty, false))
	t.Run("publicKeyProperty too long", testInterceptedValidatorInfoPropertyLen(publicKeyProperty, true))

	t.Run("listProperty too short", testInterceptedValidatorInfoPropertyLen(listProperty, false))
	t.Run("listProperty too long", testInterceptedValidatorInfoPropertyLen(listProperty, true))

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgInterceptedValidatorInfo()
		ivi, _ := NewInterceptedValidatorInfo(args)
		require.False(t, check.IfNil(ivi))
		assert.Nil(t, ivi.CheckValidity())
	})
}

func testInterceptedValidatorInfoPropertyLen(property string, tooLong bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		value := []byte("")
		expectedError := process.ErrPropertyTooShort
		if tooLong {
			value = make([]byte, 130)
			expectedError = process.ErrPropertyTooLong
		}

		args := createMockArgInterceptedValidatorInfo()
		ivi, _ := NewInterceptedValidatorInfo(args)
		require.False(t, check.IfNil(ivi))

		switch property {
		case publicKeyProperty:
			ivi.shardValidatorInfo.PublicKey = value
		case listProperty:
			ivi.shardValidatorInfo.List = string(value)
		default:
			assert.True(t, false)
		}

		err := ivi.CheckValidity()
		assert.True(t, strings.Contains(err.Error(), expectedError.Error()))
	}
}

func TestInterceptedValidatorInfo_Getters(t *testing.T) {
	t.Parallel()

	args := createMockArgInterceptedValidatorInfo()
	ivi, _ := NewInterceptedValidatorInfo(args)
	require.False(t, check.IfNil(ivi))

	validatorInfo := createMockShardValidatorInfo()
	validatorInfoBuff, _ := args.Marshalizer.Marshal(validatorInfo)
	hash := args.Hasher.Compute(string(validatorInfoBuff))

	assert.True(t, ivi.IsForCurrentShard())
	assert.Equal(t, validatorInfo, ivi.ValidatorInfo())
	assert.Equal(t, hash, ivi.Hash())
	assert.Equal(t, interceptedValidatorInfoType, ivi.Type())

	identifiers := ivi.Identifiers()
	assert.Equal(t, 1, len(identifiers))
	assert.Equal(t, hash, identifiers[0])

	str := ivi.String()
	assert.True(t, strings.Contains(str, fmt.Sprintf("pk=%s", logger.DisplayByteSlice(ivi.shardValidatorInfo.PublicKey))))
	assert.True(t, strings.Contains(str, fmt.Sprintf("shard=%d", validatorInfo.ShardId)))
	assert.True(t, strings.Contains(str, fmt.Sprintf("list=%s", validatorInfo.List)))
	assert.True(t, strings.Contains(str, fmt.Sprintf("index=%d", validatorInfo.Index)))
	assert.True(t, strings.Contains(str, fmt.Sprintf("tempRating=%d", validatorInfo.TempRating)))
}
