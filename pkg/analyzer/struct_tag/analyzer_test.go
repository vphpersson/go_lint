package struct_tag_test

import (
	"testing"

	"github.com/vphpersson/go_lint/pkg/analyzer/struct_tag"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		packageName string
	}{
		{name: "option words as tag names are flagged", packageName: "flagged"},
		{name: "correct and unrelated tags are not flagged", packageName: "clean"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			analysistest.Run(t, analysistest.TestData(), struct_tag.Analyzer, testCase.packageName)
		})
	}
}
