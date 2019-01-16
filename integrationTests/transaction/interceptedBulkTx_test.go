package transaction

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptBulkTransactions(t *testing.T) {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	keyGen := schnorr.NewKeyGenerator()
	sk, pk := keyGen.GeneratePair()

	dPool := createTestDataPool()

	startingNonce := uint64(6)

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")
	accntAdapter := adbCreateAccountsDB()

	//TODO change when injecting a messenger is possible
	n, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithAccountsAdapter(accntAdapter),
		node.WithDataPool(dPool),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(&sharding.OneShardCoordinator{}),
		node.WithBlockChain(createTestBlockChain()),
		node.WithUint64ByteSliceConverter(uint64ByteSlice.NewBigEndianConverter()),
		node.WithPublicKey(pk),
		node.WithPrivateKey(sk),
	)

	n.Start()
	defer n.Stop()

	//set the account's nonce to startingNonce
	nodePubKeyBytes, _ := pk.ToByteArray()
	nodeAddress, _ := addrConverter.CreateAddressFromPublicKeyBytes(nodePubKeyBytes)
	nodeAccount, _ := accntAdapter.GetJournalizedAccount(nodeAddress)
	nodeAccount.SetNonceWithJournal(startingNonce)
	accntAdapter.Commit()

	noOfTx := 50

	_ = n.BindInterceptorsResolvers()

	wg := sync.WaitGroup{}
	wg.Add(noOfTx)

	chanDone := make(chan bool)

	go func() {
		wg.Wait()

		chanDone <- true
	}()

	txHashes := make([][]byte, 0)

	//wire up handler
	dPool.Transactions().RegisterHandler(func(key []byte) {
		txHashes = append(txHashes, key)
		wg.Done()
	})

	err := n.GenerateAndSendBulkTransactions(createDummyHexAddress(64), *big.NewInt(1), uint64(noOfTx))
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
		return
	}

	assert.Equal(t, noOfTx, len(txHashes))

	bitmap := make([]bool, noOfTx+int(startingNonce))
	//set for each nonce from found tx a true flag in bitmap
	for i := 0; i < noOfTx; i++ {
		tx, _ := dPool.Transactions().ShardDataStore(0).Get(txHashes[i])

		assert.NotNil(t, tx)
		bitmap[tx.(*transaction.Transaction).Nonce] = true
	}

	//for the first startingNonce values, the bitmap should be false
	//for the rest, true
	for i := 0; i < noOfTx+int(startingNonce); i++ {
		if i < int(startingNonce) {
			assert.False(t, bitmap[i])
			continue
		}

		assert.True(t, bitmap[i])
	}
}
