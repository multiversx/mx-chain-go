package configs

import (
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/realcomponents"
	"github.com/stretchr/testify/require"
)

func TestNewProcessorRunnerChainArguments(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	outputConfig, err := CreateChainSimulatorConfigs(ArgsChainSimulatorConfigs{
		NumOfShards:               3,
		OriginalConfigsPath:       "../../../cmd/node/config",
		GenesisAddressWithStake:   "erd10z6sdhwfy8jtuf87j5gnq7lt7fd2wfmhkg8zfzf79lrapzq265yqlnmtm7",
		GenesisAddressWithBalance: "erd1rhrm20mmf2pugzxc3twlu3fa264hxeefnglsy4ads4dpccs9s3jsg6qdrz",
	})
	require.Nil(t, err)

	pr := realcomponents.NewProcessorRunner(t, *outputConfig.Configs)
	pr.Close(t)
}
