// Package analyzer reports struct tags whose name part is a serialization option word, e.g. `json:"omitempty"` where `json:",omitempty"` was intended. Such a field is not omitted; it is serialized under the option word as its wire name, which can expose fields the developer meant to hide. See https://blog.trailofbits.com/2025/06/17/unexpected-security-footguns-in-gos-parsers/.
package struct_tag

import (
	"go/ast"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const omitemptyOptionName = "omitempty"

var tagChecks = []struct {
	key         string
	optionNames map[string]struct{}
}{
	{key: "json", optionNames: map[string]struct{}{omitemptyOptionName: {}, "omitzero": {}}},
	{key: "yaml", optionNames: map[string]struct{}{omitemptyOptionName: {}}},
	{key: "xml", optionNames: map[string]struct{}{omitemptyOptionName: {}}},
}

var Analyzer = &analysis.Analyzer{
	Name: "structtaglint",
	Doc:  "Reports struct tags whose name part is a serialization option word (e.g. `json:\"omitempty\"` instead of `json:\",omitempty\"`).",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			structType, ok := node.(*ast.StructType)
			if !ok || structType.Fields == nil {
				return true
			}

			for _, field := range structType.Fields.List {
				fieldTag := field.Tag
				if fieldTag == nil {
					continue
				}

				tagValue, err := strconv.Unquote(fieldTag.Value)
				if err != nil {
					continue
				}

				for _, tagCheck := range tagChecks {
					keyValue, ok := reflect.StructTag(tagValue).Lookup(tagCheck.key)
					if !ok {
						continue
					}

					name, _, _ := strings.Cut(keyValue, ",")
					if _, ok := tagCheck.optionNames[name]; !ok {
						continue
					}

					pass.Reportf(
						fieldTag.Pos(),
						"%s tag name %q is an option word; the field is serialized under the name %q, not omitted; write %q to apply the option",
						tagCheck.key,
						name,
						name,
						","+name,
					)
				}
			}

			return true
		})
	}

	return nil, nil
}
