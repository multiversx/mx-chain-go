package node

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	gosync "sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	block2 "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/pkg/errors"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(2000 * time.Millisecond)

// ConsensusTopic is the topic used in consensus algorithm
const ConsensusTopic topicName = "consensus"

type topicName string

var log = logger.NewDefaultLogger()

// Option represents a functional configuration parameter that can operate
//  over the None struct.
type Option func(*Node) error

// Node is a structure that passes the configuration parameters and initializes
//  required services as requested
type Node struct {
	marshalizer              marshal.Marshalizer
	ctx                      context.Context
	hasher                   hashing.Hasher
	initialNodesPubkeys      []string
	initialNodesBalances     map[string]*big.Int
	roundDuration            uint64
	consensusGroupSize       int
	messenger                p2p.Messenger
	syncer                   ntp.SyncTimer
	blockProcessor           process.BlockProcessor
	genesisTime              time.Time
	elasticSubrounds         bool
	accounts                 state.AccountsAdapter
	addrConverter            state.AddressConverter
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	processorCreator         process.ProcessorFactory

	privateKey       crypto.PrivateKey
	publicKey        crypto.PublicKey
	singleSignKeyGen crypto.KeyGenerator
	singlesig        crypto.SingleSigner
	multisig         crypto.MultiSigner
	forkDetector     process.ForkDetector

	blkc             *blockchain.BlockChain
	dataPool         data.TransientDataHolder
	shardCoordinator sharding.ShardCoordinator

	isRunning bool
}

// ApplyOptions can set up different configurable options of a Node instance
func (n *Node) ApplyOptions(opts ...Option) error {
	if n.IsRunning() {
		return errors.New("cannot apply options while node is running")
	}
	for _, opt := range opts {
		err := opt(n)
		if err != nil {
			return errors.New("error applying option: " + err.Error())
		}
	}
	return nil
}

// NewNode creates a new Node instance
func NewNode(opts ...Option) (*Node, error) {
	node := &Node{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		err := opt(node)
		if err != nil {
			return nil, errors.New("error applying option: " + err.Error())
		}
	}
	return node, nil
}

// IsRunning will return the current state of the node
func (n *Node) IsRunning() bool {
	return n.isRunning
}

// Start will create a new messenger and and set up the Node state as running
func (n *Node) Start() error {
	err := n.P2PBootstrap()
	if err == nil {
		n.isRunning = true
	}
	return err
}

// Stop closes the messenger and undos everything done in Start
func (n *Node) Stop() error {
	if !n.IsRunning() {
		return nil
	}
	err := n.messenger.Close()
	if err != nil {
		return err
	}

	n.messenger = nil
	return nil
}

// P2PBootstrap will try to connect to many peers as possible
func (n *Node) P2PBootstrap() error {
	if n.messenger == nil {
		return ErrNilMessenger
	}
	n.messenger.Bootstrap(n.ctx)
	return nil
}

// CreateShardedStores instantiate sharded cachers for Transactions and Headers
func (n *Node) CreateShardedStores() error {
	if n.shardCoordinator == nil {
		return ErrNilShardCoordinator
	}

	if n.dataPool == nil {
		return ErrNilDataPool
	}

	transactionsDataStore := n.dataPool.Transactions()
	headersDataStore := n.dataPool.Headers()

	if transactionsDataStore == nil {
		return errors.New("nil transaction sharded data store")
	}

	if headersDataStore == nil {
		return errors.New("nil header sharded data store")
	}

	shards := n.shardCoordinator.NoShards()

	for i := uint32(0); i < shards; i++ {
		transactionsDataStore.CreateShardStore(i)
		headersDataStore.CreateShardStore(i)
	}

	return nil
}

// StartConsensus will start the consesus service for the current node
func (n *Node) StartConsensus() error {

	genesisHeader, genesisHeaderHash, err := n.createGenesisBlock()
	if err != nil {
		return err
	}
	n.blkc.GenesisBlock = genesisHeader
	n.blkc.GenesisHeaderHash = genesisHeaderHash

	round := n.createRound()
	chr := n.createChronology(round)

	boot, err := n.createBootstrap(round)

	if err != nil {
		return err
	}

	rndc := n.createRoundConsensus()
	rth := n.createRoundThreshold()
	rnds := n.createRoundStatus()
	cns := n.createConsensus(rndc, rth, rnds, chr)
	sposWrk, err := n.createConsensusWorker(cns, boot)

	if err != nil {
		return err
	}

	topic := n.createConsensusTopic(sposWrk)

	err = n.messenger.AddTopic(topic)

	if err != nil {
		log.Debug(fmt.Sprintf(err.Error()))
	}

	n.addSubroundsToChronology(sposWrk)

	go sposWrk.Cns.Chr.StartRounds()

	return nil
}

