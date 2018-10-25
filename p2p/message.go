package p2p

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

// Message is the main type of object used for exchanging data
type Message struct {
	marsh    *marshal.Marshalizer
	Payload  []byte
	Type     string
	PubKey   []byte
	Sig      []byte
	isSigned bool

	Hops  int
	Peers []string
}

// NewMessage creates a new Message object
func NewMessage(peerID string, payload []byte, mrsh marshal.Marshalizer) *Message {
	if mrsh == nil {
		panic("Nil marshalizer when creating a new Message!")
	}

	return &Message{Payload: payload, Hops: 0, Peers: []string{peerID}, marsh: &mrsh}
}

// ToByteArray will convert the message into its corresponding slice of bytes representation. Uses a marshalizer implmementation
func (m *Message) ToByteArray() ([]byte, error) {
	if m.marsh == nil {
		return nil, errors.New("Uninitialized marshalizer!")
	}

	return (*m.marsh).Marshal(m)
}

// CreateFromByteArray recreates the object based on the corresponding slice of bytes
func CreateFromByteArray(mrsh marshal.Marshalizer, buff []byte) (*Message, error) {
	m := &Message{}
	m.marsh = &mrsh

	if mrsh == nil {
		return nil, errors.New("Uninitialized marshalizer!")
	}

	err := mrsh.Unmarshal(m, buff)

	return m, err
}

// AddHop adds the peerID to the traversed peers, incrementing the hop counter
func (m *Message) AddHop(peerID string) {
	m.Hops++
	m.Peers = append(m.Peers, peerID)
}

// Signed returns true if the message was signed
// False means that the message was unsigned
// A signed message that was tampered (signature is not verified) will be automatically discarded
func (m *Message) Signed() bool {
	return m.isSigned
}

// Signs the message with the private key
func (m *Message) Sign(sk crypto.PrivKey) error {
	if sk == nil {
		return errors.New("Invalid private key!")
	}

	pk := sk.GetPublic()

	pkey, err := crypto.MarshalPublicKey(pk)
	if err != nil {
		return err
	}

	sig, err := sk.Sign(append(m.Payload, []byte(m.Type)...))
	if err != nil {
		return err
	}

	m.PubKey = pkey
	m.Sig = sig
	m.isSigned = true
	return nil
}

// Verify returns true in one of the following cases:
// 1. The message was not signed
// 2. There is a peer in the list of traversed peers
// 3. The message was signed and the signature verifies with the public key provided and the first ID from
//    traversed peers list is obtained from the public key
func (m *Message) Verify() (bool, error) {
	if m.Sig == nil || m.PubKey == nil {
		return false, nil
	}

	if len(m.Sig) == 0 || len(m.PubKey) == 0 {
		return false, nil
	}

	if len(m.Peers) == 0 {
		m.isSigned = false
		return false, nil
	}

	pubKey, err := crypto.UnmarshalPublicKey(m.PubKey)
	if err != nil {
		return false, err
	}

	id, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return false, err
	}

	if id.Pretty() != m.Peers[0] {
		m.isSigned = false
		return false, errors.New("wrong id/public key pair")
	}

	verif, err := pubKey.Verify(append(m.Payload, []byte(m.Type)...), m.Sig)
	if err != nil {
		return false, err
	}

	if !verif {
		m.isSigned = false
		return false, errors.New("wrong signature")
	}

	return true, nil
}

// VerifyAndSetSigned verifies the message and saves the signed value into message.isSigned
func (m *Message) VerifyAndSetSigned() error {
	signed, err := m.Verify()

	if err != nil {
		m.isSigned = false
		return err
	}

	m.isSigned = signed
	return nil
}
