// Package url_concat reports URLs and query strings built by string concatenation or fmt.Sprintf. Concatenation escapes nothing, so a value containing "/", "?", "&", "#" or "@" changes the structure of the URL rather than its content; build a url.URL and encode parameters with url.Values instead.
package url_concat

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

// hole stands in, within the text of a built string, for an operand whose value is not known.
const hole = "\x00"

const concatenationDescription = "string concatenation"

const (
	urlMessage   = "URL built by %s; build a url.URL and call String(), e.g. new(url.URL{Scheme: \"https\", Host: host, Path: path, RawQuery: url.Values{...}.Encode()}).String()"
	queryMessage = "query string built by %s; encode the parameters with url.Values{...}.Encode()"
)

var Analyzer = &analysis.Analyzer{
	Name: "urlconcat",
	Doc:  "Reports URLs and query strings built by string concatenation or fmt.Sprintf rather than with net/url.",
	Run:  run,
}

// urlArguments maps functions that take a URL as a string to the index of that argument.
var urlArguments = map[string]int{
	"net/url.Parse":                  0,
	"net/url.ParseRequestURI":        0,
	"net/http.NewRequest":            1,
	"net/http.NewRequestWithContext": 2,
	"net/http.Get":                   0,
	"net/http.Head":                  0,
	"net/http.Post":                  0,
	"net/http.PostForm":              0,
	"net/http.Redirect":              2,
	"(*net/http.Client).Get":         0,
	"(*net/http.Client).Head":        0,
	"(*net/http.Client).Post":        0,
	"(*net/http.Client).PostForm":    0,
}

// formatArguments maps formatting functions to the index of their format argument.
var formatArguments = map[string]int{
	"fmt.Sprintf": 0,
	"fmt.Appendf": 1,
}

// urlStringMethods are the methods that render a url.URL as a string.
var urlStringMethods = map[string]struct{}{
	"(*net/url.URL).String":   {},
	"(*net/url.URL).Redacted": {},
}

type finding int

const (
	findingNone finding = iota
	findingURL
	findingQuery
)

func (f finding) message() string {
	if f == findingQuery {
		return queryMessage
	}
	return urlMessage
}

// part is one piece of a built string: a constant text, or a hole for an operand whose value is not known. A hole's expression is nil when the operand cannot be identified.
type part struct {
	text       string
	isHole     bool
	expression ast.Expr
}

type checker struct {
	pass *analysis.Pass
	// examined holds the expressions already examined, as part of a larger one or as a URL argument, so that each is reported at most once.
	examined map[ast.Node]struct{}
}

func run(pass *analysis.Pass) (any, error) {
	if pass.TypesInfo == nil {
		return nil, nil
	}

	c := &checker{pass: pass, examined: make(map[ast.Node]struct{})}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			if _, ok := c.examined[node]; ok {
				return true
			}

			switch typed := node.(type) {
			case *ast.CallExpr:
				c.checkCall(typed)
			case *ast.BinaryExpr:
				c.checkConcatenation(typed)
			case *ast.AssignStmt:
				c.checkAssignment(typed)
			case *ast.CompositeLit:
				c.checkCompositeLiteral(typed)
			}

			return true
		})
	}

	return nil, nil
}

func (c *checker) calleeName(call *ast.CallExpr) string {
	function := typeutil.StaticCallee(c.pass.TypesInfo, call)
	if function == nil {
		return ""
	}
	return function.FullName()
}

func (c *checker) constantString(expression ast.Expr) (string, bool) {
	typeAndValue, ok := c.pass.TypesInfo.Types[expression]
	if !ok || typeAndValue.Value == nil || typeAndValue.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(typeAndValue.Value), true
}

func isString(t types.Type) bool {
	if t == nil {
		return false
	}
	basic, ok := t.Underlying().(*types.Basic)
	return ok && basic.Info()&types.IsString != 0
}

