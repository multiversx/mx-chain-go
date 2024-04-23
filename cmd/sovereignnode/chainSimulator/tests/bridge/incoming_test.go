package bridge

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	eventIDDepositIncomingTransfer = "deposit"
	topicIDDepositIncomingTransfer = "deposit"
	hashSize                       = 32
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
			ConsensusGroupSize:     2,
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

	//logger.SetLogLevel("*:DEBUG,process:TRACE")

	nonce := uint64(9999999)
	incomingHeader := &sovereign.IncomingHeader{
		Header:         createHeaderV2(nonce, generateRandomHash(), generateRandomHash()),
		IncomingEvents: []*transaction.Event{createTransactionEvent()},
	}
	firstHeaderHash, _ := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), incomingHeader.Header)
	err = nodeHandler.GetIncomingHeaderHandler().AddHeader([]byte("hash"), incomingHeader)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	nonce++
	incomingHeader2 := &sovereign.IncomingHeader{
		Header:         createHeaderV2(nonce, firstHeaderHash, incomingHeader.Header.GetRandSeed()),
		IncomingEvents: []*transaction.Event{createTransactionEvent()},
	}
	err = nodeHandler.GetIncomingHeaderHandler().AddHeader([]byte("hash2"), incomingHeader2)
	require.Nil(t, err)

	err = cs.GenerateBlocks(10)
	require.Nil(t, err)

	address := "erd1crq888sv7c3j4y6d9etvlngs3q0tr3endufgls245j5y9yk0ulmqlsuute"
	account, _, err := nodeHandler.GetFacadeHandler().GetAccount(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, account)

	esdts, _, err := nodeHandler.GetFacadeHandler().GetAllESDTTokens(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	require.NotNil(t, esdts)
	require.True(t, len(esdts) > 0)
}

func createTransactionEvent() *transaction.Event {
	addressBytes, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
	tokenId, _ := hex.DecodeString("53564e2d373330613336")
	tokenData, _ := hex.DecodeString("000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	eventData, _ := hex.DecodeString("0000000000000001c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f600")

	return &transaction.Event{
		Identifier: []byte(eventIDDepositIncomingTransfer),
		Topics:     [][]byte{[]byte(topicIDDepositIncomingTransfer), addressBytes, tokenId, []byte(""), tokenData},
		Data:       eventData,
	}
}

func createHeaderV2(nonce uint64, prevHash []byte, prevRandSeed []byte) *block.HeaderV2 {
	return &block.HeaderV2{
		Header: &block.Header{
			PrevHash:     prevHash,
			Nonce:        nonce,
			Round:        nonce,
			RandSeed:     generateRandomHash(),
			PrevRandSeed: prevRandSeed,
			ChainID:      []byte(configs.ChainID),
		},
	}
}

func generateRandomHash() []byte {
	randomBytes := make([]byte, hashSize)
	_, _ = rand.Read(randomBytes)
	return randomBytes
}
