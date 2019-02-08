package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func TestFactory_NewbnFactoryNilShouldFail(t *testing.T) {
	bnf, err := bn.NewbnFactory(
		nil,
	)

	assert.Nil(t, bnf)
	assert.Equal(t, err, spos.ErrNilWorker)
}

func TestFactory_NewbnFactoryShouldWork(t *testing.T) {
	wrk := initWorker()

	bnf, _ := bn.NewbnFactory(
		wrk,
	)

	assert.NotNil(t, bnf)
}

func TestFactory_GenerateSubroundsShouldWork(t *testing.T) {
	wrk := initWorker()

	bnf, _ := bn.NewbnFactory(
		wrk,
	)

	bnf.GenerateSubrounds()
	assert.Equal(t, 8, len(wrk.SPoS.Chr.SubroundHandlers()))
}

func TestWorker_GetMessageTypeName(t *testing.T) {
	r := bn.GetMessageTypeName(bn.MtBlockBody)
	assert.Equal(t, "(BLOCK_BODY)", r)

	r = bn.GetMessageTypeName(bn.MtBlockHeader)
	assert.Equal(t, "(BLOCK_HEADER)", r)

	r = bn.GetMessageTypeName(bn.MtCommitmentHash)
	assert.Equal(t, "(COMMITMENT_HASH)", r)

	r = bn.GetMessageTypeName(bn.MtBitmap)
	assert.Equal(t, "(BITMAP)", r)

	r = bn.GetMessageTypeName(bn.MtCommitment)
	assert.Equal(t, "(COMMITMENT)", r)

	r = bn.GetMessageTypeName(bn.MtSignature)
	assert.Equal(t, "(SIGNATURE)", r)

	r = bn.GetMessageTypeName(bn.MtUnknown)
	assert.Equal(t, "(UNKNOWN)", r)

	r = bn.GetMessageTypeName(bn.MessageType(-1))
	assert.Equal(t, "Undefined message type", r)
}

func TestWorker_GetSubroundName(t *testing.T) {
	r := bn.GetSubroundName(bn.SrStartRound)
	assert.Equal(t, "(START_ROUND)", r)

	r = bn.GetSubroundName(bn.SrBlock)
	assert.Equal(t, "(BLOCK)", r)

	r = bn.GetSubroundName(bn.SrCommitmentHash)
	assert.Equal(t, "(COMMITMENT_HASH)", r)

	r = bn.GetSubroundName(bn.SrBitmap)
	assert.Equal(t, "(BITMAP)", r)

	r = bn.GetSubroundName(bn.SrCommitment)
	assert.Equal(t, "(COMMITMENT)", r)

	r = bn.GetSubroundName(bn.SrSignature)
	assert.Equal(t, "(SIGNATURE)", r)

	r = bn.GetSubroundName(bn.SrEndRound)
	assert.Equal(t, "(END_ROUND)", r)

	r = bn.GetSubroundName(bn.SrAdvance)
	assert.Equal(t, "(ADVANCE)", r)

	r = bn.GetSubroundName(chronology.SubroundId(-1))
	assert.Equal(t, "Undefined subround", r)
}
