package txcache

import (
	"math"
)

var six = uint64(6)
var nonceModulus = uint64(math.MaxUint32)

// noncesTracker is a helper struct to track nonces for a sender,
// so we can check if the sequence of nonces "is spotless" (has no gaps and no duplicates).
//
// Notes:
//
//	(a) math.MaxUint32 * math.MaxUint32 < math.MaxUint64.
//	(b) however, math.MaxUint32 * (2 * math.MaxUint32 + 1) > math.MaxUint64
//	(c) we use modular arithmetic, with modulus = nonceModulus (see above).
//	(d) memory footprint: 4 * 8 bytes = 32 bytes.
type noncesTracker struct {
	sumOfAddedNonces            uint64
	sumOfRemovedNonces          uint64
	sumOfSquaresOfAddedNonces   uint64
	sumOfSquaresOfRemovedNonces uint64
}

func newNoncesTracker() *noncesTracker {
	return &noncesTracker{}
}

func (tracker *noncesTracker) addNonce(nonce uint64) {
	nonce = tracker.mod(nonce)
	nonceSquared := tracker.mod(nonce * nonce)

	tracker.sumOfAddedNonces = tracker.mod(tracker.sumOfAddedNonces + nonce)
	tracker.sumOfSquaresOfAddedNonces = tracker.mod(tracker.sumOfSquaresOfAddedNonces + nonceSquared)
}

func (tracker *noncesTracker) removeNonce(nonce uint64) {
	nonce = tracker.mod(nonce)
	nonceSquared := tracker.mod(nonce * nonce)

	tracker.sumOfRemovedNonces = tracker.mod(tracker.sumOfRemovedNonces + nonce)
	tracker.sumOfSquaresOfRemovedNonces = tracker.mod(tracker.sumOfSquaresOfRemovedNonces + nonceSquared)
}

func (tracker *noncesTracker) computeExpectedSumOfNonces(firstNonce uint64, count uint64) uint64 {
	firstNonce = tracker.mod(firstNonce)
	lastNonce := firstNonce + count - 1
	result := (firstNonce + lastNonce) * count / 2
	return tracker.mod(result)
}

// Computes [lastNonce * (lastNonce + 1) * (2 * lastNonce + 1) - firstNonce * (firstNonce + 1) * (2 * firstNonce + 1)] / 6 * 6
func (tracker *noncesTracker) computeExpectedSumOfSquaresOfNoncesTimesSix(firstNonce uint64, count uint64) uint64 {
	firstNonce = tracker.mod(firstNonce)
	lastNonce := firstNonce + count - 1
	nonceBeforeFirst := firstNonce - 1

	firstTerm := lastNonce
	firstTerm = tracker.mod(firstTerm * (lastNonce + 1))
	// See note (b) above.
	firstTerm = tracker.mod(firstTerm * tracker.mod(2*lastNonce+1))

	secondTerm := nonceBeforeFirst
	secondTerm = tracker.mod(secondTerm * (nonceBeforeFirst + 1))
	// See note (b) above.
	secondTerm = tracker.mod(secondTerm * tracker.mod(2*nonceBeforeFirst+1))

	result := tracker.modStrict(int64(firstTerm) - int64(secondTerm))
	return uint64(result)
}

func (tracker *noncesTracker) mod(value uint64) uint64 {
	return value % nonceModulus
}

// See:
// - https://stackoverflow.com/questions/43018206/modulo-of-negative-integers-in-go
func (tracker *noncesTracker) modStrict(value int64) uint64 {
	return uint64((value%int64(nonceModulus) + int64(nonceModulus)) % int64(nonceModulus))
}

func (tracker *noncesTracker) isSpotlessSequence(firstNonce uint64, count uint64) bool {
	sumOfNonces := tracker.modStrict(int64(tracker.sumOfAddedNonces) - int64(tracker.sumOfRemovedNonces))
	expectedSumOfNonces := tracker.computeExpectedSumOfNonces(firstNonce, count)
	if sumOfNonces != expectedSumOfNonces {
		return false
	}

	sumOfSquaresOfNonces := tracker.modStrict(int64(tracker.sumOfSquaresOfAddedNonces) - int64(tracker.sumOfSquaresOfRemovedNonces))
	sumOfSquaresOfNoncesTimesSix := tracker.mod(sumOfSquaresOfNonces * six)
	expectedSumOfSquaresOfNoncesTimesSix := tracker.computeExpectedSumOfSquaresOfNoncesTimesSix(firstNonce, count)
	return sumOfSquaresOfNoncesTimesSix == expectedSumOfSquaresOfNoncesTimesSix
}
