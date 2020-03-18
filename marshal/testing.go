package marshal

// MarshalizersAvailableForTesting represents all marshalizers registered that will be used in tests.
// In this manner we assure that any modification on a serializable DTO will always be checked against
// all registered marshalizers
var MarshalizersAvailableForTesting = map[string]Marshalizer{
	"json":     &JsonMarshalizer{},
	"protobuf": &GogoProtoMarshalizer{},
}
