package transactionAPI

import (
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	rewardTxData "github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/sharding"
)

type txUnmarshaller struct {
	shardCoordinator       sharding.Coordinator
	addressPubKeyConverter core.PubkeyConverter
	marshalizer            marshal.Marshalizer
	dataFieldParser        DataFieldParser
}

func newTransactionUnmarshaller(
	marshalizer marshal.Marshalizer,
	addressPubKeyConverter core.PubkeyConverter,
	dataFieldParser DataFieldParser,
	shardCoordinator sharding.Coordinator,
) *txUnmarshaller {
	return &txUnmarshaller{
		marshalizer:            marshalizer,
		addressPubKeyConverter: addressPubKeyConverter,
		dataFieldParser:        dataFieldParser,
		shardCoordinator:       shardCoordinator,
	}
}

func (tu *txUnmarshaller) unmarshalReceipt(receiptBytes []byte) (*transaction.ApiReceipt, error) {
	rec := &receipt.Receipt{}
	err := tu.marshalizer.Unmarshal(rec, receiptBytes)
	if err != nil {
		return nil, err
	}

	senderAddress := tu.addressPubKeyConverter.SilentEncode(rec.SndAddr, log)

	return &transaction.ApiReceipt{
		Value:   rec.Value,
		SndAddr: senderAddress,
		Data:    string(rec.Data),
		TxHash:  hex.EncodeToString(rec.TxHash),
	}, nil
}

func (tu *txUnmarshaller) unmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	var apiTx *transaction.ApiTransactionResult
	var err error

	switch txType {
	case transaction.TxTypeNormal:
		var tx transaction.Transaction
		err = tu.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		apiTx = tu.prepareNormalTx(&tx)
	case transaction.TxTypeInvalid:
		var tx transaction.Transaction
		err = tu.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		apiTx = tu.prepareInvalidTx(&tx)
	case transaction.TxTypeReward:
		var tx rewardTxData.RewardTx
		err = tu.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		apiTx = tu.prepareRewardTx(&tx)

	case transaction.TxTypeUnsigned:
		var tx smartContractResult.SmartContractResult
		err = tu.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		apiTx = tu.prepareUnsignedTx(&tx)
	}
	if err != nil {
		return nil, err
	}

	res := tu.dataFieldParser.Parse(apiTx.Data, apiTx.Tx.GetSndAddr(), apiTx.Tx.GetRcvAddr(), tu.shardCoordinator.NumberOfShards())
	apiTx.Operation = res.Operation
	apiTx.Function = res.Function
	apiTx.ESDTValues = res.ESDTValues
	apiTx.Tokens = res.Tokens
	apiTx.Receivers, err = tu.addressPubKeyConverter.EncodeSlice(res.Receivers)
	if err != nil {
		log.Warn("bech32PubkeyConverter.EncodeSlice() failed while encoding apiSCR.Receivers with", "err", err)
	}

	apiTx.ReceiversShardIDs = res.ReceiversShardID
	apiTx.IsRelayed = res.IsRelayed

	return apiTx, nil
}

func (tu *txUnmarshaller) prepareNormalTx(tx *transaction.Transaction) *transaction.ApiTransactionResult {
	receiverAddress := tu.addressPubKeyConverter.SilentEncode(tx.RcvAddr, log)
	senderAddress := tu.addressPubKeyConverter.SilentEncode(tx.SndAddr, log)

	apiTx := &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeNormal),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         receiverAddress,
		ReceiverUsername: tx.RcvUserName,
		Sender:           senderAddress,
		SenderUsername:   tx.SndUserName,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
		Options:          tx.Options,
		Version:          tx.Version,
		ChainID:          string(tx.ChainID),
	}

	if len(tx.GuardianAddr) > 0 {
		apiTx.GuardianAddr = tu.addressPubKeyConverter.SilentEncode(tx.GuardianAddr, log)
		apiTx.GuardianSignature = hex.EncodeToString(tx.GuardianSignature)
	}

	if len(tx.RelayerAddr) > 0 {
		apiTx.RelayerAddress = tu.addressPubKeyConverter.SilentEncode(tx.RelayerAddr, log)
	}

	return apiTx
}

