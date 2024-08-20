package processBlock

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"
)

func TestSovereignChainSimulator_BlockTrackerPoolsCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			ConsensusGroupSize:     2,
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	headerNonce := uint64(9999999)
	prevHeader := createHeaderV2(headerNonce, generateRandomHash(), generateRandomHash())

	for round := 1; round <= 10; round++ {
		err = cs.GenerateBlocks(1)
		require.Nil(t, err)

		selfNotarizedHeaders, _, err := nodeHandler.GetProcessComponents().BlockTracker().GetSelfNotarizedHeader(core.SovereignChainShardId, 0)
		require.Nil(t, err)
		require.NotNil(t, selfNotarizedHeaders)
		require.Equal(t, uint64(round)-1, selfNotarizedHeaders.GetNonce()) // in round X, header with nonce X-1 is notarized

		crossTrackedHeaders, _ := nodeHandler.GetProcessComponents().BlockTracker().GetTrackedHeaders(core.MainChainShardId)
		require.Nil(t, err)
		checkCrossTrackedHeaders(t, round, crossTrackedHeaders, headerNonce)

		crossNotarizedHeader, _, err := nodeHandler.GetProcessComponents().BlockTracker().GetCrossNotarizedHeader(core.MainChainShardId, 0)
		require.Nil(t, err)
		checkCrossNotarizedHeader(t, round, crossNotarizedHeader, headerNonce)

		incomingHeader, headerHash := createIncomingHeader(nodeHandler, &headerNonce, prevHeader, nil)
		err = nodeHandler.GetIncomingHeaderSubscriber().AddHeader(headerHash, incomingHeader)
		require.Nil(t, err)

		prevHeader = incomingHeader.Header
	}
}

func checkCrossTrackedHeaders(t *testing.T, round int, crossTrackedHeaders []data.HeaderHandler, headerNonce uint64) {
	if round == 1 {
		require.Nil(t, crossTrackedHeaders)
	} else {
		require.NotNil(t, crossTrackedHeaders)
		require.Equal(t, headerNonce, crossTrackedHeaders[len(crossTrackedHeaders)-1].GetNonce())
		require.LessOrEqual(t, len(crossTrackedHeaders), 4) // saved 3 notarized + 1 tracked
	}
}

func checkCrossNotarizedHeader(t *testing.T, round int, crossNotarizedHeader data.HeaderHandler, headerNonce uint64) {
	require.NotNil(t, crossNotarizedHeader)

	if round == 1 { // first cross header is dummy with nonce 0
		require.Equal(t, uint64(0), crossNotarizedHeader.GetNonce())
	} else if round == 2 { // first cross notarized header is 10000000
		require.Equal(t, headerNonce, crossNotarizedHeader.GetNonce())
	} else { // cross header with nonce-1 is notarized
		require.Equal(t, headerNonce-1, crossNotarizedHeader.GetNonce())
	}
}
