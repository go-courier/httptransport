package generator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/fatih/color"
	"github.com/go-courier/codegen"
	"github.com/go-courier/oas"
)

func NewClientGenerator(serviceName string, u *url.URL) *ClientGenerator {
	return &ClientGenerator{
		ServiceName: serviceName,
		URL:         u,
		openAPI:     &oas.OpenAPI{},
	}
}

type ClientGenerator struct {
	ServiceName string
	URL         *url.URL
	openAPI     *oas.OpenAPI
}

func (g *ClientGenerator) Load() {
	if g.URL == nil {
		panic(fmt.Errorf("missing spec-url or file"))
		return
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

func (g *ClientGenerator) Output(cwd string) {
	pkgName := codegen.LowerSnakeCase("Client-" + g.ServiceName)
	rootPath := path.Join(cwd, pkgName)

	{
		file := codegen.NewFile(pkgName, path.Join(rootPath, "client.go"))
		NewServiceClientGenerator(g.ServiceName, file).Scan(g.openAPI)
		file.WriteFile()
	}

	{
		file := codegen.NewFile(pkgName, path.Join(rootPath, "operations.go"))
		NewOperationGenerator(g.ServiceName, file).Scan(g.openAPI)
		file.WriteFile()
	}

	{
		file := codegen.NewFile(pkgName, path.Join(rootPath, "types.go"))
		NewTypeGenerator(g.ServiceName, file).Scan(g.openAPI)
		file.WriteFile()
	}

	log.Printf("generated client of %s into %s", g.ServiceName, color.MagentaString(rootPath))
}