func isURLType(t types.Type) bool {
	if t == nil {
		return false
	}
	if pointer, ok := types.Unalias(t).(*types.Pointer); ok {
		t = pointer.Elem()
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}
	object := named.Obj()
	return object != nil && object.Pkg() != nil && object.Pkg().Path() == "net/url" && object.Name() == "URL"
}

// concatenation returns the expression as a string concatenation whose value is not constant.
func (c *checker) concatenation(expression ast.Expr) (*ast.BinaryExpr, bool) {
	binary, ok := ast.Unparen(expression).(*ast.BinaryExpr)
	if !ok || binary.Op != token.ADD || !isString(c.pass.TypesInfo.TypeOf(binary)) {
		return nil, false
	}
	if _, isConstant := c.constantString(binary); isConstant {
		return nil, false
	}
	return binary, true
}

// formatCall returns the expression as a call to a formatting function, with the function's name.
func (c *checker) formatCall(expression ast.Expr) (*ast.CallExpr, string, bool) {
	call, ok := ast.Unparen(expression).(*ast.CallExpr)
	if !ok {
		return nil, "", false
	}
	name := c.calleeName(call)
	if _, ok := formatArguments[name]; !ok {
		return nil, "", false
	}
	return call, name, true
}

// built describes how an expression builds a string, if it is a concatenation or a formatting call.
func (c *checker) built(expression ast.Expr) (string, bool) {
	if _, ok := c.concatenation(expression); ok {
		return concatenationDescription, true
	}
	if _, name, ok := c.formatCall(expression); ok {
		return name, true
	}
	return "", false
}

// examine marks an expression, and the concatenations it is made of, as examined.
func (c *checker) examine(expression ast.Expr) {
	expression = ast.Unparen(expression)
	c.examined[expression] = struct{}{}
	if binary, ok := c.concatenation(expression); ok {
		c.examine(binary.X)
		c.examine(binary.Y)
	}
}

// concatenationParts flattens a chain of concatenations into its operands, marking the chain examined.
func (c *checker) concatenationParts(expression ast.Expr) []part {
	expression = ast.Unparen(expression)
	if text, ok := c.constantString(expression); ok {
		return []part{{text: text}}
	}

	binary, ok := c.concatenation(expression)
	if !ok {
		return []part{{isHole: true, expression: expression}}
	}

	c.examined[binary] = struct{}{}
	return append(c.concatenationParts(binary.X), c.concatenationParts(binary.Y)...)
}

// formatParts splits a format string at its verbs, each a hole for the argument it formats. Holes are matched to arguments only while the verbs consume them in order.
func formatParts(format string, arguments []ast.Expr) []part {
	var parts []part
	var text strings.Builder
	argumentIndex := 0
	matchArguments := arguments != nil

	flushText := func() {
		if text.Len() > 0 {
			parts = append(parts, part{text: text.String()})
			text.Reset()
		}
	}

	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			text.WriteByte(format[i])
			continue
		}

		i++
		if i < len(format) && format[i] == '%' {
			text.WriteByte('%')
			continue
		}

		for i < len(format) && strings.IndexByte("+-# 0123456789.*[]", format[i]) >= 0 {
			if format[i] == '*' || format[i] == '[' {
				matchArguments = false
			}
			i++
		}

		flushText()

		var expression ast.Expr
		if matchArguments && argumentIndex < len(arguments) {
			expression = arguments[argumentIndex]
		}
		argumentIndex++
		parts = append(parts, part{isHole: true, expression: expression})
	}

	flushText()

	return parts
}

func isSeparator(b byte) bool {
	return strings.IndexByte(" \t\n\r\v\f\"'`<>", b) >= 0
}

func isSchemeByte(b byte) bool {
	return 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z' || '0' <= b && b <= '9' || b == '+' || b == '-' || b == '.'
}

// isPathByte reports whether a byte may end a path, as the byte before a "?" that starts a query.
func isPathByte(b byte) bool {
	return 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z' || '0' <= b && b <= '9' || strings.IndexByte("/._~%-", b) >= 0 || b == hole[0]
}

