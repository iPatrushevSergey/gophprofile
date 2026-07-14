package port

// IDGenerator creates new avatar identifiers.
type IDGenerator interface {
	NewID() (string, error)
}
