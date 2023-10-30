package configs

import (
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/realcomponents"
)

func TestNewProcessorRunnerChainArguments(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outputConfig := CreateChainSimulatorConfigs(t, ArgsChainSimulatorConfigs{
		NumOfShards:               3,
		OriginalConfigsPath:       "../../../cmd/node/config",
		GenesisAddressWithStake:   "erd10z6sdhwfy8jtuf87j5gnq7lt7fd2wfmhkg8zfzf79lrapzq265yqlnmtm7",
		GenesisAddressWithBalance: "erd1rhrm20mmf2pugzxc3twlu3fa264hxeefnglsy4ads4dpccs9s3jsg6qdrz",
	})

	pr := realcomponents.NewProcessorRunner(t, *outputConfig.Configs)
	pr.Close(t)
}
