package keyBuilder

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisabledKeyBuilder(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	builder := NewDisabledKeyBuilder()
	require.NotNil(t, builder)

	builder.BuildKey([]byte("key"))

	key, err := builder.GetKey()
	require.Nil(t, err)
	require.True(t, bytes.Equal(key, []byte{}))

	clonedBuilder := builder.Clone()
	require.Equal(t, &disabledKeyBuilder{}, clonedBuilder)
}
