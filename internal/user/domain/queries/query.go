package queries

// Query represents a user read operation
type Query interface {
	QueryType() string
}