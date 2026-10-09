package clean

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// The recommended form, including concatenation within the Path field.
func structured(host, userID string, page int) string {
	return new(url.URL{
		Scheme:   "https",
		Host:     host,
		Path:     "/v1/users/" + userID,
		RawQuery: url.Values{"page": {strconv.Itoa(page)}}.Encode(),
	}).String()
}

func joined(base *url.URL, userID string) string {
	return base.JoinPath("users", userID).String()
}

// A constant has no operand to escape.
const constantURL = "https://example.com" + "/v1"

func message(address string, err error) string {
	return "could not reach " + address + ": " + err.Error()
}

func complete(err error) string {
	return "see https://example.com/docs: " + err.Error()
}

func urlInMessage(baseURL string, err error) string {
	return baseURL + ": " + err.Error()
}

func urlLast(baseURL string) string {
	return "base: " + baseURL
}

func sqlPlaceholders(n int) string {
	return "IN (?" + strings.Repeat(", ?", n-1) + ")"
}

func sqlComparison(column string) string {
	return column + " = ?" + " AND deleted = false"
}

func filePath(directory, name string) string {
	return directory + "/" + name
}

func hostPort(host, port string) string {
	return host + ":" + port
}

func notAURLName(curl, security string) string {
	return curl + "/bin/" + security
}

func errorMessage(host string, err error) error {
	return fmt.Errorf("fetch https://%s: %w", host, err)
}

func preformatted(format, host string) string {
	return fmt.Sprintf(format, host)
}

func spread(arguments ...any) string {
	return fmt.Sprintf("%s/users", arguments...)
}

func parsedVariable(raw string) (*url.URL, error) {
	return url.Parse(raw)
}

func rawQueryEncoded(u *url.URL, values url.Values) {
	u.RawQuery = values.Encode()
}

func nul(name string) string {
	return "a\x00b" + name + "\x00"
}

func numbers(a, b int) int {
	a += b
	return a + b
}

func urlPattern(domain string) *regexp.Regexp {
	return regexp.MustCompile(`^https://dd\.[a-f0-9]+\.` + regexp.QuoteMeta(domain) + `$`)
}
