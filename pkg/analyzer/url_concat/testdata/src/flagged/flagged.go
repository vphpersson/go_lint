package flagged

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const apiBase = "https://api.example.com"

type client struct {
	baseURL    string
	address    string
	httpClient *http.Client
}

func schemeLiteral(host string) string {
	return "https://" + host + "/v1" // want `URL built by string concatenation; build a url.URL`
}

func constantBase(userID string) string {
	return apiBase + "/users/" + userID // want `URL built by string concatenation`
}

func schemeOperand(scheme, host string) string {
	return scheme + "://" + host // want `URL built by string concatenation`
}

func connectionString(user, password, host string) string {
	return "postgres://" + user + ":" + password + "@" + host // want `URL built by string concatenation`
}

func parenthesized(host string) string {
	return ("https://" + host) + ("/v1/" + host) // want `URL built by string concatenation`
}

func sprintfScheme(host string, port int) string {
	return fmt.Sprintf("https://%s:%d/health", host, port) // want `URL built by fmt.Sprintf`
}

func appendfScheme(buffer []byte, host string) []byte {
	return fmt.Appendf(buffer, "https://%s", host) // want `URL built by fmt.Appendf`
}

func urlNamedBase(c *client, userID string) string {
	return c.baseURL + "/users/" + userID // want `URL built by string concatenation`
}

func urlNamedSprintf(baseURL, userID string) string {
	return fmt.Sprintf("%s/users/%s", baseURL, userID) // want `URL built by fmt.Sprintf`
}

func urlNamedPrefix(urlPrefix, path string) string {
	return urlPrefix + path // want `URL built by string concatenation`
}

func snakeCaseName(api_url string) string {
	return api_url + "/v1" // want `URL built by string concatenation`
}

func renderedURL(u *url.URL, path string) string {
	return u.String() + path // want `URL built by string concatenation`
}

func queryParameter(page int) string {
	return "/items?page=" + strconv.Itoa(page) // want `query string built by string concatenation; encode the parameters with url.Values`
}

func appendedQuery(path, query string) string {
	return path + "?" + query // want `query string built by string concatenation`
}

func formBody(clientID, secret string) string {
	return "grant_type=client_credentials&client_id=" + clientID + "&client_secret=" + secret // want `query string built by string concatenation`
}

func sprintfQuery(path string, limit int) string {
	return fmt.Sprintf("%s?limit=%d", path, limit) // want `query string built by fmt.Sprintf`
}

func appendAssign(endpointURL, userID string) string {
	endpointURL += "/users/" + userID // want `URL built by string concatenation`
	return endpointURL
}

func queryAppendAssign(query, page string) string {
	query += "&page=" + page // want `query string built by string concatenation`
	return query
}

func rawQueryAssign(u *url.URL, page string) {
	u.RawQuery = "page=" + page // want `query string built by string concatenation`
}

func rawQueryAppendAssign(u *url.URL, page string) {
	u.RawQuery += "&page=" + page // want `query string built by string concatenation`
}

func rawQueryField(page string) *url.URL {
	return &url.URL{Scheme: "https", Host: "example.com", RawQuery: fmt.Sprintf("page=%s", page)} // want `query string built by fmt.Sprintf`
}

func parsed(c *client) (*url.URL, error) {
	return url.Parse(c.address + "/v1") // want `URL built by string concatenation`
}

func request(c *client, path string) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, c.address+path, nil) // want `URL built by string concatenation`
}

func clientGet(c *client, address string) (*http.Response, error) {
	return c.httpClient.Get(fmt.Sprint(address) + "/status") // want `URL built by string concatenation`
}

func redirect(w http.ResponseWriter, r *http.Request, next string) {
	http.Redirect(w, r, "/login?next="+url.QueryEscape(next), http.StatusFound) // want `URL built by string concatenation`
}

func sprintfArgument(address string) (*http.Response, error) {
	return http.Get(fmt.Sprintf("%s/status", address)) // want `URL built by fmt.Sprintf`
}

// The Sprintf inside is part of the reported concatenation, not a second finding.
func nested(baseURL, query string) string {
	return fmt.Sprintf("%s/users", baseURL) + "?q=" + query // want `query string built by string concatenation`
}
