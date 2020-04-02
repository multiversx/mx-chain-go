package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/debug/resolver"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptorResolverFactory_DisabledShouldWork(t *testing.T) {
	t.Parallel()

	irdh, err := NewInterceptorResolverFactory(
		config.InterceptorResolverDebugConfig{
			Enabled: false,
		},
	)

	assert.Nil(t, err)
	expected, _ := resolver.NewDisabledInterceptorResolver()
	assert.IsType(t, expected, irdh)
}

func TestNewInterceptorResolverFactory_InterceptorResolver(t *testing.T) {
	t.Parallel()

	irdh, err := NewInterceptorResolverFactory(
		config.InterceptorResolverDebugConfig{
			Enabled:   true,
			CacheSize: 1000,
		},
	)

	assert.Nil(t, err)
	expected, _ := resolver.NewInterceptorResolver(config.InterceptorResolverDebugConfig{
		Enabled:   false,
		CacheSize: 1,
	})
	assert.IsType(t, expected, irdh)
}
