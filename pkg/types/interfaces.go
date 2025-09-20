package types

import (
	"github.com/charmbracelet/bubbles/table"
)

type ResourceData interface {
	GetName() string
	GetNamespace() string
	GetColumns() table.Row
}
