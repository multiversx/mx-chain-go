package factory

import (
	"fmt"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage/pathmanager"
)

// ArgCreatePathManager is used to pass the strings to the path manager factory function
type ArgCreatePathManager struct {
	WorkingDir string
	ChainID    string
}

// CreatePathManager crates a path manager from provided working directory and chain ID
func CreatePathManager(arg ArgCreatePathManager) (*pathmanager.PathManager, error) {
	return CreatePathManagerFromSinglePathString(filepath.Join(arg.WorkingDir, common.DefaultDBPath, arg.ChainID))
}

// CreatePathManagerFromSinglePathString crates a path manager from provided path string
func CreatePathManagerFromSinglePathString(dbPathWithChainID string) (*pathmanager.PathManager, error) {
	pathTemplateForPruningStorer := filepath.Join(
		dbPathWithChainID,
		fmt.Sprintf("%s_%s", common.DefaultEpochString, common.PathEpochPlaceholder),
		fmt.Sprintf("%s_%s", common.DefaultShardString, common.PathShardPlaceholder),
		common.PathIdentifierPlaceholder)

	pathTemplateForStaticStorer := filepath.Join(
		dbPathWithChainID,
		common.DefaultStaticDbString,
		fmt.Sprintf("%s_%s", common.DefaultShardString, common.PathShardPlaceholder),
		common.PathIdentifierPlaceholder)

	return pathmanager.NewPathManager(pathTemplateForPruningStorer, pathTemplateForStaticStorer, dbPathWithChainID)
}