// GetBalance gets the balance for a specific address
func (n *Node) GetBalance(addressHex string) (*big.Int, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}

	address, err := n.addrConverter.CreateAddressFromHex(addressHex)
	if err != nil {
		return nil, errors.New("invalid address, could not decode from hex: " + err.Error())
	}
	account, err := n.accounts.GetExistingAccount(address)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param: " + err.Error())
	}

	if account == nil {
		return big.NewInt(0), nil
	}

	return account.BaseAccount().Balance, nil
}

// GenerateAndSendBulkTransactions is a method for generating and propagating a set
// of transactions to be processed. It is mainly used for demo purposes
func (n *Node) GenerateAndSendBulkTransactions(receiverHex string, value *big.Int, noOfTx uint64) error {
	if noOfTx == 0 {
		return errors.New("can not generate and broadcast 0 transactions")
	}

	if n.publicKey == nil {
		return ErrNilPublicKey
	}

	if n.singlesig == nil {
		return ErrNilSingleSig
	}

	senderAddressBytes, err := n.publicKey.ToByteArray()
	if err != nil {
		return err
	}

	if n.addrConverter == nil {
		return ErrNilAddressConverter
	}
	senderAddress, err := n.addrConverter.CreateAddressFromPublicKeyBytes(senderAddressBytes)
	if err != nil {
		return err
	}

	receiverAddress, err := n.addrConverter.CreateAddressFromHex(receiverHex)
	if err != nil {
		return errors.New("could not create receiver address from provided param: " + err.Error())
	}

	if n.accounts == nil {
		return ErrNilAccountsAdapter
	}
	senderAccount, err := n.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return errors.New("could not fetch sender account from provided param: " + err.Error())
	}
	newNonce := uint64(0)
	if senderAccount != nil {
		newNonce = senderAccount.BaseAccount().Nonce
	}

	wg := gosync.WaitGroup{}
	wg.Add(int(noOfTx))

	mutTransactions := gosync.RWMutex{}
	transactions := make([][]byte, 0)

	mutErrFound := gosync.Mutex{}
	var errFound error

	for nonce := newNonce; nonce < newNonce+noOfTx; nonce++ {
		go func(crtNonce uint64) {
			_, signedTxBuff, err := n.generateAndSignTx(
				crtNonce,
				value,
				receiverAddress.Bytes(),
				senderAddressBytes,
				nil,
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

	topic := n.messenger.GetTopic(string(factory.TransactionTopic))
	if topic == nil {
		return errors.New("could not get transaction topic")
	}

	if len(transactions) != int(noOfTx) {
		return errors.New(fmt.Sprintf("generated only %d from required %d transactions", len(transactions), noOfTx))
	}

	for i := 0; i < len(transactions); i++ {
		err = topic.BroadcastBuff(transactions[i])
		time.Sleep(time.Microsecond * 100)

		if err != nil {
			return errors.New("could not broadcast transaction: " + err.Error())
		}
	}

	return nil
}

// createRound method creates a round object
func (n *Node) createRound() *chronology.Round {
	rnd := chronology.NewRound(
		n.genesisTime,
		n.syncer.CurrentTime(n.syncer.ClockOffset()),
		time.Millisecond*time.Duration(n.roundDuration))

	return rnd
}

// createChronology method creates a chronology object
func (n *Node) createChronology(round *chronology.Round) *chronology.Chronology {
	chr := chronology.NewChronology(
		!n.elasticSubrounds,
		round,
		n.genesisTime,
		n.syncer)

	return chr
}

func (n *Node) createBootstrap(round *chronology.Round) (*sync.Bootstrap, error) {
	bootstrap, err := sync.NewBootstrap(n.dataPool, n.blkc, round, n.blockProcessor, WaitTime, n.marshalizer, n.forkDetector)

	if err != nil {
		return nil, err
	}

	resH, err := n.processorCreator.ResolverContainer().Get(string(factory.HeadersTopic))
	if err != nil {
		return nil, errors.New("cannot find headers topic resolver")
	}
	hdrRes := resH.(*block2.HeaderResolver)

	resT, err := n.processorCreator.ResolverContainer().Get(string(factory.TxBlockBodyTopic))
	if err != nil {
		return nil, errors.New("cannot find tx block body topic resolver")

	}
	gbbrRes := resT.(*block2.GenericBlockBodyResolver)

	bootstrap.RequestHeaderHandler = createRequestHeaderHandler(hdrRes)
	bootstrap.RequestTxBodyHandler = cerateRequestTxBodyHandler(gbbrRes)

	bootstrap.StartSync()

	return bootstrap, nil
}

func createRequestHeaderHandler(hdrRes *block2.HeaderResolver) func(nonce uint64) {
	return func(nonce uint64) {
		err := hdrRes.RequestHeaderFromNonce(nonce)

		log.Info(fmt.Sprintf("requested header with nonce %d from network\n", nonce))
		if err != nil {
			log.Error("RequestHeaderFromNonce error:  ", err.Error())
		}
	}
}

func cerateRequestTxBodyHandler(gbbrRes *block2.GenericBlockBodyResolver) func(hash []byte) {
	return func(hash []byte) {
		err := gbbrRes.RequestBlockBodyFromHash(hash)

		log.Info(fmt.Sprintf("requested tx body with hash %s from network\n", toB64(hash)))
		if err != nil {
			log.Error("RequestBlockBodyFromHash error: ", err.Error())
			return
		}
	}
}

// createRoundConsensus method creates a RoundConsensus object
func (n *Node) createRoundConsensus() *spos.RoundConsensus {

	selfId, err := n.publicKey.ToByteArray()

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	nodes := n.initialNodesPubkeys[0:n.consensusGroupSize]
	rndc := spos.NewRoundConsensus(
		nodes,
		string(selfId))

	rndc.ResetRoundState()

	return rndc
}

// createRoundThreshold method creates a RoundThreshold object
func (n *Node) createRoundThreshold() *spos.RoundThreshold {
	rth := spos.NewRoundThreshold()

	pbftThreshold := n.consensusGroupSize*2/3 + 1

	rth.SetThreshold(spos.SrBlock, 1)
	rth.SetThreshold(spos.SrCommitmentHash, pbftThreshold)
	rth.SetThreshold(spos.SrBitmap, pbftThreshold)
	rth.SetThreshold(spos.SrCommitment, pbftThreshold)
	rth.SetThreshold(spos.SrSignature, pbftThreshold)

	return rth
}

// createRoundStatus method creates a RoundStatus object
func (n *Node) createRoundStatus() *spos.RoundStatus {
	rnds := spos.NewRoundStatus()

	rnds.ResetRoundStatus()

	return rnds
}

// createConsensus method creates a Consensus object
func (n *Node) createConsensus(rndc *spos.RoundConsensus, rth *spos.RoundThreshold, rnds *spos.RoundStatus, chr *chronology.Chronology,
) *spos.Consensus {
	cns := spos.NewConsensus(
		nil,
		rndc,
		rth,
		rnds,
		chr)

	return cns
}

// createConsensusWorker method creates a ConsensusWorker object
func (n *Node) createConsensusWorker(cns *spos.Consensus, boot *sync.Bootstrap) (*spos.SPOSConsensusWorker, error) {
	sposWrk, err := spos.NewConsensusWorker(
		cns,
		n.blkc,
		n.hasher,
		n.marshalizer,
		n.blockProcessor,
		boot,
		n.singlesig,
		n.multisig,
		n.singleSignKeyGen,
		n.privateKey,
		n.publicKey,
	)

	if err != nil {
		return nil, err
	}

	sposWrk.SendMessage = n.sendMessage
	sposWrk.BroadcastBlockBody = n.broadcastBlockBody
	sposWrk.BroadcastHeader = n.broadcastHeader

	return sposWrk, nil
}

// createConsensusTopic creates a consensus topic for node
func (n *Node) createConsensusTopic(sposWrk *spos.SPOSConsensusWorker) *p2p.Topic {
	t := p2p.NewTopic(string(ConsensusTopic), &spos.ConsensusData{}, n.marshalizer)
	t.AddDataReceived(sposWrk.ReceivedMessage)
	return t
}

// addSubroundsToChronology adds subrounds to chronology
func (n *Node) addSubroundsToChronology(sposWrk *spos.SPOSConsensusWorker) {
	roundDuration := sposWrk.Cns.Chr.Round().TimeDuration()

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrStartRound),
		chronology.SubroundId(spos.SrBlock), int64(roundDuration*5/100),
		sposWrk.Cns.GetSubroundName(spos.SrStartRound),
		sposWrk.DoStartRoundJob,
		sposWrk.ExtendStartRound,
		sposWrk.Cns.CheckStartRoundConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrBlock),
		chronology.SubroundId(spos.SrCommitmentHash),
		int64(roundDuration*25/100),
		sposWrk.Cns.GetSubroundName(spos.SrBlock),
		sposWrk.DoBlockJob,
		sposWrk.ExtendBlock,
		sposWrk.Cns.CheckBlockConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrCommitmentHash),
		chronology.SubroundId(spos.SrBitmap),
		int64(roundDuration*40/100),
		sposWrk.Cns.GetSubroundName(spos.SrCommitmentHash),
		sposWrk.DoCommitmentHashJob,
		sposWrk.ExtendCommitmentHash,
		sposWrk.Cns.CheckCommitmentHashConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrBitmap),
		chronology.SubroundId(spos.SrCommitment),
		int64(roundDuration*55/100),
		sposWrk.Cns.GetSubroundName(spos.SrBitmap),
		sposWrk.DoBitmapJob,
		sposWrk.ExtendBitmap,
		sposWrk.Cns.CheckBitmapConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrCommitment),
		chronology.SubroundId(spos.SrSignature),
		int64(roundDuration*70/100),
		sposWrk.Cns.GetSubroundName(spos.SrCommitment),
		sposWrk.DoCommitmentJob,
		sposWrk.ExtendCommitment,
		sposWrk.Cns.CheckCommitmentConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrSignature),
		chronology.SubroundId(spos.SrEndRound),
		int64(roundDuration*85/100),
		sposWrk.Cns.GetSubroundName(spos.SrSignature),
		sposWrk.DoSignatureJob,
		sposWrk.ExtendSignature,
		sposWrk.Cns.CheckSignatureConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrEndRound),
		chronology.SubroundId(spos.SrAdvance),
		int64(roundDuration*95/100),
		sposWrk.Cns.GetSubroundName(spos.SrEndRound),
		sposWrk.DoEndRoundJob,
		sposWrk.ExtendEndRound,
		sposWrk.Cns.CheckEndRoundConsensus))

	sposWrk.Cns.Chr.AddSubround(spos.NewSubround(
		chronology.SubroundId(spos.SrAdvance),
		-1,
		int64(roundDuration*100/100),
		sposWrk.Cns.GetSubroundName(spos.SrAdvance),
		sposWrk.DoAdvanceJob,
		nil,
		sposWrk.Cns.CheckAdvanceConsensus))
}

