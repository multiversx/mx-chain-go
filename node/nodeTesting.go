package node

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

const maxGoRoutinesSendMessage = 30

//TODO move this funcs in a new benchmarking/stress-test binary

// GenerateAndSendBulkTransactions is a method for generating and propagating a set
// of transactions to be processed. It is mainly used for demo purposes
func (n *Node) GenerateAndSendBulkTransactions(receiverHex string, value *big.Int, noOfTxs uint64) error {
	if atomic.LoadInt32(&n.currentSendingGoRoutines) >= maxGoRoutinesSendMessage {
		return ErrSystemBusyGeneratingTransactions
	}

	err := n.generateBulkTransactionsChecks(noOfTxs)
	if err != nil {
		return err
	}

	newNonce, senderAddressBytes, recvAddressBytes, senderShardId, err := n.generateBulkTransactionsPrepareParams(receiverHex)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(int(noOfTxs))

	mutTransactions := sync.RWMutex{}
	transactions := make([][]byte, 0)

	mutErrFound := sync.Mutex{}
	var errFound error

	dataPacker, err := partitioning.NewSimpleDataPacker(n.marshalizer)
	if err != nil {
		return err
	}

	for nonce := newNonce; nonce < newNonce+noOfTxs; nonce++ {
		go func(crtNonce uint64) {
			_, signedTxBuff, err := n.generateAndSignSingleTx(
				crtNonce,
				value,
				recvAddressBytes,
				senderAddressBytes,
				"",
			)

			if err != nil {
				mutErrFound.Lock()
				errFound = errors.New(fmt.Sprintf("failure generating transaction %d: %s", crtNonce, err.Error()))
				mutErrFound.Unlock()

				wg.Done()
				return
			}

			mutTransactions.Lock()
			transactions = append(transactions, signedTxBuff)
			mutTransactions.Unlock()
			wg.Done()
		}(nonce)
	}

	wg.Wait()

	if errFound != nil {
		return errFound
	}

	if len(transactions) != int(noOfTxs) {
		return errors.New(fmt.Sprintf("generated only %d from required %d transactions", len(transactions), noOfTxs))
	}

	//the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.TransactionTopic + n.shardCoordinator.CommunicationIdentifier(senderShardId)

	packets, err := dataPacker.PackDataInChunks(transactions, core.MaxBulkTransactionSize)
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
				log.Error(err.Error())
			}

			atomic.AddInt32(&n.currentSendingGoRoutines, -1)
		}(buff)
	}

	return nil
}

// GenerateAndSendBulkTransactionsOneByOne is a method for generating and propagating a set
// of transactions to be processed. It is mainly used for demo purposes
func (n *Node) GenerateAndSendBulkTransactionsOneByOne(receiverHex string, value *big.Int, noOfTx uint64) error {
	err := n.generateBulkTransactionsChecks(noOfTx)
	if err != nil {
		return err
	}

	newNonce, senderAddressBytes, recvAddressBytes, senderShardId, err := n.generateBulkTransactionsPrepareParams(receiverHex)
	if err != nil {
		return err
	}

	generated := 0
	identifier := factory.TransactionTopic + n.shardCoordinator.CommunicationIdentifier(senderShardId)
	for nonce := newNonce; nonce < newNonce+noOfTx; nonce++ {
		_, signedTxBuff, err := n.generateAndSignTxBuffArray(
			nonce,
			value,
			recvAddressBytes,
			senderAddressBytes,
			"",
		)
		if err != nil {
			return err
		}

		generated++

		n.messenger.BroadcastOnChannel(
			SendTransactionsPipe,
			identifier,
			signedTxBuff,
		)
	}

	if generated != int(noOfTx) {
		return errors.New(fmt.Sprintf("generated only %d from required %d transactions", generated, noOfTx))
	}

	return nil
}

func (n *Node) generateBulkTransactionsChecks(noOfTx uint64) error {
	if noOfTx == 0 {
		return errors.New("can not generate and broadcast 0 transactions")
	}
	if n.txSignPubKey == nil {
		return ErrNilPublicKey
	}
	if n.txSingleSigner == nil {
		return ErrNilSingleSig
	}
	if n.addrConverter == nil {
		return ErrNilAddressConverter
	}
	if n.shardCoordinator == nil {
		return ErrNilShardCoordinator
	}
	if n.accounts == nil {
		return ErrNilAccountsAdapter
	}

	return nil
}

