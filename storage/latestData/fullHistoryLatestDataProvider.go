package latestData

import "github.com/ElrondNetwork/elrond-go/storage"

type fullHistoryLatestDataProvider struct {
	*latestDataProvider
}

// NewFullHistoryLatestDataProvider returns a new instance of fullHistoryLatestDataProvider
func NewFullHistoryLatestDataProvider(args ArgsLatestDataProvider) (*fullHistoryLatestDataProvider, error) {
	return &fullHistoryLatestDataProvider{
		&latestDataProvider{
			generalConfig:         args.GeneralConfig,
			parentDir:             args.ParentDir,
			directoryReader:       args.DirectoryReader,
			defaultShardString:    args.DefaultShardString,
			defaultEpochString:    args.DefaultEpochString,
			bootstrapDataProvider: args.BootstrapDataProvider,
		}}, nil
}

// Get will return a struct containing the latest usable data for the full archive node in storage
func (ldp *fullHistoryLatestDataProvider) Get() (storage.LatestDataFromStorage, error) {
	epochDirs, err := ldp.getEpochDirs()
	if err != nil {
		return storage.LatestDataFromStorage{}, err
	}

	for index := range epochDirs {
		parentDir, lastEpoch, errGetDir := ldp.getParentDirAndLastEpochWithIndex(index)
		if errGetDir != nil {
			err = errGetDir
		}

		latestDataFromStorage, errGetEpoch := ldp.getLastEpochAndRoundFromStorage(parentDir, lastEpoch)
		if errGetEpoch == nil {
			return latestDataFromStorage, nil
		}
	}

	return storage.LatestDataFromStorage{}, err
}

// GetParentDirAndLastEpoch returns the parent directory and last usable epoch for the full archive node
func (ldp *fullHistoryLatestDataProvider) GetParentDirAndLastEpoch() (string, uint32, error) {
	epochDirs, err := ldp.getEpochDirs()
	if err != nil {
		return "", 0, err
	}

	for index := range epochDirs {
		parentDir, lastEpoch, errGetDir := ldp.getParentDirAndLastEpochWithIndex(index)
		if errGetDir != nil {
			continue
		}

		_, errGetEpoch := ldp.getLastEpochAndRoundFromStorage(parentDir, lastEpoch)
		if errGetEpoch != nil {
			err = errGetEpoch
			continue
		}
		return parentDir, lastEpoch, nil
	}

	return "", 0, err
}

func (ldp *fullHistoryLatestDataProvider) getParentDirAndLastEpochWithIndex(index int) (string, uint32, error) {
	epochDirs, err := ldp.getEpochDirs()
	if err != nil {
		return "", 0, err
	}

	lastEpoch, err := ldp.GetLastEpochFromDirNames(epochDirs, index)
	if err != nil {
		return "", 0, err
	}

	return ldp.parentDir, lastEpoch, nil

}
