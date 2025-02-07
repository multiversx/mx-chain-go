package incomingHeader

import (
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/common"
)

type eventData struct {
	nonce                uint64
	functionCallWithArgs []byte
	gasLimit             uint64
}

type depositEventProc struct {
	marshaller    marshal.Marshalizer
	hasher        hashing.Hasher
	dataCodec     SovereignDataCodec
	topicsChecker TopicsChecker
}

// ProcessEvent will process incoming token deposit events and return an incoming scr info
func (dep *depositEventProc) ProcessEvent(event data.EventHandler) (*EventResult, error) {
	topics := event.GetTopics()
	err := dep.topicsChecker.CheckValidity(topics)
	if err != nil {
		return nil, err
	}

	receivedEventData, err := dep.createEventData(event.GetData())
	if err != nil {
		return nil, err
	}

	scrData, err := dep.createSCRData(topics)
	if err != nil {
		return nil, err
	}

	scrData = append(scrData, receivedEventData.functionCallWithArgs...)
	scr := &smartContractResult.SmartContractResult{
		Nonce:          receivedEventData.nonce,
		OriginalTxHash: nil, // TODO:  Implement this in MX-14321 task
		RcvAddr:        topics[1],
		SndAddr:        core.ESDTSCAddress,
		Data:           scrData,
		Value:          big.NewInt(0),
		GasLimit:       receivedEventData.gasLimit,
	}

	hash, err := core.CalculateHash(dep.marshaller, dep.hasher, scr)
	if err != nil {
		return nil, err
	}

	return &EventResult{
		SCR: &SCRInfo{
			SCR:  scr,
			Hash: hash,
		},
	}, nil
}

func (dep *depositEventProc) createEventData(data []byte) (*eventData, error) {
	evData, err := dep.dataCodec.DeserializeEventData(data)
	if err != nil {
		return nil, err
	}

	gasLimit, functionCallWithArgs := extractSCTransferInfo(evData.TransferData)
	return &eventData{
		nonce:                evData.Nonce,
		functionCallWithArgs: functionCallWithArgs,
		gasLimit:             gasLimit,
	}, nil
}

func extractSCTransferInfo(transferData *sovereign.TransferData) (uint64, []byte) {
	gasLimit := uint64(0)
	functionCallWithArgs := make([]byte, 0)
	if transferData != nil {
		gasLimit = transferData.GasLimit

		functionCallWithArgs = append(functionCallWithArgs, []byte("@")...)
		functionCallWithArgs = append(functionCallWithArgs, hex.EncodeToString(transferData.Function)...)
		functionCallWithArgs = append(functionCallWithArgs, extractArguments(transferData.Args)...)
	}

	return gasLimit, functionCallWithArgs
}

func extractArguments(arguments [][]byte) []byte {
	if len(arguments) == 0 {
		return make([]byte, 0)
	}

	args := make([]byte, 0)
	for _, arg := range arguments {
		args = append(args, []byte("@")...)
		args = append(args, hex.EncodeToString(arg)...)
	}

	return args
}

func (dep *depositEventProc) createSCRData(topics [][]byte) ([]byte, error) {
	numTokensToTransfer := len(topics[tokensIndex:]) / numTransferTopics
	numTokensToTransferBytes := big.NewInt(int64(numTokensToTransfer)).Bytes()

	ret := []byte(core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(numTokensToTransferBytes))

	for idx := tokensIndex; idx < len(topics); idx += numTransferTopics {
		tokenData, err := dep.getTokenDataBytes(topics[idx+1], topics[idx+2])
		if err != nil {
			return nil, err
		}

		transfer := []byte("@" +
			hex.EncodeToString(topics[idx]) + // tokenID
			"@" + hex.EncodeToString(topics[idx+1]) + // nonce
			"@" + hex.EncodeToString(tokenData)) // value/tokenData

		ret = append(ret, transfer...)
	}

	return ret, nil
}

func (dep *depositEventProc) getTokenDataBytes(tokenNonce []byte, tokenData []byte) ([]byte, error) {
	esdtTokenData, err := dep.dataCodec.DeserializeTokenData(tokenData)
	if err != nil {
		return nil, err
	}

	if esdtTokenData.TokenType == core.Fungible {
		return esdtTokenData.Amount.Bytes(), nil
	}

	nonce, err := common.ByteSliceToUint64(tokenNonce)
	if err != nil {
		return nil, err
	}

	digitalToken := &esdt.ESDigitalToken{
		Type:  uint32(esdtTokenData.TokenType),
		Value: esdtTokenData.Amount,
		TokenMetaData: &esdt.MetaData{
			Nonce:      nonce,
			Name:       esdtTokenData.Name,
			Creator:    esdtTokenData.Creator,
			Royalties:  uint32(esdtTokenData.Royalties.Uint64()),
			Hash:       esdtTokenData.Hash,
			URIs:       esdtTokenData.Uris,
			Attributes: esdtTokenData.Attributes,
		},
	}

	return dep.marshaller.Marshal(digitalToken)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (dep *depositEventProc) IsInterfaceNil() bool {
	return dep == nil
}
