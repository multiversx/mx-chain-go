package outport

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/stretchr/testify/require"
)

func TestNewOutport(t *testing.T) {
	t.Parallel()

	outportHandler := NewOutport()

	require.False(t, outportHandler.IsInterfaceNil())
	require.False(t, outportHandler.HasDrivers())
}

func TestOutport_SaveAccounts(t *testing.T) {
	t.Parallel()

	called1 := false
	called2 := false
	driver1 := &mock.DriverStub{
		SaveAccountsCalled: func(blockTimestamp uint64, acc []data.UserAccountHandler) error {
			called1 = true

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveAccountsCalled: func(blockTimestamp uint64, acc []data.UserAccountHandler) error {
			called2 = true

			return nil
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.SaveAccounts(0, []data.UserAccountHandler{})
	require.True(t, called1)
	require.True(t, called2)
	require.Nil(t, err)
}

func TestOutport_SaveAccountsErrorsShouldReturnError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	called1 := false
	called2 := false
	driver1 := &mock.DriverStub{
		SaveAccountsCalled: func(blockTimestamp uint64, acc []data.UserAccountHandler) error {
			called1 = true

			return expectedError
		},
	}
	driver2 := &mock.DriverStub{
		SaveAccountsCalled: func(blockTimestamp uint64, acc []data.UserAccountHandler) error {
			called2 = true

			return nil
		},
	}

	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.SaveAccounts(0, []data.UserAccountHandler{})
	require.Equal(t, expectedError, err)
	require.True(t, called1)
	require.True(t, called2)
}

func TestOutport_SaveBlock(t *testing.T) {
	t.Parallel()

	called := false
	driver1 := &mock.DriverStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) error {
			called = true

			return nil
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)

	err := outportHandler.SaveBlock(&indexer.ArgsSaveBlockData{})
	require.True(t, called)
	require.Nil(t, err)
}

func TestOutport_SaveBlockErrorsShouldReturnError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	called1 := false
	called2 := false
	driver1 := &mock.DriverStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) error {
			called1 = true

			return expectedError
		},
	}
	driver2 := &mock.DriverStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) error {
			called2 = true

			return nil
		},
	}

	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.SaveBlock(&indexer.ArgsSaveBlockData{})
	require.Equal(t, expectedError, err)
	require.True(t, called1)
	require.True(t, called2)
}

func TestOutport_SaveRoundsInfo(t *testing.T) {
	t.Parallel()

	called1 := false
	driver1 := &mock.DriverStub{
		SaveRoundsInfoCalled: func(roundsInfos []*indexer.RoundInfo) error {
			called1 = true

			return nil
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)

	err := outportHandler.SaveRoundsInfo(nil)
	require.True(t, called1)
	require.Nil(t, err)
}

func TestOutport_SaveRoundsInfoErrorsShouldReturnError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	called1 := false
	called2 := false
	driver1 := &mock.DriverStub{
		SaveRoundsInfoCalled: func(roundsInfos []*indexer.RoundInfo) error {
			called1 = true

			return expectedError
		},
	}
	driver2 := &mock.DriverStub{
		SaveRoundsInfoCalled: func(roundsInfos []*indexer.RoundInfo) error {
			called2 = true

			return nil
		},
	}

	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.SaveRoundsInfo(nil)
	require.Equal(t, expectedError, err)
	require.True(t, called1)
	require.True(t, called2)
}

func TestOutport_SaveValidatorsPubKeys(t *testing.T) {
	t.Parallel()

	called := false
	driver := &mock.DriverStub{
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
			called = true

			return nil
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver)

	err := outportHandler.SaveValidatorsPubKeys(nil, 0)
	require.True(t, called)
	require.Nil(t, err)
}

func TestOutport_SaveValidatorsPubKeysErrorsShouldReturnError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	called1 := false
	called2 := false
	driver1 := &mock.DriverStub{
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
			called1 = true

			return expectedError
		},
	}
	driver2 := &mock.DriverStub{
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
			called2 = true

			return nil
		},
	}

	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.SaveValidatorsPubKeys(nil, 0)
	require.Equal(t, expectedError, err)
	require.True(t, called1)
	require.True(t, called2)
}

func TestOutport_SaveValidatorsRating(t *testing.T) {
	t.Parallel()

	called := false
	driver := &mock.DriverStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []*indexer.ValidatorRatingInfo) error {
			called = true

			return nil
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver)

	err := outportHandler.SaveValidatorsRating("", nil)
	require.True(t, called)
	require.Nil(t, err)
}

func TestOutport_SaveValidatorsRatingErrorsShouldReturnError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	called1 := false
	called2 := false
	driver1 := &mock.DriverStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []*indexer.ValidatorRatingInfo) error {
			called1 = true

			return expectedError
		},
	}
	driver2 := &mock.DriverStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []*indexer.ValidatorRatingInfo) error {
			called2 = true

			return nil
		},
	}

	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.SaveValidatorsRating("", nil)
	require.Equal(t, expectedError, err)
	require.True(t, called1)
	require.True(t, called2)
}

func TestOutport_RevertBlock(t *testing.T) {
	t.Parallel()

	called := false
	driver := &mock.DriverStub{
		RevertBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			called = true

			return nil
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver)

	err := outportHandler.RevertIndexedBlock(nil, nil)
	require.True(t, called)
	require.Nil(t, err)
}

func TestOutport_RevertBlockErrorsShouldReturnError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	called1 := false
	called2 := false
	driver1 := &mock.DriverStub{
		RevertBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			called1 = true

			return expectedError
		},
	}
	driver2 := &mock.DriverStub{
		RevertBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			called2 = true

			return nil
		},
	}

	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.RevertIndexedBlock(nil, nil)
	require.Equal(t, expectedError, err)
	require.True(t, called1)
	require.True(t, called2)
}

func TestOutport_FinalizedBlock(t *testing.T) {
	t.Parallel()

	called := false
	driver := &mock.DriverStub{
		FinalizedBlockCalled: func(headerHash []byte) error {
			called = true

			return nil
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver)

	err := outportHandler.FinalizedBlock(nil)
	require.True(t, called)
	require.Nil(t, err)
}

func TestOutport_FinalizedBlockErrorsShouldReturnError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	called1 := false
	called2 := false
	driver1 := &mock.DriverStub{
		FinalizedBlockCalled: func(headerHash []byte) error {
			called1 = true

			return expectedError
		},
	}
	driver2 := &mock.DriverStub{
		FinalizedBlockCalled: func(headerHash []byte) error {
			called2 = true

			return nil
		},
	}

	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.FinalizedBlock(nil)
	require.Equal(t, expectedError, err)
	require.True(t, called1)
	require.True(t, called2)
}

func TestOutport_SubscribeDriver(t *testing.T) {
	t.Parallel()

	outportHandler := NewOutport()

	err := outportHandler.SubscribeDriver(nil)
	require.Equal(t, ErrNilDriver, err)
}

func TestOutport_Close(t *testing.T) {
	t.Parallel()

	outportHandler := NewOutport()

	localErr := errors.New("local err")
	driver1 := &mock.DriverStub{
		CloseCalled: func() error {
			return localErr
		},
	}
	driver2 := &mock.DriverStub{
		CloseCalled: func() error {
			return nil
		},
	}

	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.Close()
	require.Equal(t, localErr, err)
}
