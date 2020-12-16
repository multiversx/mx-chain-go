package node

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

const maxGoRoutinesSendMessage = 30

var minTxGasPrice = uint64(100)
var minTxGasLimit = uint64(1000)

//TODO move this funcs in a new benchmarking/stress-test binary

// GenerateAndSendBulkTransactions is a method for generating and propagating a set
// of transactions to be processed. It is mainly used for demo purposes
func (n *Node) GenerateAndSendBulkTransactions(
	receiverHex string,
	value *big.Int,
	numOfTxs uint64,
	sk crypto.PrivateKey,
	whiteList func([]*transaction.Transaction),
	chainID []byte,
	minTxVersion uint32,
) error {
	if sk == nil {
		return ErrNilPrivateKey
	}

	if atomic.LoadInt32(&n.currentSendingGoRoutines) >= maxGoRoutinesSendMessage {
		return ErrSystemBusyGeneratingTransactions
	}

	err := n.generateBulkTransactionsChecks(numOfTxs)
	if err != nil {
		return err
	}

	newNonce, senderAddressBytes, recvAddressBytes, senderShardId, err := n.generateBulkTransactionsPrepareParams(receiverHex, sk)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(int(numOfTxs))

	mutTransactions := sync.RWMutex{}
	txsBuff := make([][]byte, 0)

	mutErrFound := sync.Mutex{}
	var errFound error

	dataPacker, err := partitioning.NewSimpleDataPacker(n.internalMarshalizer)
	if err != nil {
		return err
	}

	txs := make([]*transaction.Transaction, 0)
	for nonce := newNonce; nonce < newNonce+numOfTxs; nonce++ {
		go func(crtNonce uint64) {
			tx, txBuff, errGenTx := n.generateAndSignSingleTx(
				crtNonce,
				value,
				recvAddressBytes,
				senderAddressBytes,
				"",
				sk,
				chainID,
				minTxVersion,
			)

			if errGenTx != nil {
				mutErrFound.Lock()
				errFound = fmt.Errorf("failure generating transaction %d: %s", crtNonce, errGenTx.Error())
				mutErrFound.Unlock()

				wg.Done()
				return
			}

			mutTransactions.Lock()
			txsBuff = append(txsBuff, txBuff)
			txs = append(txs, tx)
			mutTransactions.Unlock()
			wg.Done()
		}(nonce)
	}

	wg.Wait()

	if errFound != nil {
		return errFound
	}

	if len(txsBuff) != int(numOfTxs) {
		return fmt.Errorf("generated only %d from required %d transactions", len(txsBuff), numOfTxs)
	}

	if whiteList != nil {
		whiteList(txs)
	}

	//the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.TransactionTopic + n.shardCoordinator.CommunicationIdentifier(senderShardId)

	packets, err := dataPacker.PackDataInChunks(txsBuff, core.MaxBulkTransactionSize)
	if err != nil {
		return err
	}

	atomic.AddInt32(&n.currentSendingGoRoutines, int32(len(packets)))
	for _, buff := range packets {
		go func(bufferToSend []byte) {
			err = n.messenger.BroadcastOnChannelBlocking(
				SendTransactionsPipe,
				identifier,
				bufferToSend,
			)
			if err != nil {
				log.Debug("BroadcastOnChannelBlocking", "error", err.Error())
			}

			atomic.AddInt32(&n.currentSendingGoRoutines, -1)
		}(buff)
	}

	return nil
}

func (n *Node) generateBulkTransactionsChecks(numOfTxs uint64) error {
	if numOfTxs == 0 {
		return errors.New("can not generate and broadcast 0 transactions")
	}
	if check.IfNil(n.txSingleSigner) {
		return ErrNilSingleSig
	}
	if check.IfNil(n.addressPubkeyConverter) {
		return ErrNilPubkeyConverter
	}
	if check.IfNil(n.shardCoordinator) {
		return ErrNilShardCoordinator
	}
	if check.IfNil(n.accounts) {
		return ErrNilAccountsAdapter
	}

	return nil
}

