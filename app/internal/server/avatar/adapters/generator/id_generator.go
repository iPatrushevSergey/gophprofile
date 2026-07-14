package generator

import (
	"github.com/google/uuid"

	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

var _ appport.IDGenerator = (*IDGenerator)(nil)

// IDGenerator generates identifiers as UUID v7 strings.
type IDGenerator struct{}

// NewIDGenerator returns a UUID-based id generator.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// NewID returns a new identifier.
func (g *IDGenerator) NewID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
