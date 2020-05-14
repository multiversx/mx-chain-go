package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTryCatch_WorksWhenNoError(t *testing.T) {
	tryCalled := false
	catchCalled := false

	try := func() {
		tryCalled = true
	}

	catch := func(err error) {
		catchCalled = true
	}

	TryCatch(try, catch, "message")

	require.True(t, tryCalled)
	require.False(t, catchCalled)
}

func TestTryCatch_CatchesRuntimeError(t *testing.T) {
	var caughtError error

	try := func() {
		bytes := make([]byte, 42)
		// Causes runtime error.
		bytes[42]++
	}

	catch := func(err error) {
		caughtError = err
	}

	TryCatch(try, catch, "message")

	require.NotNil(t, caughtError)
}

func TestTryCatch_CatchesCustomError(t *testing.T) {
	var caughtError error

	try := func() {
		panic("untyped error")
	}

	catch := func(err error) {
		caughtError = err
	}

	TryCatch(try, catch, "!thisMessage!")

	require.NotNil(t, caughtError)
	require.Contains(t, caughtError.Error(), "!thisMessage!")
	require.Contains(t, caughtError.Error(), "untyped error")
}

func TestTryCatch_CatchesCustomErrorTyped(t *testing.T) {
	var caughtError error
	customError := fmt.Errorf("error")

	try := func() {
		panic(customError)
	}

	catch := func(err error) {
		caughtError = err
	}

	TryCatch(try, catch, "!thisMessage!")

	require.NotNil(t, caughtError)
	require.Equal(t, customError, caughtError)
}

func BenchmarkTryCatch(b *testing.B) {
	var foo int

	tryGoodFunction := func() {
		foo = foo * 2
	}

	tryBadFunction := func() {
		foo = 1 / foo
	}

	// ~1x
	foo = 1
	b.Run("no try-catch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000000; j++ {
				tryGoodFunction()
			}
		}
	})

	// ~20x (~15_000_000 calls per second on an average computer)
	foo = 1
	b.Run("optimistic (no panic)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000000; j++ {
				TryCatch(tryGoodFunction, func(err error) {}, "tried by failed")
			}
		}
	})

	// ~40x
	foo = 0
	b.Run("pesimistic (divide by zero)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < 1000000; j++ {
				TryCatch(tryBadFunction, func(err error) {}, "tried by failed")
			}
		}
	})
}