func (n *Node) generateBulkTransactionsPrepareParams(receiverHex string, sk crypto.PrivateKey) (uint64, []byte, []byte, uint32, error) {
	senderAddressBytes, err := sk.GeneratePublic().ToByteArray()
	if err != nil {
		return 0, nil, nil, 0, err
	}

	receiverAddress, err := n.addressPubkeyConverter.Decode(receiverHex)
	if err != nil {
		return 0, nil, nil, 0, errors.New("could not create receiver address from provided param: " + err.Error())
	}

	senderShardId := n.shardCoordinator.ComputeId(senderAddressBytes)

	newNonce := uint64(0)
	if senderShardId != n.shardCoordinator.SelfId() {
		return newNonce, senderAddressBytes, receiverAddress, senderShardId, nil
	}

	senderAccount, err := n.accounts.GetExistingAccount(senderAddressBytes)
	if err != nil {
		return 0, nil, nil, 0, errors.New("could not fetch sender account from provided param: " + err.Error())
	}

	acc, ok := senderAccount.(state.UserAccountHandler)
	if !ok {
		return 0, nil, nil, 0, errors.New("wrong account type")
	}
	newNonce = acc.GetNonce()

	return newNonce, senderAddressBytes, receiverAddress, senderShardId, nil
}

func (n *Node) generateAndSignSingleTx(
	nonce uint64,
	value *big.Int,
	rcvAddrBytes []byte,
	sndAddrBytes []byte,
	dataField string,
	sk crypto.PrivateKey,
	chainID []byte,
	minTxVersion uint32,
) (*transaction.Transaction, []byte, error) {
	if check.IfNil(n.internalMarshalizer) {
		return nil, nil, ErrNilMarshalizer
	}
	if check.IfNil(n.txSignMarshalizer) {
		return nil, nil, ErrNilMarshalizer
	}
	if check.IfNil(sk) {
		return nil, nil, ErrNilPrivateKey
	}

	tx := transaction.Transaction{
		Nonce:    nonce,
		Value:    new(big.Int).Set(value),
		GasLimit: minTxGasLimit,
		GasPrice: minTxGasPrice,
		RcvAddr:  rcvAddrBytes,
		SndAddr:  sndAddrBytes,
		Data:     []byte(dataField),
		ChainID:  chainID,
		Version:  minTxVersion,
	}

	marshalizedTx, err := tx.GetDataForSigning(n.addressPubkeyConverter, n.txSignMarshalizer)
	if err != nil {
		return nil, nil, errors.New("could not marshal transaction")
	}

	sig, err := n.txSingleSigner.Sign(sk, marshalizedTx)
	if err != nil {
		return nil, nil, errors.New("could not sign the transaction")
	}

	tx.Signature = sig
	txBuff, err := n.internalMarshalizer.Marshal(&tx)
	if err != nil {
		return nil, nil, err
	}

	return &tx, txBuff, err
}

func (n *Node) generateAndSignTxBuffArray(
	nonce uint64,
	value *big.Int,
	rcvAddrBytes []byte,
	sndAddrBytes []byte,
	data string,
	sk crypto.PrivateKey,
	chainID []byte,
	minTxVersion uint32,
) (*transaction.Transaction, []byte, error) {
	tx, txBuff, err := n.generateAndSignSingleTx(nonce, value, rcvAddrBytes, sndAddrBytes, data, sk, chainID, minTxVersion)
	if err != nil {
		return nil, nil, err
	}

	signedMarshalizedTx, err := n.internalMarshalizer.Marshal(&batch.Batch{Data: [][]byte{txBuff}})
	if err != nil {
		return nil, nil, errors.New("could not marshal signed transaction")
	}

	return tx, signedMarshalizedTx, nil
}

//GenerateTransaction generates a new transaction with sender, receiver, amount and code
func (n *Node) GenerateTransaction(senderHex string, receiverHex string, value *big.Int, transactionData string, privateKey crypto.PrivateKey, chainID []byte, minTxVersion uint32) (*transaction.Transaction, error) {
	if check.IfNil(n.addressPubkeyConverter) {
		return nil, ErrNilPubkeyConverter
	}
	if check.IfNil(n.accounts) {
		return nil, ErrNilAccountsAdapter
	}
	if check.IfNil(privateKey) {
		return nil, errors.New("initialize PrivateKey first")
	}

	receiverAddress, err := n.addressPubkeyConverter.Decode(receiverHex)
	if err != nil {
		return nil, errors.New("could not create receiver address from provided param")
	}
	senderAddress, err := n.addressPubkeyConverter.Decode(senderHex)
	if err != nil {
		return nil, errors.New("could not create sender address from provided param")
	}
	senderAccount, err := n.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param")
	}

	newNonce := uint64(0)
	acc, ok := senderAccount.(state.UserAccountHandler)
	if !ok {
		return nil, errors.New("wrong account type")
	}
	newNonce = acc.GetNonce()

	tx, _, err := n.generateAndSignTxBuffArray(
		newNonce,
		value,
		receiverAddress,
		senderAddress,
		transactionData,
		privateKey,
		chainID,
		minTxVersion,
	)

	return tx, err
}
