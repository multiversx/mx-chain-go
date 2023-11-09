package pathmanager

import (
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-go/storage"
)

var _ storage.PathManagerHandler = (*PathManager)(nil)

// PathManager will handle creation of paths for storers
type PathManager struct {
	databasePath        string
	pruningPathTemplate string
	staticPathTemplate  string
}

// NewPathManager will return a new instance of PathManager if the provided arguments are fine
func NewPathManager(pruningPathTemplate string, staticPathTemplate string, databasePath string) (*PathManager, error) {
	if len(pruningPathTemplate) == 0 {
		return nil, ErrEmptyPruningPathTemplate
	}
	if !strings.Contains(pruningPathTemplate, storage.PathEpochPlaceholder) ||
		!strings.Contains(pruningPathTemplate, storage.PathShardPlaceholder) ||
		!strings.Contains(pruningPathTemplate, storage.PathIdentifierPlaceholder) {
		return nil, ErrInvalidPruningPathTemplate
	}

	if len(staticPathTemplate) == 0 {
		return nil, ErrEmptyStaticPathTemplate
	}
	if !strings.Contains(staticPathTemplate, storage.PathShardPlaceholder) ||
		!strings.Contains(staticPathTemplate, storage.PathIdentifierPlaceholder) {
		return nil, ErrInvalidStaticPathTemplate
	}

	if len(databasePath) == 0 {
		return nil, ErrInvalidDatabasePath
	}

	return &PathManager{
		pruningPathTemplate: pruningPathTemplate,
		staticPathTemplate:  staticPathTemplate,
		databasePath:        databasePath,
	}, nil
}

// PathForEpoch will return the new path for a pruning storer
func (pm *PathManager) PathForEpoch(shardId string, epoch uint32, identifier string) string {
	path := pm.pruningPathTemplate
	path = strings.Replace(path, storage.PathEpochPlaceholder, fmt.Sprintf("%d", epoch), 1)
	path = strings.Replace(path, storage.PathShardPlaceholder, shardId, 1)
	path = strings.Replace(path, storage.PathIdentifierPlaceholder, identifier, 1)

	return path
}

// PathForStatic will return the path for a static storer
func (pm *PathManager) PathForStatic(shardId string, identifier string) string {
	path := pm.staticPathTemplate
	path = strings.Replace(path, storage.PathShardPlaceholder, shardId, 1)
	path = strings.Replace(path, storage.PathIdentifierPlaceholder, identifier, 1)

	return path
}

// DatabasePath returns the path for the databases directory
func (pm *PathManager) DatabasePath() string {
	return pm.databasePath
}

// IsInterfaceNil returns true if there is no value under the interface
func (pm *PathManager) IsInterfaceNil() bool {
	return pm == nil
}
