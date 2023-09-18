package processor

import (
	"encoding/hex"
	stdErr "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	sovereignData "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/stretchr/testify/require"
)

func createSovHdrProcArgs() *ArgsSovereignHeaderInterceptorProcessor {
	return &ArgsSovereignHeaderInterceptorProcessor{
		BlockBlackList:           &testscommon.TimeCacheStub{},
		Hasher:                   &hashingMocks.HasherMock{},
		Marshaller:               &marshallerMock.MarshalizerMock{},
		IncomingHeaderSubscriber: &sovereign.IncomingHeaderSubscriberStub{},
		HeadersPool:              &mock.HeadersCacherStub{},
	}
}

func TestNewSovereignHdrInterceptorProcessor(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createSovHdrProcArgs()
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Nil(t, err)
		require.False(t, sovHdrProc.IsInterfaceNil())
		require.NotPanics(t, func() { sovHdrProc.RegisterHandler(nil) })
	})
	t.Run("nil args, should return error", func(t *testing.T) {
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(nil)
		require.Equal(t, process.ErrNilArgumentStruct, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil black list cache, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.BlockBlackList = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, process.ErrNilBlackListCacher, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.Hasher = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, errors.ErrNilHasher, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.Marshaller = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, errors.ErrNilMarshalizer, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil incoming header subscriber, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.IncomingHeaderSubscriber = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, errors.ErrNilIncomingHeaderSubscriber, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil headers pool, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.HeadersPool = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, process.ErrNilHeadersDataPool, err)
		require.Nil(t, sovHdrProc)
	})
}

func TestSovereignHeaderInterceptorProcessor_Validate(t *testing.T) {
	t.Parallel()

	extendedHdrHash := []byte("hash")
	wasHashComputed := false

	args := createSovHdrProcArgs()
	args.Hasher = &testscommon.HasherStub{
		ComputeCalled: func(s string) []byte {
			wasHashComputed = true
			return extendedHdrHash
		},
	}

	extendedHeader := &block.ShardHeaderExtended{}
	interceptedHdr := &testscommon.SovereignInterceptedExtendedHeaderDataStub{
		HashCalled: func() []byte {
			return extendedHdrHash
		},
		GetExtendedHeaderCalled: func() data.ShardHeaderExtendedHandler {
			return extendedHeader
		},
	}
	args.IncomingHeaderSubscriber = &sovereign.IncomingHeaderSubscriberStub{
		CreateExtendedHeaderCalled: func(header sovereignData.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
			require.Equal(t, extendedHeader, header)
			return extendedHeader, nil
		},
	}

	sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
	err := sovHdrProc.Validate(interceptedHdr, "")
	require.Nil(t, err)
	require.True(t, wasHashComputed)
}

