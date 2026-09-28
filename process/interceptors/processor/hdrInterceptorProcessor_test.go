package processor_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
)

func createMockHdrArgument() *processor.ArgHdrInterceptorProcessor {
	roundExclusions, _ := common.NewRoundExclusionHandler(nil)
	arg := &processor.ArgHdrInterceptorProcessor{
		Headers:             &mock.HeadersCacherStub{},
		Proofs:              &dataRetriever.ProofsPoolMock{},
		BlockBlackList:      &testscommon.TimeCacheStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		RoundExclusions:     roundExclusions,
	}

	return arg
}

// ------- NewHdrInterceptorProcessor

func TestNewHdrInterceptorProcessor_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	hip, err := processor.NewHdrInterceptorProcessor(nil)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewHdrInterceptorProcessor_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.Headers = nil
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewHdrInterceptorProcessor_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.BlockBlackList = nil
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilBlackListCacher, err)
}

func TestNewHdrInterceptorProcessor_NilProofsPoolShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.Proofs = nil
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilProofsPool, err)
}

func TestNewHdrInterceptorProcessor_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.EnableEpochsHandler = nil
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewHdrInterceptorProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.False(t, check.IfNil(hip))
	assert.Nil(t, err)
	assert.False(t, hip.IsInterfaceNil())
}

// ------- Validate

func TestHdrInterceptorProcessor_ValidateNilHdrShouldErr(t *testing.T) {
	t.Parallel()

	hip, _ := processor.NewHdrInterceptorProcessor(createMockHdrArgument())

	err := hip.Validate(nil, "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestHdrInterceptorProcessor_ValidateHeaderIsBlackListedShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.BlockBlackList = &testscommon.TimeCacheStub{
		HasCalled: func(key string) bool {
			return true
		},
	}
	hip, _ := processor.NewHdrInterceptorProcessor(arg)

	hdrInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.GetHdrHandlerStub
	}{
		InterceptedDataStub: testscommon.InterceptedDataStub{
			HashCalled: func() []byte {
				return make([]byte, 0)
			},
		},
		GetHdrHandlerStub: mock.GetHdrHandlerStub{
			HeaderHandlerCalled: func() data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{}
			},
		},
	}
	err := hip.Validate(hdrInterceptedData, "")

	assert.Equal(t, process.ErrHeaderIsBlackListed, err)
}

func TestHdrInterceptorProcessor_ValidateReturnsNil(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.BlockBlackList = &testscommon.TimeCacheStub{}
	hip, _ := processor.NewHdrInterceptorProcessor(arg)

	hdrInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.GetHdrHandlerStub
	}{
		InterceptedDataStub: testscommon.InterceptedDataStub{
			HashCalled: func() []byte {
				return make([]byte, 0)
			},
		},
		GetHdrHandlerStub: mock.GetHdrHandlerStub{
			HeaderHandlerCalled: func() data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{}
			},
		},
	}
	err := hip.Validate(hdrInterceptedData, "")

	assert.Nil(t, err)
}

func TestHdrInterceptorProcessor_ValidateExcludedRoundShouldReturnBeforeBlacklistAccess(t *testing.T) {
	t.Parallel()

	const excludedRound = uint64(42)
	roundExclusions, err := common.NewRoundExclusionHandler([]config.HardforkRoundExclusionConfig{
		{StartRound: excludedRound, EndRound: excludedRound},
	})
	assert.NoError(t, err)
	arg := createMockHdrArgument()
	arg.RoundExclusions = roundExclusions
	arg.BlockBlackList = &testscommon.TimeCacheStub{
		SweepCalled: func() {
			assert.Fail(t, "blacklist should not be accessed")
		},
	}
	hip, err := processor.NewHdrInterceptorProcessor(arg)
	assert.NoError(t, err)
	hdrInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.GetHdrHandlerStub
	}{
		GetHdrHandlerStub: mock.GetHdrHandlerStub{
			HeaderHandlerCalled: func() data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{RoundField: excludedRound}
			},
		},
	}

	err = hip.Validate(hdrInterceptedData, "")

	assert.ErrorIs(t, err, common.ErrRoundExcluded)
}

// ------- Save

func TestHdrInterceptorProcessor_SaveNilDataShouldErr(t *testing.T) {
	t.Parallel()

	hip, _ := processor.NewHdrInterceptorProcessor(createMockHdrArgument())

	_, err := hip.Save(nil, "", "", "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestHdrInterceptorProcessor_SaveShouldWork(t *testing.T) {
	t.Parallel()

	minNonceWithProof := uint64(2)
	hdrInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.GetHdrHandlerStub
	}{
		InterceptedDataStub: testscommon.InterceptedDataStub{
			HashCalled: func() []byte {
				return []byte("hash")
			},
		},
		GetHdrHandlerStub: mock.GetHdrHandlerStub{
			HeaderHandlerCalled: func() data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{
					GetNonceCalled: func() uint64 {
						return minNonceWithProof
					},
				}
			},
		},
	}

	wasAddedHeaders := false

	arg := createMockHdrArgument()
	arg.Headers = &mock.HeadersCacherStub{
		AddCalled: func(headerHash []byte, header data.HeaderHandler) {
			wasAddedHeaders = true
		},
	}
	arg.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag
		},
	}

	hip, _ := processor.NewHdrInterceptorProcessor(arg)
	chanCalled := make(chan struct{}, 1)
	hip.RegisterHandler(func(topic string, hash []byte, data interface{}) {
		chanCalled <- struct{}{}
	})

	_, err := hip.Save(hdrInterceptedData, "", "", "")

	assert.Nil(t, err)
	assert.True(t, wasAddedHeaders)

	timeout := time.Second * 2
	select {
	case <-chanCalled:
	case <-time.After(timeout):
		assert.Fail(t, "save did not notify handler in a timely fashion")
	}
}

func TestHdrInterceptorProcessor_SaveExcludedRoundShouldNotStoreOrNotify(t *testing.T) {
	t.Parallel()

	const excludedRound = uint64(42)
	roundExclusions, err := common.NewRoundExclusionHandler([]config.HardforkRoundExclusionConfig{
		{StartRound: excludedRound, EndRound: excludedRound},
	})
	assert.NoError(t, err)
	arg := createMockHdrArgument()
	arg.RoundExclusions = roundExclusions
	arg.Headers = &mock.HeadersCacherStub{
		AddCalled: func(headerHash []byte, header data.HeaderHandler) {
			assert.Fail(t, "excluded header should not be stored")
		},
	}
	hip, err := processor.NewHdrInterceptorProcessor(arg)
	assert.NoError(t, err)
	hip.RegisterHandler(func(topic string, hash []byte, data interface{}) {
		assert.Fail(t, "excluded header should not be announced")
	})
	hdrInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.GetHdrHandlerStub
	}{
		GetHdrHandlerStub: mock.GetHdrHandlerStub{
			HeaderHandlerCalled: func() data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{RoundField: excludedRound}
			},
		},
	}

	saved, err := hip.Save(hdrInterceptedData, "", "", "")

	assert.False(t, saved)
	assert.ErrorIs(t, err, common.ErrRoundExcluded)
}

func TestHdrInterceptorProcessor_RegisterHandlerNilHandler(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	hip, _ := processor.NewHdrInterceptorProcessor(arg)

	hip.RegisterHandler(nil)
	assert.Equal(t, 0, len(hip.RegisteredHandlers()))
}

// ------- IsInterfaceNil

func TestHdrInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var hip *processor.HdrInterceptorProcessor

	assert.True(t, check.IfNil(hip))
}
