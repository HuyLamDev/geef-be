package queries

// Query represents a project read operation
type Query interface {
	QueryType() string
}