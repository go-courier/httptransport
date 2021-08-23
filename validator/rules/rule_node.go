package rules

import (
	"bytes"
)

type RuleNode interface {
	node()
	Bytes() []byte
}

func NewRule(name string) *Rule {
	return &Rule{
		Name: name,
	}
}

type Rule struct {
	RAW []byte

	Name   string
	Params []RuleNode

	Range          []*RuleLit
	ExclusiveLeft  bool
	ExclusiveRight bool

	ValueMatrix [][]*RuleLit

	Pattern string

	Optional     bool
	DefaultValue []byte

	RuleNode
}

func (r *Rule) ComputedValues() []*RuleLit {
	return computedValueMatrix(r.ValueMatrix)
}

func computedValueMatrix(valueMatrix [][]*RuleLit) []*RuleLit {
	switch len(valueMatrix) {
	case 0:
		return nil
	case 1:
		return valueMatrix[0]
	default:
		rowI := valueMatrix[0]
		rowJ := valueMatrix[1]

		nI := len(rowI)
		nJ := len(rowJ)

		values := make([]*RuleLit, nI*nJ)

		for i := range rowI {
			for j := range rowJ {
				values[i*nJ+j] = NewRuleLit(append(append([]byte{}, rowI[i].Bytes()...), rowJ[j].Bytes()...))
			}
		}

		return computedValueMatrix(append([][]*RuleLit{values}, valueMatrix[2:]...))
	}
}

func (r *Rule) Bytes() []byte {
	if r == nil {
		return nil
	}

	buf := &bytes.Buffer{}
	buf.WriteByte('@')
	buf.WriteString(r.Name)

	if len(r.Params) > 0 {
		buf.WriteByte('<')
		for i, p := range r.Params {
			if i > 0 {
				buf.WriteByte(',')
			}
			if p != nil {
				buf.Write(p.Bytes())
			}
		}
		buf.WriteByte('>')
	}

	if len(r.Range) > 0 {
		if r.ExclusiveLeft {
			buf.WriteRune('(')
		} else {
			buf.WriteRune('[')
		}
		for i, p := range r.Range {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(p.Bytes())
		}
		if r.ExclusiveRight {
			buf.WriteRune(')')
		} else {
			buf.WriteRune(']')
		}
	}

	for i := range r.ValueMatrix {
		values := r.ValueMatrix[i]

		buf.WriteByte('{')

		for i, p := range values {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(p.Bytes())
		}

		buf.WriteByte('}')
	}

	if r.Pattern != "" {
		buf.Write(Slash([]byte(r.Pattern)))
	}

	if r.Optional {
		if r.DefaultValue != nil {
			buf.WriteByte(' ')
			buf.WriteByte('=')
			buf.WriteByte(' ')

			buf.Write(SingleQuote(r.DefaultValue))
		} else {
			buf.WriteByte('?')
		}
	}

	return buf.Bytes()
}

func NewRuleLit(lit []byte) *RuleLit {
	return &RuleLit{
		Lit: lit,
	}
}

type RuleLit struct {
	Lit []byte
	RuleNode
}

func (lit *RuleLit) Append(b []byte) {
	lit.Lit = append(lit.Lit, b...)
}

func (lit *RuleLit) Bytes() []byte {
	if lit == nil {
		return nil
	}
	return lit.Lit
}
