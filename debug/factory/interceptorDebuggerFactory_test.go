package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug/handler"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptorResolverDebuggerFactory_DisabledShouldWork(t *testing.T) {
	t.Parallel()

	idf, err := NewInterceptorDebuggerFactory(
		config.InterceptorResolverDebugConfig{
			Enabled: false,
		},
		&testscommon.SyncTimerStub{},
	)

	assert.Nil(t, err)
	expected := handler.NewDisabledInterceptorDebugHandler()
	assert.IsType(t, expected, idf)
}

func TestNewInterceptorResolverDebuggerFactory_InterceptorResolver(t *testing.T) {
	t.Parallel()

	idf, err := NewInterceptorDebuggerFactory(
		config.InterceptorResolverDebugConfig{
			Enabled:   true,
			CacheSize: 1000,
		},
		&testscommon.SyncTimerStub{},
	)

	assert.Nil(t, err)
	expected, _ := handler.NewInterceptorDebugHandler(config.InterceptorResolverDebugConfig{
		Enabled:   false,
		CacheSize: 1,
	}, &testscommon.SyncTimerStub{})
	assert.IsType(t, expected, idf)
}
