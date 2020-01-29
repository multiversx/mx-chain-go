package libp2p

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/p2p"
	pubsub "github.com/ElrondNetwork/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-core/peer"
)

type traverseInfo struct {
	peer      p2p.PeerID
	timestamp time.Time
	tag       string
}

// Message is a data holder struct
type Message struct {
	from         []byte
	data         []byte
	seqNo        []byte
	topicIds     []string
	signature    []byte
	key          []byte
	peer         p2p.PeerID
	traverseInfo []*traverseInfo
}

// NewMessage returns a new instance of a Message object
func NewMessage(message *pubsub.Message) *Message {
	msg := &Message{
		from:         message.From,
		data:         message.Data,
		seqNo:        message.Seqno,
		topicIds:     message.TopicIDs,
		signature:    message.Signature,
		key:          message.Key,
		traverseInfo: make([]*traverseInfo, len(message.Traverse)),
	}

	for i, t := range message.Traverse {
		id, _ := peer.IDFromBytes(t.Key)
		msg.traverseInfo[i] = &traverseInfo{
			peer:      p2p.PeerID(id),
			timestamp: time.Unix(0, *t.Timestamp),
			tag:       *t.Tag,
		}
	}

	id, err := peer.IDFromBytes(msg.from)
	if err != nil {
		log.Trace("creating new p2p message", "error", err.Error())
	} else {
		msg.peer = p2p.PeerID(id)
	}

	return msg
}

// From returns the message originator's peer ID
func (m *Message) From() []byte {
	return m.from
}

// Data returns the message payload
func (m *Message) Data() []byte {
	return m.data
}

// SeqNo returns the message sequence number
func (m *Message) SeqNo() []byte {
	return m.seqNo
}

// TopicIDs returns the topic on which the message was sent
func (m *Message) TopicIDs() []string {
	return m.topicIds
}

// Signature returns the message signature
func (m *Message) Signature() []byte {
	return m.signature
}

// Key returns the message public key (if it can not be recovered from From field)
func (m *Message) Key() []byte {
	return m.key
}

// Peer returns the peer that originated the message
func (m *Message) Peer() p2p.PeerID {
	return m.peer
}

func (m *Message) TraverseInfoTable() string {
	hdr := []string{"pid", "timestamp", "tag"}
	lds := make([]*display.LineData, 0)

	for _, ti := range m.traverseInfo {
		ld := display.NewLineData(false, []string{
			ti.peer.Pretty(),
			ti.timestamp.Format("05.000000"),
			ti.tag,
		})

		lds = append(lds, ld)
	}
	tbl, _ := display.CreateTableString(hdr, lds)

	return tbl
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *Message) IsInterfaceNil() bool {
	return m == nil
}
