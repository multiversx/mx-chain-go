package notifier_test

import (
	"fmt"
	"testing"

	mock3 "github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	mock4 "github.com/ElrondNetwork/elrond-go/process/slash/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash/notifier"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestSlashingNotifier_CreateShardSlashingTransaction(t *testing.T) {
	notifier, _ := notifier.NewSlashingNotifier(
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{},
		&testscommon.PubkeyConverterMock{},
		&mock.SignerMock{},
		&mock3.BaseAccountMock{},
		testscommon.HasherMock{},
		&testscommon.MarshalizerMock{})

	proof := &mock4.MultipleHeaderProposalProofStub{
		GetHeadersCalled: func() []*interceptedBlocks.InterceptedHeader {
			return []*interceptedBlocks.InterceptedHeader{}
		},
	}

	tx, _ := notifier.CreateShardSlashingTransaction(proof)
	require.NotNil(t, tx)
	fmt.Println(tx.GetData())
}
