package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const apiInterface = "127.0.0.1:8080"

func TestNewFixedPortAPIConfigurator(t *testing.T) {
	t.Parallel()

	instance := NewFixedPortAPIConfigurator(apiInterface, map[uint32]int{0: 123})
	require.NotNil(t, instance)

	interf := instance.RestApiInterface(0)
	require.Equal(t, fmt.Sprintf("%s:123", apiInterface), interf)
}
