package generator

import (
	"go/types"
	"reflect"
	"strings"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/packagesx"

	"github.com/go-courier/httptransport"
)

const (
	XID           = "x-id"
	XGoVendorType = `x-go-vendor-type`
	XGoStarLevel  = `x-go-star-level`
	XGoFieldName  = `x-go-field-name`

	XTagValidate = `x-tag-validate`
	XTagMime     = `x-tag-mime`
	XTagJSON     = `x-tag-json`
	XTagXML      = `x-tag-xml`
	XTagName     = `x-tag-name`

	XEnumLabels = `x-enum-labels`
	// Deprecated  use XEnumLabels
	XEnumOptions = `x-enum-options`
	XStatusErrs  = `x-status-errors`
)

var (
	pkgImportPathHttpTransport = packagesx.ImportGoPath(reflect.TypeOf(httptransport.HttpRouteMeta{}).PkgPath())
	pkgImportPathHttpx         = packagesx.ImportGoPath(reflect.TypeOf(httpx.Response{}).PkgPath())
	pkgImportPathCourier       = packagesx.ImportGoPath(reflect.TypeOf(courier.Router{}).PkgPath())
)

func isRouterType(typ types.Type) bool {
	return strings.HasSuffix(typ.String(), pkgImportPathCourier+".Router")
}

func isHttpxResponse(typ types.Type) bool {
	return strings.HasSuffix(typ.String(), pkgImportPathHttpx+".Response")
}

func isFromHttpTransport(typ types.Type) bool {
	return strings.Contains(typ.String(), pkgImportPathHttpTransport+".")
}

func tagValueAndFlagsByTagString(tagString string) (string, map[string]bool) {
	valueAndFlags := strings.Split(tagString, ",")
	v := valueAndFlags[0]
	tagFlags := map[string]bool{}
	if len(valueAndFlags) > 1 {
		for _, flag := range valueAndFlags[1:] {
			tagFlags[flag] = true
		}
	}
	return v, tagFlags
}

func filterMarkedLines(comments []string) []string {
	lines := make([]string, 0)
	for _, line := range comments {
		if !strings.HasPrefix(line, "@") {
			lines = append(lines, line)
		}
	}
	return lines
}

func dropMarkedLines(lines []string) string {
	return strings.Join(filterMarkedLines(lines), "\n")
}
