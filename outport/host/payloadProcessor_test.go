package host

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestPayloadProcessor_New(t *testing.T) {
	t.Parallel()

	pp, err := newPayloadProcessor(&testscommon.LoggerStub{})
	require.Nil(t, err)
	require.NotNil(t, pp)
	require.False(t, pp.IsInterfaceNil())
}

func TestPayloadProcessor_SetHandlerFunc(t *testing.T) {
	t.Parallel()

	pp, _ := newPayloadProcessor(&testscommon.LoggerStub{})

	// set nil handler func
	err := pp.SetHandlerFuncForTopic(nil, "topic")
	require.Equal(t, errNilHandlerFunc, err)

	// set empty topic
	err = pp.SetHandlerFuncForTopic(func() error { return nil }, "")
	require.Equal(t, errEmptyTopic, err)

	called := false
	hFunc := func() error {
		called = true
		return nil
	}
	err = pp.SetHandlerFuncForTopic(hFunc, outport.TopicSettings)
	require.Nil(t, err)

	// wrong topic should ignore
	err = pp.ProcessPayload([]byte(""), outport.TopicSaveAccounts, "1.0")
	require.Nil(t, err)
	require.False(t, called)

	// should call handler func
	err = pp.ProcessPayload([]byte(""), outport.TopicSettings, "1.0")
	require.Nil(t, err)
	require.True(t, called)
}
