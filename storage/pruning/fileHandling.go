package pruning

import (
	"io"
	"os"
	"strings"
)

func removeDirectoryIfEmpty(path string) {
	elementsSplitBySeparator := strings.Split(path, string(os.PathSeparator))
	// the structure is this way :
	// workspace/db/Epoch_X/Shard_Y/DbName
	// we need to remove the last 3 if everything is empty

	epochDirectory := ""
	for idx := 0; idx < len(elementsSplitBySeparator)-2; idx++ {
		epochDirectory += elementsSplitBySeparator[idx] + string(os.PathSeparator)
	}

	shardDirectory := epochDirectory + elementsSplitBySeparator[len(elementsSplitBySeparator)-2]
	if isDirectoryEmpty(shardDirectory) {
		err := os.RemoveAll(shardDirectory)
		if err != nil {
			log.Debug("delete old db directory", "error", err.Error())
		}

		err = os.RemoveAll(epochDirectory)
		if err != nil {
			log.Debug("delete old db directory", "error", err.Error())
		}
	}
}

func isDirectoryEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Debug("pruning db - file close", "error", err.Error())
		}
	}()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	return err == io.EOF
}
