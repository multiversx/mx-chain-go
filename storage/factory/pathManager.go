package factory

import (
	"fmt"
	"path/filepath"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/pathmanager"
)

// ArgCreatePathManager is used to pass the strings to the path manager factory function
type ArgCreatePathManager struct {
	WorkingDir string
	ChainID    string
}

// CreatePathManager crates a path manager from provided working directory and chain ID
func CreatePathManager(arg ArgCreatePathManager) (*pathmanager.PathManager, error) {
	return CreatePathManagerFromSinglePathString(filepath.Join(arg.WorkingDir, storage.DefaultDBPath, arg.ChainID))
}

// CreatePathManagerFromSinglePathString crates a path manager from provided path string
func CreatePathManagerFromSinglePathString(dbPathWithChainID string) (*pathmanager.PathManager, error) {
	pathTemplateForPruningStorer := filepath.Join(
		dbPathWithChainID,
		fmt.Sprintf("%s_%s", storage.DefaultEpochString, storage.PathEpochPlaceholder),
		fmt.Sprintf("%s_%s", storage.DefaultShardString, storage.PathShardPlaceholder),
		storage.PathIdentifierPlaceholder)

	pathTemplateForStaticStorer := filepath.Join(
		dbPathWithChainID,
		storage.DefaultStaticDbString,
		fmt.Sprintf("%s_%s", storage.DefaultShardString, storage.PathShardPlaceholder),
		storage.PathIdentifierPlaceholder)

	return pathmanager.NewPathManager(pathTemplateForPruningStorer, pathTemplateForStaticStorer, dbPathWithChainID)
}
