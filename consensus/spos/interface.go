package spos

// SposFactory defines an interface for SPoS implementations
type SposFactory interface {
	GenerateSubrounds()
}
