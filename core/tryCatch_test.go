package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_TryCatch_WorksWhenNoError(t *testing.T) {
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

func Test_TryCatch_CatchesRuntimeError(t *testing.T) {
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

func Test_TryCatch_CatchesCustomError(t *testing.T) {
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

func Test_TryCatch_CatchesCustomErrorTyped(t *testing.T) {
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
