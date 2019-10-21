package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-courier/packagesx"
)

func ExampleNewRouterScanner() {
	cwd, _ := os.Getwd()
	pkg, _ := packagesx.Load(filepath.Join(cwd, "./__examples__/router_scanner"))

	router := pkg.Var("Router")

	scanner := NewRouterScanner(pkg)
	routes := scanner.Router(router).Routes(pkg)

	for _, r := range routes {
		fmt.Println(r.String())
	}
	// Output:
	// GET /root/:id httptransport.GroupOperator auth.Auth main.Get
	// HEAD /root/group/health httptransport.GroupOperator httptransport.GroupOperator httptransport.GroupOperator group.Health
}
