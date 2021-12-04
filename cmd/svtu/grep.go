package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/msoap/byline"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

type Greper struct {
	Viper *viper.Viper
	In    io.Reader
	Out   io.Writer
	Err   io.Writer
	Fs    afero.Fs
}

func newGrepCmd(greper Greper) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "grep PATTERNS [FILE...]",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ranges := greper.Viper.GetStringSlice("range")
			if len(ranges) == 0 {
				if len(args) > 0 {
					ranges = []string{args[0]}
					args = args[1:]
				} else {
					return errors.New("no RANGE specified")
				}
			}
			paths := args
			svRanges := []semver.Range{}
			for _, r := range ranges {
				svRange, err := semver.ParseRange(r)
				if err != nil {
					return err
				}
				svRanges = append(svRanges, svRange)
			}

			return greper.Run(svRanges, paths...)
		},
	}
	// Use `StringSliceP` instead of `StringArrayP` due to [1] & [2].
	//
	// [1]: https://github.com/spf13/viper/issues/246
	// [2]: https://github.com/spf13/viper/pull/398
	cmd.Flags().StringSliceP("range", "e", nil, "Use RANGE as the version range.")
	cmd.Flags().IntP("goroutines", "j", 0, "Set the number of grep worker goroutines to use.")
	return cmd
}

func grep(reader io.Reader, svRanges []semver.Range) ([]semver.Version, error) {
	linesReader := byline.NewReader(reader)
	semvers := []semver.Version{}

	err := linesReader.EachString(func(line string) {
		line = strings.TrimSuffix(line, "\n")
		v, err := semver.Parse(line)
		if err != nil {
			return
		}
		ok := true
		for _, r := range svRanges {
			if !(r(v)) {
				ok = false
				break
			}
		}
		if ok {
			semvers = append(semvers, v)
		}
	}).Discard()

	if err != nil {
		return nil, err
	}

	return semvers, nil
}

func (c Greper) Run(svRanges []semver.Range, paths ...string) error {
	g, ctx := errgroup.WithContext(context.Background())
	readerC := make(chan io.ReadCloser)

	g.Go(func() error {
		defer close(readerC)
		if len(paths) == 0 {
			paths = []string{"-"}
		}
		for i := 0; i < len(paths); i++ {
			path := paths[i]
			var reader io.ReadCloser
			var err error
			if path == "-" {
				reader = ioutil.NopCloser(c.In)
			} else {
				reader, err = c.Fs.Open(path)
			}
			if err != nil {
				return err
			}
			select {
			case readerC <- reader:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	semversC := make(chan []semver.Version)
	jobNum := c.Viper.GetInt("goroutines")
	if jobNum <= 0 {
		jobNum = runtime.NumCPU()
	}
	for i := 0; i < jobNum; i++ {
		g.Go(func() error {
			for reader := range readerC {
				semvers, err := grep(reader, svRanges)

				if err != nil {
					reader.Close()
					return err
				}

				if err := reader.Close(); err != nil {
					return err
				}

				select {
				case semversC <- semvers:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	go func() {
		defer close(semversC)
		g.Wait() //nolint:errcheck // We will check the return value later.
	}()

	for semvers := range semversC {
		for _, semver := range semvers {
			fmt.Fprintln(c.Out, semver)
		}
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
