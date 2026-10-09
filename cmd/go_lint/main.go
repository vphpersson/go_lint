// Command go_lint runs the analyzers on their own, for use outside golangci-lint.
package main

import (
	"github.com/vphpersson/go_lint/pkg/analyzer/struct_tag"
	"github.com/vphpersson/go_lint/pkg/analyzer/url_concat"
	"github.com/vphpersson/go_lint/pkg/analyzer/variadic_append"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		struct_tag.Analyzer,
		url_concat.Analyzer,
		variadic_append.Analyzer,
	)
}