func (n *Node) generateAndSignTx(
	nonce uint64,
	value *big.Int,
	rcvAddrBytes []byte,
	sndAddrBytes []byte,
	dataBytes []byte,
) (*transaction.Transaction, []byte, error) {

	tx := transaction.Transaction{
		Nonce:   nonce,
		Value:   value,
		RcvAddr: rcvAddrBytes,
		SndAddr: sndAddrBytes,
		Data:    dataBytes,
	}

	if n.marshalizer == nil {
		return nil, nil, ErrNilMarshalizer
	}

	if n.privateKey == nil {
		return nil, nil, ErrNilPrivateKey
	}

	marshalizedTx, err := n.marshalizer.Marshal(&tx)
	if err != nil {
		return nil, nil, errors.New("could not marshal transaction")
	}

	sig, err := n.singlesig.Sign(n.privateKey, marshalizedTx)
	if err != nil {
		return nil, nil, errors.New("could not sign the transaction")
	}
	tx.Signature = sig

	signedMarshalizedTx, err := n.marshalizer.Marshal(&tx)
	if err != nil {
		return nil, nil, errors.New("could not marshal signed transaction")
	}

	return &tx, signedMarshalizedTx, nil
}

//GenerateTransaction generates a new transaction with sender, receiver, amount and code
func (n *Node) GenerateTransaction(senderHex string, receiverHex string, value *big.Int, transactionData string) (*transaction.Transaction, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}

	if n.privateKey == nil {
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
	if senderAccount != nil {
		newNonce = senderAccount.BaseAccount().Nonce
	}

	tx, _, err := n.generateAndSignTx(
		newNonce,
		value,
		receiverAddress.Bytes(),
		senderAddress.Bytes(),
		[]byte(transactionData))

	return tx, err
}

