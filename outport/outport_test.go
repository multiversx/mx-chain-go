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
		SaveAccountsCalled: func(blockTimestamp uint64, acc []data.UserAccountHandler) {
			called1 = true
		},
	}
	driver2 := &mock.DriverStub{
		SaveAccountsCalled: func(blockTimestamp uint64, acc []data.UserAccountHandler) {
			called2 = true
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.SaveAccounts(0, []data.UserAccountHandler{})
	require.True(t, called1)
	require.True(t, called2)
}

func TestOutport_SaveBlock(t *testing.T) {
	t.Parallel()

	called := false
	driver1 := &mock.DriverStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) {
			called = true
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)

	outportHandler.SaveBlock(&indexer.ArgsSaveBlockData{})
	require.True(t, called)
}

func TestOutport_SaveRoundsInfo(t *testing.T) {
	t.Parallel()

	called1 := false
	driver1 := &mock.DriverStub{
		SaveRoundsInfoCalled: func(roundsInfos []*indexer.RoundInfo) {
			called1 = true
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver1)

	outportHandler.SaveRoundsInfo(nil)
	require.True(t, called1)
}

func TestOutport_SaveValidatorsPubKeys(t *testing.T) {
	t.Parallel()

	called := false
	driver := &mock.DriverStub{
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
			called = true
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver)

	outportHandler.SaveValidatorsPubKeys(nil, 0)
	require.True(t, called)
}

func TestOutport_SaveValidatorsRating(t *testing.T) {
	t.Parallel()

	called := false
	driver := &mock.DriverStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []*indexer.ValidatorRatingInfo) {
			called = true
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver)

	outportHandler.SaveValidatorsRating("", nil)
	require.True(t, called)
}

func TestOutport_RevertBlock(t *testing.T) {
	t.Parallel()

	called := false
	driver := &mock.DriverStub{
		RevertBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) {
			called = true
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver)

	outportHandler.RevertIndexedBlock(nil, nil)
	require.True(t, called)
}

func TestOutport_FinalizedBlock(t *testing.T) {
	t.Parallel()

	called := false
	driver := &mock.DriverStub{
		FinalizedBlockCalled: func(headerHash []byte) {
			called = true
		},
	}
	outportHandler := NewOutport()
	_ = outportHandler.SubscribeDriver(driver)

	outportHandler.FinalizedBlock(nil)
	require.True(t, called)
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
