package parse

import (
	"fmt"
)

type Node interface {
	Type() NodeType
	String() (string, error)
}

// NodeType identifies the type of a node.
type NodeType int

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeText NodeType = iota
	NodeSubstitution
	NodeVariable
)

type TextNode struct {
	NodeType
	Text string
}

func NewText(text string) *TextNode {
	return &TextNode{NodeText, text}
}

func (t *TextNode) String() (string, error) {
	return t.Text, nil
}

type VariableNode struct {
	NodeType
	Ident       string
	Env         Env
	Restrict    *Restrictions
	AllowedVars map[string]bool // nil means all vars allowed
	OriginalSrc string          // Original source like "$VAR" for literal output when not allowed
}

func NewVariable(ident string, env Env, restrict *Restrictions, allowedVars map[string]bool) *VariableNode {
	return &VariableNode{
		NodeType:    NodeVariable,
		Ident:       ident,
		Env:         env,
		Restrict:    restrict,
		AllowedVars: allowedVars,
	}
}

func (t *VariableNode) String() (string, error) {
	// If filtering is enabled and this var is not in the allowed list,
	// return original source as literal text
	if t.AllowedVars != nil && !t.AllowedVars[t.Ident] {
		if t.OriginalSrc != "" {
			return t.OriginalSrc, nil
		}
		return "$" + t.Ident, nil
	}
	if err := t.validateNoUnset(); err != nil {
		return "", err
	}
	value := t.Env.Get(t.Ident)
	if err := t.validateNoEmpty(value); err != nil {
		return "", err
	}
	return value, nil
}

func (t *VariableNode) isSet() bool {
	return t.Env.Has(t.Ident)
}

func (t *VariableNode) validateNoUnset() error {
	if t.Restrict.NoUnset && !t.isSet() {
		return fmt.Errorf("variable ${%s} not set", t.Ident)
	}
	return nil
}

func (t *VariableNode) validateNoEmpty(value string) error {
	if t.Restrict.NoEmpty && value == "" && t.isSet() {
		return fmt.Errorf("variable ${%s} set but empty", t.Ident)
	}
	return nil
}

type SubstitutionNode struct {
	NodeType
	ExpType     itemType
	Variable    *VariableNode
	Default     Node   // Default could be variable or text
	OriginalSrc string // Original source like "${VAR:-default}" for literal output when not allowed
}

func (t *SubstitutionNode) String() (string, error) {
	// If filtering is enabled and this var is not in the allowed list,
	// return original source as literal text
	if t.Variable.AllowedVars != nil && !t.Variable.AllowedVars[t.Variable.Ident] {
		if t.OriginalSrc != "" {
			return t.OriginalSrc, nil
		}
		// Fallback: reconstruct basic form
		return "${" + t.Variable.Ident + "}", nil
	}
	if t.ExpType >= itemPlus && t.Default != nil {
		switch t.ExpType {
		case itemColonDash, itemColonEquals:
			if s, _ := t.Variable.String(); s != "" {
				return s, nil
			}
			return t.Default.String()
		case itemPlus, itemColonPlus:
			if t.Variable.isSet() {
				return t.Default.String()
			}
			return "", nil
		default:
			if !t.Variable.isSet() {
				return t.Default.String()
			}
		}
	}
	return t.Variable.String()
}
