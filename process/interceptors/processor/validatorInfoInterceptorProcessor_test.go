package processor_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockValidatorInfo() state.ValidatorInfo {
	return state.ValidatorInfo{
		PublicKey: []byte("provided pk"),
		ShardId:   123,
		List:      string(common.EligibleList),
		Index:     10,
		Rating:    10,
	}
}

func createMockInterceptedValidatorInfo() process.InterceptedData {
	args := peer.ArgInterceptedValidatorInfo{
		Marshalizer: testscommon.MarshalizerMock{},
		Hasher:      &hashingMocks.HasherMock{},
	}
	args.DataBuff, _ = args.Marshalizer.Marshal(createMockValidatorInfo())
	ivi, _ := peer.NewInterceptedValidatorInfo(args)

	return ivi
}

func createMockArgValidatorInfoInterceptorProcessor() processor.ArgValidatorInfoInterceptorProcessor {
	return processor.ArgValidatorInfoInterceptorProcessor{
		ValidatorInfoPool: testscommon.NewShardedDataStub(),
	}
}

func TestNewValidatorInfoInterceptorProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil cache should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgValidatorInfoInterceptorProcessor()
		args.ValidatorInfoPool = nil

		proc, err := processor.NewValidatorInfoInterceptorProcessor(args)
		assert.Equal(t, process.ErrNilValidatorInfoPool, err)
		assert.True(t, check.IfNil(proc))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		proc, err := processor.NewValidatorInfoInterceptorProcessor(createMockArgValidatorInfoInterceptorProcessor())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(proc))
	})
}

func TestValidatorInfoInterceptorProcessor_Save(t *testing.T) {
	t.Parallel()

	t.Run("invalid data should error", func(t *testing.T) {
		t.Parallel()

		proc, err := processor.NewValidatorInfoInterceptorProcessor(createMockArgValidatorInfoInterceptorProcessor())
		assert.Nil(t, err)
		assert.Equal(t, process.ErrWrongTypeAssertion, proc.Save(nil, "", ""))
	})
	t.Run("invalid validator info should error", func(t *testing.T) {
		t.Parallel()

		providedData := mock.NewInterceptedMetaBlockMock(nil, []byte("hash")) // unable to cast to intercepted validator info
		wasCalled := false
		args := createMockArgValidatorInfoInterceptorProcessor()
		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
				wasCalled = true
			},
		}

		proc, _ := processor.NewValidatorInfoInterceptorProcessor(args)
		require.False(t, check.IfNil(proc))

		assert.Equal(t, process.ErrWrongTypeAssertion, proc.Save(providedData, "", ""))
		assert.False(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedEpoch := uint32(15)
		providedEpochStr := fmt.Sprintf("%d", providedEpoch)
		providedData := createMockInterceptedValidatorInfo()
		wasHasOrAddCalled := false
		args := createMockArgValidatorInfoInterceptorProcessor()
		providedBuff, _ := testscommon.MarshalizerMock{}.Marshal(createMockValidatorInfo())
		hasher := hashingMocks.HasherMock{}
		providedHash := hasher.Compute(string(providedBuff))

		args.ValidatorInfoPool = &testscommon.ShardedDataStub{
			AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
				assert.Equal(t, providedHash, key)
				wasHasOrAddCalled = true
			},
		}

		proc, _ := processor.NewValidatorInfoInterceptorProcessor(args)
		require.False(t, check.IfNil(proc))

		assert.Nil(t, proc.Save(providedData, "", providedEpochStr))
		assert.True(t, wasHasOrAddCalled)
	})
}

func TestValidatorInfoInterceptorProcessor_Validate(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	args := createMockArgValidatorInfoInterceptorProcessor()
	proc, _ := processor.NewValidatorInfoInterceptorProcessor(args)
	require.False(t, check.IfNil(proc))

	assert.Nil(t, proc.Validate(createMockInterceptedValidatorInfo(), ""))
}

func TestValidatorInfoInterceptorProcessor_RegisterHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	proc, err := processor.NewValidatorInfoInterceptorProcessor(createMockArgValidatorInfoInterceptorProcessor())
	assert.Nil(t, err)
	assert.False(t, check.IfNil(proc))

	proc.RegisterHandler(nil)
}
