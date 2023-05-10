package host

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	outportStubs "github.com/multiversx/mx-chain-go/testscommon/outport"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

var cannotSendOnRouteErr = errors.New("cannot send on route")

var log = logger.GetOrCreate("test")

func getMockArgs() ArgsHostDriver {
	return ArgsHostDriver{
		Marshaller: &marshal.JsonMarshalizer{},
		SenderHost: &outportStubs.SenderHostStub{},
		Log:        log,
	}
}

func TestNewWebsocketOutportDriverNodePart(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()
		args.Marshaller = nil

		o, err := NewHostDriver(args)
		require.Nil(t, o)
		require.Equal(t, core.ErrNilMarshalizer, err)
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()
		args.Log = nil

		o, err := NewHostDriver(args)
		require.Nil(t, o)
		require.Equal(t, core.ErrNilLogger, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()

		o, err := NewHostDriver(args)
		require.NotNil(t, o)
		require.NoError(t, err)
		require.False(t, o.IsInterfaceNil())
	})
}

func TestWebsocketOutportDriverNodePart_SaveBlock(t *testing.T) {
	t.Parallel()

	t.Run("SaveBlock - should error", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return cannotSendOnRouteErr
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveBlock(&outport.OutportBlock{})
		require.True(t, errors.Is(err, cannotSendOnRouteErr))
	})

	t.Run("SaveBlock - should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			require.Nil(t, r)
		}()
		args := getMockArgs()
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveBlock(&outport.OutportBlock{})
		require.NoError(t, err)
	})
}

func TestWebsocketOutportDriverNodePart_FinalizedBlock(t *testing.T) {
	t.Parallel()

	t.Run("Finalized block - should error", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return cannotSendOnRouteErr
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.FinalizedBlock(&outport.FinalizedBlock{HeaderHash: []byte("header hash")})
		require.True(t, errors.Is(err, cannotSendOnRouteErr))
	})

	t.Run("Finalized block - should work", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return nil
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.FinalizedBlock(&outport.FinalizedBlock{HeaderHash: []byte("header hash")})
		require.NoError(t, err)
	})
}

func TestWebsocketOutportDriverNodePart_RevertIndexedBlock(t *testing.T) {
	t.Parallel()

	t.Run("RevertIndexedBlock - should error", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return cannotSendOnRouteErr
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.RevertIndexedBlock(nil)
		require.True(t, errors.Is(err, cannotSendOnRouteErr))
	})

	t.Run("RevertIndexedBlock block - should work", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return nil
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.RevertIndexedBlock(nil)
		require.NoError(t, err)
	})
}

func TestWebsocketOutportDriverNodePart_SaveAccounts(t *testing.T) {
	t.Parallel()

	t.Run("SaveAccounts - should error", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return cannotSendOnRouteErr
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveAccounts(nil)
		require.True(t, errors.Is(err, cannotSendOnRouteErr))
	})

	t.Run("SaveAccounts block - should work", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return nil
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveAccounts(nil)
		require.NoError(t, err)
	})
}

func TestWebsocketOutportDriverNodePart_SaveRoundsInfo(t *testing.T) {
	t.Parallel()

	t.Run("SaveRoundsInfo - should error", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return cannotSendOnRouteErr
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveRoundsInfo(nil)
		require.True(t, errors.Is(err, cannotSendOnRouteErr))
	})

	t.Run("SaveRoundsInfo block - should work", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return nil
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveRoundsInfo(nil)
		require.NoError(t, err)
	})
}

func TestWebsocketOutportDriverNodePart_SaveValidatorsPubKeys(t *testing.T) {
	t.Parallel()

	t.Run("SaveValidatorsPubKeys - should error", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return cannotSendOnRouteErr
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveValidatorsPubKeys(nil)
		require.True(t, errors.Is(err, cannotSendOnRouteErr))
	})

	t.Run("SaveValidatorsPubKeys block - should work", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return nil
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveValidatorsPubKeys(nil)
		require.NoError(t, err)
	})
}

func TestWebsocketOutportDriverNodePart_SaveValidatorsRating(t *testing.T) {
	t.Parallel()

	t.Run("SaveValidatorsRating - should error", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return cannotSendOnRouteErr
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveValidatorsRating(nil)
		require.True(t, errors.Is(err, cannotSendOnRouteErr))
	})

	t.Run("SaveValidatorsRating block - should work", func(t *testing.T) {
		args := getMockArgs()
		args.SenderHost = &outportStubs.SenderHostStub{
			SendCalled: func(_ []byte, _ string) error {
				return nil
			},
		}
		o, err := NewHostDriver(args)
		require.NoError(t, err)

		err = o.SaveValidatorsRating(nil)
		require.NoError(t, err)
	})
}

func TestWebsocketOutportDriverNodePart_SaveBlock_PayloadCheck(t *testing.T) {
	t.Parallel()

	mockArgs := getMockArgs()

	outportBlock := &outport.OutportBlock{BlockData: &outport.BlockData{Body: &block.Body{}}}
	marshaledData, err := mockArgs.Marshaller.Marshal(outportBlock)
	require.Nil(t, err)

	mockArgs.SenderHost = &outportStubs.SenderHostStub{
		SendCalled: func(payload []byte, _ string) error {
			require.Equal(t, marshaledData, payload)

			return nil
		},
	}
	o, err := NewHostDriver(mockArgs)
	require.NoError(t, err)

	err = o.SaveBlock(outportBlock)
	require.NoError(t, err)
}

func TestWebsocketOutportDriverNodePart_Close(t *testing.T) {
	t.Parallel()

	closedWasCalled := false
	args := getMockArgs()
	args.SenderHost = &outportStubs.SenderHostStub{
		CloseCalled: func() error {
			closedWasCalled = true
			return nil
		},
	}

	o, err := NewHostDriver(args)
	require.NoError(t, err)

	err = o.Close()
	require.NoError(t, err)
	require.True(t, closedWasCalled)
}
