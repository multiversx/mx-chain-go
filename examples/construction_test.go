package examples

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/require"
)

var addressEncoder, _ = pubkeyConverter.NewBech32PubkeyConverter(32)
var signingMarshalizer = &marshal.JsonMarshalizer{}
var signer = &singlesig.Ed25519Signer{}
var signingCryptoSuite = ed25519.NewEd25519()
var contentMarshalizer = &marshal.GogoProtoMarshalizer{}
var contentHasher = &blake2b.Blake2b{}

const alicePrivateKeyHex = "413f42575f7f26fad3317a778771212fdb80245850981e48b58a4f25e344e8f9"

func TestConstructTransaction_NoDataNoValue(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    89,
		Value:    big.NewInt(0),
		RcvAddr:  getPubkeyOfAddress(t, "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"),
		SndAddr:  getPubkeyOfAddress(t, "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"),
		GasPrice: 1000000000,
		GasLimit: 50000,
		ChainID:  []byte("local-testnet"),
		Version:  1,
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "b56769014f2bdc5cf9fc4a05356807d71fcf8775c819b0f1b0964625b679c918ffa64862313bfef86f99b38cb84fcdb16fa33ad6eb565276616723405cd8f109", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "0859120200001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340d08603520d6c6f63616c2d746573746e657458016240b56769014f2bdc5cf9fc4a05356807d71fcf8775c819b0f1b0964625b679c918ffa64862313bfef86f99b38cb84fcdb16fa33ad6eb565276616723405cd8f109", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "eb30c50c8831885ebcfac986d27e949ec02cf25676e22a009b7a486e5431ec2e", hex.EncodeToString(txHash))
}

func TestConstructTransaction_WithDataNoValue(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    90,
		Value:    big.NewInt(0),
		RcvAddr:  getPubkeyOfAddress(t, "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"),
		SndAddr:  getPubkeyOfAddress(t, "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"),
		GasPrice: 1000000000,
		GasLimit: 80000,
		Data:     []byte("hello"),
		ChainID:  []byte("local-testnet"),
		Version:  1,
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "e47fd437fc17ac9a69f7bf5f85bafa9e7628d851c4f69bd9fedc7e36029708b2e6d168d5cd652ea78beedd06d4440974ca46c403b14071a1a148d4188f6f2c0d", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "085a120200001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc034080f1044a0568656c6c6f520d6c6f63616c2d746573746e657458016240e47fd437fc17ac9a69f7bf5f85bafa9e7628d851c4f69bd9fedc7e36029708b2e6d168d5cd652ea78beedd06d4440974ca46c403b14071a1a148d4188f6f2c0d", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "95ed9ac933712d7d77721d75eecfc7896873bb0d746417153812132521636872", hex.EncodeToString(txHash))
}

func TestConstructTransaction_WithDataWithValue(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    91,
		Value:    stringToBigInt("10000000000000000000"),
		RcvAddr:  getPubkeyOfAddress(t, "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"),
		SndAddr:  getPubkeyOfAddress(t, "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"),
		GasPrice: 1000000000,
		GasLimit: 100000,
		Data:     []byte("for the book"),
		ChainID:  []byte("local-testnet"),
		Version:  1,
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "9074789e0b4f9b2ac24b1fd351a4dd840afcfeb427b0f93e2a2d429c28c65ee9f4c288ca4dbde79de0e5bcf8c1a5d26e1b1c86203faea923e0edefb0b5099b0c", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "085b1209008ac7230489e800001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340a08d064a0c666f722074686520626f6f6b520d6c6f63616c2d746573746e6574580162409074789e0b4f9b2ac24b1fd351a4dd840afcfeb427b0f93e2a2d429c28c65ee9f4c288ca4dbde79de0e5bcf8c1a5d26e1b1c86203faea923e0edefb0b5099b0c", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "af53e0fc86612d5068862716b5169effdf554951ecc89849b0e836eb0b63fa3e", hex.EncodeToString(txHash))
}

func TestConstructTransaction_WithDataWithLargeValue(t *testing.T) {
	tx := &transaction.Transaction{
		Nonce:    92,
		Value:    stringToBigInt("123456789000000000000000000000"),
		RcvAddr:  getPubkeyOfAddress(t, "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"),
		SndAddr:  getPubkeyOfAddress(t, "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"),
		GasPrice: 1000000000,
		GasLimit: 100000,
		Data:     []byte("for the spaceship"),
		ChainID:  []byte("local-testnet"),
		Version:  1,
	}

	tx.Signature = computeTransactionSignature(t, alicePrivateKeyHex, tx)
	require.Equal(t, "39938d15812708475dfc8125b5d41dbcea0b2e3e7aabbbfceb6ce4f070de3033676a218b73facd88b1432d7d4accab89c6130b3abe5cc7bbbb5146e61d355b03", hex.EncodeToString(tx.Signature))

	data, _ := contentMarshalizer.Marshal(tx)
	require.Equal(t, "085c120e00018ee90ff6181f3761632000001a208049d639e5a6980d1cd2392abcce41029cda74a1563523a202f09641cc2618f82a200139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1388094ebdc0340a08d064a11666f722074686520737061636573686970520d6c6f63616c2d746573746e65745801624039938d15812708475dfc8125b5d41dbcea0b2e3e7aabbbfceb6ce4f070de3033676a218b73facd88b1432d7d4accab89c6130b3abe5cc7bbbb5146e61d355b03", hex.EncodeToString(data))

	txHash := contentHasher.Compute(string(data))
	require.Equal(t, "e4a6048d92409cfe50f12e81218cb92f39966c618979a693b8d16320a06061c1", hex.EncodeToString(txHash))
}

func stringToBigInt(input string) *big.Int {
	result := big.NewInt(0)
	_, _ = result.SetString(input, 10)
	return result
}

func getPubkeyOfAddress(t *testing.T, address string) []byte {
	pubkey, err := addressEncoder.Decode(address)
	require.Nil(t, err)
	return pubkey
}

func computeTransactionSignature(t *testing.T, senderSeedHex string, tx *transaction.Transaction) []byte {
	keyGenerator := signing.NewKeyGenerator(signingCryptoSuite)

	senderSeed, err := hex.DecodeString(senderSeedHex)
	require.Nil(t, err)

	privateKey, err := keyGenerator.PrivateKeyFromByteArray(senderSeed)
	require.Nil(t, err)

	dataToSign, err := tx.GetDataForSigning(addressEncoder, signingMarshalizer)
	require.Nil(t, err)

	signature, err := signer.Sign(privateKey, dataToSign)
	require.Nil(t, err)
	require.Len(t, signature, 64)

	return signature
}
