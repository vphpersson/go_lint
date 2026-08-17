// Package gclplugin registers the analyzers as golangci-lint module plugins. They are registered
// separately, each under its own name, so that a configuration may enable one without the other.
package gclplugin

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/vphpersson/go_lint/pkg/analyzer/struct_tag"
	"github.com/vphpersson/go_lint/pkg/analyzer/variadic_append"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("structtaglint", newStructTagPlugin)
	register.Plugin("variadicappend", newVariadicAppendPlugin)
}

// plugin serves any of the analyzers: what a plugin does is decided by the analyzer it is built
// with, and none of them needs settings of its own.
type plugin struct {
	analyzer *analysis.Analyzer
	loadMode string
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{p.analyzer}, nil
}

func (p *plugin) GetLoadMode() string {
	return p.loadMode
}

func newStructTagPlugin(_ any) (register.LinterPlugin, error) {
	return &plugin{analyzer: struct_tag.Analyzer, loadMode: register.LoadModeSyntax}, nil
}

// newVariadicAppendPlugin asks for type information: an identifier is matched against the parameter
// object it resolves to, rather than against a name.
func newVariadicAppendPlugin(_ any) (register.LinterPlugin, error) {
	return &plugin{analyzer: variadic_append.Analyzer, loadMode: register.LoadModeTypesInfo}, nil
}
