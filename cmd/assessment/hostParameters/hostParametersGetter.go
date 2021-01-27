package hostParameters

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type hostParametersGetter struct {
	versionString string
}

// NewHostParameterGetter will create a structure that is able to get and format the host's relevant parameters
func NewHostParameterGetter(version string) *hostParametersGetter {
	return &hostParametersGetter{
		versionString: version,
	}
}

// GetHostInfo is able to get all the known parameters of a host
func (hpg *hostParametersGetter) GetHostInfo() *HostInfo {
	hi := &HostInfo{
		AppVersion: hpg.versionString,
	}

	hpg.applyCpuInfo(hi)
	hpg.applyMemInfo(hi)

	return hi
}

func (hpg *hostParametersGetter) applyCpuInfo(hi *HostInfo) {
	rawCpuInfo, err := cpu.Info()
	if err != nil {
		hi.CPUModel = fmt.Sprintf("[ERR:%s]", err)
		return
	}

	if len(rawCpuInfo) == 0 {
		hi.CPUModel = "[ERR:no logical cpus]"
		return
	}

	hi.CPUNumLogical = len(rawCpuInfo)
	hi.CPUModel = rawCpuInfo[0].ModelName
	hi.CPUMaxFreqInMHz = int(rawCpuInfo[0].Mhz)
	hi.CPUFlags = rawCpuInfo[0].Flags
	sort.Slice(hi.CPUFlags, func(i, j int) bool {
		return strings.Compare(hi.CPUFlags[i], hi.CPUFlags[j]) < 0
	})
}

func (hpg *hostParametersGetter) applyMemInfo(hi *HostInfo) {
	vms, err := mem.VirtualMemory()
	if err != nil {
		hi.MemorySize = fmt.Sprintf("[ERR:%s]", err)
	}

	hi.MemorySize = core.ConvertBytes(vms.Total)
}
