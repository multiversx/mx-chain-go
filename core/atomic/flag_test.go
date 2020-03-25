package atomic

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlag_Set(t *testing.T) {
	var flag Flag
	var wg sync.WaitGroup

	require.False(t, flag.IsSet())

	wg.Add(2)

	go func() {
		flag.Set()
		wg.Done()
	}()

	go func() {
		flag.Set()
		wg.Done()
	}()

	wg.Wait()
	require.True(t, flag.IsSet())
}

func TestFlag_Unset(t *testing.T) {
	var flag Flag
	var wg sync.WaitGroup

	flag.Set()
	require.True(t, flag.IsSet())

	wg.Add(2)

	go func() {
		flag.Unset()
		wg.Done()
	}()

	go func() {
		flag.Unset()
		wg.Done()
	}()

	wg.Wait()
	require.False(t, flag.IsSet())
}

func TestFlag_Toggle(t *testing.T) {
	var flag Flag
	var wg sync.WaitGroup

	// First, Toggle(true)
	wg.Add(2)

	go func() {
		flag.Toggle(true)
		wg.Done()
	}()

	go func() {
		flag.Toggle(true)
		wg.Done()
	}()

	wg.Wait()
	require.True(t, flag.IsSet())

	// Then, Toggle(false)
	wg.Add(2)

	go func() {
		flag.Toggle(false)
		wg.Done()
	}()

	go func() {
		flag.Toggle(false)
		wg.Done()
	}()

	wg.Wait()
	require.False(t, flag.IsSet())
}
