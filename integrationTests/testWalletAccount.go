package integrationTests

import (
	"encoding/hex"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"math/big"
)

type TestWalletAccount struct {
	SingleSigner  crypto.SingleSigner
	SkTxSign      crypto.PrivateKey
	PkTxSign      crypto.PublicKey
	PkTxSignBytes []byte
	KeygenTxSign  crypto.KeyGenerator

	Address state.AddressContainer
	Balance *big.Int
}

func CreateTestWalletAccount(coordinator sharding.Coordinator, shardId uint32) *TestWalletAccount {
	testWalletAccount := &TestWalletAccount{}
	testWalletAccount.initCrypto(coordinator, shardId)
	return testWalletAccount
}

func (twa *TestWalletAccount) initCrypto(coordinator sharding.Coordinator, shardId uint32) {
	suite := kyber.NewBlakeSHA256Ed25519()
	twa.SingleSigner = &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	for {
		pkBytes, _ := pk.ToByteArray()
		addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(pkBytes)
		if coordinator.ComputeId(addr) == shardId {
			break
		}
		sk, pk = keyGen.GeneratePair()
	}

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s in shard %d\n", hex.EncodeToString(pkBuff), shardId)

	twa.SkTxSign = sk
	twa.PkTxSign = pk
	twa.PkTxSignBytes, _ = pk.ToByteArray()
	twa.KeygenTxSign = keyGen
	twa.Address, _ = TestAddressConverter.CreateAddressFromPublicKeyBytes(twa.PkTxSignBytes)
}
