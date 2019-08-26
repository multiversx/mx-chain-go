package machine

import (
	"os"

	"github.com/shirou/gopsutil/process"
)

func getCurrentProcess() (*process.Process, error) {
	checkPid := os.Getpid()
	ret, err := process.NewProcess(int32(checkPid))
	if err != nil {
		return nil, err
	}

	return ret, nil
}
