package generator

import (
	"sort"
	"strings"

	"github.com/go-courier/httptransport/openapi/generator"
	"github.com/go-courier/oas"
)

func mayPrefixDeprecated(desc string, deprecated bool) []string {
	comments := []string{desc}
	if deprecated {
		comments = append([]string{"@deprecated"}, comments...)
	}
	return comments
}

func eachOperation(openapi *oas.OpenAPI, mapper func(method string, path string, op *oas.Operation)) {
	ops := map[string]struct {
		Method string
		Path   string
		*oas.Operation
	}{}

	operationIDs := make([]string, 0)

	for path := range openapi.Paths.Paths {
		pathItems := openapi.Paths.Paths[path]
		for method := range pathItems.Operations.Operations {
			op := pathItems.Operations.Operations[method]

			if strings.HasPrefix(op.OperationId, "OpenAPI") {
				continue
			}

			if strings.HasPrefix(op.OperationId, "ER") {
				continue
			}

			ops[op.OperationId] = struct {
				Method string
				Path   string
				*oas.Operation
			}{
				Method:    strings.ToUpper(string(method)),
				Path:      toColonPath(path),
				Operation: op,
			}
			operationIDs = append(operationIDs, op.OperationId)
		}
	}

	sort.Strings(operationIDs)

	for _, id := range operationIDs {
		op := ops[id]
		mapper(op.Method, op.Path, op.Operation)
	}
}

func requestBodyMediaType(requestBody *oas.RequestBody) *oas.MediaType {
	if requestBody == nil {
		return nil
	}

	for contentType := range requestBody.Content {
		mediaType := requestBody.Content[contentType]
		return mediaType
	}
	return nil
}

func mediaTypeAndStatusErrors(responses *oas.Responses) (*oas.MediaType, []string) {
	if responses == nil {
		return nil, nil
	}

	response := (*oas.Response)(nil)

	statusErrors := make([]string, 0)

	for code := range responses.Responses {
		if isOk(code) {
			response = responses.Responses[code]
		} else {
			extensions := responses.Responses[code].Extensions

			if extensions != nil {
				if errors, ok := extensions[generator.XStatusErrs]; ok {
					if errs, ok := errors.([]interface{}); ok {
						for _, err := range errs {
							statusErrors = append(statusErrors, err.(string))
						}
					}
				}
			}
		}
	}

	sort.Strings(statusErrors)

	if response == nil {
		return nil, nil
	}

	for contentType := range response.Content {
		mediaType := response.Content[contentType]
		return mediaType, statusErrors
	}

	return nil, statusErrors
}
