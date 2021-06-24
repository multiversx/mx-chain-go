package olddbclean

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewOldDatabaseCleaner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description string
		getArgs     func() Args
		expectedErr error
	}{
		{
			description: "nil epoch start notifier",
			getArgs: func() Args {
				args := createMockArgs()
				args.EpochStartNotifier = nil

				return args
			},
			expectedErr: storage.ErrNilEpochStartNotifier,
		},
		{
			description: "nil storage list provider",
			getArgs: func() Args {
				args := createMockArgs()
				args.StorageListProvider = nil

				return args
			},
			expectedErr: storage.ErrNilStorageListProvider,
		},
		{
			description: "should work",
			getArgs: func() Args {
				return createMockArgs()
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t1 *testing.T) {
			args := tt.getArgs()
			odc, err := NewOldDatabaseCleaner(args)
			require.Equal(t1, tt.expectedErr, err)
			if err == nil {
				require.False(t1, check.IfNil(odc))
			} else {
				require.True(t1, check.IfNil(odc))
			}
		})
	}
}

func TestOldDatabaseCleaner_EpochChangeRemoveFileIsCalled(t *testing.T) {
	t.Parallel()

	var handlerFunc epochStart.ActionHandler
	args := createMockArgs()
	args.EpochStartNotifier = &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			handlerFunc = handler
		},
	}

	directoryReader := &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(_ string) ([]string, error) {
			return []string{"Epoch_0", "Epoch_1"}, nil
		},
	}
	fileRemoverWasCalled := false
	fileRemover := func(file string) error {
		fileRemoverWasCalled = true
		return nil
	}
	odc, _ := NewOldDatabaseCleaner(args)
	odc.fileRemover = fileRemover
	odc.directoryReader = directoryReader
	require.False(t, check.IfNil(odc))

	handlerFunc.EpochStartAction(&block.Header{Epoch: 5})
	require.False(t, fileRemoverWasCalled)
	handlerFunc.EpochStartAction(&block.Header{Epoch: 6})
	require.True(t, fileRemoverWasCalled)
}

func TestOldDatabaseCleaner_EpochChangeDirectoryReadFailsShouldNotRemove(t *testing.T) {
	t.Parallel()

	var handlerFunc epochStart.ActionHandler
	args := createMockArgs()
	args.EpochStartNotifier = &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			handlerFunc = handler
		},
	}

	directoryReader := &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(_ string) ([]string, error) {
			return nil, errors.New("local err")
		},
	}
	fileRemoverWasCalled := false
	fileRemover := func(file string) error {
		fileRemoverWasCalled = true
		return nil
	}
	odc, _ := NewOldDatabaseCleaner(args)
	odc.fileRemover = fileRemover
	odc.directoryReader = directoryReader
	require.False(t, check.IfNil(odc))

	handlerFunc.EpochStartAction(&block.Header{Epoch: 5})
	require.False(t, fileRemoverWasCalled)
	handlerFunc.EpochStartAction(&block.Header{Epoch: 6})
	require.False(t, fileRemoverWasCalled)
}

func TestOldDatabaseCleaner_EpochChangeNoEpochDirectory(t *testing.T) {
	t.Parallel()

	var handlerFunc epochStart.ActionHandler
	args := createMockArgs()
	args.EpochStartNotifier = &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			handlerFunc = handler
		},
	}

	directoryReader := &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(_ string) ([]string, error) {
			return []string{}, nil
		},
	}
	fileRemoverWasCalled := false
	fileRemover := func(file string) error {
		fileRemoverWasCalled = true
		return nil
	}
	odc, _ := NewOldDatabaseCleaner(args)
	odc.fileRemover = fileRemover
	odc.directoryReader = directoryReader
	require.False(t, check.IfNil(odc))

	handlerFunc.EpochStartAction(&block.Header{Epoch: 5})
	require.False(t, fileRemoverWasCalled)
	handlerFunc.EpochStartAction(&block.Header{Epoch: 6})
	require.False(t, fileRemoverWasCalled)
}

func TestOldDatabaseCleaner_EpochChangeShouldNotRemoveIfNewOldestEpochIsOlder(t *testing.T) {
	t.Parallel()

	var handlerFunc epochStart.ActionHandler
	args := createMockArgs()
	args.EpochStartNotifier = &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			handlerFunc = handler
		},
	}

	directoryReader := &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(_ string) ([]string, error) {
			return []string{"Epoch_0", "Epoch_1", "Epoch_2", "Epoch_3"}, nil
		},
	}

	removedFiles := make([]string, 0)
	fileRemover := func(file string) error {
		removedFiles = append(removedFiles, file)
		return nil
	}

	args.StorageListProvider = getStorageListProviderWithOldEpoch(4)
	odc, _ := NewOldDatabaseCleaner(args)
	odc.fileRemover = fileRemover
	odc.directoryReader = directoryReader
	require.False(t, check.IfNil(odc))

	handlerFunc.EpochStartAction(&block.Header{Epoch: 5})
	require.Empty(t, removedFiles)

	odc.storageListProvider = getStorageListProviderWithOldEpoch(3)
	handlerFunc.EpochStartAction(&block.Header{Epoch: 6})
	require.Equal(t,
		[]string{},
		removedFiles,
	)
}

func TestOldDatabaseCleaner_EpochChange(t *testing.T) {
	t.Parallel()

	var handlerFunc epochStart.ActionHandler
	args := createMockArgs()
	args.EpochStartNotifier = &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			handlerFunc = handler
		},
	}

	directoryReader := &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(_ string) ([]string, error) {
			return []string{"Epoch_0", "Epoch_1", "WrongDir", "Epoch_XYZ", "Epoch_2", "Epoch_3"}, nil
		},
	}

	removedFiles := make([]string, 0)
	fileRemover := func(file string) error {
		removedFiles = append(removedFiles, file)
		return nil
	}

	args.StorageListProvider = getStorageListProviderWithOldEpoch(2)
	odc, _ := NewOldDatabaseCleaner(args)
	odc.fileRemover = fileRemover
	odc.directoryReader = directoryReader
	require.False(t, check.IfNil(odc))

	handlerFunc.EpochStartAction(&block.Header{Epoch: 5})
	require.Empty(t, removedFiles)
	handlerFunc.EpochStartAction(&block.Header{Epoch: 6})
	require.Equal(t,
		[]string{"db/D/Epoch_0", "db/D/Epoch_1"},
		removedFiles,
	)
}

func getStorageListProviderWithOldEpoch(epoch uint32) StorageListProviderHandler {
	return &mock.StorageListProviderStub{
		GetAllStorersCalled: func() map[dataRetriever.UnitType]storage.Storer {
			return map[dataRetriever.UnitType]storage.Storer{
				dataRetriever.TransactionUnit: &testscommon.StorerStub{
					GetOldestEpochCalled: func() (uint32, error) {
						return epoch + 1, nil
					},
				},
				dataRetriever.BlockHeaderUnit: &testscommon.StorerStub{
					GetOldestEpochCalled: func() (uint32, error) {
						return epoch, nil
					},
				},
			}
		},
	}
}

func createMockArgs() Args {
	return Args{
		DatabasePath:        "db/D",
		StorageListProvider: &mock.StorageListProviderStub{},
		EpochStartNotifier: &mock.EpochStartNotifierStub{
			RegisterHandlerCalled: func(_ epochStart.ActionHandler) {},
		},
	}
}
