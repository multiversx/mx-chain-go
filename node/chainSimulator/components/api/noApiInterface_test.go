package api

import (
	"testing"

	"github.com/multiversx/mx-chain-go/facade"
	"github.com/stretchr/testify/require"
)

func TestNewNoApiInterface(t *testing.T) {
	t.Parallel()

	instance := NewNoApiInterface()
	require.NotNil(t, instance)

	interf := instance.RestApiInterface(0)
	require.Equal(t, facade.DefaultRestPortOff, interf)
}