// SendTransaction will send a new transaction on the topic channel
func (n *Node) SendTransaction(
	nonce uint64,
	senderHex string,
	receiverHex string,
	value *big.Int,
	transactionData string,
	signature []byte) (*transaction.Transaction, error) {

	sender, err := n.addrConverter.CreateAddressFromHex(senderHex)
	if err != nil {
		return nil, err
	}
	receiver, err := n.addrConverter.CreateAddressFromHex(receiverHex)
	if err != nil {
		return nil, err
	}

	tx := transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		RcvAddr:   receiver.Bytes(),
		SndAddr:   sender.Bytes(),
		Data:      []byte(transactionData),
		Signature: signature,
	}

	topic := n.messenger.GetTopic(string(factory.TransactionTopic))

	if topic == nil {
		return nil, errors.New("could not get transaction topic")
	}

	marshalizedTx, err := n.marshalizer.Marshal(&tx)
	if err != nil {
		return nil, errors.New("could not marshal transaction")
	}

	err = topic.BroadcastBuff(marshalizedTx)
	if err != nil {
		return nil, errors.New("could not broadcast transaction: " + err.Error())
	}
	return &tx, nil
}

//GetTransaction gets the transaction
func (n *Node) GetTransaction(hash string) (*transaction.Transaction, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// GetCurrentPublicKey will return the current node's public key
func (n *Node) GetCurrentPublicKey() string {
	if n.publicKey != nil {
		pkey, _ := n.publicKey.ToByteArray()
		return fmt.Sprintf("%x", pkey)
	}
	return ""
}

// GetAccount will return acount details for a given address
func (n *Node) GetAccount(address string) (*state.Account, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}

	addr, err := n.addrConverter.CreateAddressFromHex(address)
	if err != nil {
		return nil, errors.New("could not create address object from provided string")
	}
	account, err := n.accounts.GetExistingAccount(addr)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param")
	}
	return account.BaseAccount(), nil
}