func isUpper(b byte) bool {
	return 'A' <= b && b <= 'Z'
}

// isURLName reports whether a name designates a URL: url or uri alone, or as a word at either end, as in baseURL, redirectUri, urlPrefix and api_url.
func isURLName(name string) bool {
	for _, word := range []string{"url", "uri"} {
		if strings.EqualFold(name, word) {
			return true
		}
		if len(name) <= len(word) {
			continue
		}

		suffixStart := len(name) - len(word)
		if strings.EqualFold(name[suffixStart:], word) && (isUpper(name[suffixStart]) || name[suffixStart-1] == '_') {
			return true
		}
		if strings.EqualFold(name[:len(word)], word) && (isUpper(name[len(word)]) || name[len(word)] == '_') {
			return true
		}
	}
	return false
}

// isURLValued reports whether an operand holds a URL, judged by its name or by its being a rendered url.URL.
func (c *checker) isURLValued(expression ast.Expr) bool {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.Ident:
		return isURLName(typed.Name)
	case *ast.SelectorExpr:
		return isURLName(typed.Sel.Name)
	case *ast.CallExpr:
		_, ok := urlStringMethods[c.calleeName(typed)]
		return ok
	default:
		return false
	}
}

func hasScheme(token string) bool {
	index := strings.Index(token, "://")
	if index <= 0 {
		return false
	}
	previous := token[index-1]
	return isSchemeByte(previous) || previous == hole[0]
}

// startsWithParameterHole reports whether text starts with a query parameter whose value is a hole, as "page=" followed by an operand.
func startsWithParameterHole(text string) bool {
	name, value, ok := strings.Cut(text, "=")
	return ok && name != "" && !strings.ContainsAny(name, "?&#") && strings.HasPrefix(value, hole)
}

// isQuery reports whether a token builds a query: a "?" after a path followed by an operand, or a parameter whose value is one.
func isQuery(token string) bool {
	for i := range len(token) {
		if token[i] != '?' && token[i] != '&' {
			continue
		}
		if token[i] == '?' && i > 0 && isPathByte(token[i-1]) && strings.HasPrefix(token[i+1:], hole) {
			return true
		}
		if startsWithParameterHole(token[i+1:]) {
			return true
		}
	}
	return false
}

// classifyToken classifies one whitespace-delimited token of a built string, given the operands of its holes in order.
func (c *checker) classifyToken(token string, holes []ast.Expr) finding {
	if !strings.Contains(token, hole) {
		return findingNone
	}

	if hasScheme(token) {
		return findingURL
	}

	// An operand holding a URL, extended by a path, query, fragment or another operand.
	holeIndex := 0
	for i := range len(token) {
		if token[i] != hole[0] {
			continue
		}
		expression := holes[holeIndex]
		holeIndex++
		if expression != nil && i+1 < len(token) && strings.IndexByte("/?#&"+hole, token[i+1]) >= 0 && c.isURLValued(expression) {
			return findingURL
		}
	}

	if isQuery(token) {
		return findingQuery
	}

	return findingNone
}

// classify joins the parts of a built string into text with holes and classifies each token of it, a URL finding taking precedence over a query one.
func (c *checker) classify(parts []part) finding {
	var builder strings.Builder
	holes := make([]ast.Expr, 0, len(parts))
	for _, p := range parts {
		if p.isHole {
			builder.WriteString(hole)
			holes = append(holes, p.expression)
			continue
		}
		// A NUL in the text itself would be taken for a hole.
		builder.WriteString(strings.ReplaceAll(p.text, hole, " "))
	}

	text := builder.String()
	result := findingNone
	holeIndex := 0
	start := 0
	for i := 0; i <= len(text); i++ {
		if i < len(text) && !isSeparator(text[i]) {
			continue
		}

		token := text[start:i]
		start = i + 1

		holeCount := strings.Count(token, hole)
		tokenHoles := holes[holeIndex : holeIndex+holeCount]
		holeIndex += holeCount

		switch c.classifyToken(token, tokenHoles) {
		case findingURL:
			return findingURL
		case findingQuery:
			result = findingQuery
		case findingNone:
		}
	}

	return result
}

