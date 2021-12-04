package svtu_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/msoap/byline"
	"github.com/sebdah/goldie/v2"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/dochang/svtu/internal/svtu"
)

func TestGrepCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdin    []string
		expected []string
	}{
		{
			name: "0",
			args: []string{">=0.0.0", "foo", "-", "bar"},
			stdin: []string{
				"1.1.0",
				"5.0.0",
			},
			expected: []string{
				"2.0.1",
				"3.5.0",
				"1.1.0",
				"5.0.0",
				"0.0.1",
				"0.0.0",
			},
		},
	}

	dataPathPrefix := "testdata"
	assert := assert.New(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var writer strings.Builder
			stdin := strings.NewReader(strings.Join(test.stdin, "\n"))
			dataPath := filepath.Join(dataPathPrefix, t.Name())
			fs := afero.NewBasePathFs(afero.NewOsFs(), dataPath)
			greper := svtu.Greper{
				Viper: viper.New(),
				In:    stdin,
				Out:   &writer,
				Err:   os.Stderr,
				Fs:    fs,
			}
			grepCmd := svtu.NewGrepCmd(greper)
			err := greper.Viper.BindPFlags(grepCmd.Flags())
			assert.NoError(err)
			grepCmd.SetArgs(append([]string{"--goroutines", "1"}, test.args...))
			err = grepCmd.Execute()
			assert.NoError(err)
			g := goldie.New(
				t,
				goldie.WithTestNameForDir(true),
				goldie.WithSubTestNameForDir(true),
			)

			g.Assert(t, "grep", []byte(writer.String()))

			output := byline.NewReader(strings.NewReader(writer.String()))
			result, err := output.MapString(func(line string) string {
				return strings.TrimSuffix(line, "\n")
			}).ReadAllSliceString()
			assert.NoError(err)
			assert.ElementsMatch(result, test.expected)
		})
	}
}
