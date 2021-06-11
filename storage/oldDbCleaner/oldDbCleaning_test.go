package oldDbCleaner

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/stretchr/testify/require"
)

func TestCleanupOldStorageForShuffleOut_ShouldExitIfDontCleanOldData(t *testing.T) {
	prunStorerCfg := config.StoragePruningConfig{
		CleanOldEpochsData: false,
	}

	err := CleanOldStorageForShuffleOut(
		Args{
			PruningStorageConfig: prunStorerCfg,
		},
	)
	require.NoError(t, err)
}

func TestCleanupOldStorageForShuffleOut_EmptyDirectoriesList(t *testing.T) {
	prunStorerCfg := config.StoragePruningConfig{
		CleanOldEpochsData: true,
	}
	dirReader := &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(_ string) ([]string, error) {
			return []string{}, nil
		},
	}

	err := CleanOldStorageForShuffleOut(
		Args{
			PruningStorageConfig: prunStorerCfg,
			DirectoryReader:      dirReader,
		},
	)
	require.NoError(t, err)
}

func TestCleanupOldStorageForShuffleOut_ShouldUseDefaultDirReaderAndFileRemover(t *testing.T) {
	prunStorerCfg := config.StoragePruningConfig{
		CleanOldEpochsData: true,
	}

	err := CleanOldStorageForShuffleOut(
		Args{
			PruningStorageConfig: prunStorerCfg,
		},
	)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "no such file"))
}

func TestCleanupOldStorageForShuffleOut_ShouldNotErrWithSomeWrongDirectoryNames(t *testing.T) {
	dirReader := &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(_ string) ([]string, error) {
			return []string{"Epoch+0", "Epoch_3", "Epoch+1", "Epoch_2", "Epoch+4", "Epoch_5"}, nil
		},
	}

	removedFiles := make([]string, 0)
	fileRemover := func(file string) error {
		removedFiles = append(removedFiles, file)
		return nil
	}

	prunStorerCfg := config.StoragePruningConfig{
		CleanOldEpochsData: true,
		NumEpochsToKeep:    2,
	}

	err := CleanOldStorageForShuffleOut(
		Args{
			DirectoryReader:      dirReader,
			FileRemover:          fileRemover,
			PruningStorageConfig: prunStorerCfg,
			CurrentEpoch:         6,
			WorkingDir:           "workDir",
			ChainID:              "T",
		},
	)

	require.NoError(t, err)

	expectedRes := []string{
		"workDir/db/T/Epoch_2",
		"workDir/db/T/Epoch_3",
	}

	require.Equal(t, expectedRes, removedFiles)
}

func TestCleanupOldStorageForShuffleOut_ShouldNotErrWithAllDirectoryNamesWrong(t *testing.T) {
	dirReader := &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(_ string) ([]string, error) {
			return []string{"Epoch+0", "Epoch+3", "Epoch+1", "Epoch+2", "Epoch+4", "Epoch+5"}, nil
		},
	}

	removedFiles := make([]string, 0)
	fileRemover := func(file string) error {
		removedFiles = append(removedFiles, file)
		return nil
	}

	prunStorerCfg := config.StoragePruningConfig{
		CleanOldEpochsData: true,
		NumEpochsToKeep:    2,
	}

	err := CleanOldStorageForShuffleOut(
		Args{
			DirectoryReader:      dirReader,
			FileRemover:          fileRemover,
			PruningStorageConfig: prunStorerCfg,
			CurrentEpoch:         6,
			WorkingDir:           "workDir",
			ChainID:              "T",
		},
	)

	require.NoError(t, err)

	expectedRes := make([]string, 0)

	require.Equal(t, expectedRes, removedFiles)
}

func TestCleanupOldStorageForShuffleOut_ShouldWork(t *testing.T) {
	dirReader := &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(_ string) ([]string, error) {
			return []string{"Epoch_0", "Epoch_3", "Epoch_1", "Epoch_2", "Epoch_4", "Epoch_5"}, nil
		},
	}

	removedFiles := make([]string, 0)
	fileRemover := func(file string) error {
		removedFiles = append(removedFiles, file)
		return nil
	}

	prunStorerCfg := config.StoragePruningConfig{
		CleanOldEpochsData: true,
		NumEpochsToKeep:    2,
	}

	closeHandler := func() error {
		return CleanOldStorageForShuffleOut(
			Args{
				DirectoryReader:      dirReader,
				FileRemover:          fileRemover,
				PruningStorageConfig: prunStorerCfg,
				CurrentEpoch:         6,
				WorkingDir:           "workDir",
				ChainID:              "T",
			},
		)
	}
	err := closeHandler()
	require.NoError(t, err)

	expectedRes := []string{
		"workDir/db/T/Epoch_0",
		"workDir/db/T/Epoch_1",
		"workDir/db/T/Epoch_2",
		"workDir/db/T/Epoch_3",
	}

	require.Equal(t, expectedRes, removedFiles)
}