func (n *Node) generateBulkTransactionsPrepareParams(receiverHex string) (uint64, []byte, []byte, uint32, error) {
	senderAddressBytes, err := n.txSignPubKey.ToByteArray()
	if err != nil {
		return 0, nil, nil, 0, err
	}

	senderAddress, err := n.addrConverter.CreateAddressFromPublicKeyBytes(senderAddressBytes)
	if err != nil {
		return 0, nil, nil, 0, err
	}

	receiverAddress, err := n.addrConverter.CreateAddressFromHex(receiverHex)
	if err != nil {
		return 0, nil, nil, 0, errors.New("could not create receiver address from provided param: " + err.Error())
	}

	senderShardId := n.shardCoordinator.ComputeId(senderAddress)

	newNonce := uint64(0)
	if senderShardId != n.shardCoordinator.SelfId() {
		return newNonce, senderAddressBytes, receiverAddress.Bytes(), senderShardId, nil
	}

	senderAccount, err := n.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return 0, nil, nil, 0, errors.New("could not fetch sender account from provided param: " + err.Error())
	}

	acc, ok := senderAccount.(*state.Account)
	if !ok {
		return 0, nil, nil, 0, errors.New("wrong account type")
	}
	newNonce = acc.Nonce

	return newNonce, senderAddressBytes, receiverAddress.Bytes(), senderShardId, nil
}

func (n *Node) generateAndSignSingleTx(
	nonce uint64,
	value *big.Int,
	rcvAddrBytes []byte,
	sndAddrBytes []byte,
	data string,
) (*transaction.Transaction, []byte, error) {

	if n.marshalizer == nil {
		return nil, nil, ErrNilMarshalizer
	}
	if n.txSignPrivKey == nil {
		return nil, nil, ErrNilPrivateKey
	}

	tx := transaction.Transaction{
		Nonce:    nonce,
		Value:    value,
		GasLimit: 100,
		GasPrice: 10,
		RcvAddr:  rcvAddrBytes,
		SndAddr:  sndAddrBytes,
		Data:     []byte(data),
	}

	marshalizedTx, err := n.marshalizer.Marshal(&tx)
	if err != nil {
		return nil, nil, errors.New("could not marshal transaction")
	}

	sig, err := n.txSingleSigner.Sign(n.txSignPrivKey, marshalizedTx)
	if err != nil {
		return nil, nil, errors.New("could not sign the transaction")
	}

	tx.Signature = sig
	txBuff, err := n.marshalizer.Marshal(&tx)
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
) (*transaction.Transaction, []byte, error) {

	tx, txBuff, err := n.generateAndSignSingleTx(nonce, value, rcvAddrBytes, sndAddrBytes, data)
	if err != nil {
		return nil, nil, err
	}

	signedMarshalizedTx, err := n.marshalizer.Marshal([][]byte{txBuff})
	if err != nil {
		return nil, nil, errors.New("could not marshal signed transaction")
	}

	return tx, signedMarshalizedTx, nil
}

//GenerateTransaction generates a new transaction with sender, receiver, amount and code
func (n *Node) GenerateTransaction(senderHex string, receiverHex string, value *big.Int, transactionData string) (*transaction.Transaction, error) {
	if n.addrConverter == nil || n.addrConverter.IsInterfaceNil() {
		return nil, ErrNilAddressConverter
	}
	if n.accounts == nil || n.accounts.IsInterfaceNil() {
		return nil, ErrNilAccountsAdapter
	}
	if n.txSignPrivKey == nil {
		return nil, errors.New("initialize PrivateKey first")
	}

	receiverAddress, err := n.addrConverter.CreateAddressFromHex(receiverHex)
	if err != nil {
		return nil, errors.New("could not create receiver address from provided param")
	}
	senderAddress, err := n.addrConverter.CreateAddressFromHex(senderHex)
	if err != nil {
		return nil, errors.New("could not create sender address from provided param")
	}
	senderAccount, err := n.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param")
	}

	newNonce := uint64(0)
	acc, ok := senderAccount.(*state.Account)
	if !ok {
		return nil, errors.New("wrong account type")
	}
	newNonce = acc.Nonce

	tx, _, err := n.generateAndSignTxBuffArray(
		newNonce,
		value,
		receiverAddress.Bytes(),
		senderAddress.Bytes(),
		transactionData)

	return tx, err
}
