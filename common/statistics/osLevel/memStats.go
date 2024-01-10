package osLevel

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
)

const (
	procFile     = "/proc/self/status"
	vmPeakName   = "VmPeak"
	vmSizeName   = "VmSize"
	vmLckName    = "VmLck"
	vmPinName    = "VmPin"
	vmHWMName    = "VmHWM"
	vmRSSName    = "VmRSS"
	rssAnonName  = "RssAnon"
	rssFileName  = "RssFile"
	rssShmemName = "RssShmem"
	vmDataName   = "VmData"
	vmStkName    = "VmStk"
	vmExeName    = "VmExe"
	vmLibName    = "VmLib"
	vmPTEName    = "VmPTE"
	vmSwapName   = "VmSwap"
	idxValue     = 1
	kbSize       = uint64(1024)
)

// MemStats is the DTO holding the memory status for the current process
type MemStats struct {
	fieldsSet bool
	VmPeak    uint64
	VmSize    uint64
	VmLck     uint64
	VmPin     uint64
	VmHWM     uint64
	VmRSS     uint64
	RssAnon   uint64
	RssFile   uint64
	RssShmem  uint64
	VmData    uint64
	VmStk     uint64
	VmExe     uint64
	VmLib     uint64
	VmPTE     uint64
	VmSwap    uint64
}

// String will convert the values stored in a one line string value
func (memStats *MemStats) String() string {
	return fmt.Sprintf("VmPeak: %s, VmSize: %s, VmLck: %s, VmPin: %s, VmHWM: %s, VmRSS: %s, RssAnon: %s, "+
		"RssFile: %s, RssShmem: %s, VmData: %s, VmStk: %s, VmExe: %s, VmLib: %s, VmPTE: %s, VmSwap: %s",
		core.ConvertBytes(memStats.VmPeak),
		core.ConvertBytes(memStats.VmSize),
		core.ConvertBytes(memStats.VmLck),
		core.ConvertBytes(memStats.VmPin),
		core.ConvertBytes(memStats.VmHWM),
		core.ConvertBytes(memStats.VmRSS),
		core.ConvertBytes(memStats.RssAnon),
		core.ConvertBytes(memStats.RssFile),
		core.ConvertBytes(memStats.RssShmem),
		core.ConvertBytes(memStats.VmData),
		core.ConvertBytes(memStats.VmStk),
		core.ConvertBytes(memStats.VmExe),
		core.ConvertBytes(memStats.VmLib),
		core.ConvertBytes(memStats.VmPTE),
		core.ConvertBytes(memStats.VmSwap),
	)
}

// ReadCurrentMemStats will read the current mem stats at the OS level
func ReadCurrentMemStats() (*MemStats, error) {
	buff, err := os.ReadFile(procFile)
	if err != nil {
		return nil, err
	}

	return parseResponse(string(buff))
}

func parseResponse(buff string) (*MemStats, error) {
	memStats := &MemStats{}
	lines := strings.Split(buff, "\n")
	for _, line := range lines {
		err := parseLine(line, memStats)
		if err != nil {
			return nil, err
		}
	}

	if !memStats.fieldsSet {
		return nil, errUnparsableData
	}

	return memStats, nil
}

func parseLine(line string, memStats *MemStats) error {
	lineComponents := splitComponents(line)
	if len(lineComponents) == 0 {
		return nil
	}

	fieldSet := true
	var err error
	switch lineComponents[0] {
	case vmPeakName:
		memStats.VmPeak, err = parseValue(lineComponents)
	case vmSizeName:
		memStats.VmSize, err = parseValue(lineComponents)
	case vmLckName:
		memStats.VmLck, err = parseValue(lineComponents)
	case vmPinName:
		memStats.VmPin, err = parseValue(lineComponents)
	case vmHWMName:
		memStats.VmHWM, err = parseValue(lineComponents)
	case vmRSSName:
		memStats.VmRSS, err = parseValue(lineComponents)
	case rssAnonName:
		memStats.RssAnon, err = parseValue(lineComponents)
	case rssFileName:
		memStats.RssFile, err = parseValue(lineComponents)
	case rssShmemName:
		memStats.RssShmem, err = parseValue(lineComponents)
	case vmDataName:
		memStats.VmData, err = parseValue(lineComponents)
	case vmStkName:
		memStats.VmStk, err = parseValue(lineComponents)
	case vmExeName:
		memStats.VmExe, err = parseValue(lineComponents)
	case vmLibName:
		memStats.VmLib, err = parseValue(lineComponents)
	case vmPTEName:
		memStats.VmPTE, err = parseValue(lineComponents)
	case vmSwapName:
		memStats.VmSwap, err = parseValue(lineComponents)
	default:
		fieldSet = false
	}

	memStats.fieldsSet = memStats.fieldsSet || fieldSet

	return err
}

func splitComponents(line string) []string {
	lineComponents := strings.Split(line, " ")
	filteredComponents := make([]string, 0)
	for _, component := range lineComponents {
		subcomponents := strings.Split(component, "\t")
		subcomponents = filterSubComponents(subcomponents)

		filteredComponents = append(filteredComponents, subcomponents...)
	}

	return filteredComponents
}

func filterSubComponents(subcomponents []string) []string {
	filteredComponents := make([]string, 0, len(subcomponents))
	for _, subComp := range subcomponents {
		subComp = strings.Trim(subComp, "\r\t: ")
		if len(subComp) > 0 {
			filteredComponents = append(filteredComponents, subComp)
		}
	}

	return filteredComponents
}

func parseValue(lineComponents []string) (uint64, error) {
	if len(lineComponents) != 3 {
		return 0, fmt.Errorf("%w, invalid string for %s", errWrongNumberOfComponents, strings.Join(lineComponents, " "))
	}

	valueString := lineComponents[idxValue]
	val, err := strconv.Atoi(valueString)
	if err != nil {
		return 0, fmt.Errorf("%w, invalid string for %s", err, strings.Join(lineComponents, " "))
	}

	return uint64(val) * kbSize, nil
}
