package bridge

import (
	"encoding/hex"
	"testing"
	"time"

	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	eventIDDepositIncomingTransfer = "deposit"
	topicIDDepositIncomingTransfer = "deposit"
)

func TestIncomingOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	epochConfig, economicsConfig, sovereignExtraConfig, err := sovereignChainSimulator.LoadSovereignConfigs(sovereignConfigPath)
	require.Nil(t, err)

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignExtraConfig: *sovereignExtraConfig,
		ChainSimulatorArgs: chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			NumOfShards:            1,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			MetaChainMinNodes:      0,
			InitialRound:           100,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EconomicsConfig = economicsConfig
				cfg.EpochConfig = epochConfig
				cfg.GeneralConfig.SovereignConfig = *sovereignExtraConfig
				cfg.GeneralConfig.VirtualMachine.Execution.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}
				cfg.GeneralConfig.VirtualMachine.Querying.WasmVMVersions = []config.WasmVMVersionByEpoch{{StartEpoch: 0, Version: "v1.5"}}
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	addressBytes, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
	tokenId, _ := hex.DecodeString("53564e2d373330613336")
	tokenData, _ := hex.DecodeString("000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	eventData, _ := hex.DecodeString("0000000000000001c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f600")

	logger.SetLogLevel("*:DEBUG,process:TRACE")

	incomingHeader := &sovereign.IncomingHeader{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Round: 9999999,
				Nonce: 9999999,
			},
		},
		IncomingEvents: []*transaction.Event{
			{
				Identifier: []byte(eventIDDepositIncomingTransfer),
				Topics:     [][]byte{[]byte(topicIDDepositIncomingTransfer), addressBytes, tokenId, []byte(""), tokenData},
				Data:       eventData,
			},
		},
	}
	prevhash, _ := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), incomingHeader.Header)
	err = nodeHandler.GetIncomingHeaderHandler().AddHeader([]byte("hash"), incomingHeader)
	require.Nil(t, err)

	incomingHeader2 := &sovereign.IncomingHeader{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Round:    9999999 + 1,
				Nonce:    9999999 + 1,
				PrevHash: prevhash,
			},
		},
		IncomingEvents: []*transaction.Event{
			{
				Identifier: []byte(eventIDDepositIncomingTransfer),
				Topics:     [][]byte{[]byte(topicIDDepositIncomingTransfer), addressBytes, tokenId, []byte(""), tokenData},
				Data:       eventData,
			},
		},
	}

	err = nodeHandler.GetIncomingHeaderHandler().AddHeader([]byte("hash2"), incomingHeader2)
	require.Nil(t, err)

	err = cs.GenerateBlocks(20)
	require.Nil(t, err)

	address := "erd1crq888sv7c3j4y6d9etvlngs3q0tr3endufgls245j5y9yk0ulmqlsuute"
	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, account)

	esdts, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdts)

	//oneEgld := big.NewInt(1000000000000000000)
	//initialMinting := big.NewInt(0).Mul(oneEgld, big.NewInt(100))
	//wallet0, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	//require.Nil(t, err)
	//
	//err = cs.GenerateBlocks(10)
	//
	//w, _, err := nodeHandler.GetFacadeHandler().GetAccount(wallet0.Bech32, coreAPI.AccountQueryOptions{})
	//require.NotNil(t, w)
}
