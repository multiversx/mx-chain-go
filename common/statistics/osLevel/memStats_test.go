//go:build darwin || linux
// +build darwin linux

package osLevel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCurrentMemStats(t *testing.T) {
	t.Parallel()

	memStats, err := ReadCurrentMemStats()
	assert.Nil(t, err)
	sumCounters := memStats.VmPeak + memStats.VmSize + memStats.VmLck + memStats.VmPin + memStats.VmHWM +
		memStats.VmRSS + memStats.RssAnon + memStats.RssFile + memStats.RssShmem + memStats.VmData +
		memStats.VmStk + memStats.VmExe + memStats.VmLib + memStats.VmPTE + memStats.VmSwap

	assert.True(t, sumCounters > 0)
}