func (n *Node) createGenesisBlock() (*block.Header, []byte, error) {
	blockBody, err := n.blockProcessor.CreateGenesisBlockBody(n.initialNodesBalances, 0)
	if err != nil {
		return nil, nil, err
	}

	marshalizedBody, err := n.marshalizer.Marshal(blockBody)
	if err != nil {
		return nil, nil, err
	}
	blockBodyHash := n.hasher.Compute(string(marshalizedBody))
	header := &block.Header{
		Nonce:         0,
		ShardId:       blockBody.ShardID,
		TimeStamp:     uint64(n.genesisTime.Unix()),
		BlockBodyHash: blockBodyHash,
		BlockBodyType: block.StateBlock,
		Signature:     blockBodyHash,
	}

	marshalizedHeader, err := n.marshalizer.Marshal(header)

	if err != nil {
		return nil, nil, err
	}

	blockHeaderHash := n.hasher.Compute(string(marshalizedHeader))

	return header, blockHeaderHash, nil
}

func (n *Node) sendMessage(cnsDta *spos.ConsensusData) {
	topic := n.messenger.GetTopic(string(ConsensusTopic))

	if topic == nil {
		log.Debug(fmt.Sprintf("could not get consensus topic"))
		return
	}

	err := topic.Broadcast(cnsDta)

	if err != nil {
		log.Debug(fmt.Sprintf("could not broadcast message: " + err.Error()))
	}
}

func (n *Node) broadcastBlockBody(msg []byte) {
	topic := n.messenger.GetTopic(string(factory.TxBlockBodyTopic))

	if topic == nil {
		log.Debug(fmt.Sprintf("could not get tx block body topic"))
		return
	}

	err := topic.BroadcastBuff(msg)

	if err != nil {
		log.Debug(fmt.Sprintf("could not broadcast message: " + err.Error()))
	}
}

func (n *Node) broadcastHeader(msg []byte) {
	topic := n.messenger.GetTopic(string(factory.HeadersTopic))

	if topic == nil {
		log.Debug(fmt.Sprintf("could not get header topic"))
		return
	}

	err := topic.BroadcastBuff(msg)

	if err != nil {
		log.Debug(fmt.Sprintf("could not broadcast message: " + err.Error()))
	}
}

func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
}
