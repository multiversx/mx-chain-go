package marshal

var DefMarshalizer Marshalizer

func init() {
	DefMarshalizer = &JsonMarshalizer{}
}
