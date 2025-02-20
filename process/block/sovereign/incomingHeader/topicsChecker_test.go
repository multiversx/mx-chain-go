package incomingHeader

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTopicsChecker(t *testing.T) {
	t.Parallel()

	tc := NewTopicsChecker()
	require.False(t, tc.IsInterfaceNil())
}
func TestTopicsChecker_CheckValidity(t *testing.T) {
	t.Parallel()

	tc := NewTopicsChecker()

	topics1 := [][]byte{[]byte("topic1")}
	err := tc.CheckValidity(topics1)
	require.ErrorContains(t, err, errInvalidNumTopicsIncomingEvent.Error())

	topics5 := [][]byte{[]byte("topic1"), []byte("topic2"), []byte("topic3"), []byte("topic4"), []byte("topic5")}
	err = tc.CheckValidity(topics5)
	require.NoError(t, err)
}
