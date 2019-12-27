package pruning

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// removeDirectoryIfEmpty will clean the directories after all persisters for one epoch were destroyed
// the structure is this way :
// workspace/db/Epoch_X/Shard_Y/DbName
// we need to remove the last 3 directories if everything is empty
func removeDirectoryIfEmpty(path string) {
	elementsSplitBySeparator := strings.Split(path, string(os.PathSeparator))

	epochDirectory := ""
	for idx := 0; idx < len(elementsSplitBySeparator)-2; idx++ {
		epochDirectory += elementsSplitBySeparator[idx] + string(os.PathSeparator)
	}

	if len(elementsSplitBySeparator) > 2 { // if length is less than 2, the path is incorrect
		shardDirectory := epochDirectory + elementsSplitBySeparator[len(elementsSplitBySeparator)-2]
		if isDirectoryEmpty(shardDirectory) {
			err := os.RemoveAll(shardDirectory)
			if err != nil {
				log.Debug("delete old db directory", "error", err.Error())
			}

			if isDirectoryEmpty(epochDirectory) {
				err = os.RemoveAll(epochDirectory)
				if err != nil {
					log.Debug("delete old db directory", "error", err.Error())
				}
			}
		}
	}
}

func isDirectoryEmpty(name string) bool {
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		return false
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Debug("pruning db - file close", "error", err.Error())
		}
	}()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	return err == io.EOF
}
