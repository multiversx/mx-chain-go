package trigger

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewImportStartHandler_EmptyStringShouldErr(t *testing.T) {
	t.Parallel()

	workingDir := "working directory"
	version := ""
	ish, err := NewImportStartHandler(workingDir, version)

	require.True(t, check.IfNil(ish))
	assert.Equal(t, update.ErrEmptyVersionString, err)
}

func TestNewImportStartHandler(t *testing.T) {
	t.Parallel()

	workingDir := "working directory"
	version := "v1"
	ish, err := NewImportStartHandler(workingDir, version)

	require.False(t, check.IfNil(ish))
	assert.Nil(t, err)
	assert.Equal(t, workingDir, ish.workingDir)
}

func TestImportStartHandler_SetGetRemove(t *testing.T) {
	t.Parallel()

	version := "v1"
	ish, _ := NewImportStartHandler(t.TempDir(), version)

	err := ish.SetStartImport()
	assert.Nil(t, err)

	ish.SetVersion("v2")
	shouldStart := ish.ShouldStartImport()
	assert.True(t, shouldStart)

	afterImport := ish.IsAfterExportBeforeImport()
	assert.True(t, afterImport)

	err = ish.ResetStartImport()
	assert.Nil(t, err)

	shouldStart = ish.ShouldStartImport()
	assert.False(t, shouldStart)
}

func TestImportStartHandler_DoubleSetShouldNotError(t *testing.T) {
	t.Parallel()

	version := "v1"
	ish, _ := NewImportStartHandler(t.TempDir(), version)

	err := ish.SetStartImport()
	assert.Nil(t, err)
	ish.SetVersion("v2")
	assert.True(t, ish.ShouldStartImport())

	err = ish.SetStartImport()
	assert.Nil(t, err)
	ish.SetVersion("v2")
	assert.False(t, ish.ShouldStartImport())
	ish.SetVersion("v3")
	assert.True(t, ish.ShouldStartImport())
}
