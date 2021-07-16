package machine

import (
	"os"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/shirou/gopsutil/process"
)

var log = logger.GetOrCreate("statistics/machine")

// GetCurrentProcess returns details about the current process
func GetCurrentProcess() (*process.Process, error) {
	checkPid := os.Getpid()
	ret, err := process.NewProcess(int32(checkPid))
	if err != nil {
		return nil, err
	}

	return ret, nil
}
