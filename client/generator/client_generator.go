package generator

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/go-courier/codegen"
	"github.com/go-courier/oas"
	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

func OptionVendorImportByGoMod() GenOptionFn {
	return func(o *GenOption) {
		o.VendorImportByGoMod = true
	}
}

type GenOptionFn = func(o *GenOption)

type GenOption struct {
	VendorImportByGoMod bool `name:"vendor-import-by-go-mod" usage:"when enable vendor only import pkg exists in go mod"`
}

func NewClientGenerator(serviceName string, u *url.URL, opts ...GenOptionFn) *ClientGenerator {
	g := &ClientGenerator{
		ServiceName: serviceName,
		URL:         u,
		openAPI:     &oas.OpenAPI{},
	}

	for _, o := range opts {
		o(&g.GenOption)
	}

	return g
}

type ClientGenerator struct {
	ServiceName string
	URL         *url.URL
	openAPI     *oas.OpenAPI

	GenOption
}

func (g *ClientGenerator) Load() {
	if g.URL == nil {
		panic(errors.Errorf("missing spec-url or file"))
	}

	if g.URL.Scheme == "file" {
		g.loadByFile()
	} else {
		g.loadBySpecURL()
	}
}

func (g *ClientGenerator) loadByFile() {
	data, err := ioutil.ReadFile(g.URL.Path)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, g.openAPI); err != nil {
		panic(err)
	}
}

func (g *ClientGenerator) loadBySpecURL() {
	hc := http.Client{}
	req, err := http.NewRequest("GET", g.URL.String(), nil)
	if err != nil {
		panic(err)
	}

	resp, err := hc.Do(req)
	if err != nil {
		panic(err)
	}

	if err := json.NewDecoder(resp.Body).Decode(g.openAPI); err != nil {
		panic(err)
	}
}

func (g *ClientGenerator) vendorImportsByGoMod(cwd string) map[string]bool {
	vendorImports := map[string]bool{}

	if g.VendorImportByGoMod {
		pkgs, err := packages.Load(nil, "std")
		if err != nil {
			panic(err)
		}

		for _, p := range pkgs {
			vendorImports[p.PkgPath] = true
		}

		d := cwd

		for d != "/" {
			gmodfile := filepath.Join(d, "go.mod")

			if data, err := os.ReadFile(gmodfile); err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
			} else {
				f, _ := modfile.Parse(gmodfile, data, nil)

				vendorImports[f.Module.Mod.Path] = true

				for _, r := range f.Require {
					vendorImports[r.Mod.Path] = true
				}

				break
			}

			d = filepath.Join(d, "../")
		}
	}

	return vendorImports
}

type contextVendorImports int

func WithVendorImports(ctx context.Context, vendorImports map[string]bool) context.Context {
	return context.WithValue(ctx, contextVendorImports(1), vendorImports)
}

func VendorImportsFromContext(ctx context.Context) map[string]bool {
	if v, ok := ctx.Value(contextVendorImports(1)).(map[string]bool); ok {
		return v
	}
	return map[string]bool{}
}

func (g *ClientGenerator) Output(cwd string) {
	pkgName := codegen.LowerSnakeCase("Client-" + g.ServiceName)
	rootPath := path.Join(cwd, pkgName)

	ctx := WithVendorImports(context.Background(), g.vendorImportsByGoMod(cwd))

	{
		file := codegen.NewFile(pkgName, path.Join(rootPath, "client.go"))
		NewServiceClientGenerator(g.ServiceName, file).Scan(ctx, g.openAPI)
		_, _ = file.WriteFile()
	}

	{
		file := codegen.NewFile(pkgName, path.Join(rootPath, "operations.go"))
		NewOperationGenerator(g.ServiceName, file).Scan(ctx, g.openAPI)
		_, _ = file.WriteFile()
	}

	{
		file := codegen.NewFile(pkgName, path.Join(rootPath, "types.go"))
		NewTypeGenerator(g.ServiceName, file).Scan(ctx, g.openAPI)
		_, _ = file.WriteFile()
	}

	log.Printf("generated client of %s into %s", g.ServiceName, color.MagentaString(rootPath))
}
