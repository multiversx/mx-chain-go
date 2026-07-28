package api

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
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

func TestFreePortAPIConfigurator_ShouldKeepAllocatedPortsUnique(t *testing.T) {
	t.Parallel()

	const reservedPort = 12345
	instance := NewFreePortAPIConfigurator(apiInterface, reservedPort)

	const numInterfaces = 100
	interfaces := make(chan string, numInterfaces)
	var wg sync.WaitGroup
	wg.Add(numInterfaces)

	for idx := 0; idx < numInterfaces; idx++ {
		go func() {
			defer wg.Done()
			interfaces <- instance.RestApiInterface(0)
		}()
	}

	wg.Wait()
	close(interfaces)

	ports := make(map[int]struct{}, numInterfaces)
	for restInterface := range interfaces {
		separatorIndex := strings.LastIndex(restInterface, ":")
		require.Positive(t, separatorIndex)

		port, err := strconv.Atoi(restInterface[separatorIndex+1:])
		require.NoError(t, err)
		require.NotEqual(t, reservedPort, port)
		_, alreadyAllocated := ports[port]
		require.False(t, alreadyAllocated)
		ports[port] = struct{}{}
	}

	require.Len(t, ports, numInterfaces)
}

func TestFreePortAPIConfigurator_ShouldSkipReservedPort(t *testing.T) {
	t.Parallel()

	const (
		reservedPort = 12345
		freePort     = 23456
	)
	instance := NewFreePortAPIConfigurator(apiInterface, reservedPort)
	candidates := []int{reservedPort, freePort}
	instance.portProvider = func() int {
		port := candidates[0]
		candidates = candidates[1:]
		return port
	}

	restInterface := instance.RestApiInterface(0)

	require.Equal(t, fmt.Sprintf("%s:%d", apiInterface, freePort), restInterface)
	require.Empty(t, candidates)
}
