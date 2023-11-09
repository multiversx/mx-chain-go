package multisig

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	llsig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"
	"github.com/stretchr/testify/require"
)

const metaConsensusSize = uint16(400)
const shardConsensusSize = uint16(63)

type multiSignerSetup struct {
	privKeys          [][]byte
	pubKeys           [][]byte
	partialSignatures [][][]byte
	messages          []string
	aggSignatures     [][]byte
}

func Benchmark_MultiSigForMetachain(b *testing.B) {
	numSigners := metaConsensusSize
	numMessages := 50
	benchmarkMultiSigners(b, numSigners, numMessages)
}

func Benchmark_MultiSigForShard(b *testing.B) {
	numSigners := shardConsensusSize
	numMessages := 50
	benchmarkMultiSigners(b, numSigners, numMessages)
}

func benchmarkMultiSigners(b *testing.B, numSigners uint16, numMessages int) {
	setupKOSK, multiSignerKOSK := createMultiSigSetupKOSK(numSigners, numMessages)
	multiSignerNoKOSK := createMultiSignerNoKOSK()

	// signature shares, public keys and messages should be the same for both multi signers
	setupNoKOSK := *setupKOSK
	setupNoKOSK.aggSignatures = aggregateSignatures(&setupNoKOSK, multiSignerNoKOSK)
	setupKOSK.aggSignatures = aggregateSignatures(setupKOSK, multiSignerKOSK)

	b.Run("AggregateSigs KOSK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = multiSignerKOSK.AggregateSigs(setupKOSK.pubKeys, setupKOSK.partialSignatures[i%numMessages])
		}
	})
	b.Run("VerifyAggregatedSig KOSK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := multiSignerKOSK.VerifyAggregatedSig(setupKOSK.pubKeys, []byte(setupKOSK.messages[i%numMessages]), setupKOSK.aggSignatures[i%numMessages])
			require.Nil(b, err)
		}
	})
	b.Run("Verify signature shares KOSK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < int(numSigners); j++ {
				err := multiSignerKOSK.VerifySignatureShare(setupKOSK.pubKeys[j], []byte(setupKOSK.messages[i%numMessages]), setupKOSK.partialSignatures[i%numMessages][j])
				require.Nil(b, err)
			}
		}
	})
	b.Run("AggregateSigs NoKOSK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = multiSignerNoKOSK.AggregateSigs(setupNoKOSK.pubKeys, setupNoKOSK.partialSignatures[i%numMessages])
		}
	})
	b.Run("VerifyAggregatedSig NoKOSK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := multiSignerNoKOSK.VerifyAggregatedSig(setupNoKOSK.pubKeys, []byte(setupNoKOSK.messages[i%numMessages]), setupNoKOSK.aggSignatures[i%numMessages])
			require.Nil(b, err)
		}
	})
	b.Run("Verify signature shares NoKOSK", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < int(numSigners); j++ {
				err := multiSignerNoKOSK.VerifySignatureShare(setupNoKOSK.pubKeys[j], []byte(setupNoKOSK.messages[i%numMessages]), setupNoKOSK.partialSignatures[i%numMessages][j])
				require.Nil(b, err)
			}
		}
	})
}

func createMultiSigSetupKOSK(numSigners uint16, numMessages int) (*multiSignerSetup, crypto.MultiSigner) {
	var multiSigner crypto.MultiSigner
	setup := &multiSignerSetup{}
	suite := mcl.NewSuiteBLS12()
	setup.privKeys, setup.pubKeys, multiSigner = createKeysAndMultiSignerBlsKOSK(numSigners, suite)
	setup.messages, setup.partialSignatures = createMessagesAndPartialSignatures(numMessages, setup.privKeys, multiSigner)

	return setup, multiSigner
}

func createMessagesAndPartialSignatures(numMessages int, privKeys [][]byte, multiSigner crypto.MultiSigner) ([]string, [][][]byte) {
	partialSignatures := make([][][]byte, numMessages)
	messages := make([]string, numMessages)

	for i := 0; i < numMessages; i++ {
		messages[i] = fmt.Sprintf("message%d", i)
		signatures := createSignaturesShares(privKeys, multiSigner, []byte(messages[i]))
		partialSignatures[i] = signatures
	}

	return messages, partialSignatures
}

func aggregateSignatures(
	setup *multiSignerSetup,
	multiSigner crypto.MultiSigner,
) [][]byte {
	aggSignatures := make([][]byte, len(setup.messages))
	for i := 0; i < len(setup.messages); i++ {
		aggSignatures[i], _ = multiSigner.AggregateSigs(setup.pubKeys, setup.partialSignatures[i])
	}

	return aggSignatures
}

func createMultiSignerNoKOSK() crypto.MultiSigner {
	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	hashSize := hashing.BlsHashSize
	hasher, _ := blake2b.NewBlake2bWithSize(hashSize)
	llSigner := &llsig.BlsMultiSigner{Hasher: hasher}
	multiSignerNoKOSK, _ := multisig.NewBLSMultisig(llSigner, kg)
	return multiSignerNoKOSK
}