func (c *checker) report(node ast.Node, f finding, description string, parts []part) {
	// Operands that are themselves built strings are covered by this report.
	for _, p := range parts {
		if p.expression != nil {
			if _, ok := c.built(p.expression); ok {
				c.examine(p.expression)
			}
		}
	}
	c.pass.Reportf(node.Pos(), f.message(), description)
}

// reportBuiltArgument reports an expression used where a URL or query is expected, if it is built by concatenation or formatting.
func (c *checker) reportBuiltArgument(expression ast.Expr, f finding) {
	description, ok := c.built(expression)
	if !ok {
		return
	}
	c.examine(expression)
	c.pass.Reportf(expression.Pos(), f.message(), description)
}

func (c *checker) checkCall(call *ast.CallExpr) {
	name := c.calleeName(call)

	// A pattern that matches URLs is not a URL.
	if strings.HasPrefix(name, "regexp.") {
		for _, argument := range call.Args {
			c.examine(argument)
		}
		return
	}

	if index, ok := urlArguments[name]; ok && index < len(call.Args) {
		c.reportBuiltArgument(call.Args[index], findingURL)
		return
	}

	formatIndex, ok := formatArguments[name]
	if !ok || formatIndex >= len(call.Args) {
		return
	}

	format, ok := c.constantString(call.Args[formatIndex])
	if !ok {
		return
	}

	// A spread argument list cannot be matched to the verbs.
	var arguments []ast.Expr
	if !call.Ellipsis.IsValid() {
		arguments = call.Args[formatIndex+1:]
	}

	parts := formatParts(format, arguments)
	if f := c.classify(parts); f != findingNone {
		c.report(call, f, name, parts)
	}
}

func (c *checker) checkConcatenation(binary *ast.BinaryExpr) {
	if _, ok := c.concatenation(binary); !ok {
		return
	}

	parts := c.concatenationParts(binary)
	if f := c.classify(parts); f != findingNone {
		c.report(binary, f, concatenationDescription, parts)
	}
}

func (c *checker) isRawQuerySelector(expression ast.Expr) bool {
	selector, ok := ast.Unparen(expression).(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "RawQuery" && isURLType(c.pass.TypesInfo.TypeOf(selector.X))
}

func (c *checker) checkAssignment(assignment *ast.AssignStmt) {
	if len(assignment.Lhs) != len(assignment.Rhs) {
		return
	}

	for i, left := range assignment.Lhs {
		right := assignment.Rhs[i]

		if c.isRawQuerySelector(left) {
			if assignment.Tok == token.ADD_ASSIGN {
				c.examine(right)
				c.pass.Reportf(assignment.Pos(), queryMessage, concatenationDescription)
			} else {
				c.reportBuiltArgument(right, findingQuery)
			}
			continue
		}

		if assignment.Tok != token.ADD_ASSIGN || !isString(c.pass.TypesInfo.TypeOf(left)) {
			continue
		}

		parts := append([]part{{isHole: true, expression: left}}, c.concatenationParts(right)...)
		if f := c.classify(parts); f != findingNone {
			c.report(assignment, f, concatenationDescription, parts)
		}
	}
}

func (c *checker) checkCompositeLiteral(literal *ast.CompositeLit) {
	if !isURLType(c.pass.TypesInfo.TypeOf(literal)) {
		return
	}

	for _, element := range literal.Elts {
		keyValue, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if key, ok := keyValue.Key.(*ast.Ident); ok && key.Name == "RawQuery" {
			c.reportBuiltArgument(keyValue.Value, findingQuery)
		}
	}
}
