package storageMarker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisabledStorageMarker_MarkStorerAsSyncedAndActive(t *testing.T) {
	t.Parallel()

	dsm := NewDisabledStorageMarker()
	assert.NotNil(t, dsm)

	dsm.MarkStorerAsSyncedAndActive(nil)
}
