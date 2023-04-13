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
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
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

	return &transaction.ApiReceipt{
		Value:   rec.Value,
		SndAddr: tu.addressPubKeyConverter.Encode(rec.SndAddr),
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
	apiTx.Receivers = datafield.EncodeBytesSlice(tu.addressPubKeyConverter.Encode, res.Receivers)
	apiTx.ReceiversShardIDs = res.ReceiversShardID
	apiTx.IsRelayed = res.IsRelayed

	return apiTx, nil
}

func (tu *txUnmarshaller) prepareNormalTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	apiTx := &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeNormal),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         tu.addressPubKeyConverter.Encode(tx.RcvAddr),
		ReceiverUsername: tx.RcvUserName,
		Sender:           tu.addressPubKeyConverter.Encode(tx.SndAddr),
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
		apiTx.GuardianAddr = tu.addressPubKeyConverter.Encode(tx.GuardianAddr)
		apiTx.GuardianSignature = hex.EncodeToString(tx.GuardianSignature)
	}

	return apiTx, nil
}

func (tu *txUnmarshaller) prepareInvalidTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	apiTx := &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeInvalid),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         tu.addressPubKeyConverter.Encode(tx.RcvAddr),
		ReceiverUsername: tx.RcvUserName,
		Sender:           tu.addressPubKeyConverter.Encode(tx.SndAddr),
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
		apiTx.GuardianAddr = tu.addressPubKeyConverter.Encode(tx.GuardianAddr)
		apiTx.GuardianSignature = hex.EncodeToString(tx.GuardianSignature)
	}

	return apiTx, nil
}

func (tu *txUnmarshaller) prepareRewardTx(tx *rewardTxData.RewardTx) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Tx:          tx,
		Type:        string(transaction.TxTypeReward),
		Round:       tx.GetRound(),
		Epoch:       tx.GetEpoch(),
		Value:       tx.GetValue().String(),
		Sender:      "metachain",
		Receiver:    tu.addressPubKeyConverter.Encode(tx.GetRcvAddr()),
		SourceShard: core.MetachainShardId,
	}, nil
}

func (tu *txUnmarshaller) prepareUnsignedTx(tx *smartContractResult.SmartContractResult) (*transaction.ApiTransactionResult, error) {
	txResult := &transaction.ApiTransactionResult{
		Tx:                      tx,
		Type:                    string(transaction.TxTypeUnsigned),
		Nonce:                   tx.GetNonce(),
		Value:                   tx.GetValue().String(),
		Receiver:                tu.addressPubKeyConverter.Encode(tx.GetRcvAddr()),
		Sender:                  tu.addressPubKeyConverter.Encode(tx.GetSndAddr()),
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

	return txResult, nil
}

func (tu *txUnmarshaller) getEncodedAddress(address []byte) string {
	if len(address) == tu.addressPubKeyConverter.Len() {
		return tu.addressPubKeyConverter.Encode(address)
	}

	return ""
}

func bigIntToStr(value *big.Int) string {
	if value != nil {
		return value.String()
	}

	return ""
}
