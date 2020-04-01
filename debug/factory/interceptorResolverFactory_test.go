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
		config.DebugConfig{
			Enabled:     false,
			CachersSize: 0,
		},
	)

	assert.Nil(t, err)
	expected, _ := resolver.NewDisabledInterceptorResolver()
	assert.IsType(t, expected, irdh)
}

func TestNewInterceptorResolverFactory_InterceptorResolver(t *testing.T) {
	t.Parallel()

	irdh, err := NewInterceptorResolverFactory(
		config.DebugConfig{
			Enabled:     true,
			CachersSize: 10000,
		},
	)

	assert.Nil(t, err)
	expected, _ := resolver.NewInterceptorResolver(1)
	assert.IsType(t, expected, irdh)
}
