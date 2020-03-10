package data

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

func TestAuthMessage_MarshalUnmarshalShouldWork(t *testing.T) {
	llw := generateAuthMessage()

	for marshName, marsh := range marshal.MarshalizersAvailableForTesting {
		testMarshalUnmarshal(t, marshName, marsh, llw)
	}
}

func generateAuthMessage() AuthMessage {
	return AuthMessage{
		AuthMessagePb: AuthMessagePb{
			Message:   []byte("test message"),
			Sig:       []byte("sig"),
			Pubkey:    []byte("pubkey"),
			Timestamp: 11223344,
		},
	}
}

func testMarshalUnmarshal(t *testing.T, marshName string, marsh marshal.Marshalizer, am AuthMessage) {
	objCopyForAssert := am

	buff, err := marsh.Marshal(&am)
	assert.Nil(t, err)

	objRecovered := &AuthMessage{}
	err = marsh.Unmarshal(objRecovered, buff)

	assert.Nil(t, err)
	assert.Equal(t, objCopyForAssert, *objRecovered, fmt.Sprintf("for marshalizer %v", marshName))
}
