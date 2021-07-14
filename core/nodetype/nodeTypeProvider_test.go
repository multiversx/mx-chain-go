package nodetype

import (
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestNewNodeTypeProvider(t *testing.T) {
	t.Parallel()

	ntp := NewNodeTypeProvider(core.NodeTypeObserver)
	require.False(t, check.IfNil(ntp))
}

func TestNodeTypeProvider_SetterAndGetterType(t *testing.T) {
	t.Parallel()

	ntp := NewNodeTypeProvider(core.NodeTypeObserver)
	require.Equal(t, core.NodeTypeObserver, ntp.GetType())

	ntp.SetType(core.NodeTypeObserver)
	require.Equal(t, core.NodeTypeObserver, ntp.GetType())

	ntp.SetType(core.NodeTypeValidator)
	require.Equal(t, core.NodeTypeValidator, ntp.GetType())

	ntp.SetType(core.NodeTypeObserver)
	require.Equal(t, core.NodeTypeObserver, ntp.GetType())
}

func TestNodeTypeProvider_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	ntp := NewNodeTypeProvider(core.NodeTypeObserver)

	defer func() {
		r := recover()
		require.Empty(t, r)
	}()

	numGoRoutines := 100

	wg := sync.WaitGroup{}
	wg.Add(numGoRoutines)
	for i := 0; i < numGoRoutines; i++ {
		go func(idx int) {
			modRes := idx % 3
			switch modRes {
			case 0:
				_ = ntp.GetType()
			case 1:
				ntp.SetType(core.NodeTypeObserver)
			case 2:
				ntp.SetType(core.NodeTypeValidator)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
