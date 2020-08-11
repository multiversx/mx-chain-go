package marshaling

import "strings"

// MarshalizerKind is the kind of a message (that is passed from the Node and the outport sink)
type MarshalizerKind uint32

const (
	// JSON is a marshalizer kind
	JSON MarshalizerKind = iota
	// Gob is a marshalizer kind
	Gob
)

// ParseKind gets a kind from a string
func ParseKind(str string) MarshalizerKind {
	str = strings.ToUpper(str)
	str = strings.Trim(str, " ")

	switch str {
	case "JSON":
		return JSON
	case "GOB":
		return Gob
	default:
		return JSON
	}
}

// Marshalizer deals with messages serialization
type Marshalizer interface {
	Marshal(data interface{}) ([]byte, error)
	Unmarshal(data interface{}, dataBytes []byte) error
	IsInterfaceNil() bool
}
