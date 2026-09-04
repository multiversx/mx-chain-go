package sync

import (
	"fmt"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/process"
)

func BenchmarkBaseForkDetector_ConcurrentProbableHighestUpdates(b *testing.B) {
	for _, numRecords := range []int{1, 100, 1000, 5000} {
		b.Run(fmt.Sprintf("records=%d", numRecords), func(b *testing.B) {
			bfd := newBranchAwareForkDetector(0, 10, []byte("genesis"))
			updates := make([]*headerInfo, 0, numRecords*3)
			previousHash := []byte("genesis")
			for index := 0; index < numRecords; index++ {
				hash := []byte(fmt.Sprintf("header-%d", index))
				record := &headerInfo{
					epoch:    1,
					nonce:    uint64(index + 11),
					round:    uint64(index + 11),
					hash:     hash,
					prevHash: previousHash,
					state:    process.BHReceived,
				}
				bfd.headers[record.nonce] = []*headerInfo{record}
				updates = append(updates,
					record,
					&headerInfo{
						epoch:    record.epoch,
						nonce:    record.nonce,
						round:    record.round,
						hash:     record.hash,
						state:    process.BHReceived,
						hasProof: true,
					},
					&headerInfo{
						epoch:    record.epoch,
						nonce:    record.nonce,
						round:    record.round,
						hash:     record.hash,
						prevHash: record.prevHash,
						state:    process.BHNotarized,
						hasProof: true,
					},
				)
				previousHash = hash
			}

			const sampleCount = 8192
			samples := make([]atomic.Int64, sampleCount)
			var operation atomic.Uint64

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					started := time.Now()
					current := operation.Add(1) - 1
					record := updates[current%uint64(len(updates))]
					bfd.appendHeaderInfo(record)
					bfd.recomputeProbableHighestNonce()
					samples[current%sampleCount].Store(time.Since(started).Nanoseconds())
				}
			})
			b.StopTimer()

			durations := make([]int64, 0, sampleCount)
			for index := range samples {
				if duration := samples[index].Load(); duration > 0 {
					durations = append(durations, duration)
				}
			}
			sort.Slice(durations, func(left int, right int) bool {
				return durations[left] < durations[right]
			})
			if len(durations) > 0 {
				p99Index := (len(durations) - 1) * 99 / 100
				b.ReportMetric(float64(durations[p99Index]), "p99-ns")
			}
		})
	}
}
