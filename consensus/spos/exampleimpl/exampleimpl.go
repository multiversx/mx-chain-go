package exampleimpl

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"time"
)

const (
	srStartRound chronology.SubroundId = iota
	srBlock
	srCommitmentHash
	srBitmap
	srCommitment
	srSignature
	srEndRound
)

const roundTimeDuration = time.Duration(10 * time.Millisecond)

// #################### <START_ROUND> ####################

type SRStartRound struct {
	Hits int
}

func (sr *SRStartRound) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoStartRound with %d hits\n", sr.Hits)
	return true
}

func (sr *SRStartRound) Current() chronology.SubroundId {
	return srStartRound
}

func (sr *SRStartRound) Next() chronology.SubroundId {
	return srBlock
}

func (sr *SRStartRound) EndTime() int64 {
	return int64(5 * roundTimeDuration / 100)
}

func (sr *SRStartRound) Name() string {
	return "<START_ROUND>"
}

// #################### <BLOCK> ####################

type SRBlock struct {
	Hits int
}

func (sr *SRBlock) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoBlock with %d hits\n", sr.Hits)
	return true
}

func (sr *SRBlock) Current() chronology.SubroundId {
	return srBlock
}

func (sr *SRBlock) Next() chronology.SubroundId {
	return srCommitmentHash
}

func (sr *SRBlock) EndTime() int64 {
	return int64(25 * roundTimeDuration / 100)
}

func (sr *SRBlock) Name() string {
	return "<BLOCK>"
}

// #################### <COMMITMENT_HASH> ####################

type SRCommitmentHash struct {
	Hits int
}

func (sr *SRCommitmentHash) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoCommitmentHash with %d hits\n", sr.Hits)
	return true
}

func (sr *SRCommitmentHash) Current() chronology.SubroundId {
	return srCommitmentHash
}

func (sr *SRCommitmentHash) Next() chronology.SubroundId {
	return srBitmap
}

func (sr *SRCommitmentHash) EndTime() int64 {
	return int64(40 * roundTimeDuration / 100)
}

func (sr *SRCommitmentHash) Name() string {
	return "<COMMITMENT_HASH>"
}

// #################### <BITMAP> ####################

type SRBitmap struct {
	Hits int
}

func (sr *SRBitmap) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoBitmap with %d hits\n", sr.Hits)
	return true
}

func (sr *SRBitmap) Current() chronology.SubroundId {
	return srBitmap
}

func (sr *SRBitmap) Next() chronology.SubroundId {
	return srCommitment
}

func (sr *SRBitmap) EndTime() int64 {
	return int64(55 * roundTimeDuration / 100)
}

func (sr *SRBitmap) Name() string {
	return "<BITMAP>"
}

// #################### <COMMITMENT> ####################

type SRCommitment struct {
	Hits int
}

func (sr *SRCommitment) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoCommitment with %d hits\n", sr.Hits)
	return true
}

func (sr *SRCommitment) Current() chronology.SubroundId {
	return srCommitment
}

func (sr *SRCommitment) Next() chronology.SubroundId {
	return srSignature
}

func (sr *SRCommitment) EndTime() int64 {
	return int64(70 * roundTimeDuration / 100)
}

func (sr *SRCommitment) Name() string {
	return "<COMMITMENT>"
}

// #################### <SIGNATURE> ####################

type SRSignature struct {
	Hits int
}

func (sr *SRSignature) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoSignature with %d hits\n", sr.Hits)
	return true
}

func (sr *SRSignature) Current() chronology.SubroundId {
	return srSignature
}

func (sr *SRSignature) Next() chronology.SubroundId {
	return srEndRound
}

func (sr *SRSignature) EndTime() int64 {
	return int64(85 * roundTimeDuration / 100)
}

func (sr *SRSignature) Name() string {
	return "<SIGNATURE>"
}

// #################### <END_ROUND> ####################

type SREndRound struct {
	Hits int
}

func (sr *SREndRound) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoEndRound with %d hits\n", sr.Hits)
	return true
}

func (sr *SREndRound) Current() chronology.SubroundId {
	return srEndRound
}

func (sr *SREndRound) Next() chronology.SubroundId {
	return srStartRound
}

func (sr *SREndRound) EndTime() int64 {
	return int64(100 * roundTimeDuration / 100)
}

func (sr *SREndRound) Name() string {
	return "<END_ROUND>"
}
