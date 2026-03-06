package cmd

import (
	"os"

	"github.com/ArttuOll/go-csv/cmd/parser"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-csv",
	Short: "A streaming CSV parser",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := parser.Parse(cmd, args)
		if err != nil {
			return err
		}

		return nil
	},
	Args: cobra.ExactArgs(1),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().Bool("header-line", false, "Indicates the presence of a header line in the file")
}
