package transactions

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

func isScResultSuccessful(scResultData []byte) bool {
	okReturnDataNewVersion := []byte("@" + hex.EncodeToString([]byte(vmcommon.Ok.String())))
	okReturnDataOldVersion := []byte("@" + vmcommon.Ok.String()) // backwards compatible
	return bytes.Contains(scResultData, okReturnDataNewVersion) || bytes.Contains(scResultData, okReturnDataOldVersion)
}

func isSCRForSenderWithRefund(dbScResult *types.ScResult, tx *types.Transaction) bool {
	isForSender := dbScResult.Receiver == tx.Sender
	isRightNonce := dbScResult.Nonce == tx.Nonce+1
	isFromCurrentTx := dbScResult.PreTxHash == tx.Hash
	isScrDataOk := isDataOk(dbScResult.Data)

	return isFromCurrentTx && isForSender && isRightNonce && isScrDataOk
}

func isDataOk(data []byte) bool {
	okEncoded := hex.EncodeToString([]byte("ok"))
	dataFieldStr := "@" + okEncoded

	return strings.HasPrefix(string(data), dataFieldStr)
}

func stringValueToBigInt(strValue string) *big.Int {
	value, ok := big.NewInt(0).SetString(strValue, 10)
	if !ok {
		return big.NewInt(0)
	}

	return value
}

func isRelayedTx(tx *types.Transaction) bool {
	return strings.HasPrefix(string(tx.Data), "relayedTx") && len(tx.SmartContractResults) > 0
}

func addToAlteredAddresses(
	tx *types.Transaction,
	alteredAddresses map[string]*types.AlteredAccount,
	miniBlock *block.MiniBlock,
	selfShardID uint32,
	isRewardTx bool,
) {
	isESDTTx := tx.EsdtTokenIdentifier != "" && tx.EsdtValue != ""

	if selfShardID == miniBlock.SenderShardID && !isRewardTx {
		alteredAddresses[tx.Sender] = &types.AlteredAccount{
			IsSender:        true,
			IsESDTOperation: isESDTTx,
			TokenIdentifier: tx.EsdtTokenIdentifier,
		}
	}

	if tx.Status == transaction.TxStatusInvalid.String() {
		// ignore receiver if we have an invalid transaction
		return
	}

	if selfShardID == miniBlock.ReceiverShardID || miniBlock.ReceiverShardID == core.AllShardId {
		alteredAddresses[tx.Receiver] = &types.AlteredAccount{
			IsSender:        false,
			IsESDTOperation: isESDTTx,
			TokenIdentifier: tx.EsdtTokenIdentifier,
		}
	}
}

func isCrossShardDstMe(tx *types.Transaction, selfShardID uint32) bool {
	return tx.SenderShard != tx.ReceiverShard && tx.ReceiverShard == selfShardID
}

func isIntraShardOrInvalid(tx *types.Transaction, selfShardID uint32) bool {
	return (tx.SenderShard == tx.ReceiverShard && tx.ReceiverShard == selfShardID) || tx.Status == transaction.TxStatusInvalid.String()
}
