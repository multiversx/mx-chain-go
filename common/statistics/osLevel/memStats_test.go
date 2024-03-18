//go:build darwin || linux

package osLevel

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCurrentMemStats(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping test on darwin")
	}

	t.Parallel()

	memStats, err := ReadCurrentMemStats()
	assert.Nil(t, err)

	sumCounters := memStats.SumFields()
	assert.True(t, sumCounters > 0)
}

func TestParseResponse(t *testing.T) {
	t.Parallel()

	t.Run("unparsable data", func(t *testing.T) {
		t.Parallel()

		memStats, err := parseResponse("unparsable data")
		assert.Nil(t, memStats)
		assert.Equal(t, errUnparsableData, err)
	})
	t.Run("invalid line should error", func(t *testing.T) {
		t.Parallel()

		dataString := `
VmPeak:	 1 kB
VmSize:	 2              kB
VmHWM:
HugetlbPages:	       1 kB`

		memStats, err := parseResponse(dataString)
		assert.Nil(t, memStats)
		assert.ErrorIs(t, err, errWrongNumberOfComponents)
	})
	t.Run("invalid value in line should error", func(t *testing.T) {
		t.Parallel()

		dataString := `
VmPeak:	 1 kB
VmSize:	 2              kB
VmHWM: NaN kB
HugetlbPages:	       1 kB`

		memStats, err := parseResponse(dataString)
		assert.Nil(t, memStats)
		assert.Equal(t, "strconv.Atoi: parsing \"NaN\": invalid syntax, invalid string for VmHWM NaN kB", err.Error())
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dataString := `
VmPeak:	 1 kB
VmSize:	 2              kB
VmLck: 4 kB
VmPin:			8	kB
VmHWM:		16 kB
VmRSS: 32		kB
RssAnon: 64 kB
RssFile: 128 kB
RssShmem: 256 kB
VmData:	512 kB
VmStk: 1024 kB
VmExe: 2048 kB
VmLib: 4096 kB
VmPTE: 8192 kB
VmSwap: 16384 kB	
HugetlbPages:	       1 kB`

		expectedString := "VmPeak: 1.00 KB, VmSize: 2.00 KB, VmLck: 4.00 KB, VmPin: 8.00 KB, VmHWM: 16.00 KB, VmRSS: 32.00 KB, " +
			"RssAnon: 64.00 KB, RssFile: 128.00 KB, RssShmem: 256.00 KB, VmData: 512.00 KB, VmStk: 1.00 MB, VmExe: 2.00 MB, " +
			"VmLib: 4.00 MB, VmPTE: 8.00 MB, VmSwap: 16.00 MB"

		memStats, err := parseResponse(dataString)
		assert.Nil(t, err)
		assert.Equal(t, uint64(32767*1024), memStats.SumFields())
		assert.Equal(t, expectedString, memStats.String())
	})
}
