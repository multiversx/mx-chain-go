package pathmanager

import (
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go/storage"
)

const (
	shardPlaceholder      = "[S]"
	epochPlaceholder      = "[E]"
	identifierPlaceholder = "[I]"
)

// PathManager will handle creation of paths for storers
type PathManager struct {
	pruningPathTemplate string
	staticPathTemplate  string
}

// NewPathManager will return a new instance of PathManager if the provided arguments are fine
func NewPathManager(pruningPathTemplate string, staticPathTemplate string) (*PathManager, error) {
	if len(pruningPathTemplate) == 0 {
		return nil, storage.ErrEmptyPruningPathTemplate
	}
	if !strings.Contains(pruningPathTemplate, epochPlaceholder) ||
		!strings.Contains(pruningPathTemplate, shardPlaceholder) ||
		!strings.Contains(pruningPathTemplate, identifierPlaceholder) {
		return nil, storage.ErrInvalidPruningPathTemplate
	}

	if len(staticPathTemplate) == 0 {
		return nil, storage.ErrEmptyStaticPathTemplate
	}
	if !strings.Contains(staticPathTemplate, shardPlaceholder) ||
		!strings.Contains(staticPathTemplate, identifierPlaceholder) {
		return nil, storage.ErrInvalidStaticPathTemplate
	}

	return &PathManager{
		pruningPathTemplate: pruningPathTemplate,
		staticPathTemplate:  staticPathTemplate,
	}, nil
}

// PathForEpoch will return the new path for a pruning storer
func (pm *PathManager) PathForEpoch(shardId string, epoch uint32, identifier string) string {
	path := pm.pruningPathTemplate
	path = strings.Replace(path, epochPlaceholder, fmt.Sprintf("%d", epoch), 1)
	path = strings.Replace(path, shardPlaceholder, shardId, 1)
	path = strings.Replace(path, identifierPlaceholder, identifier, 1)

	return path
}

// PathForStatic will return the path for a static storer
func (pm *PathManager) PathForStatic(shardId string, identifier string) string {
	path := pm.staticPathTemplate
	path = strings.Replace(path, shardPlaceholder, shardId, 1)
	path = strings.Replace(path, identifierPlaceholder, identifier, 1)

	return path
}

// IsInterfaceNil returns true if there is no value under the interface
func (pm *PathManager) IsInterfaceNil() bool {
	return pm == nil
}
