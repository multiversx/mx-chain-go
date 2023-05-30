package host

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"
)

func TestPayloadProcessor_New(t *testing.T) {
	t.Parallel()

	pp, err := newPayloadProcessor()
	require.Nil(t, err)
	require.NotNil(t, pp)
	require.False(t, pp.IsInterfaceNil())
}

func TestPayloadProcessor_SetHandlerFunc(t *testing.T) {
	t.Parallel()

	pp, _ := newPayloadProcessor()

	// set nil handler func
	err := pp.SetHandlerFunc(nil)
	require.Equal(t, errNilHandlerFunc, err)

	called := false
	hFunc := func() {
		called = true
	}
	_ = pp.SetHandlerFunc(hFunc)

	// wrong topic should ignore
	err = pp.ProcessPayload([]byte(""), outport.TopicSaveAccounts)
	require.Nil(t, err)
	require.False(t, called)

	// should call handler func
	err = pp.ProcessPayload([]byte(""), outport.TopicSettings)
	require.Nil(t, err)
	require.True(t, called)
}
