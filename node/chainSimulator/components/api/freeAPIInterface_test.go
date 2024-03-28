package api

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFreePortAPIConfigurator(t *testing.T) {
	t.Parallel()

	instance := NewFreePortAPIConfigurator(apiInterface)
	require.NotNil(t, instance)

	interf := instance.RestApiInterface(0)
	require.True(t, strings.Contains(interf, fmt.Sprintf("%s:", apiInterface)))
}
