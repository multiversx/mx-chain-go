package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug/handler"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptorResolverDebuggerFactory_DisabledShouldWork(t *testing.T) {
	t.Parallel()

	idf, err := NewInterceptorDebuggerFactory(
		config.InterceptorResolverDebugConfig{
			Enabled: false,
		},
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
	)

	assert.Nil(t, err)
	expected, _ := handler.NewInterceptorDebugHandler(config.InterceptorResolverDebugConfig{
		Enabled:   false,
		CacheSize: 1,
	})
	assert.IsType(t, expected, idf)
}
