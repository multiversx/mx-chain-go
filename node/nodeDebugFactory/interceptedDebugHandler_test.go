package nodeDebugFactory

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func TestCreateInterceptedDebugHandler_NilNodeWrapperShouldErr(t *testing.T) {
	t.Parallel()

	err := CreateInterceptedDebugHandler(
		nil,
		&mock.InterceptorsContainerStub{},
		&mock.ResolversFinderStub{},
		config.InterceptorResolverDebugConfig{},
	)

	assert.Equal(t, ErrNilNodeWrapper, err)
}

func TestCreateInterceptedDebugHandler_NilInterceptorsShouldErr(t *testing.T) {
	t.Parallel()

	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{},
		nil,
		&mock.ResolversFinderStub{},
		config.InterceptorResolverDebugConfig{},
	)

	assert.Equal(t, ErrNilInterceptorContainer, err)
}

func TestCreateInterceptedDebugHandler_NilReolversShouldErr(t *testing.T) {
	t.Parallel()

	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{},
		&mock.InterceptorsContainerStub{},
		nil,
		config.InterceptorResolverDebugConfig{},
	)

	assert.Equal(t, ErrNilResolverContainer, err)
}

func TestCreateInterceptedDebugHandler_InvalidDebugConfigShouldErr(t *testing.T) {
	t.Parallel()

	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{},
		&mock.InterceptorsContainerStub{},
		&mock.ResolversFinderStub{},
		config.InterceptorResolverDebugConfig{
			Enabled:   true,
			CacheSize: 0,
		},
	)

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Must provide a positive size"))
}

func TestCreateInterceptedDebugHandler_SettingOnInterceptorsErrShouldErr(t *testing.T) {
	t.Parallel()

	interceptorsIterateCalled := false
	resolversIterateCalled := false
	addQueryHandlerCalled := false
	expectedErr := errors.New("expected err")
	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{
			AddQueryHandlerCalled: func(name string, handler debug.QueryHandler) error {
				addQueryHandlerCalled = true
				return nil
			},
		},
		&mock.InterceptorsContainerStub{
			IterateCalled: func(handler func(key string, interceptor process.Interceptor) bool) {
				handler("key", &mock.InterceptorStub{
					SetInterceptedDebugHandlerCalled: func(handler process.InterceptedDebugger) error {
						return expectedErr
					},
				})
				interceptorsIterateCalled = true
			},
		},
		&mock.ResolversFinderStub{
			IterateCalled: func(handler func(key string, resolver dataRetriever.Resolver) bool) {
				handler("key", &mock.HeaderResolverStub{})
				resolversIterateCalled = true
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
}

func TestCreateInterceptedDebugHandler_SettingOnResolverErrShouldErr(t *testing.T) {
	t.Parallel()

	interceptorsIterateCalled := false
	resolversIterateCalled := false
	addQueryHandlerCalled := false
	expectedErr := errors.New("expected err")
	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{
			AddQueryHandlerCalled: func(name string, handler debug.QueryHandler) error {
				addQueryHandlerCalled = true
				return nil
			},
		},
		&mock.InterceptorsContainerStub{
			IterateCalled: func(handler func(key string, interceptor process.Interceptor) bool) {
				handler("key", &mock.InterceptorStub{})
				interceptorsIterateCalled = true
			},
		},
		&mock.ResolversFinderStub{
			IterateCalled: func(handler func(key string, resolver dataRetriever.Resolver) bool) {
				handler("key", &mock.HeaderResolverStub{
					SetResolverDebugHandlerCalled: func(handler dataRetriever.ResolverDebugHandler) error {
						return expectedErr
					},
				})
				resolversIterateCalled = true
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
}

func TestCreateInterceptedDebugHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	interceptorsIterateCalled := false
	resolversIterateCalled := false
	addQueryHandlerCalled := false
	err := CreateInterceptedDebugHandler(
		&mock.NodeWrapperStub{
			AddQueryHandlerCalled: func(name string, handler debug.QueryHandler) error {
				addQueryHandlerCalled = true
				return nil
			},
		},
		&mock.InterceptorsContainerStub{
			IterateCalled: func(handler func(key string, interceptor process.Interceptor) bool) {
				handler("key", &mock.InterceptorStub{})
				interceptorsIterateCalled = true
			},
		},
		&mock.ResolversFinderStub{
			IterateCalled: func(handler func(key string, resolver dataRetriever.Resolver) bool) {
				handler("key", &mock.HeaderResolverStub{})
				resolversIterateCalled = true
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
}
