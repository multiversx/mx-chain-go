package transactionAPI

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	rewardTxData "github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

type txUnmarshaller struct {
	addressPubKeyConverter core.PubkeyConverter
	marshalizer            marshal.Marshalizer
	dataFieldParser        DataFieldParser
}

func newTransactionUnmarshaller(
	marshalizer marshal.Marshalizer,
	addressPubKeyConverter core.PubkeyConverter,
	dataFieldParser DataFieldParser,
) *txUnmarshaller {
	return &txUnmarshaller{
		marshalizer:            marshalizer,
		addressPubKeyConverter: addressPubKeyConverter,
		dataFieldParser:        dataFieldParser,
	}
}

func (tu *txUnmarshaller) unmarshalReceipt(receiptBytes []byte) (*transaction.ApiReceipt, error) {
	rec := &receipt.Receipt{}
	err := tu.marshalizer.Unmarshal(rec, receiptBytes)
	if err != nil {
		return nil, err
	}

	encodedSndAddr, err := tu.addressPubKeyConverter.Encode(rec.SndAddr)
	if err != nil {
		return nil, err
	}

	return &transaction.ApiReceipt{
		Value:   rec.Value,
		SndAddr: encodedSndAddr,
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
		apiTx, err = tu.prepareNormalTx(&tx)
	case transaction.TxTypeInvalid:
		var tx transaction.Transaction
		err = tu.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		apiTx, err = tu.prepareInvalidTx(&tx)
	case transaction.TxTypeReward:
		var tx rewardTxData.RewardTx
		err = tu.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		apiTx, err = tu.prepareRewardTx(&tx)

	case transaction.TxTypeUnsigned:
		var tx smartContractResult.SmartContractResult
		err = tu.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		apiTx, err = tu.prepareUnsignedTx(&tx)
	}
	if err != nil {
		return nil, err
	}

	res := tu.dataFieldParser.Parse(apiTx.Data, apiTx.Tx.GetSndAddr(), apiTx.Tx.GetRcvAddr())
	apiTx.Operation = res.Operation
	apiTx.Function = res.Function
	apiTx.ESDTValues = res.ESDTValues
	apiTx.Tokens = res.Tokens
	apiTx.Receivers, err = tu.addressPubKeyConverter.EncodeSlice(res.Receivers)
	if err != nil {
		return nil, err
	}

	apiTx.ReceiversShardIDs = res.ReceiversShardID
	apiTx.IsRelayed = res.IsRelayed

	return apiTx, nil
}

func (tu *txUnmarshaller) prepareNormalTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	encodedRcvAddr, err := tu.addressPubKeyConverter.Encode(tx.RcvAddr)
	if err != nil {
		return nil, err
	}

	encodedSndAddr, err := tu.addressPubKeyConverter.Encode(tx.SndAddr)
	if err != nil {
		return nil, err
	}

	return &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeNormal),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         encodedRcvAddr,
		ReceiverUsername: tx.RcvUserName,
		Sender:           encodedSndAddr,
		SenderUsername:   tx.SndUserName,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
	}, nil
}

func (tu *txUnmarshaller) prepareInvalidTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	encodedRcvAddr, err := tu.addressPubKeyConverter.Encode(tx.RcvAddr)
	if err != nil {
		return nil, err
	}

	encodedSndAddr, err := tu.addressPubKeyConverter.Encode(tx.SndAddr)
	if err != nil {
		return nil, err
	}

	return &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeInvalid),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         encodedRcvAddr,
		ReceiverUsername: tx.RcvUserName,
		Sender:           encodedSndAddr,
		SenderUsername:   tx.SndUserName,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
	}, nil
}

func (tu *txUnmarshaller) prepareRewardTx(tx *rewardTxData.RewardTx) (*transaction.ApiTransactionResult, error) {
	encodedRcvAddr, err := tu.addressPubKeyConverter.Encode(tx.GetRcvAddr())
	if err != nil {
		return nil, err
	}

	return &transaction.ApiTransactionResult{
		Tx:          tx,
		Type:        string(transaction.TxTypeReward),
		Round:       tx.GetRound(),
		Epoch:       tx.GetEpoch(),
		Value:       tx.GetValue().String(),
		Sender:      "metachain",
		Receiver:    encodedRcvAddr,
		SourceShard: core.MetachainShardId,
	}, nil
}

func (tu *txUnmarshaller) prepareUnsignedTx(tx *smartContractResult.SmartContractResult) (*transaction.ApiTransactionResult, error) {
	encodedRcvAddr, err := tu.addressPubKeyConverter.Encode(tx.GetRcvAddr())
	if err != nil {
		return nil, err
	}

	encodedSndAddr, err := tu.addressPubKeyConverter.Encode(tx.GetSndAddr())
	if err != nil {
		return nil, err
	}

	txResult := &transaction.ApiTransactionResult{
		Tx:                      tx,
		Type:                    string(transaction.TxTypeUnsigned),
		Nonce:                   tx.GetNonce(),
		Value:                   tx.GetValue().String(),
		Receiver:                encodedRcvAddr,
		Sender:                  encodedSndAddr,
		GasPrice:                tx.GetGasPrice(),
		GasLimit:                tx.GetGasLimit(),
		Data:                    tx.GetData(),
		Code:                    string(tx.GetCode()),
		CodeMetadata:            tx.GetCodeMetadata(),
		PreviousTransactionHash: hex.EncodeToString(tx.GetPrevTxHash()),
		OriginalTransactionHash: hex.EncodeToString(tx.GetOriginalTxHash()),
		ReturnMessage:           string(tx.GetReturnMessage()),
	}
	encodedOriginalSndAddr, err := tu.addressPubKeyConverter.Encode(tx.GetOriginalSender())
	if err != nil {
		return nil, err
	}
	if len(tx.GetOriginalSender()) == tu.addressPubKeyConverter.Len() {
		txResult.OriginalSender = encodedOriginalSndAddr
	}

	return txResult, nil
}
