package spos

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/stretchr/testify/assert"
)

func SndWithSuccess(chronology.Subround) bool {
	fmt.Printf("message was sent with success\n")
	return true
}

func SndWithoutSuccess(chronology.Subround) bool {
	fmt.Printf("message was NOT sent with success\n")
	return false
}

func RcvWithSuccess(rcvMsg *[]byte, chr *chronology.Chronology) bool {
	fmt.Printf("message was consumed with success\n")
	return true
}

func RcvWithoutSuccess(rcvMsg *[]byte, chr *chronology.Chronology) bool {
	fmt.Printf("message was NOT consumed with success\n")
	return false
}

func RcvWithoutSuccessAndCancel(rcvMsg *[]byte, chr *chronology.Chronology) bool {

	fmt.Printf("message was NOT consumed with success and this round should be canceled\n")
	chr.SetSelfSubround(chronology.SrCanceled)
	return false
}

func TestConsensus(t *testing.T) {

	cns := NewConsensus(true, &Validators{}, &Threshold{}, &RoundStatus{}, &chronology.Chronology{})
	cns.ResetRoundStatus()

	assert.Equal(t, cns.RoundStatus.Block, ssNotFinished)
}
