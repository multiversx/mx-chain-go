package processor_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockHdrArgument() *processor.ArgHdrInterceptorProcessor {
	arg := &processor.ArgHdrInterceptorProcessor{
		Headers:        &mock.HeadersCacherStub{},
		BlockBlackList: &mock.BlackListHandlerStub{},
	}

	return arg
}

//------- NewHdrInterceptorProcessor

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

func TestNewHdrInterceptorProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.False(t, check.IfNil(hip))
	assert.Nil(t, err)
	assert.False(t, hip.IsInterfaceNil())
}

//------- Validate

func TestHdrInterceptorProcessor_ValidateNilHdrShouldErr(t *testing.T) {
	t.Parallel()

	hip, _ := processor.NewHdrInterceptorProcessor(createMockHdrArgument())

	err := hip.Validate(nil, "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestHdrInterceptorProcessor_ValidateHeaderIsBlackListedShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.BlockBlackList = &mock.BlackListHandlerStub{
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
	}
	err := hip.Validate(hdrInterceptedData, "")

	assert.Equal(t, process.ErrHeaderIsBlackListed, err)
}

func TestHdrInterceptorProcessor_ValidateReturnsNil(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.BlockBlackList = &mock.BlackListHandlerStub{}
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
	}
	err := hip.Validate(hdrInterceptedData, "")

	assert.Nil(t, err)
}

//------- Save

func TestHdrInterceptorProcessor_SaveNilDataShouldErr(t *testing.T) {
	t.Parallel()

	hip, _ := processor.NewHdrInterceptorProcessor(createMockHdrArgument())

	err := hip.Save(nil, "", "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestHdrInterceptorProcessor_SaveShouldWork(t *testing.T) {
	t.Parallel()

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
				return &testscommon.HeaderHandlerStub{}
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

	hip, _ := processor.NewHdrInterceptorProcessor(arg)
	chanCalled := make(chan struct{}, 1)
	hip.RegisterHandler(func(topic string, hash []byte, data interface{}) {
		chanCalled <- struct{}{}
	})

	err := hip.Save(hdrInterceptedData, "", "")

	assert.Nil(t, err)
	assert.True(t, wasAddedHeaders)

	timeout := time.Second * 2
	select {
	case <-chanCalled:
	case <-time.After(timeout):
		assert.Fail(t, "save did not notify handler in a timely fashion")
	}
}

func TestHdrInterceptorProcessor_RegisterHandlerNilHandler(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	hip, _ := processor.NewHdrInterceptorProcessor(arg)

	hip.RegisterHandler(nil)
	assert.Equal(t, 0, len(hip.RegisteredHandlers()))
}

//------- IsInterfaceNil

func TestHdrInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var hip *processor.HdrInterceptorProcessor

	assert.True(t, check.IfNil(hip))
}
