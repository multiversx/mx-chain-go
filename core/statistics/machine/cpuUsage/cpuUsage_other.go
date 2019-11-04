package cpuUsage

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func getCPUSample() (uint64, uint64, uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, 0
	}

	lines := strings.Split(string(contents), "\n")
	var totalTime uint64
	if len(lines) == 0 {
		return 0, 0, 0
	}

	cpuData := strings.Fields(lines[0])
	for _, data := range cpuData {
		value, err := strconv.ParseUint(data, 10, 64)
		if err != nil {
			continue
		}
		totalTime += value
	}

	checkPid := os.Getpid()
	command := "/proc/" + fmt.Sprintf("%d", checkPid) + "/stat"
	contents, err = ioutil.ReadFile(command)
	if err != nil {
		return 0, 0, 0
	}

	lines = strings.Split(string(contents), "\n")
	if len(lines) == 0 {
		return 0, 0, 0
	}

	processData := strings.Fields(lines[0])

	uTime, _ := strconv.ParseUint(processData[13], 10, 64)
	sTime, _ := strconv.ParseUint(processData[14], 10, 64)

	return totalTime, uTime, sTime
}

// GetCpuUsage will return Cpu usage of current process in percentage
func GetCpuUsage() float64 {
	totalTimeBefore, uTimeBefore, sTimeBefore := getCPUSample()
	time.Sleep(time.Second)
	totalTimeAfter, uTimeAfter, sTimeAfter := getCPUSample()

	userUtil := 100 * float64(uTimeAfter-uTimeBefore) / float64(totalTimeAfter-totalTimeBefore)
	sysUtil := 100 * float64(sTimeAfter-sTimeBefore) / float64(totalTimeAfter-totalTimeBefore)

	return userUtil + sysUtil
}
