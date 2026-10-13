package grpcdriver

import (
	"context"
	"errors"
	"testing"
	"time"

	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/outport/grpcadapter"
	"github.com/multiversx/mx-chain-go/outport/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewGRPCDriver(t *testing.T) {
	t.Parallel()

	marshaller := &testscommon.MarshallerStub{}
	client := &mock.OutportGRPCClientStub{}

	driver, err := NewGRPCDriver(client, marshaller, time.Second)

	require.NoError(t, err)
	require.NotNil(t, driver)
	require.Same(t, marshaller, driver.GetMarshaller())

	_, err = NewGRPCDriver(nil, marshaller, time.Second)
	require.ErrorIs(t, err, ErrNilGRPCClient)

	_, err = NewGRPCDriver(client, nil, time.Second)
	require.ErrorIs(t, err, ErrNilMarshaller)

	var nilClient *mock.OutportGRPCClientStub
	_, err = NewGRPCDriver(nilClient, marshaller, time.Second)
	require.ErrorIs(t, err, ErrNilGRPCClient)

	driver, err = NewGRPCDriver(client, marshaller, 0)
	require.NoError(t, err)
	require.Equal(t, defaultGRPCCallTimeout, driver.(*grpcDriver).requestTimeout)
}

func TestGrpcDriverDelegatesCalls(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")

	tests := []struct {
		name   string
		client func(t *testing.T) grpcadapter.OutportClient
		call   func(driver *grpcDriver) error
	}{
		{
			name: "SaveBlock",
			client: func(t *testing.T) grpcadapter.OutportClient {
				expected := &outportcore.OutportBlock{ShardID: 1}
				return &mock.OutportGRPCClientStub{
					SaveBlockCalled: func(ctx context.Context, in *outportcore.OutportBlock) error {
						require.Equal(t, expected, in)
						_, hasDeadline := ctx.Deadline()
						require.True(t, hasDeadline)
						return expectedErr
					},
				}
			},
			call: func(driver *grpcDriver) error {
				return driver.SaveBlock(&outportcore.OutportBlock{ShardID: 1})
			},
		},
		{
			name: "RevertIndexedBlock",
			client: func(t *testing.T) grpcadapter.OutportClient {
				expected := &outportcore.BlockData{ShardID: 2}
				return &mock.OutportGRPCClientStub{
					RevertIndexedBlockCalled: func(ctx context.Context, in *outportcore.BlockData) error {
						require.Equal(t, expected, in)
						require.NotNil(t, ctx)
						return expectedErr
					},
				}
			},
			call: func(driver *grpcDriver) error {
				return driver.RevertIndexedBlock(&outportcore.BlockData{ShardID: 2})
			},
		},
		{
			name: "SaveRoundsInfo",
			client: func(t *testing.T) grpcadapter.OutportClient {
				expected := &outportcore.RoundsInfo{}
				return &mock.OutportGRPCClientStub{
					SaveRoundsInfoCalled: func(ctx context.Context, in *outportcore.RoundsInfo) error {
						require.Equal(t, expected, in)
						require.NotNil(t, ctx)
						return expectedErr
					},
				}
			},
			call: func(driver *grpcDriver) error {
				return driver.SaveRoundsInfo(&outportcore.RoundsInfo{})
			},
		},
		{
			name: "SaveValidatorsPubKeys",
			client: func(t *testing.T) grpcadapter.OutportClient {
				expected := &outportcore.ValidatorsPubKeys{ShardID: 3}
				return &mock.OutportGRPCClientStub{
					SaveValidatorsPubKeysCalled: func(ctx context.Context, in *outportcore.ValidatorsPubKeys) error {
						require.Equal(t, expected, in)
						require.NotNil(t, ctx)
						return expectedErr
					},
				}
			},
			call: func(driver *grpcDriver) error {
				return driver.SaveValidatorsPubKeys(&outportcore.ValidatorsPubKeys{ShardID: 3})
			},
		},
		{
			name: "SaveValidatorsRating",
			client: func(t *testing.T) grpcadapter.OutportClient {
				expected := &outportcore.ValidatorsRating{ShardID: 4}
				return &mock.OutportGRPCClientStub{
					SaveValidatorsRatingCalled: func(ctx context.Context, in *outportcore.ValidatorsRating) error {
						require.Equal(t, expected, in)
						require.NotNil(t, ctx)
						return expectedErr
					},
				}
			},
			call: func(driver *grpcDriver) error {
				return driver.SaveValidatorsRating(&outportcore.ValidatorsRating{ShardID: 4})
			},
		},
		{
			name: "SaveAccounts",
			client: func(t *testing.T) grpcadapter.OutportClient {
				expected := &outportcore.Accounts{ShardID: 5}
				return &mock.OutportGRPCClientStub{
					SaveAccountsCalled: func(ctx context.Context, in *outportcore.Accounts) error {
						require.Equal(t, expected, in)
						require.NotNil(t, ctx)
						return expectedErr
					},
				}
			},
			call: func(driver *grpcDriver) error {
				return driver.SaveAccounts(&outportcore.Accounts{ShardID: 5})
			},
		},
		{
			name: "FinalizedBlock",
			client: func(t *testing.T) grpcadapter.OutportClient {
				expected := &outportcore.FinalizedBlock{ShardID: 6}
				return &mock.OutportGRPCClientStub{
					FinalizedBlockEventCalled: func(ctx context.Context, in *outportcore.FinalizedBlock) error {
						require.Equal(t, expected, in)
						require.NotNil(t, ctx)
						return expectedErr
					},
				}
			},
			call: func(driver *grpcDriver) error {
				return driver.FinalizedBlock(&outportcore.FinalizedBlock{ShardID: 6})
			},
		},
	}

	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			driverAsInterface, err := NewGRPCDriver(tt.client(t), &testscommon.MarshallerStub{}, time.Second)
			require.NoError(t, err)
			driver := driverAsInterface.(*grpcDriver)

			err = tt.call(driver)

			require.ErrorIs(t, err, expectedErr)
		})
	}
}

func TestGrpcDriverNoOpMethods(t *testing.T) {
	t.Parallel()

	closeCalled := false
	driverAsInterface, err := NewGRPCDriver(&mock.OutportGRPCClientStub{
		CloseCalled: func() error {
			closeCalled = true
			return nil
		},
	}, &testscommon.MarshallerStub{}, time.Second)
	require.NoError(t, err)
	driver := driverAsInterface.(*grpcDriver)

	require.NoError(t, driver.SetCurrentSettings(outportcore.OutportConfig{}))
	require.NoError(t, driver.RegisterHandler(nil, ""))
	require.NoError(t, driver.Close())
	require.True(t, closeCalled)
}

func TestGrpcDriverIsInterfaceNil(t *testing.T) {
	t.Parallel()

	var nilDriver *grpcDriver
	require.True(t, nilDriver.IsInterfaceNil())

	driver := &grpcDriver{}
	require.False(t, driver.IsInterfaceNil())
}
