package trigger

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/update"
)

const filename = "mustimport"

type importStartHandler struct {
	workingDir     string
	currentVersion string
}

// NewImportStartHandler returns a new importStartHandler instance able to control
// the startup of the import process after a hardfork event
func NewImportStartHandler(workingDir string, version string) (*importStartHandler, error) {
	if len(version) == 0 {
		return nil, update.ErrEmptyVersionString
	}
	if !strings.Contains(version, ".") {
		log.Warn("version string is not in the correct format. Hardfork trigger will lead to unpredictable results")
	}

	ish := &importStartHandler{
		workingDir:     workingDir,
		currentVersion: version,
	}

	return ish, nil
}

func (ish *importStartHandler) getFilename() string {
	fullFilename := filepath.Join(ish.workingDir, filename)
	return filepath.Clean(fullFilename)
}

func (ish *importStartHandler) loadFileStatus() (string, bool) {
	info, err := os.Stat(ish.getFilename())
	if os.IsNotExist(err) {
		return "", false
	}
	if info.IsDir() {
		return "", false
	}

	contents, err := os.ReadFile(ish.getFilename())
	if err != nil {
		log.Error("error reading must import file",
			"file", ish.getFilename(),
			"error", err,
		)
		return "", false
	}

	return string(contents), true
}

// ShouldStartImport returns true if the import process should start
func (ish *importStartHandler) ShouldStartImport() bool {
	fileVersion, fileExists := ish.loadFileStatus()
	isDifferentVersion := fileVersion != ish.currentVersion

	return isDifferentVersion && fileExists
}

// IsAfterExportBeforeImport returns true if the export has been done but import should not occur because the new version
// is not published yet
func (ish *importStartHandler) IsAfterExportBeforeImport() bool {
	_, fileExists := ish.loadFileStatus()

	return fileExists
}

// ResetStartImport removes the file that will trigger a new import start event
func (ish *importStartHandler) ResetStartImport() error {
	log.Debug("removing must import file", "file", ish.getFilename())
	return os.Remove(ish.getFilename())
}

// SetStartImport creates the file that will trigger a new import start event
func (ish *importStartHandler) SetStartImport() error {
	if ish.ShouldStartImport() {
		err := os.Remove(ish.getFilename())
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(
		filepath.Join(ish.getFilename()),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		core.FileModeReadWrite,
	)
	if err != nil {
		return err
	}

	_, err = file.WriteString(ish.currentVersion)
	if err != nil {
		return err
	}

	return file.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ish *importStartHandler) IsInterfaceNil() bool {
	return ish == nil
}
