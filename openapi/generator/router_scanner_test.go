package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-courier/loaderx"
)

func ExampleNewRouterScanner() {
	cwd, _ := os.Getwd()
	program, pkgInfo, _ := loaderx.LoadWithTests(filepath.Join(cwd, "./__examples__/router_scanner"))

	info := loaderx.NewPackageInfo(pkgInfo)

	router := info.Var("Router")

	scanner := NewRouterScanner(program)
	routes := scanner.Router(router).Routes(program)

	for _, r := range routes {
		fmt.Println(r.String())
	}
	// Output:
	// GET /root/:id httptransport.GroupOperator auth.Auth main.Get
	// HEAD /root/group/health httptransport.GroupOperator httptransport.GroupOperator group.Health
}
