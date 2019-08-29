package machine

import (
	"os"

	"github.com/shirou/gopsutil/process"
)

// GetCurrentProcess returns details about the current process
func GetCurrentProcess() (*process.Process, error) {
	checkPid := os.Getpid()
	ret, err := process.NewProcess(int32(checkPid))
	if err != nil {
		return nil, err
	}

	return ret, nil
}
