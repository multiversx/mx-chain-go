package bridge

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
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

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ChainSimulatorArgs: &chainSimulator.ArgsChainSimulator{
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
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)

	logger.SetLogLevel("*:DEBUG,process:TRACE,block:TRACE")

	headerNonce := uint64(9999999)
	randomHeader := createHeaderV2(headerNonce, generateRandomHash(), generateRandomHash())

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	incomingHeader := createIncomingHeader(nodeHandler, &headerNonce, randomHeader, nil)
	err = nodeHandler.GetIncomingHeaderSubscriber().AddHeader(generateRandomHash(), incomingHeader)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	incomingHeader = createIncomingHeader(nodeHandler, &headerNonce, incomingHeader.Header, createTransactionsEvent())
	err = nodeHandler.GetIncomingHeaderSubscriber().AddHeader(generateRandomHash(), incomingHeader)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	incomingHeader = createIncomingHeader(nodeHandler, &headerNonce, incomingHeader.Header, nil)
	err = nodeHandler.GetIncomingHeaderSubscriber().AddHeader(generateRandomHash(), incomingHeader)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
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

func createIncomingHeader(nodeHandler process.NodeHandler, headerNonce *uint64, prevHeader *block.HeaderV2, txsEvent []*transaction.Event) *sovereign.IncomingHeader {
	*headerNonce++
	prevHeaderHash, _ := core.CalculateHash(nodeHandler.GetCoreComponents().InternalMarshalizer(), nodeHandler.GetCoreComponents().Hasher(), prevHeader)
	incomingHeader := &sovereign.IncomingHeader{
		Header:         createHeaderV2(*headerNonce, prevHeaderHash, prevHeader.GetRandSeed()),
		IncomingEvents: txsEvent,
	}
	return incomingHeader
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

func createTransactionsEvent() []*transaction.Event {
	addressBytes, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
	tokenId, _ := hex.DecodeString("53564e2d373330613336")
	tokenData, _ := hex.DecodeString("000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	eventData, _ := hex.DecodeString("0000000000000001c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f600")

	events := make([]*transaction.Event, 0)
	return append(events, &transaction.Event{
		Identifier: []byte(eventIDDepositIncomingTransfer),
		Topics:     [][]byte{[]byte(topicIDDepositIncomingTransfer), addressBytes, tokenId, []byte(""), tokenData},
		Data:       eventData,
	})
}

func generateRandomHash() []byte {
	randomBytes := make([]byte, hashSize)
	_, _ = rand.Read(randomBytes)
	return randomBytes
}
