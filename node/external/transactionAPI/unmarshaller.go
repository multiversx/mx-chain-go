package transactionAPI

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	rewardTxData "github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
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

	senderAddress, err := tu.addressPubKeyConverter.Encode(rec.SndAddr)
	if err != nil {
		return nil, err
	}

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

	res := tu.dataFieldParser.Parse(apiTx.Data, apiTx.Tx.GetSndAddr(), apiTx.Tx.GetRcvAddr(), tu.shardCoordinator.NumberOfShards())
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
	receiverAddress, err := tu.addressPubKeyConverter.Encode(tx.RcvAddr)
	if err != nil {
		return nil, err
	}

	senderAddress, err := tu.addressPubKeyConverter.Encode(tx.SndAddr)
	if err != nil {
		return nil, err
	}

	return &transaction.ApiTransactionResult{
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
	}, nil
}

func (tu *txUnmarshaller) prepareInvalidTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	receiverAddress, err := tu.addressPubKeyConverter.Encode(tx.RcvAddr)
	if err != nil {
		return nil, err
	}

	senderAddress, err := tu.addressPubKeyConverter.Encode(tx.SndAddr)
	if err != nil {
		return nil, err
	}

	return &transaction.ApiTransactionResult{
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
	}, nil
}

func (tu *txUnmarshaller) prepareRewardTx(tx *rewardTxData.RewardTx) (*transaction.ApiTransactionResult, error) {
	receiverAddress, err := tu.addressPubKeyConverter.Encode(tx.GetRcvAddr())
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
		Receiver:    receiverAddress,
		SourceShard: core.MetachainShardId,
	}, nil
}

func (tu *txUnmarshaller) prepareUnsignedTx(tx *smartContractResult.SmartContractResult) (*transaction.ApiTransactionResult, error) {
	receiverAddress, err := tu.addressPubKeyConverter.Encode(tx.GetRcvAddr())
	if err != nil {
		return nil, err
	}

	senderAddress, err := tu.addressPubKeyConverter.Encode(tx.GetSndAddr())
	if err != nil {
		return nil, err
	}

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
		RelayedValue:            bigIntToStr(tx.GetRelayedValue()),
	}

	txResult.OriginalSender, err = tu.getEncodedAddress(tx.GetOriginalSender())
	if err != nil {
		return nil, err
	}

	txResult.RelayerAddress, err = tu.getEncodedAddress(tx.GetRelayerAddr())
	if err != nil {
		return nil, err
	}

	return txResult, nil
}

func (tu *txUnmarshaller) getEncodedAddress(address []byte) (string, error) {
	if len(address) == tu.addressPubKeyConverter.Len() {
		return tu.addressPubKeyConverter.Encode(address)
	}

	return "", nil
}

func bigIntToStr(value *big.Int) string {
	if value != nil {
		return value.String()
	}

	return ""
}