func TestSovereignHeaderInterceptorProcessor_ValidateErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid intercepted data", func(t *testing.T) {
		t.Parallel()

		args := createSovHdrProcArgs()
		sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
		err := sovHdrProc.Validate(&testscommon.InterceptedDataStub{}, "")
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("black listed header", func(t *testing.T) {
		t.Parallel()

		extendedHdrHash := []byte("hash")
		args := createSovHdrProcArgs()
		args.BlockBlackList = &testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				require.Equal(t, string(extendedHdrHash), key)
				return true
			},
		}

		extendedHeader := &block.ShardHeaderExtended{}
		interceptedHdr := &testscommon.SovereignInterceptedExtendedHeaderDataStub{
			HashCalled: func() []byte {
				return extendedHdrHash
			},
			GetExtendedHeaderCalled: func() data.ShardHeaderExtendedHandler {
				return extendedHeader
			},
		}

		sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
		err := sovHdrProc.Validate(interceptedHdr, "")
		require.ErrorIs(t, err, process.ErrHeaderIsBlackListed)
	})

	t.Run("cannot create extended header", func(t *testing.T) {
		t.Parallel()

		args := createSovHdrProcArgs()
		errCreateHdr := stdErr.New("error creating hdr")
		args.IncomingHeaderSubscriber = &sovereign.IncomingHeaderSubscriberStub{
			CreateExtendedHeaderCalled: func(header sovereignData.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
				return nil, errCreateHdr
			},
		}

		sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
		err := sovHdrProc.Validate(&testscommon.SovereignInterceptedExtendedHeaderDataStub{}, "")
		require.Equal(t, errCreateHdr, err)
	})

	t.Run("cannot compute hash", func(t *testing.T) {
		t.Parallel()

		args := createSovHdrProcArgs()
		errMarshal := stdErr.New("error creating hdr")
		args.Marshaller = &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, errMarshal
			},
		}
		sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
		err := sovHdrProc.Validate(&testscommon.SovereignInterceptedExtendedHeaderDataStub{}, "")
		require.Equal(t, errMarshal, err)
	})

	t.Run("different computed hash", func(t *testing.T) {
		t.Parallel()

		extendedHdrHash := []byte("hash")
		args := createSovHdrProcArgs()
		args.Hasher = &testscommon.HasherStub{
			ComputeCalled: func(s string) []byte {
				return extendedHdrHash
			},
		}

		extendedHeader := &block.ShardHeaderExtended{}
		args.IncomingHeaderSubscriber = &sovereign.IncomingHeaderSubscriberStub{
			CreateExtendedHeaderCalled: func(header sovereignData.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
				require.Equal(t, extendedHeader, header)
				return extendedHeader, nil
			},
		}

		anotherHash := []byte("another hash")
		interceptedHdr := &testscommon.SovereignInterceptedExtendedHeaderDataStub{
			HashCalled: func() []byte {
				return anotherHash
			},
			GetExtendedHeaderCalled: func() data.ShardHeaderExtendedHandler {
				return extendedHeader
			},
		}

		sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
		err := sovHdrProc.Validate(interceptedHdr, "")
		require.ErrorIs(t, err, errors.ErrInvalidReceivedSovereignProof)
		require.True(t, strings.Contains(err.Error(), fmt.Sprintf("computed hash: %s", hex.EncodeToString(extendedHdrHash))))
		require.True(t, strings.Contains(err.Error(), fmt.Sprintf("received hash: %s", hex.EncodeToString(anotherHash))))
	})
}

func TestSovereignHeaderInterceptorProcessor_Save(t *testing.T) {
	t.Parallel()

	t.Run("add missing header", func(t *testing.T) {
		t.Parallel()

		extendedHdrHash := []byte("hash")
		extendedHeader := &block.ShardHeaderExtended{}
		wasHdrAdded := false

		args := createSovHdrProcArgs()
		args.HeadersPool = &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				require.Equal(t, extendedHdrHash, hash)
				return nil, stdErr.New("hdr not found")
			},
		}
		args.IncomingHeaderSubscriber = &sovereign.IncomingHeaderSubscriberStub{
			AddHeaderCalled: func(headerHash []byte, header sovereignData.IncomingHeaderHandler) error {
				require.Equal(t, extendedHdrHash, headerHash)
				require.Equal(t, extendedHeader, header)
				wasHdrAdded = true

				return nil
			},
		}

		interceptedHdr := &testscommon.SovereignInterceptedExtendedHeaderDataStub{
			HashCalled: func() []byte {
				return extendedHdrHash
			},
			GetExtendedHeaderCalled: func() data.ShardHeaderExtendedHandler {
				return extendedHeader
			},
		}
		sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
		err := sovHdrProc.Save(interceptedHdr, "", "")
		require.Nil(t, err)
		require.True(t, wasHdrAdded)
	})

	t.Run("invalid received data", func(t *testing.T) {
		t.Parallel()

		args := createSovHdrProcArgs()
		sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
		err := sovHdrProc.Save(&testscommon.InterceptedDataStub{}, "", "")
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("add already existing header, should skip it", func(t *testing.T) {
		t.Parallel()

		extendedHdrHash := []byte("hash")
		extendedHeader := &block.ShardHeaderExtended{}
		wasHdrAdded := false

		args := createSovHdrProcArgs()
		args.HeadersPool = &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				require.Equal(t, extendedHdrHash, hash)
				return nil, nil
			},
		}
		args.IncomingHeaderSubscriber = &sovereign.IncomingHeaderSubscriberStub{
			AddHeaderCalled: func(headerHash []byte, header sovereignData.IncomingHeaderHandler) error {
				wasHdrAdded = true
				return nil
			},
		}

		interceptedHdr := &testscommon.SovereignInterceptedExtendedHeaderDataStub{
			HashCalled: func() []byte {
				return extendedHdrHash
			},
			GetExtendedHeaderCalled: func() data.ShardHeaderExtendedHandler {
				return extendedHeader
			},
		}
		sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
		err := sovHdrProc.Save(interceptedHdr, "", "")
		require.Nil(t, err)
		require.False(t, wasHdrAdded)
	})
}
