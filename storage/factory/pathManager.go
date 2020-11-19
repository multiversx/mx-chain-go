package factory

import (
	"fmt"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/storage/pathmanager"
)

// ArgCreatePathManager is used to pass the strings to the path manager factory function
type ArgCreatePathManager struct {
	WorkingDir string
	ChainID    string
}

// CreatePathManager crates a path manager from provided working directory and chain ID
func CreatePathManager(arg ArgCreatePathManager) (*pathmanager.PathManager, error) {
	return CreatePathManagerFromSinglePathString(filepath.Join(arg.WorkingDir, core.DefaultDBPath, arg.ChainID))
}

// CreatePathManagerFromSinglePathString crates a path manager from provided path string
func CreatePathManagerFromSinglePathString(dbPathWithChainID string) (*pathmanager.PathManager, error) {
	pathTemplateForPruningStorer := filepath.Join(
		dbPathWithChainID,
		fmt.Sprintf("%s_%s", core.DefaultEpochString, core.PathEpochPlaceholder),
		fmt.Sprintf("%s_%s", core.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	pathTemplateForStaticStorer := filepath.Join(
		dbPathWithChainID,
		core.DefaultStaticDbString,
		fmt.Sprintf("%s_%s", core.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	return pathmanager.NewPathManager(pathTemplateForPruningStorer, pathTemplateForStaticStorer)
}
