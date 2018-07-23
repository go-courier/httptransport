package transformers

import (
	"fmt"
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
	pathString := ""
	for i := 0; i < len(pw.path); i++ {
		switch pw.path[i].(type) {
		case string:
			if pathString != "" {
				pathString += "."
			}
			pathString += pw.path[i].(string)
		case int:
			pathString += fmt.Sprintf("[%d]", pw.path[i].(int))
		}
	}
	return pathString
}
