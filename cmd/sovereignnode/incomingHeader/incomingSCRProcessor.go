package incomingHeader

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
)

const (
	minTopicsInEvent  = 4
	numTransferTopics = 3
)

type scrInfo struct {
	scr  *smartContractResult.SmartContractResult
	hash []byte
}

type scrProcessor struct {
	txPool     TransactionPool
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
	nonce      uint64
}

func (sp *scrProcessor) createIncomingSCRs(events []data.EventHandler) ([]*scrInfo, error) {
	scrs := make([]*scrInfo, 0, len(events))

	for idx, event := range events {
		topics := event.GetTopics()
		// TODO: Check each param validity (e.g. check that topic[0] == valid address)
		if len(topics) < minTopicsInEvent || len(topics[1:])%numTransferTopics != 0 {
			log.Error("incomingHeaderHandler.createIncomingSCRs",
				"error", errInvalidNumTopicsIncomingEvent,
				"num topics", len(topics),
				"topics", topics)
			return nil, fmt.Errorf("%w at event idx = %d; num topics = %d",
				errInvalidNumTopicsIncomingEvent, idx, len(topics))
		}

		scr := &smartContractResult.SmartContractResult{
			Nonce:          sp.nonce, // TODO:  Save this nonce in storage + load in from storage in MX-14320 task
			OriginalTxHash: nil,      // TODO:  Implement this in MX-14321 task
			RcvAddr:        topics[0],
			SndAddr:        core.ESDTSCAddress,
			Data:           createSCRData(topics),
			Value:          big.NewInt(0),
		}

		hash, err := core.CalculateHash(sp.marshaller, sp.hasher, scr)
		if err != nil {
			return nil, err
		}

		// TODO: Should we revert nonce incrementation if processing fails later in code?
		sp.nonce++
		scrs = append(scrs, &scrInfo{
			scr:  scr,
			hash: hash,
		})
	}

	return scrs, nil
}

func createSCRData(topics [][]byte) []byte {
	numTokensToTransfer := len(topics[1:]) / numTransferTopics
	numTokensToTransferBytes := big.NewInt(int64(numTokensToTransfer)).Bytes()

	ret := []byte(core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(numTokensToTransferBytes))

	for idx := 1; idx < len(topics[1:]); idx += 3 {
		transfer := []byte("@" +
			hex.EncodeToString(topics[idx]) + // tokenID
			"@" + hex.EncodeToString(topics[idx+1]) + //nonce
			"@" + hex.EncodeToString(topics[idx+2])) //value

		ret = append(ret, transfer...)
	}

	return ret
}

func (sp *scrProcessor) addSCRsToPool(scrs []*scrInfo) {
	cacheID := process.ShardCacherIdentifier(core.MainChainShardId, core.SovereignChainShardId)

	for _, scrData := range scrs {
		sp.txPool.AddData(scrData.hash, scrData.scr, scrData.scr.Size(), cacheID)
	}
}
