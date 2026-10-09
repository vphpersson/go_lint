package url_concat_test

import (
	"testing"

	"github.com/vphpersson/go_lint/pkg/analyzer/url_concat"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		packageName string
	}{
		{name: "URLs and query strings built from strings are flagged", packageName: "flagged"},
		{name: "url.URL construction and unrelated concatenation are not flagged", packageName: "clean"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			analysistest.Run(t, analysistest.TestData(), url_concat.Analyzer, testCase.packageName)
		})
	}
}
