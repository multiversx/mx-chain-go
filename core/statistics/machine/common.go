package machine

import (
	"os"
	"time"

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

func CalculateCpuLoad() float64 {
	currentProcess, err := GetCurrentProcess()
	if err != nil {
		return 0
	}

	percent, err := currentProcess.Percent(100 * time.Millisecond)
	if err != nil {
		return 0
	}

	return percent
}
