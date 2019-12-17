package svtu_test

import (
	"testing"

	"github.com/blang/semver"
	"github.com/dochang/svtu"
	"github.com/stretchr/testify/assert"
)

func TestRemoveInvalidSemver(t *testing.T) {
	tests := []struct {
		input  []string
		output []semver.Version
	}{
		{
			input: []string{"1.0.0", "foo", "3.1.0", "bar", "2"},
			output: []semver.Version{
				semver.MustParse("1.0.0"),
				semver.MustParse("3.1.0"),
			},
		},
	}
	for _, test := range tests {
		assert.Equal(t, svtu.RemoveInvalidSemver(test.input), test.output)
	}
}

func TestRemoveOutOfRange(t *testing.T) {
	tests := []struct {
		svRanges []semver.Range
		input    []semver.Version
		output   []semver.Version
	}{
		{
			svRanges: []semver.Range{
				semver.MustParseRange(">1.0.0 <2.0.0"),
				semver.MustParseRange("1.5.0"),
			},
			input: []semver.Version{
				semver.MustParse("1.1.0"),
				semver.MustParse("1.5.0"),
				semver.MustParse("1.9.0"),
			},
			output: []semver.Version{
				semver.MustParse("1.5.0"),
			},
		},
	}
	for _, test := range tests {
		assert.Equal(t, svtu.RemoveOutOfRanges(test.svRanges, test.input), test.output)
	}
}