func (tu *txUnmarshaller) prepareInvalidTx(tx *transaction.Transaction) *transaction.ApiTransactionResult {
	receiverAddress := tu.addressPubKeyConverter.SilentEncode(tx.RcvAddr, log)
	senderAddress := tu.addressPubKeyConverter.SilentEncode(tx.SndAddr, log)

	apiTx := &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeInvalid),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         receiverAddress,
		ReceiverUsername: tx.RcvUserName,
		Sender:           senderAddress,
		SenderUsername:   tx.SndUserName,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
		Options:          tx.Options,
		Version:          tx.Version,
		ChainID:          string(tx.ChainID),
	}

	if len(tx.GuardianAddr) > 0 {
		apiTx.GuardianAddr = tu.addressPubKeyConverter.SilentEncode(tx.GuardianAddr, log)
		apiTx.GuardianSignature = hex.EncodeToString(tx.GuardianSignature)
	}

	if len(tx.RelayerAddr) > 0 {
		apiTx.RelayerAddress = tu.addressPubKeyConverter.SilentEncode(tx.RelayerAddr, log)
	}

	return apiTx
}

func (tu *txUnmarshaller) prepareRewardTx(tx *rewardTxData.RewardTx) *transaction.ApiTransactionResult {
	receiverAddress := tu.addressPubKeyConverter.SilentEncode(tx.GetRcvAddr(), log)

	return &transaction.ApiTransactionResult{
		Tx:          tx,
		Type:        string(transaction.TxTypeReward),
		Round:       tx.GetRound(),
		Epoch:       tx.GetEpoch(),
		Value:       tx.GetValue().String(),
		Sender:      "metachain",
		Receiver:    receiverAddress,
		SourceShard: core.MetachainShardId,
	}
}

func (tu *txUnmarshaller) prepareUnsignedTx(tx *smartContractResult.SmartContractResult) *transaction.ApiTransactionResult {
	receiverAddress := tu.addressPubKeyConverter.SilentEncode(tx.GetRcvAddr(), log)
	senderAddress := tu.addressPubKeyConverter.SilentEncode(tx.GetSndAddr(), log)

	txResult := &transaction.ApiTransactionResult{
		Tx:                      tx,
		Type:                    string(transaction.TxTypeUnsigned),
		Nonce:                   tx.GetNonce(),
		Value:                   tx.GetValue().String(),
		Receiver:                receiverAddress,
		Sender:                  senderAddress,
		GasPrice:                tx.GetGasPrice(),
		GasLimit:                tx.GetGasLimit(),
		Data:                    tx.GetData(),
		Code:                    string(tx.GetCode()),
		CodeMetadata:            tx.GetCodeMetadata(),
		PreviousTransactionHash: hex.EncodeToString(tx.GetPrevTxHash()),
		OriginalTransactionHash: hex.EncodeToString(tx.GetOriginalTxHash()),
		ReturnMessage:           string(tx.GetReturnMessage()),
		CallType:                tx.CallType.ToString(),
		RelayerAddress:          tu.getEncodedAddress(tx.GetRelayerAddr()),
		RelayedValue:            bigIntToStr(tx.GetRelayedValue()),
		OriginalSender:          tu.getEncodedAddress(tx.GetOriginalSender()),
	}

	return txResult
}

func (tu *txUnmarshaller) getEncodedAddress(address []byte) string {
	if len(address) == tu.addressPubKeyConverter.Len() {
		return tu.addressPubKeyConverter.SilentEncode(address, log)
	}

	return ""
}

func bigIntToStr(value *big.Int) string {
	if value != nil {
		return value.String()
	}

	return ""
}
