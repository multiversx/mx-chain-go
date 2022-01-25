package transactionAPI

import (
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go-core/core"
	rewardTxData "github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

type unmarshalerAndPreparer struct {
	addressPubKeyConverter core.PubkeyConverter
	marshalizer            marshal.Marshalizer
}

func newTransactionUnmashalerAndPreparer(marshalizer marshal.Marshalizer, addressPubKeyConverter core.PubkeyConverter) *unmarshalerAndPreparer {
	return &unmarshalerAndPreparer{
		marshalizer:            marshalizer,
		addressPubKeyConverter: addressPubKeyConverter,
	}
}

func (up *unmarshalerAndPreparer) unmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	switch txType {
	case transaction.TxTypeNormal:
		var tx transaction.Transaction
		err := up.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return up.prepareNormalTx(&tx)
	case transaction.TxTypeInvalid:
		var tx transaction.Transaction
		err := up.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return up.prepareInvalidTx(&tx)
	case transaction.TxTypeReward:
		var tx rewardTxData.RewardTx
		err := up.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return up.prepareRewardTx(&tx)

	case transaction.TxTypeUnsigned:
		var tx smartContractResult.SmartContractResult
		err := up.marshalizer.Unmarshal(&tx, txBytes)
		if err != nil {
			return nil, err
		}
		return up.prepareUnsignedTx(&tx)
	}

	return &transaction.ApiTransactionResult{Type: string(transaction.TxTypeInvalid)}, nil // this shouldn't happen
}

func (up *unmarshalerAndPreparer) prepareNormalTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeNormal),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         up.addressPubKeyConverter.Encode(tx.RcvAddr),
		ReceiverUsername: tx.RcvUserName,
		Sender:           up.addressPubKeyConverter.Encode(tx.SndAddr),
		SenderUsername:   tx.SndUserName,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
	}, nil
}

func (up *unmarshalerAndPreparer) prepareInvalidTx(tx *transaction.Transaction) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Tx:               tx,
		Type:             string(transaction.TxTypeInvalid),
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         up.addressPubKeyConverter.Encode(tx.RcvAddr),
		ReceiverUsername: tx.RcvUserName,
		Sender:           up.addressPubKeyConverter.Encode(tx.SndAddr),
		SenderUsername:   tx.SndUserName,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
	}, nil
}

func (up *unmarshalerAndPreparer) prepareRewardTx(tx *rewardTxData.RewardTx) (*transaction.ApiTransactionResult, error) {
	return &transaction.ApiTransactionResult{
		Tx:          tx,
		Type:        string(transaction.TxTypeReward),
		Round:       tx.GetRound(),
		Epoch:       tx.GetEpoch(),
		Value:       tx.GetValue().String(),
		Sender:      "metachain",
		Receiver:    up.addressPubKeyConverter.Encode(tx.GetRcvAddr()),
		SourceShard: core.MetachainShardId,
	}, nil
}

func (up *unmarshalerAndPreparer) prepareUnsignedTx(tx *smartContractResult.SmartContractResult) (*transaction.ApiTransactionResult, error) {
	txResult := &transaction.ApiTransactionResult{
		Tx:                      tx,
		Type:                    string(transaction.TxTypeUnsigned),
		Nonce:                   tx.GetNonce(),
		Value:                   tx.GetValue().String(),
		Receiver:                up.addressPubKeyConverter.Encode(tx.GetRcvAddr()),
		Sender:                  up.addressPubKeyConverter.Encode(tx.GetSndAddr()),
		GasPrice:                tx.GetGasPrice(),
		GasLimit:                tx.GetGasLimit(),
		Data:                    tx.GetData(),
		Code:                    string(tx.GetCode()),
		CodeMetadata:            tx.GetCodeMetadata(),
		PreviousTransactionHash: hex.EncodeToString(tx.GetPrevTxHash()),
		OriginalTransactionHash: hex.EncodeToString(tx.GetOriginalTxHash()),
		ReturnMessage:           string(tx.GetReturnMessage()),
	}
	if len(tx.GetOriginalSender()) == up.addressPubKeyConverter.Len() {
		txResult.OriginalSender = up.addressPubKeyConverter.Encode(tx.GetOriginalSender())
	}

	return txResult, nil
}
