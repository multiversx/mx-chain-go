package hostParameters

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const versionMarker = "App version"
const modelMarker = "CPU model"
const numLogicalMarker = "CPU numLogical"
const maxFreqMarker = "CPU maxFreq"
const freqMarker = "MHz"
const flagsMarker = "CPU flags"
const maxFlagsCharsPerLine = 80
const sizeMarker = "Memory size"

type cpuInfo struct {
	numLogicalCPUs    int
	modelName         string
	maxFrequencyInMHz int
	flags             []string
}

func (ci *cpuInfo) toDisplayLines() []*display.LineData {
	lines := []*display.LineData{
		display.NewLineData(false, []string{modelMarker, ci.modelName}),
		display.NewLineData(false, []string{numLogicalMarker, fmt.Sprintf("%d", ci.numLogicalCPUs)}),
	}
	accumulator := ""
	flagMarkerWritten := false
	flagMarkerToBeWritten := ""
	for _, flag := range ci.flags {
		flagMarkerToBeWritten = ""
		accumulator = strings.Join([]string{accumulator, flag}, " ")
		if len(accumulator) > maxFlagsCharsPerLine {
			if !flagMarkerWritten {
				flagMarkerWritten = true
				flagMarkerToBeWritten = flagsMarker
			}

			lines = append(lines, display.NewLineData(false, []string{flagMarkerToBeWritten, accumulator}))
			accumulator = ""
		}
	}

	flagMarkerToBeWritten = ""
	if len(accumulator) > 0 {
		if !flagMarkerWritten {
			flagMarkerToBeWritten = flagsMarker
		}
		lines = append(lines, display.NewLineData(false, []string{flagMarkerToBeWritten, accumulator}))
	}

	lines = append(lines, display.NewLineData(true, []string{maxFreqMarker, fmt.Sprintf("%d %s", ci.maxFrequencyInMHz, freqMarker)}))

	return lines
}

type memInfo struct {
	size string
}

func (mi *memInfo) toDisplayLines() []*display.LineData {
	lines := []*display.LineData{
		display.NewLineData(true, []string{sizeMarker, mi.size}),
	}

	return lines
}

type hostParametersGetter struct {
	versionString string
}

// NewHostParameterGetter will create a structure that is able to get and format the host's relevant parameters
func NewHostParameterGetter(version string) *hostParametersGetter {
	return &hostParametersGetter{
		versionString: version,
	}
}

// GetParameterStringTable is able to get and format in a table-like string all the known parameters of a host
func (hpg *hostParametersGetter) GetParameterStringTable() string {
	ci := hpg.getCpuInfo()
	mi := hpg.getMemInfo()

	lines := make([]*display.LineData, 0)
	lines = append(lines, display.NewLineData(true, []string{versionMarker, hpg.versionString}))
	lines = append(lines, ci.toDisplayLines()...)
	lines = append(lines, mi.toDisplayLines()...)

	hdr := []string{"Parameter", "Value"}

	tbl, err := display.CreateTableString(hdr, lines)
	if err != nil {
		return fmt.Sprintf("[ERR:%s]", err)
	}

	return tbl
}

func (hpg *hostParametersGetter) getCpuInfo() cpuInfo {
	ci := cpuInfo{}
	rawCpuInfo, err := cpu.Info()
	if err != nil {
		ci.modelName = fmt.Sprintf("[ERR:%s]", err)
		return ci
	}

	if len(rawCpuInfo) == 0 {
		ci.modelName = "[ERR:no logical cpus]"
		return ci
	}

	ci.numLogicalCPUs = len(rawCpuInfo)
	ci.modelName = rawCpuInfo[0].ModelName
	ci.maxFrequencyInMHz = int(rawCpuInfo[0].Mhz)
	ci.flags = rawCpuInfo[0].Flags
	sort.Slice(ci.flags, func(i, j int) bool {
		return strings.Compare(ci.flags[i], ci.flags[j]) < 0
	})

	return ci
}

func (hpg *hostParametersGetter) getMemInfo() memInfo {
	mi := memInfo{}
	vms, err := mem.VirtualMemory()
	if err != nil {
		mi.size = fmt.Sprintf("[ERR:%s]", err)
	}

	mi.size = core.ConvertBytes(vms.Total)

	return mi
}
