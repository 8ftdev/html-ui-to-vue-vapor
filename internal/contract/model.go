// Package contract parses the constrained html-ui source convention without evaluating it.
package contract

import (
	"encoding/json"
	"fmt"
)

type Position struct{ Offset, Line, Column int }
type Field struct {
	Name, Type string
	Optional   bool
	Pos        Position
}
type Prop struct {
	Field
	Default json.RawMessage
}
type Slot struct {
	Name     string
	Optional bool
	Scope    []Field
	Pos      Position
}
type StateRead struct {
	Prop, Property string
	Pos            Position
}
type Event struct {
	Name, Type, Target string
	State              []StateRead
	Pos                Position
}
type Attribute struct {
	Name, Value string
	Pos         Position
}
type Binding struct {
	Kind, Name, Prop   string
	Guarded, Stringify bool
	Pos                Position
}
type ScopeBinding struct {
	Name, Prop string
	Pos        Position
}
type Child struct {
	Node, Slot string
	Optional   bool
	Scope      []ScopeBinding
	Pos        Position
}
type Node struct {
	ID, Tag, DOMType string
	Attributes       []Attribute
	Bindings         []Binding
	Children         []Child
	Pos              Position
}
type Component struct {
	Name, Factory, Root, RootType, Behavior string
	Version                                 int
	UI                                      *UIContract
	Props                                   []Prop
	Slots                                   []Slot
	Events                                  []Event
	Nodes                                   []Node
}
type Diagnostic struct {
	Pos     Position
	Message string
}

func (d *Diagnostic) Error() string {
	return fmt.Sprintf("input:%d:%d: %s", d.Pos.Line, d.Pos.Column, d.Message)
}
func diagnostic(pos Position, message string) error { return &Diagnostic{Pos: pos, Message: message} }
