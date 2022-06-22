package enablers

type roundFlagsHolder struct {
	example *roundFlag
}

// IsExampleEnabled provides an example how to use this component. Replace this getter with the first use case
func (holder *roundFlagsHolder) IsExampleEnabled() bool {
	return holder.example.IsSet()
}
