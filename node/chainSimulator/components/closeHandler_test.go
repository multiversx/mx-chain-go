package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// localErrorlessCloser implements errorlessCloser interface
type localErrorlessCloser struct {
	wasCalled bool
}

// Close -
func (closer *localErrorlessCloser) Close() {
	closer.wasCalled = true
}

// localCloser implements io.Closer interface
type localCloser struct {
	wasCalled     bool
	expectedError error
}

// Close -
func (closer *localCloser) Close() error {
	closer.wasCalled = true
	return closer.expectedError
}

// localCloseAllHandler implements allCloser interface
type localCloseAllHandler struct {
	wasCalled     bool
	expectedError error
}

// CloseAll -
func (closer *localCloseAllHandler) CloseAll() error {
	closer.wasCalled = true
	return closer.expectedError
}

func TestCloseHandler(t *testing.T) {
	t.Parallel()

	handler := NewCloseHandler()
	require.NotNil(t, handler)

	handler.AddComponent(nil) // for coverage only

	lec := &localErrorlessCloser{}
	handler.AddComponent(lec)

	lcNoError := &localCloser{}
	handler.AddComponent(lcNoError)

	lcWithError := &localCloser{expectedError: expectedErr}
	handler.AddComponent(lcWithError)

	lcahNoError := &localCloseAllHandler{}
	handler.AddComponent(lcahNoError)

	lcahWithError := &localCloseAllHandler{expectedError: expectedErr}
	handler.AddComponent(lcahWithError)

	err := handler.Close()
	require.True(t, strings.Contains(err.Error(), expectedErr.Error()))
}
