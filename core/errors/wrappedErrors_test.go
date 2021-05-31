package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapError(t *testing.T) {
	t.Parallel()

	we := WrapError(nil)
	require.NotNil(t, we)
	assert.Equal(t, nil, we.GetLastError())
	assert.True(t, strings.Contains(we.Error(), "/core/errors/wrappedErrors_test.go"))
	assert.True(t, strings.Contains(we.Error(), nilPlaceholder))

	definedError := errors.New("defined errors")
	we = WrapError(definedError)
	require.NotNil(t, we)
	assert.Equal(t, definedError, we.GetLastError())
	assert.True(t, strings.Contains(we.Error(), "/core/errors/wrappedErrors_test.go"))
	assert.True(t, strings.Contains(we.Error(), definedError.Error()))
	assert.True(t, we.Is(definedError))

	we1 := WrapError(definedError)
	we2 := WrapError(we1)
	require.NotNil(t, we2)
	assert.Equal(t, definedError, we2.GetLastError())
	assert.True(t, strings.Contains(we2.Error(), "/core/errors/wrappedErrors_test.go"))
	assert.True(t, strings.Contains(we2.Error(), we1.Error()))
	assert.True(t, we2.Is(definedError))
}

func TestWrappableError_WrapWithMessage(t *testing.T) {
	t.Parallel()

	we := WrapError(nil)
	we = we.WrapWithMessage("")
	require.NotNil(t, we)
	assert.NotNil(t, we.GetLastError())
	assert.True(t, strings.Contains(we.Error(), "/core/errors/wrappedErrors_test.go"))
	assert.True(t, strings.Contains(we.Error(), nilPlaceholder))

	we = WrapError(nil)
	msg := "a message"
	we = we.WrapWithMessage(msg)
	require.NotNil(t, we)
	assert.NotNil(t, we.GetLastError())
	assert.True(t, strings.Contains(we.Error(), "/core/errors/wrappedErrors_test.go"))
	assert.True(t, strings.Contains(we.Error(), nilPlaceholder))
	assert.True(t, strings.Contains(we.Error(), msg))
}

func TestWrappableError_WrapWithStackTrace(t *testing.T) {
	t.Parallel()

	we := WrapError(nil)
	we = we.WrapWithStackTrace()
	require.NotNil(t, we)
	assert.NotNil(t, we.GetLastError())
	path := "/core/errors/wrappedErrors_test.go"
	assert.True(t, strings.Contains(we.Error(), path))
	assert.True(t, strings.Contains(we.Error(), nilPlaceholder))
	assert.False(t, strings.LastIndex(we.Error(), path) == strings.Index(we.Error(), path)) //found the path more than once
}

func TestWrappableError_Is(t *testing.T) {
	t.Parallel()

	we := WrapError(nil)
	err := fmt.Errorf("error")
	assert.False(t, we.Is(err))

	we = we.WrapWithError(err)
	assert.True(t, we.Is(err))
}

func TestWrappableError_GetBaseLastError(t *testing.T) {
	t.Parallel()

	we := WrapError(nil)
	err := fmt.Errorf("error")
	assert.Nil(t, we.GetBaseError())
	assert.Nil(t, we.GetLastError())

	we = we.WrapWithError(err)
	assert.Nil(t, we.GetBaseError())
	assert.Equal(t, err, we.GetLastError())
}

func TestWrappableError_Unwrap(t *testing.T) {
	t.Parallel()

	we := WrapError(nil)
	unwrappedErr := we.Unwrap()
	assert.Nil(t, unwrappedErr)

	err := fmt.Errorf("error")
	we = WrapError(err)
	unwrappedErr = we.Unwrap()
	assert.Equal(t, err, unwrappedErr)
}
