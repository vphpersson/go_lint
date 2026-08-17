// Package analyzer reports appending to a variadic parameter. A variadic parameter is a slice over
// an array the caller may still hold: when the caller spreads a slice of its own, as in f(options...),
// the parameter shares that array, and appending to it writes into the caller's slice whenever there
// is spare capacity. The write is silent, happens only for some inputs, and surfaces as one call site
// mysteriously affecting another. Append to a copy instead, e.g. append(slices.Clone(options), x).
package variadic_append

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const appendBuiltinName = "append"

var Analyzer = &analysis.Analyzer{
	Name: "variadicappend",
	Doc:  "Reports appending to a variadic parameter, which can write into the caller's backing array.",
	Run:  run,
}

// variadicParameter returns the variadic parameter of a signature, or nil if it has none. It is
// always the last one, and it is the only parameter whose array the caller may share.
func variadicParameter(signature *types.Signature) *types.Var {
	if signature == nil || !signature.Variadic() {
		return nil
	}

	parameters := signature.Params()
	if parameters == nil || parameters.Len() == 0 {
		return nil
	}

	return parameters.At(parameters.Len() - 1)
}

// appendedIdentifier returns the identifier an append call appends to, seeing through a slice
// expression: append(parameter[:0], x) writes into the same array that append(parameter, x) does.
func appendedIdentifier(expression ast.Expr) *ast.Ident {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed
	case *ast.SliceExpr:
		identifier, _ := typed.X.(*ast.Ident)
		return identifier
	default:
		return nil
	}
}

func run(pass *analysis.Pass) (any, error) {
	typesInfo := pass.TypesInfo
	if typesInfo == nil {
		return nil, nil
	}

	// Collected by object rather than by name, so that a local variable shadowing a parameter is
	// not mistaken for it.
	variadicParameters := make(map[types.Object]struct{})

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			var signature *types.Signature

			switch typed := node.(type) {
			case *ast.FuncDecl:
				if typed.Name != nil {
					if function, ok := typesInfo.Defs[typed.Name].(*types.Func); ok {
						signature, _ = function.Type().(*types.Signature)
					}
				}
			case *ast.FuncLit:
				signature, _ = typesInfo.TypeOf(typed).(*types.Signature)
			default:
				return true
			}

			if parameter := variadicParameter(signature); parameter != nil {
				variadicParameters[parameter] = struct{}{}
			}

			return true
		})
	}

	if len(variadicParameters) == 0 {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			callExpression, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Only an append that adds something can write into the array; append(parameter) alone
			// does nothing.
			if len(callExpression.Args) < 2 {
				return true
			}

			functionIdentifier, ok := callExpression.Fun.(*ast.Ident)
			if !ok || functionIdentifier.Name != appendBuiltinName {
				return true
			}

			// The builtin, rather than something of the same name declared locally.
			if _, ok := typesInfo.Uses[functionIdentifier].(*types.Builtin); !ok {
				return true
			}

			identifier := appendedIdentifier(callExpression.Args[0])
			if identifier == nil {
				return true
			}

			object := typesInfo.Uses[identifier]
			if object == nil {
				return true
			}

			if _, ok := variadicParameters[object]; !ok {
				return true
			}

			pass.Reportf(
				callExpression.Pos(),
				"append to variadic parameter %q writes into the caller's array when it has spare capacity; append to a copy, e.g. append(slices.Clone(%s), ...)",
				identifier.Name,
				identifier.Name,
			)

			return true
		})
	}

	return nil, nil
}
