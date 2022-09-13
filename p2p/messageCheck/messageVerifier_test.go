package messagecheck_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
	messagecheck "github.com/ElrondNetwork/elrond-go/p2p/messageCheck"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewMessageVerifier(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := messagecheck.ArgsMessageVerifier{
			Marshaller: nil,
		}

		mv, err := messagecheck.NewMessageVerifier(args)
		require.Nil(t, mv)
		require.Equal(t, p2p.ErrNilMarshalizer, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := messagecheck.ArgsMessageVerifier{
			Marshaller: &mock.MarshalizerStub{},
		}

		mv, err := messagecheck.NewMessageVerifier(args)
		require.Nil(t, err)
		require.False(t, check.IfNil(mv))
	})
}

func TestSerializeDeserialize(t *testing.T) {
	t.Parallel()

	t.Run("serialize, marshal should err", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := messagecheck.ArgsMessageVerifier{
			Marshaller: &testscommon.MarshalizerStub{
				MarshalCalled: func(obj interface{}) ([]byte, error) {
					return nil, expectedErr
				},
			},
		}

		messages := []p2p.MessageP2P{
			&message.Message{
				FromField:    []byte("from1"),
				PayloadField: []byte("payload1"),
			},
		}

		mv, err := messagecheck.NewMessageVerifier(args)
		require.Nil(t, err)

		messagesBytes, err := mv.Serialize(messages)
		require.Nil(t, messagesBytes)
		require.Equal(t, expectedErr, err)
	})

	t.Run("deserialize, unmarshal should err", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := messagecheck.ArgsMessageVerifier{
			Marshaller: &testscommon.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					return expectedErr
				},
			},
		}

		mv, err := messagecheck.NewMessageVerifier(args)
		require.Nil(t, err)

		messages, err := mv.Deserialize([]byte("messages data"))
		require.Nil(t, messages)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := messagecheck.ArgsMessageVerifier{
			Marshaller: &testscommon.MarshalizerMock{},
		}

		messages := []p2p.MessageP2P{
			&message.Message{
				FromField:    []byte("from1"),
				PayloadField: []byte("payload1"),
			},
			&message.Message{
				FromField:    []byte("from2"),
				PayloadField: []byte("payload2"),
			},
		}

		mv, err := messagecheck.NewMessageVerifier(args)
		require.Nil(t, err)

		messagesBytes, err := mv.Serialize(messages)
		require.Nil(t, err)

		expectedMessages, err := mv.Deserialize(messagesBytes)
		require.Nil(t, err)

		require.Equal(t, messages, expectedMessages)
	})
}
