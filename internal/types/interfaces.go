package types

import (
	"github.com/charmbracelet/bubbles/table"
)

// ResourceData defines the interface for resource data
type ResourceData interface {
	GetName() string
	GetNamespace() string
	GetColumns() table.Row
}
