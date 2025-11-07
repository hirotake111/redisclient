package infoid

import (
	"github.com/google/uuid"
)

// Info ID represents a unique identifier for informational messages.
type InfoID uuid.UUID

func New() InfoID {
	return InfoID(uuid.New())
}
