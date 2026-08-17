package variadic_append_test

import (
	"testing"

	"github.com/vphpersson/go_lint/pkg/analyzer/variadic_append"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		packageName string
	}{
		{name: "appends to a variadic parameter are flagged", packageName: "flagged"},
		{name: "appends to copies and to other slices are not flagged", packageName: "clean"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			analysistest.Run(t, analysistest.TestData(), variadic_append.Analyzer, testCase.packageName)
		})
	}
}
