package p2p

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

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

func NewMessage(peerID string, payload []byte, mrsh marshal.Marshalizer) *Message {
	if mrsh == nil {
		panic("Nil marshalizer when creating a new Message!")
	}

	return &Message{Payload: payload, Hops: 0, Peers: []string{peerID}, marsh: &mrsh}
}

func (m *Message) ToByteArray() ([]byte, error) {
	if m.marsh == nil {
		return nil, errors.New("Uninitialized marshalizer!")
	}

	return (*m.marsh).Marshal(m)
}

func CreateFromByteArray(mrsh marshal.Marshalizer, buff []byte) (*Message, error) {
	m := &Message{}
	m.marsh = &mrsh

	if mrsh == nil {
		return nil, errors.New("Uninitialized marshalizer!")
	}

	err := mrsh.Unmarshal(m, buff)

	return m, err
}

func (m *Message) AddHop(peerID string) {
	m.Hops++
	m.Peers = append(m.Peers, peerID)
}

func (m *Message) Signed() bool {
	return m.isSigned
}

func (m *Message) Sign(params *ConnectParams) error {
	if params.PubKey == nil || params.PrivKey == nil || params.ID == "" {
		return errors.New("Invalid private key/public key/ID tuple!")
	}

	pkey, err := crypto.MarshalPublicKey(params.PubKey)
	if err != nil {
		return err
	}

	sig, err := params.PrivKey.Sign(append(m.Payload, []byte(m.Type)...))
	if err != nil {
		return err
	}

	m.PubKey = pkey
	m.Sig = sig
	m.isSigned = true
	return nil
}

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

func (m *Message) VerifyAndSetSigned() error {
	signed, err := m.Verify()

	if err != nil {
		m.isSigned = false
		return err
	}

	m.isSigned = signed
	return nil
}
