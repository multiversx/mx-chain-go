package hostParameters

import (
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go/display"
)

const versionMarker = "App version"
const modelMarker = "CPU model"
const numLogicalMarker = "CPU numLogical"
const maxFreqMarker = "CPU maxFreq"
const freqMarker = "MHz"
const flagsMarker = "CPU flags"
const maxFlagsCharsPerLine = 80
const sizeMarker = "Memory size"

// HostInfo will contain the host information
type HostInfo struct {
	AppVersion      string
	CPUModel        string
	CPUNumLogical   int
	CPUFlags        []string
	CPUMaxFreqInMHz int
	MemorySize      string
}

// ToDisplayTable will output the contained data as an ASCII table
func (hi *HostInfo) ToDisplayTable() string {
	lines := []*display.LineData{
		display.NewLineData(true, []string{versionMarker, hi.AppVersion}),
		display.NewLineData(false, []string{modelMarker, hi.CPUModel}),
		display.NewLineData(false, []string{numLogicalMarker, hi.cpuNumLogicalAsString()}),
	}
	accumulator := ""
	flagMarkerWritten := false
	flagMarkerToBeWritten := ""
	for _, flag := range hi.CPUFlags {
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

	lines = append(lines, display.NewLineData(true, []string{maxFreqMarker, hi.frequencyAsString()}))
	lines = append(lines, display.NewLineData(true, []string{sizeMarker, hi.MemorySize}))

	hdr := []string{"Parameter", "Value"}
	tbl, err := display.CreateTableString(hdr, lines)
	if err != nil {
		return fmt.Sprintf("[ERR:%s]", err)
	}

	return tbl
}

func (hi *HostInfo) cpuNumLogicalAsString() string {
	return fmt.Sprintf("%d", hi.CPUNumLogical)
}

func (hi *HostInfo) frequencyAsString() string {
	return fmt.Sprintf("%d %s", hi.CPUMaxFreqInMHz, freqMarker)
}

// ToStrings will return the contained data as strings (to be easily written, e.g. in a file)
func (hi *HostInfo) ToStrings() [][]string {
	result := make([][]string, 0)

	result = append(result, []string{versionMarker, hi.AppVersion})
	result = append(result, []string{modelMarker, hi.CPUModel})
	result = append(result, []string{numLogicalMarker, hi.cpuNumLogicalAsString()})
	result = append(result, []string{flagsMarker, strings.Join(hi.CPUFlags, " ")})
	result = append(result, []string{maxFreqMarker, hi.frequencyAsString()})
	result = append(result, []string{sizeMarker, hi.MemorySize})

	return result
}
