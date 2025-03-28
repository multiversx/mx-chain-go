package proofscache_test

import (
	"fmt"
	"testing"

	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
)

func Benchmark_AddProof_Bucket10_Pool1000(b *testing.B) {
	benchmarkAddProof(b, 10, 1000)
}

func Benchmark_AddProof_Bucket100_Pool10000(b *testing.B) {
	benchmarkAddProof(b, 100, 10000)
}

func Benchmark_AddProof_Bucket1000_Pool100000(b *testing.B) {
	benchmarkAddProof(b, 1000, 100000)
}

func benchmarkAddProof(b *testing.B, bucketSize int, nonceRange int) {
	pc := proofscache.NewProofsCache(bucketSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		proof := generateProof()
		nonce := generateRandomNonce(int64(nonceRange))

		proof.HeaderNonce = nonce
		proof.HeaderHash = []byte("hash_" + fmt.Sprintf("%d", nonce))
		b.StartTimer()

		pc.AddProof(proof)
	}
}

func Benchmark_CleanupProofs_Bucket10_Pool1000(b *testing.B) {
	benchmarkCleanupProofs(b, 10, 1000)
}

func Benchmark_CleanupProofs_Bucket100_Pool10000(b *testing.B) {
	benchmarkCleanupProofs(b, 100, 10000)
}

func Benchmark_CleanupProofs_Bucket1000_Pool100000(b *testing.B) {
	benchmarkCleanupProofs(b, 1000, 100000)
}

func benchmarkCleanupProofs(b *testing.B, bucketSize int, nonceRange int) {
	pc := proofscache.NewProofsCache(bucketSize)

	for i := uint64(0); i < uint64(nonceRange); i++ {
		proof := generateProof()
		proof.HeaderNonce = i
		proof.HeaderHash = []byte("hash_" + fmt.Sprintf("%d", i))

		pc.AddProof(proof)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.CleanupProofsBehindNonce(uint64(nonceRange))
	}
}
