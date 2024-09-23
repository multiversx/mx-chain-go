package hostParameters

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostInfo_ToDisplayTable(t *testing.T) {
	t.Parallel()

	appVersion := "appversion"
	cpuModel := "cpumodel"
	numLogicalCPU := 1
	cpuFlags := []string{"sse", "fma3"}
	cpuMaxFreqInMHz := 2
	memorySize := "100 GB"
	hi := &HostInfo{
		AppVersion:      appVersion,
		CPUModel:        cpuModel,
		CPUNumLogical:   numLogicalCPU,
		CPUFlags:        cpuFlags,
		CPUMaxFreqInMHz: cpuMaxFreqInMHz,
		MemorySize:      memorySize,
	}

	tbl := hi.ToDisplayTable()
	fmt.Println(tbl)

	stringsToContain := []string{versionMarker, modelMarker, modelMarker, maxFreqMarker, freqMarker, flagsMarker, sizeMarker,
		appVersion, cpuModel, fmt.Sprintf("%d", numLogicalCPU), strings.Join(cpuFlags, " "),
		fmt.Sprintf("%d", cpuMaxFreqInMHz), memorySize}

	for _, str := range stringsToContain {
		assert.True(t, strings.Contains(tbl, str), "string %s not contained", str)
	}
}

func TestHostInfo_ToStrings(t *testing.T) {
	t.Parallel()

	appVersion := "appversion"
	cpuModel := "cpumodel"
	numLogicalCPU := 1
	cpuFlags := []string{"sse", "fma3"}
	cpuMaxFreqInMHz := 2
	memorySize := "100 GB"
	hi := &HostInfo{
		AppVersion:      appVersion,
		CPUModel:        cpuModel,
		CPUNumLogical:   numLogicalCPU,
		CPUFlags:        cpuFlags,
		CPUMaxFreqInMHz: cpuMaxFreqInMHz,
		MemorySize:      memorySize,
	}

	data := hi.ToStrings()

	stringsToContain := []string{versionMarker, modelMarker, modelMarker, maxFreqMarker, freqMarker, flagsMarker, sizeMarker,
		appVersion, cpuModel, fmt.Sprintf("%d", numLogicalCPU), strings.Join(cpuFlags, " "),
		fmt.Sprintf("%d", cpuMaxFreqInMHz), memorySize}

	for _, str := range stringsToContain {
		found := false
		for _, line := range data {
			for _, cell := range line {
				if strings.Contains(cell, str) {
					found = true
					break
				}
			}
		}

		assert.True(t, found, "string %s not contained", str)
	}
}
