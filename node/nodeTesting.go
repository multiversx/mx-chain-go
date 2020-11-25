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

var minTxGasPrice = uint64(10)
var minTxGasLimit = uint64(1000)

//TODO remove this file and adapt integration tests using GenerateAndSendBulkTransactions

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

	dataPacker, err := partitioning.NewSimpleDataPacker(n.coreComponents.InternalMarshalizer())
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

	// TODO: might think of a way to stop waiting at a signal
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
	identifier := factory.TransactionTopic + n.processComponents.ShardCoordinator().CommunicationIdentifier(senderShardId)

	packets, err := dataPacker.PackDataInChunks(txsBuff, core.MaxBulkTransactionSize)
	if err != nil {
		return err
	}

	atomic.AddInt32(&n.currentSendingGoRoutines, int32(len(packets)))
	for _, buff := range packets {
		go func(bufferToSend []byte) {
			err = n.networkComponents.NetworkMessenger().BroadcastOnChannelBlocking(
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
	if check.IfNil(n.cryptoComponents.TxSingleSigner()) {
		return ErrNilSingleSig
	}
	if check.IfNil(n.coreComponents.AddressPubKeyConverter()) {
		return ErrNilPubkeyConverter
	}
	if check.IfNil(n.processComponents.ShardCoordinator()) {
		return ErrNilShardCoordinator
	}
	if check.IfNil(n.stateComponents.AccountsAdapter()) {
		return ErrNilAccountsAdapter
	}

	return nil
}

func (n *Node) generateBulkTransactionsPrepareParams(receiverHex string, sk crypto.PrivateKey) (uint64, []byte, []byte, uint32, error) {
	senderAddressBytes, err := sk.GeneratePublic().ToByteArray()
	if err != nil {
		return 0, nil, nil, 0, err
	}

	receiverAddress, err := n.coreComponents.AddressPubKeyConverter().Decode(receiverHex)
	if err != nil {
		return 0, nil, nil, 0, errors.New("could not create receiver address from provided param: " + err.Error())
	}

	senderShardId := n.processComponents.ShardCoordinator().ComputeId(senderAddressBytes)

	newNonce := uint64(0)
	if senderShardId != n.processComponents.ShardCoordinator().SelfId() {
		return newNonce, senderAddressBytes, receiverAddress, senderShardId, nil
	}

	senderAccount, err := n.stateComponents.AccountsAdapter().GetExistingAccount(senderAddressBytes)
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
	if check.IfNil(n.coreComponents.InternalMarshalizer()) {
		return nil, nil, ErrNilMarshalizer
	}
	if check.IfNil(n.coreComponents.TxMarshalizer()) {
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

	marshalizedTx, err := tx.GetDataForSigning(n.coreComponents.AddressPubKeyConverter(), n.coreComponents.TxMarshalizer())
	if err != nil {
		return nil, nil, errors.New("could not marshal transaction")
	}

	sig, err := n.cryptoComponents.TxSingleSigner().Sign(sk, marshalizedTx)
	if err != nil {
		return nil, nil, errors.New("could not sign the transaction")
	}

	tx.Signature = sig
	txBuff, err := n.coreComponents.InternalMarshalizer().Marshal(&tx)
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

	signedMarshalizedTx, err := n.coreComponents.InternalMarshalizer().Marshal(&batch.Batch{Data: [][]byte{txBuff}})
	if err != nil {
		return nil, nil, errors.New("could not marshal signed transaction")
	}

	return tx, signedMarshalizedTx, nil
}

//GenerateTransaction generates a new transaction with sender, receiver, amount and code
func (n *Node) GenerateTransaction(senderHex string, receiverHex string, value *big.Int, transactionData string, privateKey crypto.PrivateKey, chainID []byte, minTxVersion uint32) (*transaction.Transaction, error) {
	if check.IfNil(n.coreComponents.AddressPubKeyConverter()) {
		return nil, ErrNilPubkeyConverter
	}
	if check.IfNil(n.stateComponents.AccountsAdapter()) {
		return nil, ErrNilAccountsAdapter
	}
	if check.IfNil(privateKey) {
		return nil, errors.New("initialize PrivateKey first")
	}

	receiverAddress, err := n.coreComponents.AddressPubKeyConverter().Decode(receiverHex)
	if err != nil {
		return nil, errors.New("could not create receiver address from provided param")
	}
	senderAddress, err := n.coreComponents.AddressPubKeyConverter().Decode(senderHex)
	if err != nil {
		return nil, errors.New("could not create sender address from provided param")
	}
	senderAccount, err := n.stateComponents.AccountsAdapter().GetExistingAccount(senderAddress)
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
