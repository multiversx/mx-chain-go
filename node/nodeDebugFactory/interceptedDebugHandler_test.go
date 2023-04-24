package nodeDebugFactory

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverTests "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/assert"
)

func TestCreateInterceptedDebugHandler_NilNodeWrapperShouldErr(t *testing.T) {
	t.Parallel()

	err := CreateInterceptedDebugHandler(
		nil,
		&testscommon.InterceptorsContainerStub{},
		&dataRetrieverTests.ResolversContainerStub{},
		&dataRetrieverTests.RequestersContainerStub{},
		config.InterceptorResolverDebugConfig{},
	)

	assert.Equal(t, ErrNilNodeWrapper, err)
}

func TestCreateInterceptedDebugHandler_NilInterceptorsShouldErr(t *testing.T) {
	t.Parallel()

	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{},
		nil,
		&dataRetrieverTests.ResolversContainerStub{},
		&dataRetrieverTests.RequestersFinderStub{},
		config.InterceptorResolverDebugConfig{},
	)

	assert.Equal(t, ErrNilInterceptorContainer, err)
}

func TestCreateInterceptedDebugHandler_NilResolversShouldErr(t *testing.T) {
	t.Parallel()

	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{},
		&testscommon.InterceptorsContainerStub{},
		nil,
		&dataRetrieverTests.RequestersFinderStub{},
		config.InterceptorResolverDebugConfig{},
	)

	assert.Equal(t, ErrNilResolverContainer, err)
}

func TestCreateInterceptedDebugHandler_NilRequestersShouldErr(t *testing.T) {
	t.Parallel()

	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{},
		&testscommon.InterceptorsContainerStub{},
		&dataRetrieverTests.ResolversContainerStub{},
		nil,
		config.InterceptorResolverDebugConfig{},
	)

	assert.Equal(t, ErrNilRequestersContainer, err)
}

func TestCreateInterceptedDebugHandler_InvalidDebugConfigShouldErr(t *testing.T) {
	t.Parallel()

	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{},
		&testscommon.InterceptorsContainerStub{},
		&dataRetrieverTests.ResolversContainerStub{},
		&dataRetrieverTests.RequestersFinderStub{},
		config.InterceptorResolverDebugConfig{
			Enabled:   true,
			CacheSize: 0,
		},
	)

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "must provide a positive size"))
}

func TestCreateInterceptedDebugHandler_SettingOnInterceptorsErrShouldErr(t *testing.T) {
	t.Parallel()

	interceptorsIterateCalled := false
	resolversIterateCalled := false
	requestersIterateCalled := false
	addQueryHandlerCalled := false
	expectedErr := errors.New("expected err")
	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{
			AddQueryHandlerCalled: func(name string, handler debug.QueryHandler) error {
				addQueryHandlerCalled = true
				return nil
			},
		},
		&testscommon.InterceptorsContainerStub{
			IterateCalled: func(handler func(key string, interceptor process.Interceptor) bool) {
				handler("key", &testscommon.InterceptorStub{
					SetInterceptedDebugHandlerCalled: func(handler process.InterceptedDebugger) error {
						return expectedErr
					},
				})
				interceptorsIterateCalled = true
			},
		},
		&dataRetrieverTests.ResolversContainerStub{
			IterateCalled: func(handler func(key string, resolver dataRetriever.Resolver) bool) {
				resolversIterateCalled = true
			},
		},
		&dataRetrieverTests.RequestersFinderStub{
			IterateCalled: func(handler func(key string, resolver dataRetriever.Requester) bool) {
				requestersIterateCalled = true
			},
		},
		config.InterceptorResolverDebugConfig{
			Enabled: false,
		},
	)

	assert.True(t, errors.Is(err, expectedErr))
	assert.False(t, addQueryHandlerCalled)
	assert.True(t, interceptorsIterateCalled)
	assert.False(t, resolversIterateCalled)
	assert.False(t, requestersIterateCalled)
}

func TestCreateInterceptedDebugHandler_SettingOnResolverErrShouldErr(t *testing.T) {
	t.Parallel()

	interceptorsIterateCalled := false
	resolversIterateCalled := false
	requestersIterateCalled := false
	addQueryHandlerCalled := false
	expectedErr := errors.New("expected err")
	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{
			AddQueryHandlerCalled: func(name string, handler debug.QueryHandler) error {
				addQueryHandlerCalled = true
				return nil
			},
		},
		&testscommon.InterceptorsContainerStub{
			IterateCalled: func(handler func(key string, interceptor process.Interceptor) bool) {
				handler("key", &testscommon.InterceptorStub{})
				interceptorsIterateCalled = true
			},
		},
		&dataRetrieverTests.ResolversContainerStub{
			IterateCalled: func(handler func(key string, resolver dataRetriever.Resolver) bool) {
				handler("key", &dataRetrieverMocks.HeaderResolverStub{
					SetDebugHandlerCalled: func(handler dataRetriever.DebugHandler) error {
						return expectedErr
					},
				})
				resolversIterateCalled = true
			},
		},
		&dataRetrieverTests.RequestersFinderStub{
			IterateCalled: func(handler func(key string, resolver dataRetriever.Requester) bool) {
				requestersIterateCalled = true
			},
		},
		config.InterceptorResolverDebugConfig{
			Enabled: false,
		},
	)

	assert.True(t, errors.Is(err, expectedErr))
	assert.False(t, addQueryHandlerCalled)
	assert.True(t, interceptorsIterateCalled)
	assert.True(t, resolversIterateCalled)
	assert.False(t, requestersIterateCalled)
}

func TestCreateInterceptedDebugHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	interceptorsIterateCalled := false
	resolversIterateCalled := false
	requestersIterateCalled := false
	addQueryHandlerCalled := false
	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{
			AddQueryHandlerCalled: func(name string, handler debug.QueryHandler) error {
				addQueryHandlerCalled = true
				return nil
			},
		},
		&testscommon.InterceptorsContainerStub{
			IterateCalled: func(handler func(key string, interceptor process.Interceptor) bool) {
				handler("key", &testscommon.InterceptorStub{})
				interceptorsIterateCalled = true
			},
		},
		&dataRetrieverTests.ResolversContainerStub{
			IterateCalled: func(handler func(key string, resolver dataRetriever.Resolver) bool) {
				handler("key", &dataRetrieverMocks.HeaderResolverStub{})
				resolversIterateCalled = true
			},
		},
		&dataRetrieverTests.RequestersFinderStub{
			IterateCalled: func(handler func(key string, resolver dataRetriever.Requester) bool) {
				handler("key", &dataRetrieverTests.HeaderRequesterStub{})
				requestersIterateCalled = true
			},
		},
		config.InterceptorResolverDebugConfig{
			Enabled: false,
		},
	)

	assert.Nil(t, err)
	assert.True(t, addQueryHandlerCalled)
	assert.True(t, interceptorsIterateCalled)
	assert.True(t, resolversIterateCalled)
	assert.True(t, requestersIterateCalled)
}
