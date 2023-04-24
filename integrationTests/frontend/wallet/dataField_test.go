package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

func TestTxDataFieldContainingUTF8Characters(t *testing.T) {
	suite := ed25519.NewEd25519()
	keyGen := signing.NewKeyGenerator(suite)
	singleSigner := integrationTests.TestSingleSigner

	sk, pk := keyGen.GeneratePair()
	pkBytes, _ := pk.ToByteArray()

	nonce := uint64(3322)
	value := big.NewInt(0)
	value.SetString("999", 10)
	sndAddrBytes := bytes.Repeat([]byte{97}, 32)
	gasPrice := uint64(1122)
	gasLimit := uint64(2247)
	chainID := []byte("chain ID")
	version := uint32(1445)

	tx1 := &transaction.Transaction{
		Nonce:    nonce,
		Value:    value,
		RcvAddr:  pkBytes,
		SndAddr:  sndAddrBytes,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Data:     []byte{97, 0x80},
		ChainID:  chainID,
		Version:  version,
	}

	tx2 := &transaction.Transaction{
		Nonce:    nonce,
		Value:    value,
		RcvAddr:  pkBytes,
		SndAddr:  sndAddrBytes,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Data:     []byte{97, 0x81},
		ChainID:  chainID,
		Version:  version,
	}

	sig1 := sign(tx1, singleSigner, sk)

	sig2 := sign(tx2, singleSigner, sk)

	fmt.Println("sig1: " + hex.EncodeToString(sig1))
	fmt.Println("sig2: " + hex.EncodeToString(sig2))

	assert.NotEqual(t, sig1, sig2)
}

func sign(tx *transaction.Transaction, signer crypto.SingleSigner, sk crypto.PrivateKey) []byte {
	marshalizer := &marshal.JsonMarshalizer{}
	converter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")

	receiverAddress, _ := converter.Encode(tx.RcvAddr)

	senderAddress, _ := converter.Encode(tx.SndAddr)

	ftx := &transaction.FrontendTransaction{
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         receiverAddress,
		Sender:           senderAddress,
		SenderUsername:   nil,
		ReceiverUsername: nil,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		Data:             tx.Data,
		Signature:        "",
		ChainID:          string(tx.ChainID),
		Version:          tx.Version,
	}

	buff, _ := marshalizer.Marshal(ftx)

	fmt.Println(string(buff))

	signature, _ := signer.Sign(sk, buff)

	return signature
}
