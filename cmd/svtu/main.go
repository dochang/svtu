package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dochang/svtu/internal/svtu"
)

var (
	rootCmd = newRootCmd()
)

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "svtu",
		Short:         "SemVerTextUtils",
		Long:          `Semantic Versioning Text Utilities.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Hello, world")
			// fmt.Fprintln(os.Stdout, "Hello, world")
			return nil
		},
	}
}

func init() {
	v := viper.GetViper()
	if err := v.BindPFlags(rootCmd.Flags()); err != nil {
		log.Fatalln(err)
	}
	greper := svtu.Greper{
		Viper: v,
		In:    os.Stdin,
		Out:   os.Stdout,
		Err:   os.Stderr,
		Fs:    afero.NewOsFs(),
	}
	grepCmd := svtu.NewGrepCmd(greper)
	if err := v.BindPFlags(grepCmd.Flags()); err != nil {
		log.Fatalln(err)
	}
	rootCmd.AddCommand(grepCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
