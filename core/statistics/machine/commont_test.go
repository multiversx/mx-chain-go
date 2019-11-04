package machine

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/statistics/machine/cpuUsage"
	"github.com/stretchr/testify/assert"
)

func TestCanRead(t *testing.T) {
	t.Parallel()

	cpuUsagePercent := cpuUsage.GetCpuUsage()
	fmt.Println(cpuUsagePercent)
	assert.NotNil(t, cpuUsagePercent)
}
