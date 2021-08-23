package transformers

import (
	"fmt"
	"strings"
)

type PathWalker struct {
	path []interface{}
}

func (pw *PathWalker) Enter(i interface{}) {
	pw.path = append(pw.path, i)
}

func (pw *PathWalker) Exit() {
	pw.path = pw.path[:len(pw.path)-1]
}

func (pw *PathWalker) Paths() []interface{} {
	return pw.path
}

func (pw *PathWalker) String() string {
	b := &strings.Builder{}
	for i := 0; i < len(pw.path); i++ {
		switch x := pw.path[i].(type) {
		case string:
			if b.Len() != 0 {
				b.WriteByte('.')
			}
			b.WriteString(x)
		case int:
			b.WriteString(fmt.Sprintf("[%d]", x))
		}
	}
	return b.String()
}
